package main

import (
	"fmt"

	"rip2025/internal/app/config"
	"rip2025/internal/app/dsn"
	"rip2025/internal/app/handler"
	"rip2025/internal/app/repository"
	"rip2025/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	router := gin.Default()
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()
	fmt.Println(postgresString)

	rep, errRep := repository.NewHarvestModel(postgresString)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	hand := handler.NewHarvestController(rep)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
