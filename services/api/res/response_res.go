package res

import (
	"time"

	"github.com/Dziqha/logistics-delivery-tracker/services/api/models"
	"github.com/gofiber/fiber/v2"
)

type ResponseCode struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type AdminRegisterResponse struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type AdminLoginResponse struct {
	Username string `json:"username"`
	Token string `json:"token"`
}

type LocationResponse struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Shipment_id int64   `json:"shipment_id"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at,omitempty"`
}
type ShipmentResponse struct {
	SenderName    string `json:"sender_name" `
	ReceiverName  string `json:"receiver_name"`
	OriginAddress string `json:"origin_address"`
	DestAddress   string `json:"dest_address" `
	Status        string `json:"status"`
	Locations     []LocationResponse `json:"locations,omitempty"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
}

type NotificationResponse struct {
	ID          int64     `json:"id"`
	ShipmentID  int64     `json:"shipment_id"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Message     string    `json:"message"`
	IsRead      bool      `json:"is_read"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	TrackingCodeShipment string    `json:"tracking_code"`
}

func SuccessResponse(code int, message string, data any) *ResponseCode {
	return &ResponseCode{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

func ErrorResponse(code int, message string) *ResponseCode {
	return &ResponseCode{
		Code:    code,
		Message: message,
		Data:    nil,
	}
}

func NotFoundResponse(message string) *ResponseCode {
	return &ResponseCode{
		Code:    fiber.StatusNotFound,
		Message: message,
		Data:    nil,
	}
}

func BadRequestResponse(message string) *ResponseCode {
	return &ResponseCode{
		Code:    fiber.StatusBadRequest,
		Message: message,
		Data:    nil,
	}
}

func InternalServerErrorResponse(message string) *ResponseCode {
	return &ResponseCode{
		Code:    fiber.StatusInternalServerError,
		Message: message,
		Data:    nil,
	}
}

func UnauthorizedResponse(message string) *ResponseCode {
	return &ResponseCode{
		Code:    fiber.StatusUnauthorized,
		Message: message,
		Data:    nil,
	}
}

func ForbiddenResponse(message string) *ResponseCode {
	return &ResponseCode{
		Code:    fiber.StatusForbidden,
		Message: message,
		Data:    nil,
	}
}

func CreatedResponse(message string, data any) *ResponseCode {
	return &ResponseCode{
		Code:    fiber.StatusCreated,
		Message: message,
		Data:    data,
	}
}

func NoContentResponse(message string) *ResponseCode {
	return &ResponseCode{
		Code:    fiber.StatusNoContent,
		Message: message,
		Data:    nil,
	}
}	


func UpdatedResponse(message string, data any) *ResponseCode {
	return &ResponseCode{
		Code:    fiber.StatusOK,
		Message: message,
		Data:    data,
	}
}

func DeletedResponse(message string) *ResponseCode {
	return &ResponseCode{
		Code:    fiber.StatusOK,
		Message: message,
		Data:    nil,
	}
}


func AdminRegister(username, email string) *AdminRegisterResponse {
	return &AdminRegisterResponse{
		Username: username,
		Email:    email,
	}
}

func AdminLogin(username, token string) *AdminLoginResponse {
	return &AdminLoginResponse{
		Username: username,
		Token:    token,
	}
}

func LocationResponseData(id int64, name string, latitude, longitude float64, shipment_id int64, createdAt string, updatedAt string) *LocationResponse {
	return &LocationResponse{
		ID:        id,
		Name:      name,
		Latitude:  latitude,
		Longitude: longitude,
		Shipment_id: shipment_id,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

func ShipmentResponseData(shipment models.Shipment) *ShipmentResponse {
	return &ShipmentResponse{
		SenderName:    shipment.SenderName,
		ReceiverName:  shipment.ReceiverName,
		OriginAddress: shipment.OriginAddress,
		DestAddress:   shipment.DestAddress,
		Status:        shipment.Status,
		Locations:     []LocationResponse{
		},
		UpdatedAt:     shipment.UpdatedAt,
	}
}
