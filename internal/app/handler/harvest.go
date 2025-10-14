package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"rip2025/internal/app/ds"
	"rip2025/internal/app/middleware"
	"rip2025/internal/app/role"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	_ "rip2025/docs"
)

func (h *HarvestController) RegisterController(router *gin.Engine) {
	api := router.Group("/api")
	{
		guest := api.Group("/")
		guest.Use(middleware.WithAuthCheck(h.JWTSecret, h.redisClient, role.User, role.Admin, role.Guest))
		{
			guest.GET("/harvestResources", h.GetHarvestResources) // Guest
			guest.GET("/harvestDetailedResource/:id", h.GetHarvestResource)
			guest.POST("/users/register", h.RegisterUser)
			guest.POST("/users/login", h.LoginUser)
		}
		authUser := api.Group("/")
		authUser.Use(middleware.WithAuthCheck(h.JWTSecret, h.redisClient, role.User, role.Admin))
		{
			authUser.GET("/harvestApplication/:id", h.GetHarvestApplication) // User
			authUser.POST("/harvestApplication/:id/addResource", h.AddResourceToApplication)
			authUser.DELETE("/harvestApplication/:id/delete", h.DeleteApplication)
			authUser.GET("/harvestApplication/harvestCart", h.GetUserCart)
			authUser.POST("/harvestApplication/:id/setChanges", h.SetApplicationChanges)
			authUser.POST("/harvestApplication/:id/form", h.FormApplication)
			authUser.DELETE("/applicationResource/:id/deleteResource", h.DeleteApplicationResource)
			authUser.POST("/applicationResource/:id/setCoeff", h.SetApplicationResourceCoeff)
			authUser.POST("/users/logout", h.LogoutUser)
		}

		admin := api.Group("/")
		admin.Use(middleware.WithAuthCheck(h.JWTSecret, h.redisClient, role.Admin))
		{
			admin.POST("/harvestResources/createResource", h.CreateResource) // Admin
			admin.PUT("/harvestResources/:id/update", h.UpdateResource)
			admin.DELETE("/harvestResources/:id/delete", h.DeleteResource)
			admin.POST("/harvestResources/:id/setImage", h.SetResourceImage)
			admin.GET("/users/:id", h.GetUser)
			admin.PUT("/users/:id/setChanges", h.SetUserChanges)
			admin.GET("/harvestApplications", h.GetHarvestApplications)
			admin.PUT("/harvestApplication/:id/decline", h.DeclineApplication)
		}
	}
}

func (h *HarvestController) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/statics", "./resources")
}

func (h *HarvestController) errorResponse(ctx *gin.Context, statusCode int, message string) {
	logrus.Error(message)
	ctx.JSON(statusCode, gin.H{
		"status":  "error",
		"message": message,
	})
}

func (h *HarvestController) successResponse(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, gin.H{
		"data": data,
	})
}

// GetHarvestResources godoc
// @Summary Получить список ресурсов урожая
// @Description Получить все ресурсы урожая с возможностью поиска
// @Tags harvest
// @Accept json
// @Produce json
// @Param harvestQuery query string false "Поисковый запрос"
// @Success 200 {object} ds.StandardResponse
// @Failure 500 {object} ds.StandardResponse
// @Security BearerAuth
// @Router /harvestResources [get]
func (h *HarvestController) GetHarvestResources(ctx *gin.Context) {
	var harvestResources []ds.HarvestResource
	var err error

	searchHarvestQuery := ctx.Query("harvestQuery")
	if searchHarvestQuery == "" {
		harvestResources, err = h.HarvestModel.GetHarvestResources()
		if err != nil {
			h.errorResponse(ctx, http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		harvestResources, err = h.HarvestModel.GetHarvestResourcesByTitle(searchHarvestQuery)
		if err != nil {
			h.errorResponse(ctx, http.StatusInternalServerError, err.Error())
			return
		}
	}

	userIDVal, exists := ctx.Get("user_id")
	if !exists {
		h.successResponse(ctx, gin.H{
			"resources":   harvestResources,
			"searchQuery": searchHarvestQuery,
		})
		return
	}
	userID := userIDVal.(int)
	draftApp, err := h.HarvestModel.GetUserDraftApplication(userID)
	if err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка получения черновика заявки: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"resources":     harvestResources,
		"searchQuery":   searchHarvestQuery,
		"applicationID": draftApp.ID,
	})
}

