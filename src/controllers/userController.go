package controllers

import (
	"context"
	"estore/src/database"
	"estore/src/models"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func Agents(c *fiber.Ctx) error {
	var users []models.User

	database.DB.Where("is_agent = true").Find(&users)
	return c.JSON(users)
}

func Rankings(c *fiber.Ctx) error {
	rankings, err := database.Cache.ZRevRangeByScoreWithScores(context.Background(), "rankings", &redis.ZRangeBy{
		Min: "-inf",
		Max: "+inf",
	}).Result()

	if err != nil {
		return err
	}

	return c.JSON(rankings)
}
