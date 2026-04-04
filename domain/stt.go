package domain

import "time"

type Stt struct {
	SttDate        time.Time
	SourceFileName string
	Status         string
	TextFileName   string
}
