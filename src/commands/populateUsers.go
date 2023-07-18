package main

import (
	"estore/src/database"
	"estore/src/models"
	"github.com/go-faker/faker/v4"
)

func main() {
	database.ConnectToDB()

	for i := 0; i < 30; i++ {
		agent := models.User{
			FirstName: faker.FirstName(),
			LastName:  faker.LastName(),
			Email:     faker.Email(),
			IsAgent:   true,
		}
		agent.HashPassword("1")
		database.DB.Create(&agent)
	}
}
