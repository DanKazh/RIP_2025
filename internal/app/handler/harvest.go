package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"rip2025/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *HarvestController) RegisterController(router *gin.Engine) {
	router.GET("/harvestResources", h.GetHarvestResources)
	router.GET("/harvestDetailedResource/:id", h.GetHarvestResource)
	router.GET("/harvestApplication/:id", h.GetHarvestApplication)
	router.POST("/harvestApplication/:id/addResource", h.AddResourceToApplication)
	router.POST("/harvestApplication/:id/delete", h.DeleteApplication)
}

func (h *HarvestController) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/statics", "./resources")
}

func (h *HarvestController) errorController(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}

func (h *HarvestController) GetHarvestResources(ctx *gin.Context) {
	var harvestResources []ds.HarvestResource
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

	userID := 1
	draftApp, err := h.HarvestModel.GetUserDraftApplication(userID)
	if err != nil {
		logrus.Errorf("GetUserDraftApplication error: %v", err)
	}

	applicationCount, err := h.HarvestModel.GetHarvestApplicationCount(draftApp.ID)
	if err != nil {
		logrus.Error(err)
		applicationCount = 0
	}

	ctx.HTML(http.StatusOK, "harvestResources.html", gin.H{
		"harvestResources":        harvestResources,
		"harvestQuery":            searchHarvestQuery,
		"harvestApplicationCount": applicationCount,
		"harvestApplicationID":    draftApp.ID,
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
		h.errorController(ctx, http.StatusGone, fmt.Errorf("заявка была удалена"))
		return
	}

	harvestApplicationTotalCost, err := h.HarvestModel.GetApplicationTotalCost(id)
	if err != nil {
		logrus.Error(err)
	}

	harvestCultures, err := h.HarvestModel.GetHarvestCultures()
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "harvestApplication.html", gin.H{
		"harvestApplication":          harvestApplication,
		"harvestApplicationTotalCost": harvestApplicationTotalCost,
		"harvestCultures":             harvestCultures,
		"harvestApplicationID":        id,
	})
}

func (h *HarvestController) AddResourceToApplication(ctx *gin.Context) {
	applicationIDStr := ctx.Param("id")
	applicationID, err := strconv.Atoi(applicationIDStr)
	if err != nil {
		h.errorController(ctx, http.StatusBadRequest, fmt.Errorf("неверный ID заявки"))
		return
	}

	resourceIDStr := ctx.PostForm("resource_id")
	resourceID, err := strconv.Atoi(resourceIDStr)
	if err != nil {
		h.errorController(ctx, http.StatusBadRequest, fmt.Errorf("неверный ID услуги"))
		return
	}

	err = h.HarvestModel.AddResourceToApplication(applicationID, resourceID)
	if err != nil {
		h.errorController(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusFound, "/harvestResources")
}

func (h *HarvestController) DeleteApplication(ctx *gin.Context) {
	applicationIDStr := ctx.Param("id")
	applicationID, err := strconv.Atoi(applicationIDStr)
	if err != nil {
		h.errorController(ctx, http.StatusBadRequest, fmt.Errorf("неверный ID заявки"))
		return
	}

	err = h.HarvestModel.DeleteApplication(applicationID)
	if err != nil {
		h.errorController(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusFound, "/harvestResources")
}
