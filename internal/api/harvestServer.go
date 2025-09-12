package api

import (
	"log"
	"rip2025/internal/app/handler"
	"rip2025/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	log.Println("Starting server")

	harvestModel, err := repository.NewHarvestModel()
	if err != nil {
		logrus.Error("ошибка инициализации репозитория")
	}

	harvestController := handler.NewHarvestController(harvestModel)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	r.GET("/harvestResources", harvestController.GetHarvestResources)
	r.GET("/harvestDetailedResource/:id", harvestController.GetHarvestResource)
	r.GET("/harvestApplication/:id", harvestController.GetHarvestApplication)

	r.Run()
	log.Println("Server down")
}
