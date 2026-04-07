package service

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/johannes-kuhfuss/services_utils/logger"
	"github.com/johannes-kuhfuss/stt-service/config"
	"github.com/johannes-kuhfuss/stt-service/helper"
)

const lore = "Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet."

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
	sourceFilePath := filepath.Join(s.Cfg.Stt.SttPath, sourcePath)
	basePath := filepath.Dir(sourceFilePath)
	file := fileNameWithoutExt(filepath.Base(sourceFilePath))
	targetFilePath := filepath.Join(basePath, file+".txt")
	targetFile, err := os.Create(targetFilePath)
	if err != nil {
		msg := "Error when saving result"
		helper.AddToSttList(s.Cfg, sourceFilePath, targetFilePath, msg, "")
		s.Cfg.Metrics.SttFailureCounter.Add(context.TODO(), 1)
		logger.Error(msg, err)
		return err
	}
	defer targetFile.Close()
	targetFile.WriteString(lore)
	helper.AddToSttList(s.Cfg, sourceFilePath, targetFilePath, "Speech-To-Text extracted successfully", lore)

	s.Cfg.Metrics.SttSuccessCounter.Add(context.TODO(), 1)
	return nil
}

func fileNameWithoutExt(fileName string) string {
	if pos := strings.LastIndexByte(fileName, '.'); pos != -1 {
		return fileName[:pos]
	}
	return fileName
}
