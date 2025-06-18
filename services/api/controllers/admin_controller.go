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
		return c.Status(fiber.StatusBadRequest).JSON(res.BadRequestResponse("Invalid request body"))
	}

	if req.Username == "" || req.Password == "" || req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(res.BadRequestResponse("Username, password, and email are required"))
	}

	pwhash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(res.InternalServerErrorResponse("Failed to hash password"))
	}
	
	err = configs.DatabaseConnection().Transaction(func(tx *gorm.DB) error {
		var existingAdmin models.Admin
		if err := tx.Where("username = ?", req.Username).First(&existingAdmin).Error; err == nil {
			return gorm.ErrDuplicatedKey
		}

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
		if err == gorm.ErrDuplicatedKey {
			return c.Status(fiber.StatusConflict).JSON(res.ErrorResponse(fiber.StatusConflict, "Username already exists"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(res.InternalServerErrorResponse("Failed to register admin"))
	}

	response := res.AdminRegisterResponse{
		Username: req.Username,
		Email:    req.Email,
	}

	return c.Status(fiber.StatusCreated).JSON(res.SuccessResponse(fiber.StatusCreated, "Admin registered successfully", response))
}

func (a *AdminController) LoginAdmin(c *fiber.Ctx) error {
	var req models.AdminLogin
	var admin models.Admin
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(res.BadRequestResponse("Invalid request body"))
	}

	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(res.BadRequestResponse("Username and password are required"))
	}

	err := configs.DatabaseConnection().Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("username = ?", req.Username).First(&admin).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(res.NotFoundResponse("Admin not found"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(res.InternalServerErrorResponse("Database error"))
	}

	err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(res.UnauthorizedResponse("Invalid username or password"))
	}

	privateKey := helpers.GetPrivateKey()
	token := jwt.New(jwt.SigningMethodES256)
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = admin.Username
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(res.InternalServerErrorResponse("Failed to generate token"))
	}

	response := res.AdminLoginResponse{
		Username: admin.Username,
		Token:    signedToken,
	}
	
	return c.Status(fiber.StatusOK).JSON(res.SuccessResponse(fiber.StatusOK, "Admin logged in successfully", response))
}