// GetHarvestResource godoc
// @Summary Получить детальную информацию о ресурсе
// @Description Получить детальную информацию о конкретном ресурсе урожая
// @Tags harvest
// @Accept json
// @Produce json
// @Param id path int true "ID ресурса"
// @Success 200 {object} ds.StandardResponse
// @Failure 400 {object} ds.StandardResponse
// @Failure 404 {object} ds.StandardResponse
// @Security BearerAuth
// @Router /harvestDetailedResource/{id} [get]
func (h *HarvestController) GetHarvestResource(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный ID ресурса")
		return
	}

	harvestResource, err := h.HarvestModel.GetHarvestResource(id)
	if err != nil {
		h.errorResponse(ctx, http.StatusNotFound, "Ресурс не найден")
		return
	}

	h.successResponse(ctx, harvestResource)
}

// GetHarvestApplication godoc
// @Summary Получить информацию о заявке
// @Description Получить детальную информацию о заявке на урожай
// @Tags application
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Success 200 {object} ds.StandardResponse
// @Failure 400 {object} ds.StandardResponse
// @Failure 404 {object} ds.StandardResponse
// @Security BearerAuth
// @Router /harvestApplication/{id} [get]
func (h *HarvestController) GetHarvestApplication(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный ID заявки")
		return
	}

	harvestApplication, err := h.HarvestModel.GetHarvestApplication(id)
	if err != nil {
		h.errorResponse(ctx, http.StatusNotFound, "Заявка не найдена или удалена")
		return
	}

	harvestApplicationTotalCost, err := h.HarvestModel.GetApplicationTotalCost(id)
	if err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка расчета стоимости: "+err.Error())
		return
	}

	harvestApplicationInfo, err := h.HarvestModel.GetHarvestApplicationInfo(id)
	if err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка получения информации о заявке: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"applicationItems": harvestApplication,
		"totalCost":        harvestApplicationTotalCost,
		"applicationID":    id,
		"applicationInfo":  harvestApplicationInfo,
	})
}

// AddResourceToApplication godoc
// @Summary Добавить ресурс в заявку
// @Description Добавить ресурс урожая в существующую заявку
// @Tags application
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Param request body ds.AddResourceRequest true "Данные ресурса"
// @Success 200 {object} ds.StandardResponse
// @Failure 400 {object} ds.StandardResponse
// @Failure 500 {object} ds.StandardResponse
// @Security BearerAuth
// @Router /harvestApplication/{id}/addResource [post]
func (h *HarvestController) AddResourceToApplication(ctx *gin.Context) {
	applicationIDStr := ctx.Param("id")
	applicationID, err := strconv.Atoi(applicationIDStr)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный ID заявки")
		return
	}

	var request struct {
		ResourceID int `json:"resource_id" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный формат данных")
		return
	}

	err = h.HarvestModel.AddResourceToApplication(applicationID, request.ResourceID)
	if err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"applicationID": applicationID,
		"resourceID":    request.ResourceID,
		"message":       "Ресурс успешно добавлен в заявку",
	})
}

// DeleteApplication godoc
// @Summary Удалить заявку
// @Description Удалить заявку на урожай
// @Tags application
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Success 200 {object} ds.StandardResponse
// @Failure 400 {object} ds.StandardResponse
// @Failure 500 {object} ds.StandardResponse
// @Security BearerAuth
// @Router /harvestApplication/{id}/delete [delete]
func (h *HarvestController) DeleteApplication(ctx *gin.Context) {
	applicationIDStr := ctx.Param("id")
	applicationID, err := strconv.Atoi(applicationIDStr)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный ID заявки")
		return
	}

	err = h.HarvestModel.DeleteApplication(applicationID)
	if err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"message": "Заявка успешно удалена",
	})
}

// SetApplicationChanges godoc
// @Summary Изменить параметры заявки
// @Description Установить изменения веса и продуктивности для заявки
// @Tags application
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Param request body ds.SetApplicationChangesRequest true "Данные изменений"
// @Success 200 {object} ds.StandardResponse
// @Failure 400 {object} ds.StandardResponse
// @Failure 500 {object} ds.StandardResponse
// @Security BearerAuth
// @Router /harvestApplication/{id}/setChanges [post]
func (h *HarvestController) SetApplicationChanges(ctx *gin.Context) {
	applicationIDStr := ctx.Param("id")
	applicationID, err := strconv.Atoi(applicationIDStr)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный ID заявки")
		return
	}

	var request struct {
		Weight       int `json:"weight" binding:"required"`
		Productivity int `json:"productivity" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный формат данных")
		return
	}

	err = h.HarvestModel.SetApplicationChanges(applicationID, request.Weight, request.Productivity)
	if err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"message": "Изменения успешно сохранены",
	})
}

