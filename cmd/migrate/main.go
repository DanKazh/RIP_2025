package main

import (
	"rip2025/internal/app/ds"
	"rip2025/internal/app/dsn"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title Harvest API
// @version 1.0
// @description API для системы управления урожаем
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8081
// @BasePath /api
// @schemes http
func main() {
	_ = godotenv.Load()
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	err = db.AutoMigrate(
		&ds.HarvestResource{},
		&ds.HarvestApplication{},
		&ds.ApplicationResource{},
	)
	if err != nil {
		panic("cant migrate db")
	}
}
