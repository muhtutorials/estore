package controllers

import (
	"estore/src/database"
	"estore/src/middleware"
	"estore/src/models"
	"github.com/gofiber/fiber/v2"
	"strings"
	"time"
)

func Register(c *fiber.Ctx) error {
	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	if data["password"] != data["password_confirm"] {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": "Passwords don't match.",
		})
	}

	user := models.User{
		FirstName: data["first_name"],
		LastName:  data["last_name"],
		Email:     data["email"],
		IsAgent:   strings.Contains(c.Path(), "/api/agent"),
	}
	user.HashPassword(data["password"])

	database.DB.Create(&user)

	return c.JSON(user)
}

func Login(c *fiber.Ctx) error {
	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	var user models.User

	database.DB.Where("email = ?", data["email"]).First(&user)

	if user.Id == 0 {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": "Invalid credentials.",
		})
	}

	if err := user.CheckPassword(data["password"]); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": "Invalid credentials.",
		})
	}

	isAgent := strings.Contains(c.Path(), "/api/agent")

	var role string

	if isAgent {
		role = "agent"
	} else {
		role = "admin"
	}

	if !isAgent && user.IsAgent {
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{
			"message": "Unauthorized.",
		})
	}

	token, err := middleware.GenerateJWT(user.Id, role)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": "Invalid credentials.",
		})
	}

	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(time.Hour * 24),
		HTTPOnly: true,
	}
	c.Cookie(&cookie)

	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func User(c *fiber.Ctx) error {
	var user models.User
	database.DB.Where("id = ?", c.Locals("userId")).First(&user)

	if strings.Contains(c.Path(), "/api/agent") {
		agent := models.Agent(user)
		agent.CalculateRevenue(database.DB)
		return c.JSON(agent)
	}

	admin := models.Admin(user)
	admin.CalculateRevenue(database.DB)
	return c.JSON(admin)
}

func Logout(c *fiber.Ctx) error {
	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
	}

	c.Cookie(&cookie)

	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func UpdateInfo(c *fiber.Ctx) error {
	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	userId := middleware.ConvertUserId(c.Locals("userId"))

	user := models.User{
		FirstName: data["first_name"],
		LastName:  data["last_name"],
		Email:     data["email"],
	}
	user.Id = userId
	database.DB.Model(&user).Updates(&user)

	return c.JSON(user)
}

func UpdatePassword(c *fiber.Ctx) error {
	var data map[string]string
	var user models.User

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	if data["password"] != data["password_confirm"] {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": "Passwords don't match.",
		})
	}

	userId := middleware.ConvertUserId(c.Locals("userId"))

	database.DB.Where("id = ?", userId).First(&user)

	if err := user.CheckPassword(data["old_password"]); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": "Wrong password.",
		})
	}
	user.HashPassword(data["password"])

	database.DB.Model(&user).Updates(&user)

	return c.JSON(user)
}