// FormApplication godoc
// @Summary Отправить заявку
// @Description Отправить заявку на обработку
// @Tags application
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Success 200 {object} ds.StandardResponse
// @Failure 400 {object} ds.StandardResponse
// @Failure 500 {object} ds.StandardResponse
// @Security BearerAuth
// @Router /harvestApplication/{id}/form [post]
func (h *HarvestController) FormApplication(ctx *gin.Context) {
	applicationIDStr := ctx.Param("id")
	applicationID, err := strconv.Atoi(applicationIDStr)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный ID заявки")
		return
	}

	err = h.HarvestModel.FormApplication(applicationID)
	if err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"message": "Заявка успешно отправлена",
	})
}

// CreateResource godoc
// @Summary Создать новый ресурс
// @Description Создать новый ресурс урожая (только для администраторов)
// @Tags admin
// @Accept json
// @Produce json
// @Param resource body ds.HarvestResource true "Данные ресурса"
// @Success 200 {object} ds.StandardResponse
// @Failure 400 {object} ds.StandardResponse
// @Failure 500 {object} ds.StandardResponse
// @Security BearerAuth
// @Router /harvestResources/createResource [post]
func (h *HarvestController) CreateResource(ctx *gin.Context) {
	var resource ds.HarvestResource
	if err := ctx.ShouldBindJSON(&resource); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный формат данных: "+err.Error())
		return
	}

	// Устанавливаем дефолтные значения
	resource.ImageURL = "/images/default.jpg"
	resource.IsDeleted = false
	resource.CreatedAt = time.Now()

	if err := h.HarvestModel.CreateResource(&resource); err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка создания ресурса: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"message":  "Ресурс успешно создан",
		"resource": resource,
	})
}

// UpdateResource godoc
// @Summary Обновить ресурс
// @Description Обновить информацию о ресурсе урожая (только для администраторов)
// @Tags admin
// @Accept json
// @Produce json
// @Param id path int true "ID ресурса"
// @Param updates body ds.HarvestResource true "Обновляемые поля"
// @Success 200 {object} ds.StandardResponse
// @Failure 400 {object} ds.StandardResponse
// @Failure 500 {object} ds.StandardResponse
// @Security BearerAuth
// @Router /harvestResources/{id}/update [put]
func (h *HarvestController) UpdateResource(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный ID ресурса")
		return
	}

	var updates ds.HarvestResource
	if err := ctx.ShouldBindJSON(&updates); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный формат данных")
		return
	}

	updateMap := make(map[string]interface{})

	if updates.Name != "" {
		updateMap["name"] = updates.Name
	}
	if updates.Tariff != "" {
		updateMap["tariff"] = updates.Tariff
	}
	if updates.TariffCost != 0 {
		updateMap["tariff_cost"] = updates.TariffCost
	}
	if updates.Measurement != "" {
		updateMap["measurement"] = updates.Measurement
	}
	if updates.Description != "" {
		updateMap["description"] = updates.Description
	}
	if updates.DetailedDescription != "" {
		updateMap["detailed_description"] = updates.DetailedDescription
	}
	if updates.Requirement != 0 {
		updateMap["requirement"] = updates.Requirement
	}
	if updates.ImageURL != "" {
		updateMap["image_url"] = updates.ImageURL
	}

	if len(updateMap) == 0 {
		h.errorResponse(ctx, http.StatusBadRequest, "Нет данных для обновления")
		return
	}

	if err := h.HarvestModel.UpdateResource(id, updateMap); err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка обновления ресурса: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"message": "Ресурс успешно обновлен",
	})
}

