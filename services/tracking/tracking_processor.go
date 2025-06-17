package tracking

import (
	"encoding/json"
	"log"
	"time"

	"github.com/Dziqha/logistics-delivery-tracker/configs"
	"github.com/Dziqha/logistics-delivery-tracker/helpers"
	"github.com/Dziqha/logistics-delivery-tracker/services/api/models"
	"gorm.io/gorm"
)

// ShipmentEvent represents the event data sent from API
type ShipmentEvent struct {
	TrackingCode string    `json:"tracking_code"`
	EventType    string    `json:"event_type"`    // "created" | "status_updated" | "info_updated"
	OldStatus    string    `json:"old_status,omitempty"`
	NewStatus    string    `json:"new_status"`
	ShipmentID   uint      `json:"shipment_id"`
	Timestamp    time.Time `json:"timestamp"`
	// Data tambahan jika diperlukan
	SenderName     string  `json:"sender_name,omitempty"`
	ReceiverName   string  `json:"receiver_name,omitempty"`
	OriginAddress  string  `json:"origin_address,omitempty"`
	DestAddress    string  `json:"dest_address,omitempty"`
	Priority       string  `json:"priority,omitempty"`
	Weight         float64 `json:"weight,omitempty"`
	Dimensions     string  `json:"dimensions,omitempty"`
}

