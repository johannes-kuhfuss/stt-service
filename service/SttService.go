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

	"github.com/johannes-kuhfuss/services_utils/logger"
	"github.com/johannes-kuhfuss/stt-service/config"
	"github.com/johannes-kuhfuss/stt-service/helper"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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

func (s DefaultSttService) Extract(sourcePath string) error {
	var (
		buf           = new(bytes.Buffer{})
		mpw           = multipart.NewWriter(buf)
		extractedText string
		jsonRes       map[string]any
	)
	msg := fmt.Sprintf("Starting extraction using speaches at %v:%v...", s.Cfg.Stt.SpeachesHost, s.Cfg.Stt.SpeachesPort)
	logger.Info(msg)
	s.Cfg.RunTime.OLog.Info(msg)
	sourceFilePath := filepath.Join(s.Cfg.Stt.SttPath, sourcePath)
	f, err := os.Open(sourceFilePath)
	if err != nil {
		msg := "Error when opening source file"
		helper.AddToSttList(s.Cfg, sourceFilePath, "", msg, "")
		s.Cfg.Metrics.SttFailureCounter.Add(context.TODO(), 1)
		logger.Error(msg, err)
		s.Cfg.RunTime.OLog.Error(msg, slog.String("Error Message", err.Error()))
		return err
	}
	defer f.Close()
	fWriter, err := mpw.CreateFormFile("file", filepath.Base(sourceFilePath))
	if err != nil {
		msg := "Error when using source file"
		helper.AddToSttList(s.Cfg, sourceFilePath, "", msg, "")
		s.Cfg.Metrics.SttFailureCounter.Add(context.TODO(), 1)
		logger.Error(msg, err)
		s.Cfg.RunTime.OLog.Error(msg, slog.String("Error Message", err.Error()))
		return err
	}
	_, err = io.Copy(fWriter, f)
	if err != nil {
		msg := "Error when copying source file"
		helper.AddToSttList(s.Cfg, sourceFilePath, "", msg, "")
		s.Cfg.Metrics.SttFailureCounter.Add(context.TODO(), 1)
		logger.Error(msg, err)
		s.Cfg.RunTime.OLog.Error(msg, slog.String("Error Message", err.Error()))
		return err
	}
	err = mpw.WriteField("model", s.Cfg.Stt.SpeachesModel)
	if err != nil {
		msg := "Error adding model field"
		helper.AddToSttList(s.Cfg, sourceFilePath, "", msg, "")
		s.Cfg.Metrics.SttFailureCounter.Add(context.TODO(), 1)
		logger.Error(msg, err)
		s.Cfg.RunTime.OLog.Error(msg, slog.String("Error Message", err.Error()))
		return err
	}
	err = mpw.WriteField("reponse_format", "text")
	if err != nil {
		msg := "Error adding response_format field"
		helper.AddToSttList(s.Cfg, sourceFilePath, "", msg, "")
		s.Cfg.Metrics.SttFailureCounter.Add(context.TODO(), 1)
		logger.Error(msg, err)
		s.Cfg.RunTime.OLog.Error(msg, slog.String("Error Message", err.Error()))
		return err
	}
	err = mpw.Close()
	if err != nil {
		msg := "Error when closing form"
		helper.AddToSttList(s.Cfg, sourceFilePath, "", msg, "")
		s.Cfg.Metrics.SttFailureCounter.Add(context.TODO(), 1)
		logger.Error(msg, err)
		s.Cfg.RunTime.OLog.Error(msg, slog.String("Error Message", err.Error()))
		return err
	}
	speachesUrl := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(s.Cfg.Stt.SpeachesHost, s.Cfg.Stt.SpeachesPort),
		Path:   "/v1/audio/transcriptions",
	}
	client := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
	req, err := http.NewRequest("POST", speachesUrl.String(), buf)
	if err != nil {
		msg := "Error when creating request"
		helper.AddToSttList(s.Cfg, sourceFilePath, "", msg, "")
		s.Cfg.Metrics.SttFailureCounter.Add(context.TODO(), 1)
		logger.Error(msg, err)
		s.Cfg.RunTime.OLog.Error(msg, slog.String("Error Message", err.Error()))
		return err
	}
	req.Header.Add("Content-Type", mpw.FormDataContentType())
	resp, err := client.Do(req)
	if err != nil {
		msg := "Error when sending request"
		helper.AddToSttList(s.Cfg, sourceFilePath, "", msg, "")
		s.Cfg.Metrics.SttFailureCounter.Add(context.TODO(), 1)
		logger.Error(msg, err)
		s.Cfg.RunTime.OLog.Error(msg, slog.String("Error Message", err.Error()))
		return err
	}
	defer resp.Body.Close()
	msg = fmt.Sprintf("STT Request Response: %v", resp.Status)
	logger.Info(msg)
	s.Cfg.RunTime.OLog.Info(msg)
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			msg := "Error when reading response body"
			helper.AddToSttList(s.Cfg, sourceFilePath, "", msg, "")
			s.Cfg.Metrics.SttFailureCounter.Add(context.TODO(), 1)
			logger.Error(msg, err)
			s.Cfg.RunTime.OLog.Error(msg, slog.String("Error Message", err.Error()))
			return err
		}
		err = json.Unmarshal(bodyBytes, &jsonRes)
		if err != nil {
			msg := "Error when unmarshalling response body"
			helper.AddToSttList(s.Cfg, sourceFilePath, "", msg, "")
			s.Cfg.Metrics.SttFailureCounter.Add(context.TODO(), 1)
			logger.Error(msg, err)
			s.Cfg.RunTime.OLog.Error(msg, slog.String("Error Message", err.Error()))
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
			s.Cfg.Metrics.SttFailureCounter.Add(context.TODO(), 1)
			logger.Error(msg, err)
			s.Cfg.RunTime.OLog.Error(msg, slog.String("Error Message", err.Error()))
			return err
		}
		defer targetFile.Close()
		targetFile.WriteString(extractedText)
		helper.AddToSttList(s.Cfg, sourceFilePath, targetFilePath, "Speech-To-Text extracted successfully", extractedText)
		s.Cfg.Metrics.SttSuccessCounter.Add(context.TODO(), 1)
		msg := "STT extraction successful"
		logger.Info(msg)
		s.Cfg.RunTime.OLog.Info(msg)
		return nil
	} else {
		msg := "Error during speech-to-text processing"
		helper.AddToSttList(s.Cfg, sourceFilePath, "", msg, "")
		s.Cfg.Metrics.SttFailureCounter.Add(context.TODO(), 1)
		err := errors.New("Speaches returned error code")
		logger.Error(msg, err)
		s.Cfg.RunTime.OLog.Error(msg, slog.String("Error Message", err.Error()))
		return err
	}
}

func fileNameWithoutExt(fileName string) string {
	if pos := strings.LastIndexByte(fileName, '.'); pos != -1 {
		return fileName[:pos]
	}
	return fileName
}
