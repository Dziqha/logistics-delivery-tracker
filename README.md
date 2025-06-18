# Logistics Delivery Tracker

A comprehensive logistics delivery tracking system built with Go, Kafka, RabbitMQ, and PostgreSQL.

## Features

- **Shipment Management**: Create and track shipments with real-time status updates
- **Location Tracking**: Update shipment locations with GPS coordinates
- **Real-time Notifications**: Automated notifications via RabbitMQ
- **Event-driven Architecture**: Uses Kafka for reliable message processing
- **Admin Authentication**: JWT-based authentication for admin operations
- **RESTful API**: Clean REST API endpoints for all operations

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   REST API      │    │     Kafka       │    │   PostgreSQL    │
│   (Fiber)       │◄──►│   (Events)      │    │   (Database)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │              ┌─────────────────┐              │
         └──────────────►│   RabbitMQ      │◄─────────────┘
                        │ (Notifications) │
                        └─────────────────┘
```

## Services

1. **API Service**: Handles HTTP requests and publishes events to Kafka
2. **Tracking Service**: Processes shipment events and sends notifications
3. **Location Service**: Handles location updates and status changes
4. **Notification Service**: Processes and stores notifications

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.24+ (for local development)

### Using Docker Compose

1. Clone the repository:
```bash
git clone <repository-url>
cd logistics-delivery-tracker
```

2. Start all services:
```bash
docker-compose up -d
```

3. Check service health:
```bash
curl http://localhost:3000/health
```

### Local Development

1. Install dependencies:
```bash
go mod download
```

2. Start infrastructure services:
```bash
docker-compose up -d postgres kafka rabbitmq zookeeper
```

3. Run the application:
```bash
go run main.go
```

## API Endpoints

### Admin Authentication

#### Register Admin
```bash
POST /admin/register
Content-Type: application/json

{
  "username": "admin",
  "password": "password123",
  "email": "admin@example.com"
}
```

#### Login Admin
```bash
POST /admin/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password123"
}
```

### Shipment Management

#### Create Shipment
```bash
POST /shipment/create
Content-Type: application/json

{
  "receiver_name": "John Doe",
  "origin_address": "Jakarta, Indonesia",
  "dest_address": "Surabaya, Indonesia",
  "priority": "normal",
  "weight": 2.5,
  "dimensions": "30x20x10 cm"
}
```

#### Update Shipment Status
```bash
PUT /shipment/updateStatus/{tracking_code}
Content-Type: application/json

{
  "status": "shipped"
}
```

**Valid statuses**: `created`, `packaged`, `picked_up`, `shipped`, `transit_final`, `delivered`, `done`, `cancelled`, `failed`

### Location Tracking

#### Update Location
```bash
POST /location/update
Content-Type: application/json

{
  "shipment_id": 1,
  "name": "Jakarta Distribution Center",
  "latitude": -6.2088,
  "longitude": 106.8456,
  "status": "in_transit",
  "notes": "Package sorted and ready for dispatch"
}
```

**Valid location statuses**: `picked_up`, `in_transit`, `out_for_delivery`, `delivered`

## Event Flow

1. **Shipment Creation**:
   - API creates shipment in database
   - Publishes `created` event to Kafka `shipment` topic
   - Tracking service processes event and sends notification

2. **Status Update**:
   - API updates shipment status in database
   - Publishes `status_updated` event to Kafka `shipment` topic
   - Tracking service processes event and sends notification

3. **Location Update**:
   - API publishes location update to Kafka `location-update` topic
   - Location service processes update, saves to database
   - If status changes, publishes notification to Kafka `notification` topic
   - Notification service processes and stores notification

## Configuration

Environment variables are organized in the `env/` directory:

- `env/api.env`: API server configuration
- `env/kafka.env`: Kafka broker configuration
- `env/postgres.env`: PostgreSQL database configuration
- `env/rabbitmq.env`: RabbitMQ configuration
- `env/zookeeper.env`: Zookeeper configuration

## Monitoring

### Health Check
```bash
curl http://localhost:3000/health
```

### Service URLs
- **API**: http://localhost:3000
- **RabbitMQ Management**: http://localhost:15672 (guest/guest)
- **PostgreSQL**: localhost:5432

### Logs
```bash
# View all logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f api
docker-compose logs -f kafka
docker-compose logs -f rabbitmq
```

## Troubleshooting

### Kafka Issues
If Kafka consumers are not working:
```bash
# Restart Kafka and Zookeeper
docker-compose restart kafka zookeeper

# Check Kafka logs
docker-compose logs kafka
```

### Database Issues
```bash
# Check database connection
docker-compose exec postgres psql -U postgres -d logistics_tracker -c "\dt"

# Reset database
docker-compose down -v
docker-compose up -d
```

### RabbitMQ Issues
```bash
# Check RabbitMQ status
docker-compose exec rabbitmq rabbitmqctl status

# View queues
curl -u guest:guest http://localhost:15672/api/queues
```

## Development

### Adding New Features

1. **Database Models**: Add to `services/api/models/`
2. **API Controllers**: Add to `services/api/controllers/`
3. **Routes**: Add to `services/api/routes/`
4. **Background Services**: Add to `services/`

### Testing

```bash
# Run tests
go test ./...

# Test specific package
go test ./services/api/controllers
```

## Production Deployment

1. Update environment variables for production
2. Use proper secrets management
3. Configure proper logging
4. Set up monitoring and alerting
5. Use load balancers for high availability

## License

This project is licensed under the MIT License.