package controllers

import (
	"context"
	"encoding/json"
	"estore/src/database"
	"estore/src/models"
	"github.com/gofiber/fiber/v2"
	"sort"
	"strconv"
	"strings"
	"time"
)

func Products(c *fiber.Ctx) error {
	var products []models.Product

	database.DB.Find(&products)

	return c.JSON(products)
}

func CreateProduct(c *fiber.Ctx) error {
	var product models.Product

	if err := c.BodyParser(&product); err != nil {
		return err
	}

	database.DB.Create(&product)

	go database.ClearCache("products_frontend", "products_backend")

	return c.JSON(product)
}

func GetProduct(c *fiber.Ctx) error {
	var product models.Product

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		c.Status(fiber.StatusNotFound)
		return c.Send(nil)
	}
	product.Id = uint(id)

	result := database.DB.First(&product)
	if result.RowsAffected == 0 {
		c.Status(fiber.StatusNotFound)
		return c.Send(nil)
	}

	return c.JSON(product)
}

func UpdateProduct(c *fiber.Ctx) error {
	var product models.Product

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		c.Status(fiber.StatusNotFound)
		return c.Send(nil)
	}
	product.Id = uint(id)

	if err := c.BodyParser(&product); err != nil {
		return err
	}

	result := database.DB.Model(&product).Updates(&product)
	if result.RowsAffected == 0 {
		c.Status(fiber.StatusNotFound)
		return c.Send(nil)
	}

	go database.ClearCache("products_frontend", "products_backend")

	return c.JSON(product)
}

func DeleteProduct(c *fiber.Ctx) error {
	var product models.Product

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		c.Status(fiber.StatusNotFound)
		return c.Send(nil)
	}
	product.Id = uint(id)

	result := database.DB.Delete(&product)
	if result.RowsAffected == 0 {
		c.Status(fiber.StatusNotFound)
		return c.Send(nil)
	}

	go database.ClearCache("products_frontend", "products_backend")

	return nil
}

func ProductsFrontend(c *fiber.Ctx) error {
	var products []models.Product
	var ctx = context.Background()

	result, err := database.Cache.Get(ctx, "products_frontend").Result()

	if err != nil {
		database.DB.Find(&products)

		bytes, err := json.Marshal(products)
		if err != nil {
			panic(err)
		}

		if errorKey := database.Cache.Set(ctx, "products_frontend", bytes, 30*time.Minute).Err(); errorKey != nil {
			panic(errorKey)
		}
	} else {
		json.Unmarshal([]byte(result), &products)
	}

	return c.JSON(products)
}

func ProductsBackend(c *fiber.Ctx) error {
	var products []models.Product
	var ctx = context.Background()

	result, err := database.Cache.Get(ctx, "products_backend").Result()

	if err != nil {
		database.DB.Find(&products)

		bytes, err := json.Marshal(products)
		if err != nil {
			panic(err)
		}

		database.Cache.Set(ctx, "products_backend", bytes, 30*time.Minute)
	} else {
		json.Unmarshal([]byte(result), &products)
	}

	var searchedProducts []models.Product

	if search := c.Query("search"); search != "" {
		lowerCaseSearch := strings.ToLower(search)
		for _, product := range products {
			if strings.Contains(strings.ToLower(product.Title), lowerCaseSearch) ||
				strings.Contains(strings.ToLower(product.Description), lowerCaseSearch) {
				searchedProducts = append(searchedProducts, product)
			}
		}
	} else {
		searchedProducts = products
	}

	if sorting := c.Query("sort"); sorting != "" {
		lowerCaseSort := strings.ToLower(sorting)
		if lowerCaseSort == "asc" {
			sort.Slice(searchedProducts, func(i, j int) bool {
				return searchedProducts[i].Price < searchedProducts[j].Price
			})
		} else if lowerCaseSort == "desc" {
			sort.Slice(searchedProducts, func(i, j int) bool {
				return searchedProducts[i].Price > searchedProducts[j].Price
			})
		}
	}

	total := len(searchedProducts)
	page, _ := strconv.Atoi(c.Query("page", "1"))

	var data []models.Product

	const itemsPerPage = 10

	if total <= page*itemsPerPage && total >= (page-1)*itemsPerPage {
		data = searchedProducts[(page-1)*itemsPerPage : total]
	} else if total >= page*itemsPerPage {
		data = searchedProducts[(page-1)*itemsPerPage : page*itemsPerPage]
	} else {
		data = []models.Product{}
	}

	return c.JSON(fiber.Map{
		"data":  data,
		"items": total,
		"pages": total/itemsPerPage + 1,
	})
}
