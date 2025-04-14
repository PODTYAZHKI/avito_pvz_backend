package models

import (
	"time"

	"github.com/google/uuid"
)

type PvzCity string

const (
	Moscow PvzCity = "Москва"
	SPb    PvzCity = "Санкт-Петербург"
	Kazan  PvzCity = "Казань"
)

type Pvz struct {
	ID               uuid.UUID
	RegistrationDate time.Time
	City             PvzCity
}

type PvzWithReceptions struct {
	PVZ        Pvz
	Receptions []ReceptionWithProducts
}

func (c PvzCity) IsValid() bool {
	switch c {
	case Moscow, SPb, Kazan:
		return true
	default:
		return false
	}
}

type PvzFilter struct {
	StartDate time.Time
	EndDate   time.Time
	Page      int
	Limit     int
}

func NewPvzFilter(startDate, endDate string, page, limit int) PvzFilter {
	filter := PvzFilter{
		Page:  page,
		Limit: limit,
	}

	if startDate != "" {
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			filter.StartDate = t
		}
	}

	if endDate != "" {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			filter.EndDate = t
		}
	}

	return filter
}
