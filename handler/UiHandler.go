package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/johannes-kuhfuss/stt-service/config"
	"github.com/johannes-kuhfuss/stt-service/helper"
)

type UiHandler struct {
	Cfg *config.AppConfig
}

func NewUiHandler(cfg *config.AppConfig) UiHandler {
	return UiHandler{
		Cfg: cfg,
	}
}

func (uh *UiHandler) AboutPage(c *gin.Context) {
	c.HTML(http.StatusOK, "about.page.tmpl", gin.H{
		"title": "About",
		"data":  nil,
	})
}

func (uh *UiHandler) XcodeListPage(c *gin.Context) {
	files := helper.GetSortedXcodeList(uh.Cfg.RunTime.XcodeList)
	c.HTML(http.StatusOK, "xcodelist.page.tmpl", gin.H{
		"title": "Transcode List",
		"data":  files,
	})
}
