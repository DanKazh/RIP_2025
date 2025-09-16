package ds

import "time"

type HarvestResource struct {
	ID                  int `gorm:"primaryKey"`
	Name                string
	Tariff              string
	TariffCost          int
	Measurement         string
	Description         string
	DetailedDescription string
	ImageURL            string
	IsDeleted           bool
	CreatedAt           time.Time
	DeletedAt           *time.Time
	Requirement         float64
}

type HarvestCulture struct {
	ID           int    `gorm:"primaryKey"`
	Name         string `gorm:"unique"`
	CreatedAt    time.Time
	IsDeleted    bool
	Productivity int
}

type ApplicationResource struct {
	ApplicationID int `gorm:"primaryKey"`
	ResourceID    int `gorm:"primaryKey"`
	Ratio         int
	NeededAmount  int
	TotalCost     int
	CreatedAt     time.Time
	Resource      HarvestResource `gorm:"foreignKey:ResourceID"`
}

type HarvestApplication struct {
	ID             int `gorm:"primaryKey"`
	Status         string
	CreatorID      int
	CreatedAt      time.Time
	FormationDate  *time.Time
	CompletionDate *time.Time
	ModeratorID    *int
	CultureID      int
	Weight         int
	FullCost       int
	Notes          *string

	Culture   HarvestCulture        `gorm:"foreignKey:CultureID"`
	Resources []ApplicationResource `gorm:"foreignKey:ApplicationID"`
}
