package domain

import "time"

type Xcode struct {
	XcodeDate      time.Time
	SourceFileName string
	Status         string
	TargetFileName string
}
