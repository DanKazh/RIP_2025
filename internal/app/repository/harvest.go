package repository

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/argon2"
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

	weight := 550
	fullCost := 1
	productivity := 10

	application = ds.HarvestApplication{
		Status:       "draft",
		CreatorID:    userID,
		CreatedAt:    time.Now(),
		Weight:       &weight,
		FullCost:     &fullCost,
		Productivity: &productivity,
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

func (r *HarvestModel) GetHarvestApplicationInfo(id int) (ds.HarvestApplicationInfo, error) {
	var harvestApplicationInfo ds.HarvestApplication

	result := r.db.Where("id = ?", id).First(&harvestApplicationInfo)
	if result.Error != nil {
		return ds.HarvestApplicationInfo{}, fmt.Errorf("услуга не найдена: %w", result.Error)
	}

	applicationInfo := ds.HarvestApplicationInfo{
		Productivity: harvestApplicationInfo.Productivity,
		Weight:       harvestApplicationInfo.Weight,
	}

	return applicationInfo, nil
}

func (r *HarvestModel) GetHarvestApplication(id int) ([]map[string]interface{}, error) {
	var harvestApplication ds.HarvestApplication

	err := r.db.Preload("Resources.Resource").
		Where("id = ?", id).First(&harvestApplication).Error
	if err != nil {
		return nil, fmt.Errorf("заявка не найдена")
	}

	if harvestApplication.Status == "deleted" {
		return nil, fmt.Errorf("заявка удалена")
	}

	plannedWeight := harvestApplication.Weight

	productivity := harvestApplication.Productivity

	var applicationItems []map[string]interface{}
	for _, item := range harvestApplication.Resources {
		resource := item.Resource

		rawAmount := float64(*plannedWeight) / float64(*productivity) * float64(item.Ratio) * resource.Requirement
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
	}

	return r.db.Create(&applicationResource).Error
}

func (r *HarvestModel) DeleteApplication(applicationID int) error {
	result := r.db.Exec("UPDATE harvest_applications SET status = 'deleted' WHERE id = ?", applicationID)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("заявка не найдена или уже удалена")
	}

	return nil
}

func (r *HarvestModel) FormApplication(id int) error {
	var application ds.HarvestApplication

	result := r.db.First(&application, "id = ? AND status = ?", id, "draft")
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("заявка не найдена")
		}
		return result.Error
	}

	result = r.db.Model(&application).Update("status", "submitted")
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *HarvestModel) SetApplicationChanges(id, weight, productivity int) error {
	var application ds.HarvestApplication

	result := r.db.First(&application, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("заявка не найдена")
		}
		return result.Error
	}

	result = r.db.Model(&application).Updates(map[string]interface{}{
		"weight":       weight,
		"productivity": productivity,
	})
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *HarvestModel) CreateResource(resource *ds.HarvestResource) error {
	resource.CreatedAt = time.Now()
	resource.IsDeleted = false
	return r.db.Create(resource).Error
}

func (r *HarvestModel) UpdateResource(id int, updates map[string]interface{}) error {
	result := r.db.Model(&ds.HarvestResource{}).Where("id = ? AND is_deleted = false", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("ресурс не найден")
	}
	return nil
}

func (r *HarvestModel) DeleteResource(id int) error {
	result := r.db.Model(&ds.HarvestResource{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_deleted": true,
		"deleted_at": time.Now(),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("ресурс не найден")
	}
	return nil
}

func (r *HarvestModel) SetResourceImage(id int, imageURL string) error {
	return r.db.Model(&ds.HarvestResource{}).Where("id = ? AND is_deleted = false", id).Update("image_url", imageURL).Error
}

func (r *HarvestModel) GetUserCart(userID int) ([]ds.HarvestApplication, error) {
	var applications []ds.HarvestApplication
	err := r.db.Where("creator_id = ? AND status = 'draft'", userID).
		Preload("Resources.Resource").
		Find(&applications).Error
	return applications, err
}

func (r *HarvestModel) GetAllHarvestApplications() ([]ds.HarvestApplication, error) {
	var applications []ds.HarvestApplication
	err := r.db.Preload("Resources.Resource").
		Where("status NOT IN ?", []string{"draft", "deleted"}).
		Find(&applications).Error
	return applications, err
}

func (r *HarvestModel) DeclineApplication(applicationID int, moderatorID int, notes string) error {
	return r.db.Model(&ds.HarvestApplication{}).Where("id = ?", applicationID).Updates(map[string]interface{}{
		"status":       "rejected",
		"moderator_id": moderatorID,
		"notes":        notes,
	}).Error
}

func (r *HarvestModel) DeleteApplicationResource(applicationID, resourceID int) error {
	result := r.db.Where("application_id = ? AND resource_id = ?", applicationID, resourceID).
		Delete(&ds.ApplicationResource{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("ресурс не найден в заявке")
	}
	return nil
}

func (r *HarvestModel) SetApplicationResourceCoeff(applicationID, resourceID int, coefficient float64) error {
	if coefficient < 0.5 || coefficient > 1.5 {
		return fmt.Errorf("коэффициент должен быть между 0.5 и 1.5")
	}

	result := r.db.Model(&ds.ApplicationResource{}).
		Where("application_id = ? AND resource_id = ?", applicationID, resourceID).
		Update("ratio", coefficient)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("ресурс не найден в заявке")
	}
	return nil
}

func (r *HarvestModel) CreateUser(user *ds.User) error {
	user.CreatedAt = time.Now()
	user.IsDeleted = false
	return r.db.Create(user).Error
}

func (r *HarvestModel) GetUserByID(id int) (*ds.User, error) {
	var user ds.User
	err := r.db.Where("id = ? AND is_deleted = false", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *HarvestModel) GetUserByUsername(username string) (*ds.User, error) {
	var user ds.User
	err := r.db.Where("username = ? AND is_deleted = false", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *HarvestModel) UpdateUser(id int, updates map[string]interface{}) error {
	result := r.db.Model(&ds.User{}).Where("id = ? AND is_deleted = false", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("пользователь не найден")
	}
	return nil
}

func (r *HarvestModel) GenerateToken(user *ds.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &ds.JWTClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "harvest-app",
		},
		UserID: user.ID,
		Role:   user.Role,
	})

	return token.SignedString([]byte(r.jwtSecret))
}

func (r *HarvestModel) ParseToken(tokenString string) (*ds.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(r.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*ds.JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

func (r *HarvestModel) HashPass(plainPassword string) []byte {
	salt := make([]byte, 8)
	_, err := rand.Read(salt)
	if err != nil {
		return []byte{}
	}
	hashedPass := argon2.IDKey([]byte(plainPassword), []byte(salt), 1, 64*1024, 4, 32)
	return append(salt, hashedPass...)
}

func CheckPass(passHash []byte, plainPassword string) bool {
	salt := passHash[:8]
	userHash := argon2.IDKey([]byte(plainPassword), salt, 1, 64*1024, 4, 32)
	userHashedPassword := append(salt, userHash...)
	return bytes.Equal(userHashedPassword, passHash)
}
