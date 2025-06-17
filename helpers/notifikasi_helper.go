package helpers

import (
	"fmt"

	"github.com/Dziqha/logistics-delivery-tracker/services/api/models"
)

func MakeNotification(shipment models.Shipment, notifType, title, messagePart string) *models.NotificationCreate {
	return &models.NotificationCreate{
		ShipmentID: shipment.ID,
		Type:       notifType,
		Title:      title,
		Message:    fmt.Sprintf("Pesanan %s %s", shipment.TrackingCode, messagePart),
	}
}
