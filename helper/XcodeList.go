package helper

import (
	"sort"
	"time"

	"github.com/johannes-kuhfuss/stt-service/config"
	"github.com/johannes-kuhfuss/stt-service/domain"
	"github.com/johannes-kuhfuss/stt-service/dto"
)

func AddToXcodeList(cfg *config.AppConfig, source, target, status string) {
	xc := domain.Xcode{
		XcodeDate:      time.Now(),
		SourceFileName: source,
		Status:         status,
		TargetFileName: target,
	}

	cfg.RunTime.XcodeList = append(cfg.RunTime.XcodeList, xc)
}

func GetSortedXcodeList(list []domain.Xcode) []dto.Xcode {
	var (
		entry      dto.Xcode
		sortedList []dto.Xcode
	)

	sort.Slice(list, func(i, j int) bool {
		return list[i].XcodeDate.After((list[j].XcodeDate))
	})

	for _, el := range list {
		entry.XcodeDate = el.XcodeDate.Format("2006-01-02 15:04:05")
		entry.SourceFileName = el.SourceFileName
		entry.TargetFileName = el.TargetFileName
		entry.Status = el.Status
		sortedList = append(sortedList, entry)
	}
	return sortedList
}