// DeleteResource godoc
// @Summary Удалить ресурс
// @Description Удалить ресурс урожая (только для администраторов)
// @Tags admin
// @Accept json
// @Produce json
// @Param id path int true "ID ресурса"
// @Success 200 {object} ds.StandardResponse
// @Failure 400 {object} ds.StandardResponse
// @Failure 500 {object} ds.StandardResponse
// @Security BearerAuth
// @Router /harvestResources/{id}/delete [delete]
func (h *HarvestController) DeleteResource(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный ID ресурса")
		return
	}

	if err := h.HarvestModel.DeleteResource(id); err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка удаления ресурса: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"message": "Ресурс успешно удален",
	})
}

// SetResourceImage godoc
// @Summary Установить изображение для ресурса
// @Description Установить или обновить изображение ресурса (только для администраторов)
// @Tags admin
// @Accept json
// @Produce json
// @Param id path int true "ID ресурса"
// @Param request body ds.SetResourceImageRequest true "URL изображения"
// @Success 200 {object} ds.StandardResponse
// @Failure 400 {object} ds.StandardResponse
// @Failure 500 {object} ds.StandardResponse
// @Security BearerAuth
// @Router /harvestResources/{id}/setImage [post]
func (h *HarvestController) SetResourceImage(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный ID ресурса")
		return
	}

	var request struct {
		ImageURL string `json:"image_url" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный формат данных")
		return
	}

	if err := h.HarvestModel.SetResourceImage(id, request.ImageURL); err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка установки изображения: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"message": "Изображение успешно установлено",
	})
}

// GetUserCart godoc
// @Summary Получить корзину пользователя
// @Description Получить информацию о корзине заявок пользователя
// @Tags application
// @Accept json
// @Produce json
// @Param id path int true "ID пользователя"
// @Success 200 {object} ds.StandardResponse
// @Failure 400 {object} ds.StandardResponse
// @Security BearerAuth
// @Router /harvestApplication/{id}/harvestCart [get]
func (h *HarvestController) GetUserCart(ctx *gin.Context) {
	userIDStr, _ := ctx.Get("user_id")
	userID := userIDStr.(int)
	fmt.Println(userID)
	harvestApplication, err := h.HarvestModel.GetUserDraftApplication(userID)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный формат данных")
		return
	}
	amount, err := h.HarvestModel.GetHarvestApplicationCount(harvestApplication.ID)

	h.successResponse(ctx, gin.H{
		"application_ID": harvestApplication.ID,
		"amount":         amount,
	})
}

// GetHarvestApplications godoc
// @Summary Получить все заявки
// @Description Получить список всех заявок (только для администраторов)
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} ds.StandardResponse
// @Failure 500 {object} ds.StandardResponse
// @Security BearerAuth
// @Router /harvestApplications [get]
func (h *HarvestController) GetHarvestApplications(ctx *gin.Context) {
	applications, err := h.HarvestModel.GetAllHarvestApplications()
	if err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка получения заявок: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"applications": applications,
		"total":        len(applications),
	})
}

// DeclineApplication godoc
// @Summary Отклонить заявку
// @Description Отклонить заявку на урожай (только для администраторов)
// @Tags admin
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Param request body ds.DeclineApplicationRequest true "Данные отклонения"
// @Success 200 {object} ds.StandardResponse
// @Failure 400 {object} ds.StandardResponse
// @Failure 500 {object} ds.StandardResponse
// @Security BearerAuth
// @Router /harvestApplication/{id}/decline [put]
func (h *HarvestController) DeclineApplication(ctx *gin.Context) {
	applicationIDStr := ctx.Param("id")
	applicationID, err := strconv.Atoi(applicationIDStr)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный ID заявки")
		return
	}

	var request struct {
		ModeratorID int    `json:"moderator_id" binding:"required"`
		Notes       string `json:"notes" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный формат данных")
		return
	}

	if err := h.HarvestModel.DeclineApplication(applicationID, request.ModeratorID, request.Notes); err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка отклонения заявки: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"message": "Заявка успешно отклонена",
	})
}

