package models

import (
	"log"

	"gorm.io/gorm"
)

type Admin struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type AdminLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type AdminRegister struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func (a *Admin) TableName() string {
	return "admins"
}

func AdminMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(&Admin{})
	if err != nil {
		log.Printf("Error migrating Admin model: %v", err)
	}
	return nil
}