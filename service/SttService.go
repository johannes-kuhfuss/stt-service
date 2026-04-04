package service

import (
	"context"

	"github.com/johannes-kuhfuss/stt-service/config"
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
	s.Cfg.Metrics.SttSuccessCounter.Add(context.TODO(), 1)
	return nil
}
