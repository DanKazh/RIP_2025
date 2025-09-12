package repository

import (
	"fmt"
	"strings"
)

type HarvestModel struct {
}

func NewHarvestModel() (*HarvestModel, error) {
	return &HarvestModel{}, nil
}

type HarvestResource struct {
	ID                  int
	Name                string
	Tariff              string
	TariffCost          int
	Measurment          string
	Description         string
	DetailedDescription string
	ImageURL            string
}

type HarvestCulture struct {
	ID   int
	Name string
}

type HarvestApplicationItem struct {
	ResourceID   int
	ResourceCost int
	UsageType    string
}

type HarvestApplication struct {
	ID      int
	Items   []HarvestApplicationItem
	Culture HarvestCulture
}

func (r *HarvestModel) GetHarvestResources() ([]HarvestResource, error) {
	harvestResources := []HarvestResource{
		{
			ID:                  1,
			Name:                "Тепличная площадь",
			Tariff:              "Руб/м²",
			TariffCost:          150,
			Measurment:          "м²",
			Description:         "Площадь теплицы для выращивания культур с учетом северных условий",
			DetailedDescription: "Выбор и расчет площади теплицы для северных регионов (с коротким световым днем, низкими температурами, сильными ветрами и снеговыми нагрузками) — это критически важная задача.",
			ImageURL:            "http://127.0.0.1:9000/lab1/image.png",
		},
		{
			ID:                  2,
			Name:                "Электроэнергия",
			Tariff:              "Руб/кВт·ч",
			TariffCost:          250,
			Measurment:          "кВт·ч",
			Description:         "Энергоснабжение для освещения и оборудования теплицы",
			DetailedDescription: "Обеспечение стабильного электроснабжения для систем освещения, вентиляции и автоматизации теплицы в условиях северных регионов с учетом перепадов напряжения и низких температур.",
			ImageURL:            "http://127.0.0.1:9000/lab1/image_1.webp",
		},
		{
			ID:                  3,
			Name:                "Отопление",
			Tariff:              "Руб/Гкал",
			TariffCost:          350,
			Measurment:          "Гкал",
			Description:         "Поддержание оптимальной температуры в северных условиях",
			DetailedDescription: "Системы отопления, специально разработанные для суровых северных условий. Обеспечивают поддержание оптимальной температуры для роста растений даже в самые холодные периоды.",
			ImageURL:            "http://127.0.0.1:9000/lab1/image_2.jpg",
		},
		{
			ID:                  4,
			Name:                "Водоснабжение",
			Tariff:              "Руб/м³",
			TariffCost:          150,
			Measurment:          "м³",
			Description:         "Полив и системы орошения для тепличных культур",
			DetailedDescription: "Системы водоснабжения и капельного орошения, адаптированные для северных условий. Включают подогрев воды и защиту от замерзания трубопроводов.",
			ImageURL:            "http://127.0.0.1:9000/lab1/image_3.webp",
		},
		{
			ID:                  5,
			Name:                "Семена",
			Tariff:              "Руб/пакет",
			TariffCost:          150,
			Measurment:          "пакетов",
			Description:         "Семена, адаптированные для северных условий выращивания",
			DetailedDescription: "Специально отобранные и обработанные семена сельскохозяйственных культур, устойчивых к низким температурам и короткому световому дню северных регионов.",
			ImageURL:            "http://127.0.0.1:9000/lab1/image_4.jpg",
		},
		{
			ID:                  6,
			Name:                "Удобрения",
			Tariff:              "Руб/кг",
			TariffCost:          150,
			Measurment:          "кг",
			Description:         "Специализированные удобрения для северных теплиц",
			DetailedDescription: "Комплексные удобрения и питательные составы, разработанные специально для условий северного земледелия. Способствуют ускоренному росту и повышению урожайности в условиях недостатка естественного света и тепла.",
			ImageURL:            "http://127.0.0.1:9000/lab1/image_5.webp",
		},
	}

	if len(harvestResources) == 0 {
		return nil, fmt.Errorf("массив услуг пустой")
	}

	return harvestResources, nil
}

func (r *HarvestModel) GetHarvestCultures() ([]HarvestCulture, error) {
	harvestCultures := []HarvestCulture{
		{ID: 1, Name: "Огурцы"},
		{ID: 2, Name: "Картофель"},
		{ID: 3, Name: "Баклажаны"},
		{ID: 4, Name: "Помидоры"},
		{ID: 5, Name: "Морковь"},
	}
	return harvestCultures, nil
}

func (r *HarvestModel) GetHarvestApplication(id int) ([]map[string]interface{}, error) {
	if id != 1 {
		return nil, fmt.Errorf("заявка не найдена")
	}

	harvestApplication := &HarvestApplication{
		ID: 1,
		Culture: HarvestCulture{
			ID:   1,
			Name: "Огурцы",
		},
		Items: []HarvestApplicationItem{
			{
				ResourceID:   1,
				UsageType:    "Экономно",
				ResourceCost: 400,
			},
		},
	}

	harvestResources, err := r.GetHarvestResources()
	if err != nil {
		return nil, err
	}

	var applicationItems []map[string]interface{}
	for _, item := range harvestApplication.Items {
		for _, resource := range harvestResources {
			if resource.ID == item.ResourceID {
				applicationItem := map[string]interface{}{
					"ResourceID":         item.ResourceID,
					"ResourceImageURL":   resource.ImageURL,
					"ResourceName":       resource.Name,
					"ResourceTariffCost": resource.TariffCost,
					"ResourceTariff":     resource.Tariff,
					"ResourceMeasurment": resource.Measurment,
					"ResourceCost":       item.ResourceCost,
					"UsageType":          item.UsageType,
				}
				applicationItems = append(applicationItems, applicationItem)
				break
			}
		}
	}

	return applicationItems, nil
}

func (r *HarvestModel) GetHarvestResource(id int) (HarvestResource, error) {
	harvestResources, err := r.GetHarvestResources()
	if err != nil {
		return HarvestResource{}, err
	}

	for _, resource := range harvestResources {
		if resource.ID == id {
			return resource, nil
		}
	}
	return HarvestResource{}, fmt.Errorf("услуга не найдена")
}

func (r *HarvestModel) GetHarvestResourcesByTitle(title string) ([]HarvestResource, error) {
	harvestResources, err := r.GetHarvestResources()
	if err != nil {
		return []HarvestResource{}, err
	}

	var result []HarvestResource
	for _, resource := range harvestResources {
		if strings.Contains(strings.ToLower(resource.Name), strings.ToLower(title)) {
			result = append(result, resource)
		}
	}

	return result, nil
}

func (r *HarvestModel) GetHarvestApplicationCount(applicationID int) (int, error) {
	harvestApplication, err := r.GetHarvestApplication(applicationID)
	if err != nil {
		return 0, err
	}
	return len(harvestApplication), nil
}
