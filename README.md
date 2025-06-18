
# 📦🚚 Logistics Delivery Tracker

A comprehensive logistics delivery tracking system built with **Go**, **Kafka**, **RabbitMQ**, and **PostgreSQL**.

---

## 🚀 Features

✅ **Shipment Management** – Create & track shipments in real time  
📍 **Location Tracking** – Update package location with geo-coordinates  
🔔 **Notifications** – Real-time status alerts via RabbitMQ  
🧱 **Event-driven Architecture** – Kafka-powered messaging  
🔐 **Admin Auth** – JWT-secured admin endpoints  
🌐 **REST API** – Clean & intuitive HTTP endpoints

---

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   REST API      │    │     Kafka       │    │   PostgreSQL    │
│   (Fiber)       │───►│   (Events)      │    │   (Database)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              ▲   │                 ▲         ▲
                              │   ▼                 │         │
                        ┌─────────────────┐         │         │
                        │ Kafka Consumer  │─────────┘         │
                        │ (Tracking Proc) │                   │
                        └──────┬──────────┘                   │
                               │                              │
                               ▼                              │
                        ┌─────────────────┐                   │
                        │   RabbitMQ      │                   │
                        │ (Notifications) │                   │
                        └──────┬──────────┘                   │
                               │                              │
                               ▼                              │
                     ┌────────────────────────┐               │
                     │ Notification Consumer  │───────────────┘
                     │ (Save notif to DB)     │
                     └────────────────────────┘
```

---

## 🧩 Services

| Service               | Description                                       |
|-----------------------|---------------------------------------------------|
| **API Gateway**       | Accepts HTTP requests, publishes Kafka events     |
| **Tracking Processor**| Consumes Kafka events, updates status, sends notif |
| **Location Updater**  | Publishes location updates to Kafka               |
| **Notification Service** | Consumes from RabbitMQ, stores notification logs |

---

## ⚡ Quick Start

### 🔧 Prerequisites
- 🐳 Docker & Docker Compose
- 🐹 Go 1.24+ (for development)

### 🐳 Run with Docker Compose
```bash
git clone <repository-url>
```

```bash
cd logistics-delivery-tracker
```

```bash
docker-compose up -d
```

```bash
curl http://localhost:3000/health
```


### 💻 Local Development

#### 🔧 Install dependencies
```bash
go mod download
```

#### ▶️ Start infrastructure services

```bash
docker-compose up -d postgres kafka rabbitmq zookeeper
```

#### 🚀 Run the application

✅ **Opsi 1: Manual**

```bash
go run main.go
```

✅ **Opsi 2: Live-reload (with Air)**

```bash
air
```

> Make sure you have `air` installed:
>
> ```bash
> go install github.com/cosmtrek/air@latest
> ```

---

## 🔌 API Endpoints

### 🔐 Admin

#### ➕ Register
```http
POST /admin/register
```
```json
{
  "username": "admin",
  "password": "password123",
  "email": "admin@example.com"
}
```

#### 🔑 Login
```http
POST /admin/login
```
```json
{
  "username": "admin",
  "password": "password123"
}
```

---

### 📦 Shipment

#### 📬 Create
```http
POST /shipment/create
```
```json
{
  "receiver_name": "John Doe",
  "origin_address": "Jakarta, Indonesia",
  "dest_address": "Surabaya, Indonesia",
  "priority": "normal",
  "weight": 2.5,
  "dimensions": "30x20x10 cm"
}
```

#### 🛠 Update Status
```http
PUT /shipment/updateStatus/{tracking_code}
```
```json
{
  "status": "shipped"
}
```

Valid statuses:
```
created, packaged, picked_up, shipped, transit_final, delivered, done, cancelled, failed
```

---

### 📍 Location

#### ➕ Update Location
```http
POST /location/update
```
```json
{
  "shipment_id": 1,
  "name": "Jakarta Distribution Center",
  "latitude": -6.2088,
  "longitude": 106.8456,
  "notes": "Package sorted and ready for dispatch"
}
```

---

## 🔄 Event Flow

1. 🆕 **Shipment Created** → saved → Kafka publish → tracking consumer → notif sent  
2. 🔄 **Status Update** → saved → Kafka publish → tracking consumer → notif sent  
3. 🧭 **Location Update** → Kafka publish → location processor → save to DB → notif stored

---

## ⚙️ Configuration

Environment variables are stored in `.env/`:

| File                  | Purpose                       |
|-----------------------|-------------------------------|
| `.env/api.env`        | API server configuration      |
| `.env/kafka.env`      | Kafka broker config           |
| `.env/postgres.env`   | PostgreSQL database config    |
| `.env/rabbitmq.env`   | RabbitMQ config               |
| `.env/zookeeper.env`  | Zookeeper config              |

🛠️ Copy all example files into `.env/`:
```bash
mkdir -p .env
for f in .env.example/*.env.example; do cp "$f" ".env/$(basename "${f%.env.example}.env")"; done
```

---

## 📊 Monitoring & Logs

### 🔎 Health Check
```bash
curl http://localhost:3000/health
```

### 🌐 URLs
- API: http://localhost:3000  
- RabbitMQ Dashboard: http://localhost:15672 (guest/guest)  
- PostgreSQL: localhost:5432

### 📄 Logs
```bash
# All logs
docker-compose logs -f

# Specific service
docker-compose logs -f api
```

---

## 🛠 Troubleshooting

### 🐘 PostgreSQL
```bash
docker-compose exec postgres psql -U postgres -d logistics_tracker -c "\dt"
```

### 🐇 RabbitMQ
```bash
docker-compose exec rabbitmq rabbitmqctl status
```

```bash
curl -u guest:guest http://localhost:15672/api/queues
```

### ⚠️ Kafka Issues
```bash
docker-compose restart kafka zookeeper
docker-compose logs kafka
```

---

## 📝 License

This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for details.