// DeleteApplicationResource godoc
// @Summary Удалить ресурс из заявки
// @Description Удалить ресурс из заявки на урожай
// @Tags application
// @Accept json
// @Produce json
// @Param id path int true "ID ресурса"
// @Param request body ds.DeleteApplicationResourceRequest true "ID заявки"
// @Success 200 {object} ds.StandardResponse
// @Failure 400 {object} ds.StandardResponse
// @Failure 500 {object} ds.StandardResponse
// @Security BearerAuth
// @Router /applicationResource/{id}/deleteResource [put]
func (h *HarvestController) DeleteApplicationResource(ctx *gin.Context) {
	resourceIDStr := ctx.Param("id")
	resourceID, err := strconv.Atoi(resourceIDStr)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный ID ресурса")
		return
	}

	var request struct {
		ApplicationID int `json:"application_id" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный формат данных")
		return
	}

	if err := h.HarvestModel.DeleteApplicationResource(request.ApplicationID, resourceID); err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка удаления ресурса из заявки: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"message": "Ресурс успешно удален из заявки",
	})
}

// SetApplicationResourceCoeff godoc
// @Summary Установить коэффициент для ресурса в заявке
// @Description Установить коэффициент для ресурса в заявке на урожай
// @Tags application
// @Accept json
// @Produce json
// @Param id path int true "ID ресурса"
// @Param request body ds.SetApplicationResourceCoeffRequest true "Данные коэффициента"
// @Success 200 {object} ds.StandardResponse
// @Failure 400 {object} ds.StandardResponse
// @Failure 500 {object} ds.StandardResponse
// @Security BearerAuth
// @Router /applicationResource/{id}/setCoeff [post]
func (h *HarvestController) SetApplicationResourceCoeff(ctx *gin.Context) {
	resourceIDStr := ctx.Param("id")
	resourceID, err := strconv.Atoi(resourceIDStr)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный ID ресурса")
		return
	}

	var request struct {
		ApplicationID int     `json:"application_id" binding:"required"`
		Coefficient   float64 `json:"coefficient" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный формат данных")
		return
	}

	if err := h.HarvestModel.SetApplicationResourceCoeff(request.ApplicationID, resourceID, request.Coefficient); err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка установки коэффициента: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"message": "Коэффициент успешно установлен",
	})
}

// RegisterUser godoc
// @Summary Регистрация пользователя
// @Description Зарегистрировать нового пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ds.RegisterRequest true "Данные регистрации"
// @Success 200 {object} ds.StandardResponse
// @Failure 400 {object} ds.StandardResponse
// @Failure 500 {object} ds.StandardResponse
// @Router /users/register [post]
func (h *HarvestController) RegisterUser(ctx *gin.Context) {
	var request ds.RegisterRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный формат данных")
		return
	}

	existingUser, _ := h.HarvestModel.GetUserByUsername(request.Username)
	if existingUser != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Пользователь с таким именем уже существует")
		return
	}

	user := &ds.User{
		Username:     request.Username,
		PasswordHash: request.Password, // хэШ!!!!!!!
		Role:         "user",
	}

	if request.Role != "" {
		user.Role = request.Role
	}

	if err := h.HarvestModel.CreateUser(user); err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка создания пользователя: "+err.Error())
		return
	}

	token, err := h.HarvestModel.GenerateToken(user)
	if err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка генерации токена: "+err.Error())
		return
	}

	// Устанавливаем куку
	ctx.SetCookie(
		"harvest_jwt",               // имя куки
		token,                       // значение токена
		int(24*time.Hour.Seconds()), // время жизни (24 часа)
		"/",                         // путь
		"",                          // домен
		false,                       // secure
		true,                        // httpOnly
	)

	userResponse := gin.H{
		"id":       user.ID,
		"username": user.Username,
		"role":     user.Role,
	}

	h.successResponse(ctx, gin.H{
		"message": "Пользователь успешно создан",
		"user":    userResponse,
	})
}

// GetUser godoc
// @Summary Получить информацию о пользователе
// @Description Получить информацию о пользователе по ID (только для администраторов)
// @Tags admin
// @Accept json
// @Produce json
// @Param id path int true "ID пользователя"
// @Success 200 {object} ds.StandardResponse
// @Failure 400 {object} ds.StandardResponse
// @Failure 404 {object} ds.StandardResponse
// @Security BearerAuth
// @Router /users/{id} [get]
func (h *HarvestController) GetUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный ID пользователя")
		return
	}

	user, err := h.HarvestModel.GetUserByID(id)
	if err != nil {
		h.errorResponse(ctx, http.StatusNotFound, "Пользователь не найден")
		return
	}

	userResponse := gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"role":       user.Role,
		"created_at": user.CreatedAt,
	}

	h.successResponse(ctx, userResponse)
}

