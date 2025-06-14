package controllers

import (
	"time"

	"github.com/Dziqha/logistics-delivery-tracker/configs"
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
	if err := c.BodyParser(&req); err != nil {
		res.BadRequestResponse("Invalid request body")
	}

	compirepw := bcrypt.CompareHashAndPassword([]byte(req.Password), []byte(req.Password))
	err := configs.DatabaseConnection().Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("username = ? AND password = ?", req.Username, compirepw).First(&models.Admin{}).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return err
			}
		}
		return nil
	})

	if err != nil {
		res.UnauthorizedResponse("Invalid username or password")
	}

	token := jwt.New(jwt.SigningMethodES256)
	claims:= token.Claims.(jwt.MapClaims)
	claims["username"] = req.Username
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	tokenString, err := token.SignedString([]byte("secret"))
	if err != nil {
		res.InternalServerErrorResponse("Failed to generate token")
	}
	response := res.AdminLoginResponse{
		Username: req.Username,
		Token:    tokenString,
	}
	return c.JSON(res.SuccessResponse(fiber.StatusOK, "Admin logged in successfully", response))
}