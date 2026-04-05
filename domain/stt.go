package domain

import (
	"time"
)

type Stt struct {
	SttDate        time.Time
	SourceFileName string
	Status         string
	TextFileName   string
}

type SttMessage struct {
	Result []struct {
		Conf  float64
		Start float64
		End   float64
		Word  string
	}
	Text string
}
