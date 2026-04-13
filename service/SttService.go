package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/johannes-kuhfuss/services_utils/logger"
	"github.com/johannes-kuhfuss/stt-service/config"
	"github.com/johannes-kuhfuss/stt-service/helper"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	eMsg = "Error Message"
)

type SttExtracter interface {
	Extract(string) error
}

type DefaultSttService struct {
	Cfg *config.AppConfig
}

func NewSttService(cfg *config.AppConfig) DefaultSttService {
	return DefaultSttService{
		Cfg: cfg,
	}
}

func (s DefaultSttService) Extract(ictx context.Context, sourcePath string) error {
	var (
		buf = new(bytes.Buffer{})
		mpw = multipart.NewWriter(buf)
	)
	speachesUrl := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(s.Cfg.Stt.SpeachesHost, s.Cfg.Stt.SpeachesPort),
		Path:   "/v1/audio/transcriptions",
	}
	stc := trace.SpanContextFromContext(ictx)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	ctx = trace.ContextWithRemoteSpanContext(ctx, stc)
	tracer := otel.Tracer("stt-service")
	ctx, span := tracer.Start(ctx, "speech-to-text_request",
		trace.WithAttributes(
			attribute.String("http.url", speachesUrl.String()),
		),
	)
	defer span.End()
	msg := fmt.Sprintf("Starting extraction using speaches at %v:%v...", s.Cfg.Stt.SpeachesHost, s.Cfg.Stt.SpeachesPort)
	logger.Info(msg)
	s.Cfg.RunTime.OLog.InfoContext(ctx, msg)
	sourceFilePath := filepath.Join(s.Cfg.Stt.SttPath, sourcePath)
	f, err := os.Open(sourceFilePath)
	if err != nil {
		msg := "Error when opening source file"
		helper.AddToSttList(s.Cfg, sourceFilePath, "", msg, "")
		s.Cfg.Metrics.SttFailureCounter.Add(ctx, 1)
		logger.Error(msg, err)
		s.Cfg.RunTime.OLog.ErrorContext(ctx, msg, slog.String(eMsg, err.Error()))
		return err
	}
	defer f.Close()
	fWriter, err := mpw.CreateFormFile("file", filepath.Base(sourceFilePath))
	if err != nil {
		msg := "Error when using source file"
		helper.AddToSttList(s.Cfg, sourceFilePath, "", msg, "")
		s.Cfg.Metrics.SttFailureCounter.Add(ctx, 1)
		logger.Error(msg, err)
		s.Cfg.RunTime.OLog.ErrorContext(ctx, msg, slog.String(eMsg, err.Error()))
		return err
	}
	_, err = io.Copy(fWriter, f)
	if err != nil {
		msg := "Error when copying source file"
		helper.AddToSttList(s.Cfg, sourceFilePath, "", msg, "")
		s.Cfg.Metrics.SttFailureCounter.Add(ctx, 1)
		logger.Error(msg, err)
		s.Cfg.RunTime.OLog.ErrorContext(ctx, msg, slog.String(eMsg, err.Error()))
		return err
	}
	err = mpw.WriteField("model", s.Cfg.Stt.SpeachesModel)
	if err != nil {
		msg := "Error adding model field"
		helper.AddToSttList(s.Cfg, sourceFilePath, "", msg, "")
		s.Cfg.Metrics.SttFailureCounter.Add(ctx, 1)
		logger.Error(msg, err)
		s.Cfg.RunTime.OLog.ErrorContext(ctx, msg, slog.String(eMsg, err.Error()))
		return err
	}
	err = mpw.WriteField("reponse_format", "text")
	if err != nil {
		msg := "Error adding response_format field"
		helper.AddToSttList(s.Cfg, sourceFilePath, "", msg, "")
		s.Cfg.Metrics.SttFailureCounter.Add(ctx, 1)
		logger.Error(msg, err)
		s.Cfg.RunTime.OLog.ErrorContext(ctx, msg, slog.String(eMsg, err.Error()))
		return err
	}
	err = mpw.Close()
	if err != nil {
		msg := "Error when closing form"
		helper.AddToSttList(s.Cfg, sourceFilePath, "", msg, "")
		s.Cfg.Metrics.SttFailureCounter.Add(ctx, 1)
		logger.Error(msg, err)
		s.Cfg.RunTime.OLog.ErrorContext(ctx, msg, slog.String(eMsg, err.Error()))
		return err
	}
	client := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
	req, err := http.NewRequestWithContext(ctx, "POST", speachesUrl.String(), buf)
	if err != nil {
		msg := "Error when creating request"
		helper.AddToSttList(s.Cfg, sourceFilePath, "", msg, "")
		s.Cfg.Metrics.SttFailureCounter.Add(ctx, 1)
		logger.Error(msg, err)
		s.Cfg.RunTime.OLog.ErrorContext(ctx, msg, slog.String(eMsg, err.Error()))
		span.RecordError(err)
		return err
	}
	req.Header.Add("Content-Type", mpw.FormDataContentType())
	resp, err := client.Do(req)
	if err != nil {
		msg := "Error when sending request"
		helper.AddToSttList(s.Cfg, sourceFilePath, "", msg, "")
		s.Cfg.Metrics.SttFailureCounter.Add(ctx, 1)
		logger.Error(msg, err)
		s.Cfg.RunTime.OLog.ErrorContext(ctx, msg, slog.String(eMsg, err.Error()))
		span.RecordError(err)
		return err
	}
	defer resp.Body.Close()
	msg = fmt.Sprintf("STT Request Response: %v", resp.Status)
	logger.Info(msg)
	s.Cfg.RunTime.OLog.Info(msg)
	if resp.StatusCode == http.StatusOK {
		err := s.ProcessResult(ctx, resp, sourceFilePath)
		if err != nil {
			span.RecordError(err)
			return err
		}
		s.Cfg.Metrics.SttSuccessCounter.Add(ctx, 1)
		msg := "STT extraction successful"
		logger.Info(msg)
		s.Cfg.RunTime.OLog.InfoContext(ctx, msg)
		span.End()
		return nil
	} else {
		msg := "Error during speech-to-text processing"
		helper.AddToSttList(s.Cfg, sourceFilePath, "", msg, "")
		s.Cfg.Metrics.SttFailureCounter.Add(ctx, 1)
		err := errors.New("Speaches returned error code")
		logger.Error(msg, err)
		s.Cfg.RunTime.OLog.ErrorContext(ctx, msg, slog.String(eMsg, err.Error()))
		span.RecordError(err)
		return err
	}
}

