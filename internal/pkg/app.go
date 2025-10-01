package pkg

import (
	"fmt"

	"rip2025/internal/app/config"
	"rip2025/internal/app/handler"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Application struct {
	Config     *config.Config
	Router     *gin.Engine
	Controller *handler.HarvestController
}

func NewApp(c *config.Config, r *gin.Engine, controller *handler.HarvestController) *Application {
	return &Application{
		Config:     c,
		Router:     r,
		Controller: controller,
	}
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
