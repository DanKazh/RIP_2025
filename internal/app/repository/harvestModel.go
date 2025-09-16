package repository

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type HarvestModel struct {
	db *gorm.DB
}

func NewHarvestModel(dsn string) (*HarvestModel, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &HarvestModel{
		db: db,
	}, nil
}