func ProcessTracking() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("⚠️ Tracking processor recovered from panic: %v", r)
		}
	}()

	log.Println("🔧 Initializing Kafka shipment consumer...")
	consumer := configs.KafkaShipmentConsumer()
	defer configs.CloseKafkaConsumer(consumer)

	db := configs.DatabaseConnection()

	configs.InitRabbitMQ()
	defer configs.CloseRabbitMQ()

	log.Println("📦 Shipment event processor started...")
	log.Println("🎯 Listening for shipment events on 'shipment' topic...")

	messageCount := 0

	for {
		log.Printf("🔍 Attempting to read message... (count: %d)", messageCount)

		msg, err := consumer.ReadMessage(5 * time.Second)
		if err != nil {
			if err.Error() == "Local: Timed out" {
				log.Printf("⏱️ No messages in 5s, continuing to listen...")
				continue
			}
			log.Printf("❌ Kafka read error: %v", err)
			continue
		}

		messageCount++
		log.Printf("📨 Received message #%d from topic: %s, partition: %d, offset: %d", 
			messageCount, *msg.TopicPartition.Topic, msg.TopicPartition.Partition, msg.TopicPartition.Offset)

		// Pastikan hanya memproses topic 'shipment'
		if *msg.TopicPartition.Topic != "shipment" {
			log.Printf("⚠️ Skipping message from wrong topic: %s", *msg.TopicPartition.Topic)
			if _, err := consumer.CommitMessage(msg); err != nil {
				log.Printf("❌ Failed to commit skipped message: %v", err)
			}
			continue
		}

		log.Printf("📄 Raw message: %s", string(msg.Value))

		// PERBAIKAN 1: Parse sebagai ShipmentEvent, bukan Shipment langsung
		var event ShipmentEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			// Fallback: coba parse sebagai Shipment untuk backward compatibility
			var shipment models.Shipment
			if err2 := json.Unmarshal(msg.Value, &shipment); err2 != nil {
				log.Printf("❌ Error parsing message as event or shipment: %v | %v\nPayload: %s", err, err2, msg.Value)
				if _, err := consumer.CommitMessage(msg); err != nil {
					log.Printf("❌ Failed to commit error message: %v", err)
				}
				continue
			}
			
			// Convert shipment to event for processing
			event = ShipmentEvent{
				TrackingCode: shipment.TrackingCode,
				EventType:    "legacy", // backward compatibility
				NewStatus:    shipment.Status,
				ShipmentID:   uint(shipment.ID),
				Timestamp:    time.Now(),
			}
			log.Printf("⚠️ Processing legacy shipment format as event")
		}

		log.Printf("📦 Processing event: Type=%s, TrackingCode=%s, Status=%s->%s", 
			event.EventType, event.TrackingCode, event.OldStatus, event.NewStatus)

		var notification *models.NotificationCreate
		var shouldSendNotification bool
		var processedSuccessfully bool

		// PERBAIKAN 2: Hanya proses notifikasi, JANGAN ubah database shipment
		err = db.Transaction(func(tx *gorm.DB) error {
			var shipment models.Shipment

			// Ambil shipment yang sudah ada di database
			if err := tx.Where("tracking_code = ?", event.TrackingCode).First(&shipment).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					log.Printf("⚠️ Shipment with tracking code %s not found in database", event.TrackingCode)
					return nil // Skip processing jika shipment tidak ada
				}
				log.Printf("❌ Database error while searching shipment: %v", err)
				return err
			}

			log.Printf("📋 Found shipment in database: ID=%d, Status=%s", shipment.ID, shipment.Status)

			// PERBAIKAN 3: Cek apakah ini event yang valid untuk diproses
			switch event.EventType {
			case "created":
				// Shipment baru dibuat via API
				if shipment.Status == "created" {
					shouldSendNotification = true
					log.Printf("✅ Processing 'created' event for new shipment")
				} else {
					log.Printf("ℹ️ Shipment already exists with status %s, skipping 'created' event", shipment.Status)
				}

			case "status_updated":
				// Status diubah via API
				if event.OldStatus != "" && shipment.Status == event.NewStatus {
					// Konfirmasi bahwa status sudah diupdate di database
					shouldSendNotification = true
					log.Printf("✅ Processing 'status_updated' event: %s -> %s", event.OldStatus, event.NewStatus)
				} else {
					log.Printf("ℹ️ Status mismatch, skipping event. DB status: %s, Event: %s->%s", 
						shipment.Status, event.OldStatus, event.NewStatus)
				}

			case "info_updated":
				// Info shipment diubah tapi status sama
				log.Printf("ℹ️ Processing 'info_updated' event - no notification needed")
				processedSuccessfully = true
				return nil

			case "legacy":
				// Backward compatibility untuk format lama
				if shipment.Status == event.NewStatus {
					log.Printf("ℹ️ Legacy event - status tidak berubah (%s)", shipment.Status)
					processedSuccessfully = true
					return nil
				}
				shouldSendNotification = true
				log.Printf("✅ Processing legacy event for status: %s", event.NewStatus)

			default:
				log.Printf("⚠️ Unknown event type: %s", event.EventType)
				return nil
			}

			// PERBAIKAN 4: Buat notifikasi berdasarkan status SAAT INI di database
			if shouldSendNotification {
				log.Printf("🔔 Creating notification for status: %s", shipment.Status)

				switch shipment.Status {
				case "created":
					notification = helpers.MakeNotification(shipment, "info", "Pesanan Dibuat", "sedang dikemas")
				case "packaged":
					notification = helpers.MakeNotification(shipment, "info", "Dikemas", "sedang dalam proses pengemasan")
				case "picked_up":
					notification = helpers.MakeNotification(shipment, "info", "Kurir Menjemput", "sedang diambil oleh kurir")
				case "shipped":
					notification = helpers.MakeNotification(shipment, "info", "Dalam Perjalanan", "sedang dalam pengiriman")
				case "transit_final":
					notification = helpers.MakeNotification(shipment, "info", "Transit Akhir", "tiba di kota tujuan")
				case "delivered":
					notification = helpers.MakeNotification(shipment, "success", "Pesanan Tiba", "telah diterima oleh penerima")
				case "done":
					notification = helpers.MakeNotification(shipment, "info", "Pesanan Selesai", "selesai. Terima kasih!")
				case "cancelled":
					notification = helpers.MakeNotification(shipment, "error", "Pesanan Dibatalkan", "telah dibatalkan")
				case "failed":
					notification = helpers.MakeNotification(shipment, "error", "Pengiriman Gagal", "gagal dikirim")
				default:
					log.Printf("⚠️ Unknown status: %s, skipping notification", shipment.Status)
					notification = nil
				}

				if notification != nil {
					log.Printf("📝 Notification created: %+v", *notification)
				}
			}

			processedSuccessfully = true
			return nil
		})

		if err != nil {
			log.Printf("❌ DB transaction error: %v", err)
			continue
		}

		// Kirim notifikasi jika diperlukan
		if notification != nil && shouldSendNotification && processedSuccessfully {
			log.Printf("📤 Sending notification to RabbitMQ...")
			if err := configs.SendNotification(*notification); err != nil {
				log.Printf("❌ Gagal kirim notifikasi ke RabbitMQ: %v", err)
			} else {
				log.Printf("📨 ✅ Notifikasi berhasil dikirim: %s", notification.Title)
			}
		} else {
			log.Printf("ℹ️ No notification to send (notification: %v, shouldSend: %v, processed: %v)", 
				notification != nil, shouldSendNotification, processedSuccessfully)
		}

		// Commit message setelah berhasil diproses
		if processedSuccessfully {
			if _, err := consumer.CommitMessage(msg); err != nil {
				log.Printf("❌ Failed to commit message: %v", err)
			} else {
				log.Printf("✅ Message committed successfully")
			}
		}

		log.Println("─────────────────────────────────────")
	}
}