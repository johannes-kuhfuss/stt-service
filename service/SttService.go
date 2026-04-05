package service

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/johannes-kuhfuss/stt-service/config"
	"github.com/johannes-kuhfuss/stt-service/domain"
	"github.com/johannes-kuhfuss/stt-service/helper"
)

const buffSize = 8000

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
		sttMsg domain.SttMessage
	)
	sourceFilePath := filepath.Join(s.Cfg.Stt.SttPath, sourcePath)

	voskUrl := url.URL{Scheme: "ws", Host: s.Cfg.Stt.VoskServer, Path: ""}
	conn, _, err := websocket.DefaultDialer.Dial(voskUrl.String(), nil)
	if err != nil {
		msg := "Could not connect to Vosk server"
		helper.AddToSttList(s.Cfg, sourceFilePath, "", msg)
		s.Cfg.Metrics.SttFailureCounter.Add(context.TODO(), 1)
		s.Cfg.RunTime.OLog.Error(msg, slog.String("Error Message", err.Error()))
		return err
	}

	sourceFile, err := os.Open(sourceFilePath)
	if err != nil {
		msg := "Could not connect open source file"
		helper.AddToSttList(s.Cfg, sourceFilePath, "", msg)
		s.Cfg.Metrics.SttFailureCounter.Add(context.TODO(), 1)
		s.Cfg.RunTime.OLog.Error(msg, slog.String("Error Message", err.Error()))
		return err
	}

	for {
		buf := make([]byte, buffSize)
		data, err := sourceFile.Read(buf)

		if data == 0 && err == io.EOF {
			err = conn.WriteMessage(websocket.TextMessage, []byte("{\"eof\" : 1}"))
			if err != nil {
				msg := "Websocket error sending EOF"
				helper.AddToSttList(s.Cfg, sourceFilePath, "", msg)
				s.Cfg.Metrics.SttFailureCounter.Add(context.TODO(), 1)
				s.Cfg.RunTime.OLog.Error(msg, slog.String("Error Message", err.Error()))
				return err
			}
			break
		}
		if err != nil {
			msg := "Websocket error sending data"
			helper.AddToSttList(s.Cfg, sourceFilePath, "", msg)
			s.Cfg.Metrics.SttFailureCounter.Add(context.TODO(), 1)
			s.Cfg.RunTime.OLog.Error(msg, slog.String("Error Message", err.Error()))
			return err
		}
		err = conn.WriteMessage(websocket.BinaryMessage, buf)
		if err != nil {
			msg := "Websocket error sending data"
			helper.AddToSttList(s.Cfg, sourceFilePath, "", msg)
			s.Cfg.Metrics.SttFailureCounter.Add(context.TODO(), 1)
			s.Cfg.RunTime.OLog.Error(msg, slog.String("Error Message", err.Error()))
			return err
		}
		_, _, err = conn.ReadMessage()
		if err != nil {
			msg := "Websocket error reading data"
			helper.AddToSttList(s.Cfg, sourceFilePath, "", msg)
			s.Cfg.Metrics.SttFailureCounter.Add(context.TODO(), 1)
			s.Cfg.RunTime.OLog.Error(msg, slog.String("Error Message", err.Error()))
			return err
		}
	}

	_, msg, err := conn.ReadMessage()
	if err != nil {
		msg := "Websocket error reading result"
		helper.AddToSttList(s.Cfg, sourceFilePath, "", msg)
		s.Cfg.Metrics.SttFailureCounter.Add(context.TODO(), 1)
		s.Cfg.RunTime.OLog.Error(msg, slog.String("Error Message", err.Error()))
		return err
	}
	conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

	err = json.Unmarshal(msg, &sttMsg)
	if err != nil {
		msg := "Error when trying to interpret result"
		helper.AddToSttList(s.Cfg, sourceFilePath, "", msg)
		s.Cfg.Metrics.SttFailureCounter.Add(context.TODO(), 1)
		s.Cfg.RunTime.OLog.Error(msg, slog.String("Error Message", err.Error()))
		return err
	}

	basePath := filepath.Dir(sourceFilePath)
	file := fileNameWithoutExt(filepath.Base(sourceFilePath))
	targetFilePath := filepath.Join(basePath, file+".txt")
	targetFile, err := os.Create(targetFilePath)
	if err != nil {
		msg := "Error when saving result"
		helper.AddToSttList(s.Cfg, sourceFilePath, targetFilePath, msg)
		s.Cfg.Metrics.SttFailureCounter.Add(context.TODO(), 1)
		s.Cfg.RunTime.OLog.Error(msg, slog.String("Error Message", err.Error()))
		return err
	}
	defer targetFile.Close()
	targetFile.WriteString(sttMsg.Text)
	helper.AddToSttList(s.Cfg, sourceFilePath, targetFilePath, "Speech-To-Text extracted successfully")

	s.Cfg.Metrics.SttSuccessCounter.Add(context.TODO(), 1)
	return nil
}

func fileNameWithoutExt(fileName string) string {
	if pos := strings.LastIndexByte(fileName, '.'); pos != -1 {
		return fileName[:pos]
	}
	return fileName
}
