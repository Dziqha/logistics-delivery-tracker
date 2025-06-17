package controllers


type NotificationController struct {}

func NewNotificationController() *NotificationController {
	return &NotificationController{}
}

func (n *NotificationController) GetNotifications() {}