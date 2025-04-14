package models

import (
	"time"

	"github.com/google/uuid"
)

type ReceptionStatus string

const (
	ReceptionStatusInProgress ReceptionStatus = "in_progress"
	ReceptionStatusClosed     ReceptionStatus = "close"
)

type Reception struct {
	ID       uuid.UUID       `json:"id"`
	DateTime time.Time       `json:"dateTime"`
	PVZID    uuid.UUID       `json:"pvzId"`
	Status   ReceptionStatus `json:"status"`
}

type ReceptionWithProducts struct {
	Reception Reception
	Products []Product
}

func (r ReceptionStatus) IsValid() bool {
	switch r {
	case ReceptionStatusInProgress, ReceptionStatusClosed:
		return true
	default:
		return false
	}
}
