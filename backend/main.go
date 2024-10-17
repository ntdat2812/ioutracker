package main

import (
	"iou_tracker/infra"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Set up DB
	infra.ConnectToMongoDB()
	defer infra.DisconnectMongoDB()

	// Initialize fiber app
	app := fiber.New()

	createEndpoints(app)

	app.Listen(":3000")
}
