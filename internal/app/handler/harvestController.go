package handler

import (
	"context"
	"os"
	"rip2025/internal/app/config"
	"rip2025/internal/app/redis"
	"rip2025/internal/app/repository"
	"strconv"
	"time"
)

type HarvestController struct {
	HarvestModel *repository.HarvestModel
	JWTSecret    string
	redisClient  *redis.Client
}

func NewHarvestController(r *repository.HarvestModel) *HarvestController {
	jwtSecret := os.Getenv("JWT_SECRET")

	port, _ := strconv.Atoi(os.Getenv("REDIS_PORT"))
	redisConfig := config.RedisConfig{
		Host:        os.Getenv("REDIS_HOST"),
		Port:        port,
		Password:    os.Getenv("REDIS_PASSWORD"),
		User:        os.Getenv("REDIS_USER"),
		DialTimeout: 10 * time.Second,
		ReadTimeout: 10 * time.Second,
	}

	redisClient, err := redis.New(context.Background(), redisConfig)
	if err != nil {
		panic(err)
	}

	return &HarvestController{
		HarvestModel: r,
		JWTSecret:    jwtSecret,
		redisClient:  redisClient,
	}
}
