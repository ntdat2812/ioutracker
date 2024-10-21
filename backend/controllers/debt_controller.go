package controllers

import (
	"iou_tracker/collections"
	"iou_tracker/constants"
	"iou_tracker/util"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type debtController struct {
}

func NewDebtController() *debtController {
	return &debtController{}
}

type CreateDebtRequest struct {
	BorrowerID string    `json:"borrower_id" validate:"required"`
	Amount     float64   `json:"amount" validate:"required,gte=0"`
	Date       time.Time `json:"date" validate:"required"`
	Note       string    `json:"note"`
}

func (d *debtController) Create(c *fiber.Ctx) error {
	var req CreateDebtRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	userId := c.Locals("userID").(string)

	// validate
	validationErrors := util.Validate(req)
	if len(validationErrors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errors": validationErrors,
		})
	}

	// extract info
	lenderID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid lender ID"})
	}

	borrowerID, err := primitive.ObjectIDFromHex(req.BorrowerID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid borrower ID"})
	}

	// persist to DB
	debtCollection := collections.GetDebtCollection()

	debt := collections.Debt{
		BorrowerID: borrowerID,
		LenderID:   lenderID,
		Amount:     req.Amount,
		Date:       req.Date,
		Note:       req.Note,
		Status:     constants.UnpaidStatus,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	_, err = debtCollection.InsertOne(c.Context(), debt)
	if err != nil {
		log.Errorf("Failed to create debt, err: %v", err.Error())
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to create debt")
	}

	return c.JSON(fiber.Map{"success": true})
}

func (d *debtController) Update(c *fiber.Ctx) error {

	return c.JSON(fiber.Map{"success": true})
}

func (d *debtController) Delete(c *fiber.Ctx) error {

	return c.JSON(fiber.Map{"success": true})
}

func (d *debtController) ListByUser(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true})
}

func (d *debtController) MarkAsPaid(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true})
}

func (d *debtController) Remind(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true})
}
