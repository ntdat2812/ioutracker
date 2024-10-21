package controllers

import (
	"iou_tracker/collections"
	"iou_tracker/constants"
	"iou_tracker/util"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type debtController struct {
}

func NewDebtController() *debtController {
	return &debtController{}
}

type CreateDebtRequest struct {
	BorrowerID string  `json:"borrower_id" validate:"required"`
	Amount     float64 `json:"amount" validate:"required,gte=0"`
	Date       string  `json:"date" validate:"required"`
	Note       string  `json:"note"`
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
		Date:       util.ConvertToDate(req.Date),
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

type UpdateDebtRequest struct {
	Amount float64 `json:"amount"`
	Date   string  `json:"date"`
	Note   string  `json:"note"`
}

func (d *debtController) Update(c *fiber.Ctx) error {

	debtId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	// find existing debt
	_, err = d.findExistingDebt(c, debtId)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Debt not found"})
	}

	var req UpdateDebtRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	updateFields := bson.M{}
	if req.Amount > 0 {
		updateFields["amount"] = req.Amount
	}

	if req.Date != "" {
		updateFields["date"] = util.ConvertToDate(req.Date)
	}

	if req.Note != "" {
		updateFields["note"] = req.Note
	}

	if len(updateFields) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No fields to update"})
	}

	// update to db
	debtCollection := collections.GetDebtCollection()
	_, err = debtCollection.UpdateOne(
		c.Context(),
		bson.M{"_id": debtId},
		bson.M{"$set": updateFields},
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not update debt"})
	}

	return c.JSON(fiber.Map{"success": true, "updatedFields": updateFields})
}

func (d *debtController) Delete(c *fiber.Ctx) error {

	debtId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	// find existing debt
	_, err = d.findExistingDebt(c, debtId)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Debt not found"})
	}

	debtCollection := collections.GetDebtCollection()
	_, err = debtCollection.DeleteOne(c.Context(), bson.M{"_id": debtId})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not delete debt"})
	}

	return c.JSON(fiber.Map{"success": true})
}

func (d *debtController) ListByUser(c *fiber.Ctx) error {

	userId, err := primitive.ObjectIDFromHex(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	role := c.Query("role")

	// build filer
	filter := bson.M{}
	switch role {
	case "borrower":
		filter = bson.M{"borrower_id": userId}
	case "lender":
		filter = bson.M{"lender_id": userId}
	case "":
		filter = bson.M{"$or": []bson.M{
			{"borrower_id": userId},
			{"lender_id": userId},
		}}
	}

	// find in db
	debtCollection := collections.GetDebtCollection()
	cursor, err := debtCollection.Find(c.Context(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not fetch debts"})
	}
	defer cursor.Close(c.Context())

	debts := make([]collections.Debt, 0)
	if err = cursor.All(c.Context(), &debts); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not parse debts"})
	}

	return c.JSON(debts)
}

func (d *debtController) Remind(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true})
}

func (d *debtController) findExistingDebt(c *fiber.Ctx, debtId primitive.ObjectID) (*collections.Debt, error) {

	debtCollection := collections.GetDebtCollection()
	existingDebt := &collections.Debt{}
	err := debtCollection.FindOne(c.Context(), bson.M{"_id": debtId}).Decode(existingDebt)
	if err != nil {
		log.Errorf("failed to findExistingDebt, err: %v", err.Error())
		return nil, err
	}

	return existingDebt, nil
}
