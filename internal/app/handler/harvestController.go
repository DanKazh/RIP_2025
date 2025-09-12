package handler

import (
	"net/http"
	"rip2025/internal/app/repository"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type HarvestController struct {
	HarvestModel *repository.HarvestModel
}

func NewHarvestController(r *repository.HarvestModel) *HarvestController {
	return &HarvestController{
		HarvestModel: r,
	}
}

func (h *HarvestController) GetHarvestResources(ctx *gin.Context) {
	var harvestResources []repository.HarvestResource
	var err error

	searchHarvestQuery := ctx.Query("harvestQuery")
	if searchHarvestQuery == "" {
		harvestResources, err = h.HarvestModel.GetHarvestResources()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		harvestResources, err = h.HarvestModel.GetHarvestResourcesByTitle(searchHarvestQuery)
		if err != nil {
			logrus.Error(err)
		}
	}

	applicationCount, err := h.HarvestModel.GetHarvestApplicationCount(1)
	if err != nil {
		logrus.Error(err)
		applicationCount = 0
	}

	ctx.HTML(http.StatusOK, "harvestResources.html", gin.H{
		"harvestResources":        harvestResources,
		"harvestQuery":            searchHarvestQuery,
		"harvestApplicationCount": applicationCount,
	})
}

func (h *HarvestController) GetHarvestResource(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
	}

	harvestResource, err := h.HarvestModel.GetHarvestResource(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "harvestDetailedResource.html", gin.H{
		"harvestResource": harvestResource,
	})
}

func (h *HarvestController) GetHarvestApplication(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
	}
	harvestApplication, err := h.HarvestModel.GetHarvestApplication(id)
	if err != nil {
		logrus.Error(err)
	}

	harvestCultures, err := h.HarvestModel.GetHarvestCultures()
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "harvestApplication.html", gin.H{
		"harvestApplication": harvestApplication,
		"harvestCultures":    harvestCultures,
	})
}
