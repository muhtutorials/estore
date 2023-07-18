package controllers

import (
	"estore/src/database"
	"estore/src/middleware"
	"estore/src/models"
	"github.com/go-faker/faker/v4"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

func Links(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		c.Status(fiber.StatusNotFound)
		return c.Send(nil)
	}

	var links []models.Link

	database.DB.Where("user_id = ?", id).Find(&links)

	for i, link := range links {
		var orders []models.Order
		database.DB.Where("code = ? and complete = true", link.Code).Find(&orders)
		links[i].Orders = orders
	}

	return c.JSON(links)
}

type CreateLinkRequest struct {
	Products []int
}

func CreateLink(c *fiber.Ctx) error {
	var request CreateLinkRequest

	if err := c.BodyParser(&request); err != nil {
		return err
	}

	userId := middleware.ConvertUserId(c.Locals("userId"))

	link := models.Link{
		UserId: userId,
		Code:   faker.Username(),
	}

	for _, productId := range request.Products {
		product := models.Product{}
		product.Id = uint(productId)
		link.Products = append(link.Products, product)
	}

	database.DB.Create(&link)

	return c.JSON(link)
}

func Stats(c *fiber.Ctx) error {
	var links []models.Link

	userId := middleware.ConvertUserId(c.Locals("userId"))

	database.DB.Find(&links, models.Link{
		UserId: userId,
	})

	var result []interface{}
	var orders []models.Order

	for _, link := range links {
		database.DB.Preload("OrderItems").Find(&orders, models.Order{
			Code:     link.Code,
			Complete: true,
		})

		revenue := 0.0

		for _, order := range orders {
			revenue += order.GetTotal()
		}

		result = append(result, fiber.Map{
			"code":    link.Code,
			"count":   len(orders),
			"revenue": revenue,
		})
	}

	return c.JSON(result)
}

func GetLink(c *fiber.Ctx) error {
	code := c.Params("code")

	link := models.Link{
		Code: code,
	}

	database.DB.Preload("User").Preload("Products").First(&link)

	return c.JSON(link)
}