func (s DefaultSttService) ProcessResult(ctx context.Context, resp *http.Response, sourceFilePath string) error {
	var (
		extractedText string
		jsonRes       map[string]any
	)
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		msg := "Error when reading response body"
		helper.AddToSttList(s.Cfg, sourceFilePath, "", msg, "")
		s.Cfg.Metrics.SttFailureCounter.Add(ctx, 1)
		logger.Error(msg, err)
		s.Cfg.RunTime.OLog.ErrorContext(ctx, msg, slog.String(eMsg, err.Error()))
		return err
	}
	err = json.Unmarshal(bodyBytes, &jsonRes)
	if err != nil {
		msg := "Error when unmarshalling response body"
		helper.AddToSttList(s.Cfg, sourceFilePath, "", msg, "")
		s.Cfg.Metrics.SttFailureCounter.Add(ctx, 1)
		logger.Error(msg, err)
		s.Cfg.RunTime.OLog.ErrorContext(ctx, msg, slog.String(eMsg, err.Error()))
		return err
	}
	extractedText = jsonRes["text"].(string)
	basePath := filepath.Dir(sourceFilePath)
	file := fileNameWithoutExt(filepath.Base(sourceFilePath))
	targetFilePath := filepath.Join(basePath, file+".txt")
	targetFile, err := os.Create(targetFilePath)
	if err != nil {
		msg := "Error when saving result"
		helper.AddToSttList(s.Cfg, sourceFilePath, targetFilePath, msg, "")
		s.Cfg.Metrics.SttFailureCounter.Add(ctx, 1)
		logger.Error(msg, err)
		s.Cfg.RunTime.OLog.ErrorContext(ctx, msg, slog.String(eMsg, err.Error()))
		return err
	}
	defer targetFile.Close()
	targetFile.WriteString(extractedText)
	helper.AddToSttList(s.Cfg, sourceFilePath, targetFilePath, "Speech-To-Text extracted successfully", extractedText)
	return nil
}

func fileNameWithoutExt(fileName string) string {
	if pos := strings.LastIndexByte(fileName, '.'); pos != -1 {
		return fileName[:pos]
	}
	return fileName
}
