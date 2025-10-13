package repository

import (
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type HarvestModel struct {
	db        *gorm.DB
	jwtSecret string
}

func NewHarvestModel(dsn string) (*HarvestModel, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	return &HarvestModel{
		db:        db,
		jwtSecret: jwtSecret,
	}, nil
}
