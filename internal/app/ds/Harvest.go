package ds

import (
	"time"

	"github.com/golang-jwt/jwt"
)

type JWTClaims struct {
	jwt.StandardClaims
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
}

type HarvestResource struct {
	ID                  int        `gorm:"primaryKey" json:"id,omitempty"`
	Name                string     `json:"name,omitempty"`
	Tariff              string     `json:"tariff,omitempty"`
	TariffCost          int        `json:"tariff_cost,omitempty"`
	Measurement         string     `json:"measurement,omitempty"`
	Description         string     `json:"description,omitempty"`
	DetailedDescription string     `json:"detailed_description,omitempty"`
	ImageURL            string     `json:"image_url,omitempty"`
	IsDeleted           bool       `json:"is_deleted,omitempty"`
	CreatedAt           time.Time  `json:"created_at,omitempty"`
	DeletedAt           *time.Time `json:"deleted_at,omitempty"`
	Requirement         float64    `json:"requirement,omitempty"`
}

type ApplicationResource struct {
	ApplicationID int `gorm:"primaryKey"`
	ResourceID    int `gorm:"primaryKey"`
	Ratio         float64
	NeededAmount  int
	TotalCost     int
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
	Productivity   *int
	Weight         *int
	FullCost       *int
	Notes          *string

	Resources []ApplicationResource `gorm:"foreignKey:ApplicationID"`
}

type HarvestApplicationInfo struct {
	Productivity *int
	Weight       *int
}

type User struct {
	ID           int        `gorm:"primaryKey" json:"id"`
	Username     string     `gorm:"size:50;uniqueIndex" json:"username"`
	PasswordHash string     `gorm:"size:255" json:"-"`
	Role         string     `gorm:"size:20" json:"role"`
	CreatedAt    time.Time  `json:"created_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
	IsDeleted    bool       `gorm:"default:false" json:"is_deleted"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role,omitempty"`
}

type UpdateUserRequest struct {
	Username string `json:"username,omitempty"`
	Role     string `json:"role,omitempty"`
}
