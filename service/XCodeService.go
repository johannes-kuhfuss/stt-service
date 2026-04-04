package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/johannes-kuhfuss/stt-service/config"
	"github.com/johannes-kuhfuss/stt-service/helper"
)

type Xcoder interface {
	Xcode(string) (string, error)
}

type DefaultXcodeService struct {
	Cfg *config.AppConfig
}

func NewXcodeService(cfg *config.AppConfig) DefaultXcodeService {
	return DefaultXcodeService{
		Cfg: cfg,
	}
}

func (s DefaultXcodeService) Xcode(sourcePath string) error {
	var (
		targetPath string
		err        error
	)
	if _, err = os.Stat(s.Cfg.Xcode.FfmpegPath); errors.Is(err, os.ErrNotExist) {
		s.Cfg.RunTime.OLog.Error("ffmpeg binary not found", slog.String("Error Message", err.Error()))
		return err
	}
	filepath := path.Join(s.Cfg.Xcode.XcodePath, sourcePath)
	if _, err = os.Stat(filepath); errors.Is(err, os.ErrNotExist) {
		helper.AddToXcodeList(s.Cfg, filepath, "", "Source file not found")
		s.Cfg.RunTime.OLog.Error("source file not found", slog.String("Error Message", err.Error()))
		return err
	}
	if targetPath, err = s.runFfmpeg(filepath); err != nil {
		helper.AddToXcodeList(s.Cfg, filepath, "", "Error executing ffmpeg")
		s.Cfg.RunTime.OLog.Error("Could not execute ffmpeg successfully", slog.String("Error Message", err.Error()))
		return err
	}
	if _, err = os.Stat(targetPath); errors.Is(err, os.ErrNotExist) {
		helper.AddToXcodeList(s.Cfg, filepath, targetPath, "Target file not found. This should not happen.")
		s.Cfg.RunTime.OLog.Error("target file not found", slog.String("Error Message", err.Error()))
		return err
	}
	helper.AddToXcodeList(s.Cfg, filepath, targetPath, "Transcode successful")
	s.Cfg.RunTime.OLog.Info(fmt.Sprintf("target file created: %v", targetPath))
	return nil
}

func (s DefaultXcodeService) runFfmpeg(filePath string) (targetFilePath string, err error) {
	// command line: ffmpeg -i input.mp3 -ac 1 -ar 16000 -c:a pcm_s16le output.wav
	basePath := filepath.Dir(filePath)
	file := fileNameWithoutExt(filepath.Base(filePath))
	target := filepath.Join(basePath, file+".wav")
	ctx := context.Background()
	timeout := time.Duration(20 * time.Second)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, s.Cfg.Xcode.FfmpegPath, "-i", filePath, "-ac", "1", "-ar", "16000", "-c:a", "pcm_s16le", target)
	_, err = cmd.CombinedOutput()
	if err != nil {
		cancel()
		s.Cfg.RunTime.OLog.Error(fmt.Sprintf("Could not execute ffmpeg on file %v: ", filePath), slog.String("Error Message", err.Error()))
		return "", err
	}
	return target, nil
}

func fileNameWithoutExt(fileName string) string {
	if pos := strings.LastIndexByte(fileName, '.'); pos != -1 {
		return fileName[:pos]
	}
	return fileName
}
