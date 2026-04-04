package helper

import (
	"sort"
	"time"

	"github.com/johannes-kuhfuss/stt-service/config"
	"github.com/johannes-kuhfuss/stt-service/domain"
	"github.com/johannes-kuhfuss/stt-service/dto"
)

func AddToSttList(cfg *config.AppConfig, source, textFile, status string) {
	xc := domain.Stt{
		SttDate:        time.Now(),
		SourceFileName: source,
		Status:         status,
		TextFileName:   textFile,
	}

	cfg.RunTime.SttList = append(cfg.RunTime.SttList, xc)
}

func GetSortedSttList(list []domain.Stt) []dto.Stt {
	var (
		entry      dto.Stt
		sortedList []dto.Stt
	)

	sort.Slice(list, func(i, j int) bool {
		return list[i].SttDate.After((list[j].SttDate))
	})

	for _, el := range list {
		entry.SttDate = el.SttDate.Format("2006-01-02 15:04:05")
		entry.SourceFileName = el.SourceFileName
		entry.TextFileName = el.TextFileName
		entry.Status = el.Status
		sortedList = append(sortedList, entry)
	}
	return sortedList
}
