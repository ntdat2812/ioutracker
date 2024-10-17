package controllers

import (
	"iou_tracker/collections"
	"iou_tracker/util"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type userController struct {
	jwtHelper *util.JWTHelper
}

func NewUserController() *userController {
	return &userController{
		jwtHelper: util.NewJWTHelper(),
	}
}

type RegisterReq struct {
	Name     string `json:"name,omitempty" validate:"required"`
	Email    string `json:"email,omitempty" validate:"required,email"`
	Gender   string `json:"gender,omitempty" validate:"required"`
	Password string `json:"password,omitempty" validate:"required,min=6"`
}

func (u *userController) Register(c *fiber.Ctx) error {

	// decode
	req := new(RegisterReq)
	if err := c.BodyParser(req); err != nil {
		log.Errorf("failed to parse register body %v", err.Error())
		return c.Status(fiber.StatusBadRequest).SendString("invalid input")
	}

	// validate
	validationErrors := util.Validate(req)
	if len(validationErrors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errors": validationErrors,
		})
	}

	// check existing email
	existingUser, _ := u.fetchUserByEmail(c, req.Email)
	if existingUser != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Email is already in use",
		})
	}

	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error hashing password")
	}

	// persist to DB
	userCollection := collections.GetUserCollection()
	user := &collections.User{
		Name:      req.Name,
		Gender:    req.Gender,
		Email:     req.Email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = userCollection.InsertOne(c.Context(), user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error registering user")
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"success": true})
}

func (u *userController) List(c *fiber.Ctx) error {

	userCollection := collections.GetUserCollection()
	name := c.Query("name")
	email := c.Query("email")

	// build filters
	var filters []bson.M
	if name != "" {
		filters = append(filters, bson.M{"name": bson.M{"$regex": primitive.Regex{Pattern: name, Options: "i"}}})
	}
	if email != "" {
		filters = append(filters, bson.M{"email": bson.M{"$regex": primitive.Regex{Pattern: email, Options: "i"}}})
	}

	filter := bson.M{}
	if len(filters) > 0 {
		filter = bson.M{"$or": filters}
	}

	// fetch users
	cursor, err := userCollection.Find(c.Context(), filter)
	if err != nil {
		log.Errorf("failed to fetch user collection %v", err.Error())
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	defer cursor.Close(c.Context())

	// scan users
	users := make([]collections.User, 0)
	for cursor.Next(c.Context()) {
		var user collections.User
		if err := cursor.Decode(&user); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		user.Password = ""
		users = append(users, user)
	}

	return c.JSON(users)
}

type LoginReq struct {
	Email    string `json:"email,omitempty" validate:"required,email"`
	Password string `json:"password,omitempty" validate:"required,min=6"`
}

func (u *userController) Login(c *fiber.Ctx) error {

	// decode
	req := new(LoginReq)
	if err := c.BodyParser(req); err != nil {
		log.Errorf("failed to parse login body %v", err.Error())
		return c.Status(fiber.StatusBadRequest).SendString("invalid input")
	}

	// validate
	validationErrors := util.Validate(req)
	if len(validationErrors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errors": validationErrors,
		})
	}

	// fetch existing user
	existingUser, err := u.fetchUserByEmail(c, req.Email)
	if err != nil {
		errMessage := err.Error()
		if err == mongo.ErrNoDocuments {
			errMessage = "Invalid email or password"
		}

		return c.Status(fiber.StatusInternalServerError).SendString(errMessage)
	}

	// compare password
	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(req.Password))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid email or password")
	}

	// generate token
	accessToken, err := u.jwtHelper.GenerateToken(existingUser.ID, false)
	if err != nil {
		log.Errorf("failed to create access token, err: %v", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create access token"})
	}
	refreshToken, err := u.jwtHelper.GenerateToken(existingUser.ID, true)
	if err != nil {
		log.Errorf("failed to create refresh token, err: %v", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create refresh token"})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

type RefreshReq struct {
	RefreshToken string `json:"refresh_token"`
}

func (u *userController) RefreshToken(c *fiber.Ctx) error {
	req := new(RefreshReq)
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	claims, err := u.jwtHelper.ParseWithClaims(req.RefreshToken, true)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid refresh token"})
	}

	userID, err := primitive.ObjectIDFromHex(claims["id"].(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	// Tạo access token mới
	newAccessToken, err := u.jwtHelper.GenerateToken(userID, false)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create new access token"})
	}

	return c.JSON(fiber.Map{
		"access_token": newAccessToken,
	})
}

func (u *userController) fetchUserByEmail(c *fiber.Ctx, email string) (*collections.User, error) {
	existingUser := new(collections.User)
	userCollection := collections.GetUserCollection()
	err := userCollection.FindOne(c.Context(), bson.M{"email": email}).Decode(existingUser)
	if err != nil {
		log.Errorf("error fetchUserByEmail %v", err.Error())
		return nil, err
	}
	return existingUser, nil
}
