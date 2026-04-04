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

type XcodeHandler struct {
	Svc service.DefaultXcodeService
	Cfg *config.AppConfig
}

func NewXcodeHandler(cfg *config.AppConfig, svc service.DefaultXcodeService) XcodeHandler {
	return XcodeHandler{
		Cfg: cfg,
		Svc: svc,
	}
}

func (uh XcodeHandler) Receive(c *gin.Context) {
	var newXcodeReq dto.XcodeRequest
	if err := c.ShouldBindJSON(&newXcodeReq); err != nil {
		msg := "Invalid JSON body in transcode request"
		uh.Cfg.RunTime.OLog.Error(msg, slog.String("Error Message", err.Error()))
		apiErr := api_error.NewBadRequestError(msg)
		c.JSON(apiErr.StatusCode(), apiErr)
		return
	}
	uh.Cfg.RunTime.Sani.Sanitize(&newXcodeReq)
	if err := validateNewXcodeRequest(newXcodeReq); err != nil {
		msg := "Invalid parameter in transcode request."
		uh.Cfg.RunTime.OLog.Error(msg, slog.String("Error Message", err.Error()))
		apiErr := api_error.NewBadRequestError(msg)
		c.JSON(apiErr.StatusCode(), apiErr)
		return
	}
	uh.Svc.Xcode(newXcodeReq.SourceFilePath)

	c.JSON(http.StatusCreated, nil)
}

func validateNewXcodeRequest(req dto.XcodeRequest) error {
	if req.SourceFilePath == "" {
		return errors.New("Source File Path cannot be empty")
	}
	return nil
}
