package handler

import (
	"rip2025/internal/app/repository"
)

type HarvestController struct {
	HarvestModel *repository.HarvestModel
}

func NewHarvestController(r *repository.HarvestModel) *HarvestController {
	return &HarvestController{
		HarvestModel: r,
	}
}
