# TransJakarta Fleet Management System

Backend system untuk manajemen armada TransJakarta dengan fitur real-time tracking, geofencing, dan event processing.

## Tech Stack

- **Backend**: Go 1.23+ dengan Fiber framework
- **Database**: PostgreSQL 15
- **Message Broker**: RabbitMQ 3.12
- **MQTT Broker**: Eclipse Mosquitto 2.0
- **Containerization**: Docker & Docker Compose
- **Caching**: Redis 7 (optional)

## Prerequisites

Sebelum menjalankan `docker compose up --build`,pastikan sudah terinstall:

- **Docker** (version 20.0+)
- **Docker Compose** (version 2.0+)
- **Git**
- **Make** (untuk menjalankan Makefile commands)

Verifikasi installation:
```bash
docker --version
docker compose version
git --version
make --version
```

## Quick Start

```bash
# Clone repository
git clone https://github.com/ivanadhii/Ivan_Sistem-Management-Armada-TJ.git
cd Ivan_Sistem-Management-Armada-TJ

# Setup lengkap (recommended)
make setup

# Atau manual
docker compose up -d --build
```

## Make Commands

| Command              |                  Fungsi                     |
|----------------------|---------------------------------------------|
| `make setup`         | Setup lengkap sistem (build, start, verify) |
| `make docker-up`     | Start semua services dengan Docker          |
| `make docker-down`   | Stop semua services                         |
| `make docker-build`  | Build Docker images                         |
| `make docker-test`   | Run integration tests                       |
| `make run`           | Run server secara lokal                     |
| `make run-publisher` | Run publisher secara lokal                  |
| `make run-worker`    | Run worker secara lokal                     |
| `make build`         | Build semua aplikasi                        |
| `make test-all`      | Run semua test suites                       |
| `make monitor`       | Monitor system statistics                   |
| `make monitor-logs`  | Monitor logs real-time                      |
| `make clean`         | Cleanup containers dan volumes              |
| `make help`          | Show semua available commands               |

## API Endpoints

### Vehicle Tracking
|                   Endpoint                      | Method |                           Fungsi                              |
|-------------------------------------------------|--------|---------------------------------------------------------------|
| `/api/v1/vehicles/{vehicle_id}/location`        |   GE   | Dapatkan lokasi terkini kendaraan                             |
| `/api/v1/vehicles/{vehicle_id}/history`         |   GET  | Dapatkan history lokasi (dengan query params `start` & `end`) |
| `/api/v1/vehicles/{vehicle_id}/geofence-events` |   GET  | Dapatkan geofence events (dengan query param `limit`)         |

### System Status
|         Endpoint          | Method |                   Fungsi                   |
|---------------------------|--------|--------------------------------------------|
| `/health`                 | GET    | Health check semua services                |
| `/api/v1/stats`           | GET    | Statistik sistem (total locations, events) |
| `/api/v1/mqtt/status`     | GET    | Status koneksi MQTT                        |
| `/api/v1/rabbitmq/status` | GET    | Status koneksi RabbitMQ                    |

### Main
| Endpoint | Method |            Fungsi            |
|----------|--------|------------------------------|
|     `/`  | GET    | API info dan welcome message |

## Ports & Services

| Service                 | Port |               Fungsi                 |
|-------------------------|------|--------------------------------------|
| **API Server**          | 3000 | REST API endpoints                   |
| **PostgreSQL**          | 5432 | Database server                      |
| **MQTT Broker**         | 1883 | MQTT message broker                  |
| **MQTT WebSocket**      | 9001 | MQTT via WebSocket                   |
| **RabbitMQ**            | 5672 | Message queue                        |
| **RabbitMQ Management** | 15672| RabbitMQ web interface (guest/guest) |
| **Redis**               | 6379 | Cache server (optional)              |

## Verification

### 1. Check System Health
```bash
curl http://localhost:3000/health
```

### 2. Test Vehicle Location
```bash
curl http://localhost:3000/api/v1/vehicles/B1234XYZ/location
```

### 3. Check Statistics
```bash
curl http://localhost:3000/api/v1/stats
```

### 4. Access RabbitMQ Management
Browser: http://localhost:15672 (guest/guest)

## Vehicle Fleet

System mensimulasikan 5 kendaraan:
- **B1234XYZ**, **B5678ABC**, **B9012DEF**, **B3456GHI**, **B7890JKL**

## Geofencing Locations

Sistem mendeteksi kendaraan yang masuk radius 50m dari:
- **Monas** (National Monument)
- **Bundaran HI** (Hotel Indonesia Roundabout)
- **Grand Indonesia** (Shopping Mall)
- **Plaza Indonesia** (Shopping Mall)
- **Sarinah** (Department Store)

## Testing

```bash
# Complete integration test
make docker-test

# Specific tests
./scripts/test_mqtt.sh
./scripts/test_phase3.sh

# Postman collection
Import: postman/transjakarta-fleet.postman_collection.json
```

## Troubleshooting

### Port Conflicts
Edit `docker-compose.yml` untuk mengubah port:
```yaml
ports:
  - "3001:3000"  # API port
  - "5433:5432"  # PostgreSQL port
```

### Reset System
```bash
make clean
make setup
```

### View Logs
```bash
docker compose logs -f
docker compose logs -f server
make monitor-logs
```

## Architecture

```
Vehicle Publisher → MQTT Broker → API Server → PostgreSQL
                                     ↓
                    RabbitMQ ← Geofence Detection
                       ↓
                   Worker Service
```

---

*TransJakarta Fleet Management System - Technical Assessment*