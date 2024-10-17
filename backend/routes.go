package main

import (
	"iou_tracker/controllers"

	"github.com/gofiber/fiber/v2"
)

func createEndpoints(app *fiber.App) {

	// controllers
	userController := controllers.NewUserController()

	// api
	api := app.Group("/api")
	api.Post("/login", userController.Login)
	api.Post("/refresh", userController.RefreshToken)

	// user group
	userGroup := api.Group("/users")
	userGroup.Get("", userController.List)
	userGroup.Post("/register", userController.Register)

}
