package collections

import (
	"iou_tracker/constants"
	"iou_tracker/infra"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Debt struct {
	ID         primitive.ObjectID   `json:"id,omitempty" bson:"_id,omitempty"`
	BorrowerID primitive.ObjectID   `json:"borrower_id" bson:"borrower_id"`
	LenderID   primitive.ObjectID   `json:"lender_id" bson:"lender_id"`
	Amount     float64              `json:"amount,omitempty" bson:"amount,omitempty"`
	Date       time.Time            `json:"date,omitempty" bson:"date,omitempty"`
	Note       string               `json:"note,omitempty" bson:"note,omitempty"`
	Status     constants.DebtStatus `json:"status,omitempty" bson:"status,omitempty"`
	CreatedAt  time.Time            `json:"created_at,omitempty" bson:"created_at,omitempty"`
	UpdatedAt  time.Time            `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
}

func GetDebtCollection() *mongo.Collection {
	return infra.MongoDB.Database(infra.DB).Collection("debt")
}
