package pkg

import (
	"context"
	"fmt"

	"rip2025/internal/app/config"
	"rip2025/internal/app/handler"
	"rip2025/internal/app/redis"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Application struct {
	Config     *config.Config
	Router     *gin.Engine
	Controller *handler.HarvestController
	Redis      *redis.Client
}

func NewApp(ctx context.Context, c *config.Config, r *gin.Engine, controller *handler.HarvestController) (*Application, error) {
	redisClient, err := redis.New(ctx, c.Redis)
	if err != nil {
		return nil, err
	}

	return &Application{
		Config:     c,
		Router:     r,
		Controller: controller,
		Redis:      redisClient,
	}, nil
}

func (a *Application) RunApp() {
	logrus.Info("Server start up")

	a.Controller.RegisterController(a.Router)
	a.Controller.RegisterStatic(a.Router)

	serverAddress := fmt.Sprintf("%s:%d", a.Config.ServiceHost, a.Config.ServicePort)
	if err := a.Router.Run(serverAddress); err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("Server down")
}

func (a *Application) Close() error {
	if a.Redis != nil {
		return a.Redis.Close()
	}
	return nil
}
