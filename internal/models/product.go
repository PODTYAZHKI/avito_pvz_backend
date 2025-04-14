package models

import (
	"time"

	"github.com/google/uuid"
)

type ProductType string

const (
	ProductTypeElectronics ProductType = "электроника"
	ProductTypeClothing    ProductType = "одежда"
	ProductTypeShoes       ProductType = "обувь"
)

type Product struct {
	ID          uuid.UUID
	DateTime    time.Time
	Type        ProductType
	ReceptionID uuid.UUID
}

func (pt ProductType) IsValid() bool {
	switch pt {
	case ProductTypeElectronics, ProductTypeClothing, ProductTypeShoes:
		return true
	default:
		return false
	}
}
