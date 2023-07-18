package main

import (
	"context"
	"estore/src/database"
	"estore/src/models"
	"github.com/redis/go-redis/v9"
)

func main() {
	database.ConnectToDB()
	database.ConnectToCache()

	ctx := context.Background()

	var users []models.User

	database.DB.Find(&users, models.User{
		IsAgent: true,
	})

	for _, user := range users {
		agent := models.Agent(user)
		agent.CalculateRevenue(database.DB)

		database.Cache.ZAdd(ctx, "rankings", redis.Z{
			Score:  *agent.Revenue,
			Member: user.FullName(),
		})
	}
}
