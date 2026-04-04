package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/johannes-kuhfuss/services_utils/api_error"
	"github.com/johannes-kuhfuss/stt-service/config"
	"github.com/johannes-kuhfuss/stt-service/dto"
	"github.com/johannes-kuhfuss/stt-service/service"
)

type SttHandler struct {
	Svc service.DefaultSttService
	Cfg *config.AppConfig
}

func NewSttHandler(cfg *config.AppConfig, svc service.DefaultSttService) SttHandler {
	return SttHandler{
		Cfg: cfg,
		Svc: svc,
	}
}

func (uh SttHandler) Receive(c *gin.Context) {
	var newSttReq dto.SttRequest
	if err := c.ShouldBindJSON(&newSttReq); err != nil {
		msg := "Invalid JSON body in STT request"
		uh.Cfg.Metrics.SttFailureCounter.Add(c.Copy().Request.Context(), 1)
		uh.Cfg.RunTime.OLog.Error(msg, slog.String("Error Message", err.Error()))
		apiErr := api_error.NewBadRequestError(msg)
		c.JSON(apiErr.StatusCode(), apiErr)
		return
	}
	uh.Cfg.RunTime.Sani.Sanitize(&newSttReq)
	if err := validateNewSttRequest(newSttReq); err != nil {
		msg := "Invalid parameter in STT request."
		uh.Cfg.Metrics.SttFailureCounter.Add(c.Copy().Request.Context(), 1)
		uh.Cfg.RunTime.OLog.Error(msg, slog.String("Error Message", err.Error()))
		apiErr := api_error.NewBadRequestError(msg)
		c.JSON(apiErr.StatusCode(), apiErr)
		return
	}
	uh.Svc.Extract(newSttReq.SourceFilePath)

	c.JSON(http.StatusCreated, nil)
}

func validateNewSttRequest(req dto.SttRequest) error {
	if req.SourceFilePath == "" {
		return errors.New("Source File Path cannot be empty")
	}
	return nil
}