// SetUserChanges godoc
// @Summary Изменить данные пользователя
// @Description Изменить данные пользователя (только для администраторов)
// @Tags admin
// @Accept json
// @Produce json
// @Param id path int true "ID пользователя"
// @Param request body ds.UpdateUserRequest true "Обновляемые данные"
// @Success 200 {object} ds.StandardResponse
// @Failure 400 {object} ds.StandardResponse
// @Failure 500 {object} ds.StandardResponse
// @Security BearerAuth
// @Router /users/{id}/setChanges [put]
func (h *HarvestController) SetUserChanges(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный ID пользователя")
		return
	}

	var request ds.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный формат данных")
		return
	}

	updates := make(map[string]interface{})
	if request.Username != "" {
		existingUser, _ := h.HarvestModel.GetUserByUsername(request.Username)
		if existingUser != nil && existingUser.ID != id {
			h.errorResponse(ctx, http.StatusBadRequest, "Имя пользователя уже занято")
			return
		}
		updates["username"] = request.Username
	}
	if request.Role != "" {
		updates["role"] = request.Role
	}

	if err := h.HarvestModel.UpdateUser(id, updates); err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка обновления пользователя: "+err.Error())
		return
	}

	h.successResponse(ctx, gin.H{
		"message": "Данные пользователя успешно обновлены",
	})
}

// LoginUser godoc
// @Summary Вход пользователя
// @Description Аутентификация пользователя и получение токена
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ds.LoginRequest true "Данные входа"
// @Success 200 {object} ds.StandardResponse
// @Failure 400 {object} ds.StandardResponse
// @Failure 401 {object} ds.StandardResponse
// @Router /users/login [post]
func (h *HarvestController) LoginUser(ctx *gin.Context) {
	var request ds.LoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Неверный формат данных")
		return
	}

	user, err := h.HarvestModel.GetUserByUsername(request.Username)
	if err != nil {
		h.errorResponse(ctx, http.StatusUnauthorized, "Неверное имя пользователя или пароль")
		return
	}

	if request.Password != user.PasswordHash {
		h.errorResponse(ctx, http.StatusUnauthorized, "Неверное имя пользователя или пароль")
		return
	}

	token, err := h.HarvestModel.GenerateToken(user)
	if err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка генерации токена: "+err.Error())
		return
	}

	// Устанавливаем куку
	ctx.SetCookie(
		"harvest_jwt",               // имя куки
		token,                       // значение токена
		int(24*time.Hour.Seconds()), // время жизни (24 часа)
		"/",                         // путь
		"",                          // домен
		false,                       // secure
		true,                        // httpOnly
	)

	userResponse := gin.H{
		"id":       user.ID,
		"username": user.Username,
		"role":     user.Role,
	}

	h.successResponse(ctx, gin.H{
		"message": "Успешный вход",
		"user":    userResponse,
	})
}

// LogoutUser godoc
// @Summary Выход пользователя
// @Description Выход пользователя и инвалидация токена
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} ds.StandardResponse
// @Failure 400 {object} ds.StandardResponse
// @Failure 500 {object} ds.StandardResponse
// @Security BearerAuth
// @Router /users/logout [post]
func (h *HarvestController) LogoutUser(ctx *gin.Context) {
	// Получаем токен из куки
	token, err := ctx.Cookie("harvest_jwt")
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Токен не найден")
		return
	}

	// Добавляем токен в Redis blacklist
	err = h.redisClient.WriteJWTToBlacklist(ctx, token, 24*time.Hour) // TTL 24 часа
	if err != nil {
		h.errorResponse(ctx, http.StatusInternalServerError, "Ошибка при выходе: "+err.Error())
		return
	}

	// Удаляем куку
	ctx.SetCookie(
		"harvest_jwt",
		"",
		-1, // удаляем куку
		"/",
		"",
		false,
		true,
	)

	h.successResponse(ctx, gin.H{
		"message": "Успешный выход",
	})
}
