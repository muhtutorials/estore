package routes

import (
	"estore/src/controllers"
	"estore/src/middleware"
	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {
	api := app.Group("api")

	admin := api.Group("admin")
	admin.Post("/register", controllers.Register)
	admin.Post("/login", controllers.Login)

	adminAuthenticated := admin.Use(middleware.IsAuthenticated)
	adminAuthenticated.Get("/user", controllers.User)
	adminAuthenticated.Post("/logout", controllers.Logout)
	adminAuthenticated.Put("/users/info", controllers.UpdateInfo)
	adminAuthenticated.Put("/users/password", controllers.UpdatePassword)
	adminAuthenticated.Get("/agents", controllers.Agents)
	adminAuthenticated.Get("/products", controllers.Products)
	adminAuthenticated.Post("/products", controllers.CreateProduct)
	adminAuthenticated.Get("/products/:id", controllers.GetProduct)
	adminAuthenticated.Put("/products/:id", controllers.UpdateProduct)
	adminAuthenticated.Delete("/products/:id", controllers.DeleteProduct)
	adminAuthenticated.Get("/users/:id/links", controllers.Links)
	adminAuthenticated.Get("/users/:id/orders", controllers.Orders)

	agent := api.Group("agent")
	agent.Post("/register", controllers.Register)
	agent.Post("/login", controllers.Login)
	agent.Get("/products/frontend", controllers.ProductsFrontend)
	agent.Get("/products/backend", controllers.ProductsBackend)
	agentAuthenticated := agent.Use(middleware.IsAuthenticated)
	agentAuthenticated.Get("/user", controllers.User)
	agentAuthenticated.Post("/logout", controllers.Logout)
	agentAuthenticated.Put("/users/info", controllers.UpdateInfo)
	agentAuthenticated.Put("/users/password", controllers.UpdatePassword)
	agentAuthenticated.Post("links", controllers.CreateLink)
	agentAuthenticated.Get("stats", controllers.Stats)
	agentAuthenticated.Get("rankings", controllers.Rankings)

	checkout := api.Group("checkout")
	checkout.Get("links/:code", controllers.GetLink)
	checkout.Post("orders", controllers.CreateOrder)
	checkout.Post("orders/confirm", controllers.CompleteOrder)
}
