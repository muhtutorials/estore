package controllers

import (
	"context"
	"estore/src/database"
	"estore/src/models"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/checkout/session"
	"net/smtp"
)

func Orders(c *fiber.Ctx) error {
	var orders []models.Order

	database.DB.Preload("OrderItems").Find(&orders)

	for i, order := range orders {
		orders[i].Name = order.GetFullName()
		orders[i].Total = order.GetTotal()
	}

	return c.JSON(orders)
}

type CreateOrderRequest struct {
	Code      string
	FirstName string
	LastName  string
	Email     string
	Address   string
	Country   string
	City      string
	Zip       string
	Products  []map[string]int
}

func CreateOrder(c *fiber.Ctx) error {
	var request CreateOrderRequest

	if err := c.BodyParser(&request); err != nil {
		return err
	}

	link := models.Link{
		Code: request.Code,
	}

	database.DB.Where(&link).Preload("User").First(&link)

	if link.Id == 0 {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": "Invalid link.",
		})
	}

	order := models.Order{
		Code:       link.Code,
		UserId:     link.UserId,
		AgentEmail: link.User.Email,
		FirstName:  request.FirstName,
		LastName:   request.LastName,
		Email:      request.Email,
		Address:    request.Address,
		Country:    request.Country,
		City:       request.City,
		Zip:        request.Zip,
	}

	tx := database.DB.Begin()

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	var lineItems []*stripe.CheckoutSessionLineItemParams

	for _, requestProduct := range request.Products {
		product := models.Product{}
		product.Id = uint(requestProduct["product_id"])
		database.DB.First(&product)

		total := product.Price * float64(requestProduct["quantity"])

		orderItem := models.OrderItem{
			OrderId:      order.Id,
			ProductTitle: product.Title,
			Price:        product.Price,
			Quantity:     uint(requestProduct["quantity"]),
			AgentRevenue: 0.1 * total,
			AdminRevenue: 0.9 * total,
		}

		if err := tx.Create(&orderItem).Error; err != nil {
			tx.Rollback()
			c.Status(fiber.StatusBadRequest)
			return c.JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
			Name:        stripe.String(product.Title),
			Description: stripe.String(product.Description),
			Images:      []*string{stripe.String(product.Image)},
			Amount:      stripe.Int64(int64(product.Price) * 100),
			Currency:    stripe.String("usd"),
			Quantity:    stripe.Int64(int64(requestProduct["quantity"])),
		})
	}

	stripe.Key = "sk_test_gE56xaQlAzaeMre05WdsKWzQ00dZHiRvlr"

	params := stripe.CheckoutSessionParams{
		SuccessURL:         stripe.String("http://localhost:8000/success?source={CHECKOUT_SESSION_ID}"),
		CancelURL:          stripe.String("http://localhost:8000/error"),
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems:          lineItems,
	}

	source, err := session.New(&params)

	if err != nil {
		tx.Rollback()
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	order.TransactionId = source.ID

	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	tx.Commit()

	return c.JSON(source)
}

func CompleteOrder(c *fiber.Ctx) error {
	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	order := models.Order{}

	database.DB.Preload("OrderItems").First(&order, "transaction_id = ?", data["transaction_id"])

	if order.Id == 0 {
		c.Status(fiber.StatusNotFound)
		return c.JSON(fiber.Map{
			"message": "Order not found.",
		})
	}

	order.Complete = true
	database.DB.Save(&order)

	go func(order models.Order) {
		agentRevenue := 0.0
		adminRevenue := 0.0

		for _, orderItem := range order.OrderItems {
			agentRevenue += orderItem.AgentRevenue
			adminRevenue += orderItem.AdminRevenue
		}

		user := models.User{}
		user.Id = order.UserId

		database.DB.First(&user)

		database.Cache.ZIncrBy(context.Background(), "rankings", agentRevenue, user.FullName())

		agentMessage := []byte(fmt.Sprintf("You earned $%f from the link #%s", agentRevenue, order.Code))
		smtp.SendMail("<email server address and port>", nil, "noreply@email.com", []string{order.AgentEmail}, agentMessage)

		adminMessage := []byte(fmt.Sprintf("Order #%d with a total of $%f has been completed", order.Id, adminRevenue))
		smtp.SendMail("<email server address and port>", nil, "noreply@email.com", []string{"igorhu13@gmail.com"}, adminMessage)
	}(order)

	return c.JSON(fiber.Map{
		"message": "success",
	})
}
