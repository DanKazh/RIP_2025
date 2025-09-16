package repository

import (
	"errors"
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"

	"rip2025/internal/app/ds"
	"strings"
)

func (r *HarvestModel) GetUserDraftApplication(userID int) (*ds.HarvestApplication, error) {
	var application ds.HarvestApplication
	err := r.db.Where("creator_id = ? AND status = ?", userID, "draft").
		Order("created_at asc").
		First(&application).Error

	if err == nil {
		return &application, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	application = ds.HarvestApplication{
		Status:    "draft",
		CreatorID: userID,
		CreatedAt: time.Now(),
		Weight:    0,
		FullCost:  0,
		CultureID: 1,
	}

	if err := r.db.Create(&application).Error; err != nil {
		return nil, err
	}

	return &application, nil
}

func (r *HarvestModel) GetHarvestResources() ([]ds.HarvestResource, error) {
	var harvestResources []ds.HarvestResource

	err := r.db.Find(&harvestResources).Error
	if err != nil {
		return nil, err
	}

	if len(harvestResources) == 0 {
		return nil, fmt.Errorf("массив услуг пустой")
	}

	return harvestResources, nil
}

func (r *HarvestModel) GetHarvestCultures() ([]ds.HarvestCulture, error) {
	var harvestCultures []ds.HarvestCulture

	err := r.db.Find(&harvestCultures).Error
	if err != nil {
		return nil, err
	}

	return harvestCultures, nil
}

func (r *HarvestModel) GetHarvestApplication(id int) ([]map[string]interface{}, error) {
	var harvestApplication ds.HarvestApplication

	err := r.db.Preload("Resources.Resource").Preload("Culture").
		Where("id = ?", id).First(&harvestApplication).Error
	if err != nil {
		return nil, fmt.Errorf("заявка не найдена")
	}

	if harvestApplication.Status == "deleted" {
		return nil, fmt.Errorf("заявка удалена")
	}

	plannedWeight := harvestApplication.Weight
	if plannedWeight == 0 {
		plannedWeight = 550 // дефолтный урожай
	}

	productivity := harvestApplication.Culture.Productivity

	var applicationItems []map[string]interface{}
	for _, item := range harvestApplication.Resources {
		resource := item.Resource

		rawAmount := float64(plannedWeight) / float64(productivity) *
			float64(item.Ratio) * resource.Requirement
		neededAmount := int(math.Ceil(rawAmount))

		totalCost := neededAmount * int(resource.TariffCost)

		applicationItem := map[string]interface{}{
			"ResourceID":          item.ResourceID,
			"ResourceImageURL":    resource.ImageURL,
			"ResourceName":        resource.Name,
			"ResourceTariffCost":  resource.TariffCost,
			"ResourceTariff":      resource.Tariff,
			"ResourceMeasurement": resource.Measurement,
			"Ratio":               item.Ratio,
			"NeededAmount":        neededAmount,
			"TotalCost":           totalCost,
		}
		applicationItems = append(applicationItems, applicationItem)
	}

	return applicationItems, nil
}

func (r *HarvestModel) GetHarvestResource(id int) (ds.HarvestResource, error) {
	var resource ds.HarvestResource
	err := r.db.Where("id = ?", id).First(&resource).Error
	if err != nil {
		return ds.HarvestResource{}, fmt.Errorf("услуга не найдена")
	}
	return resource, nil
}

func (r *HarvestModel) GetHarvestResourcesByTitle(title string) ([]ds.HarvestResource, error) {
	var resources []ds.HarvestResource
	err := r.db.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(title)+"%").Find(&resources).Error
	if err != nil {
		return nil, err
	}
	return resources, nil
}

func (r *HarvestModel) GetHarvestApplicationCount(applicationID int) (int, error) {
	var count int64
	err := r.db.Model(&ds.ApplicationResource{}).Where("application_id = ?", applicationID).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *HarvestModel) GetApplicationTotalCost(id int) (int, error) {
	items, err := r.GetHarvestApplication(id)
	if err != nil {
		return 0, err
	}
	total := 0
	for _, item := range items {
		total += item["TotalCost"].(int)
	}
	return total, nil
}

func (r *HarvestModel) AddResourceToApplication(applicationID, resourceID int) error {
	var application ds.HarvestApplication
	err := r.db.Where("id = ? AND status != 'deleted'", applicationID).First(&application).Error
	if err != nil {
		return fmt.Errorf("заявка не найдена или удалена")
	}

	var resource ds.HarvestResource
	err = r.db.Where("id = ? AND is_deleted = false", resourceID).First(&resource).Error
	if err != nil {
		return fmt.Errorf("услуга не найдена")
	}

	var existingResource ds.ApplicationResource
	err = r.db.Where("application_id = ? AND resource_id = ?", applicationID, resourceID).
		First(&existingResource).Error

	if err == nil {
		return fmt.Errorf("услуга уже добавлена в заявку")
	}

	applicationResource := ds.ApplicationResource{
		ApplicationID: applicationID,
		ResourceID:    resourceID,
		Ratio:         1,
		NeededAmount:  0,
		TotalCost:     0,
		CreatedAt:     time.Now(),
	}

	return r.db.Create(&applicationResource).Error
}

func (r *HarvestModel) DeleteApplication(applicationID int) error {
	result := r.db.Exec("UPDATE harvest_applications SET status = 'deleted' WHERE id = ? AND status != 'deleted'", applicationID)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("заявка не найдена или уже удалена")
	}

	return nil
}

func (r *HarvestModel) CheckApplicationStatus(id int) (string, error) {
	var application ds.HarvestApplication
	err := r.db.Select("status").Where("id = ? AND status = 'deleted'", id).First(&application).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "active", nil
		}
		return "", err
	}
	return application.Status, nil
}
