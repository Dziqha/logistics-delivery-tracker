package controllers

import (
	"time"

	"github.com/Dziqha/logistics-delivery-tracker/configs"
	"github.com/Dziqha/logistics-delivery-tracker/helpers"
	"github.com/Dziqha/logistics-delivery-tracker/services/api/models"
	"github.com/Dziqha/logistics-delivery-tracker/services/api/res"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdminController struct{}

func NewAdminController() *AdminController {
	return &AdminController{}
}

func (a *AdminController) RegisterAdmin(c *fiber.Ctx) error {
	var req models.AdminRegister
	if err := c.BodyParser(&req); err != nil {
		res.BadRequestResponse("Invalid request body")
	}

	pwhash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	
	err := configs.DatabaseConnection().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&models.Admin{
			Username: req.Username,
			Password: string(pwhash),
			Email:    req.Email,
		}).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		res.InternalServerErrorResponse("Failed to register admin")
	}
	response := res.AdminRegisterResponse{
		Username: req.Username,
		Email:    req.Email,
	}

	return c.JSON(res.SuccessResponse(fiber.StatusCreated, "Admin registered successfully", response))
}



func (a *AdminController) LoginAdmin(c *fiber.Ctx) error {
	var req models.AdminLogin
	var admin models.Admin
	
	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.JSON(res.BadRequestResponse("Invalid request body"))
	}

	// Find admin in database
	err := configs.DatabaseConnection().Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("username = ?", req.Username).First(&admin).Error; err != nil {
			return err // Return the error to be handled outside
		}
		return nil
	})

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(res.NotFoundResponse("Admin not found"))
		}
		return c.JSON(res.InternalServerErrorResponse("Database error"))
	}

	// Compare password (note: parameters were swapped)
	err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password))
	if err != nil {
		return c.JSON(res.UnauthorizedResponse("Invalid username or password"))
	}

	// Generate JWT token
	privateKey := helpers.GetPrivateKey()
	token := jwt.New(jwt.SigningMethodES256)
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = admin.Username
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return c.JSON(res.InternalServerErrorResponse("Failed to generate token"))
	}

	// Prepare response
	response := res.AdminLoginResponse{
		Username: admin.Username, // Use admin.Username instead of req.Username
		Token:    signedToken,
	}
	
	return c.JSON(res.SuccessResponse(fiber.StatusOK, "Admin logged in successfully", response))
}
