package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"strings"
	"time"
)

const SecretKey = "secret"

func IsAuthenticated(c *fiber.Ctx) error {
	cookie := c.Cookies("jwt")
	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(cookie, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})

	if err != nil || !token.Valid {
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{
			"message": "Unauthenticated.",
		})
	}

	c.Locals("userId", claims["sub"])

	isAgent := strings.Contains(c.Path(), "/api/agent")
	if (claims["role"] == "admin" && isAgent) || (claims["role"] == "agent" && !isAgent) {
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{
			"message": "Unauthorized.",
		})
	}
	return c.Next()
}

func GenerateJWT(id uint, role string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  id,
		"role": role,
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString([]byte(SecretKey))
}

func ConvertUserId(value interface{}) uint {
	id := value.(float64)
	return uint(id)
}
