# 🚀 TransJakarta Fleet Management System
**Quick Start Guide untuk Technical Assessment**

## 📋 Prerequisites

Pastikan sistem Anda sudah terinstall:
- **Docker** (version 20.0+)
- **Docker Compose** (version 2.0+)
- **Git**
- **cURL** (untuk testing API)

### Verifikasi Prerequisites
```bash
docker --version
docker compose version
git --version
curl --version
```

## 🎯 One-Command Setup (Recommended)

### Option 1: Complete Auto Setup
```bash
# Clone repository
git clone <repository-url>
cd transjakarta-fleet

# One command setup - handles everything automatically
make setup
```

**⏱️ Waktu setup: ~3-5 menit**

Setup otomatis akan:
- ✅ Check prerequisites
- ✅ Build semua Docker images
- ✅ Start semua services
- ✅ Run database migrations
- ✅ Verify health checks
- ✅ Start vehicle simulation

## 🐳 Manual Docker Setup

### Option 2: Step-by-Step Manual
```bash
# 1. Clone repository
git clone <repository-url>
cd transjakarta-fleet

# 2. Start semua services
docker compose up -d --build

# 3. Wait for services to be ready (30-60 seconds)
sleep 60

# 4. Verify all services are running
docker compose ps
```

## ✅ Verification Steps

### 1. Check System Health
```bash
curl http://localhost:3000/health
```
**Expected Response:**
```json
{
  "status": "healthy",
  "database": "connected",
  "mqtt": "connected", 
  "rabbitmq": "connected",
  "timestamp": "2024-01-XX..."
}
```

### 2. Test Vehicle Location API
```bash
# Get latest location for vehicle B1234XYZ
curl http://localhost:3000/api/v1/vehicles/B1234XYZ/location
```
**Expected Response:**
```json
{
  "vehicle_id": "B1234XYZ",
  "latitude": -6.2088,
  "longitude": 106.8456,
  "timestamp": 1715003456
}
```

### 3. Test History API
```bash
# Get vehicle history (last 1 hour)
START_TIME=$(($(date +%s) - 3600))
END_TIME=$(date +%s)

curl "http://localhost:3000/api/v1/vehicles/B1234XYZ/history?start=$START_TIME&end=$END_TIME"
```

### 4. Check System Statistics
```bash
curl http://localhost:3000/api/v1/stats
```
**Expected Response:**
```json
{
  "total_locations": 150,
  "total_geofence_events": 5,
  "timestamp": "2024-01-XX..."
}
```

## 🎯 Testing Geofencing Features

### Check Geofence Events
```bash
# Check if vehicles entered Jakarta landmarks
curl "http://localhost:3000/api/v1/vehicles/B1234XYZ/geofence-events?limit=5"
```

### Monitor Real-time Geofence Detection
```bash
# Open new terminal and watch worker logs
docker compose logs -f worker
```

## 📊 Available Test Scripts

### Automated Integration Tests
```bash
# Complete Docker integration test
make docker-test

# MQTT integration test  
chmod +x scripts/test_mqtt.sh
./scripts/test_mqtt.sh

# Geofencing test
chmod +x scripts/test_phase3.sh
./scripts/test_phase3.sh
```

### Live Monitoring
```bash
# Real-time system monitoring
make monitor

# Live log monitoring
make monitor-logs

# Health monitoring
make monitor-health
```

## 🌐 Access Points

| Service | URL | Credentials |
|---------|-----|-------------|
| **Main API** | http://localhost:3000 | - |
| **Health Check** | http://localhost:3000/health | - |
| **RabbitMQ Management** | http://localhost:15672 | guest/guest |
| **PostgreSQL** | localhost:5432 | postgres/postgres |

## 📱 Key API Endpoints

### Vehicle Tracking
```bash
# Latest location
GET /api/v1/vehicles/{vehicle_id}/location

# Location history
GET /api/v1/vehicles/{vehicle_id}/history?start={timestamp}&end={timestamp}

# Geofence events
GET /api/v1/vehicles/{vehicle_id}/geofence-events?limit={number}
```

### System Status
```bash
# Overall health
GET /health

# System statistics
GET /api/v1/stats

# MQTT status
GET /api/v1/mqtt/status

# RabbitMQ status
GET /api/v1/rabbitmq/status
```

## 🚌 Active Vehicles

Sistem mensimulasikan 5 kendaraan TransJakarta:
- **B1234XYZ** - Vehicle 1
- **B5678ABC** - Vehicle 2  
- **B9012DEF** - Vehicle 3
- **B3456GHI** - Vehicle 4
- **B7890JKL** - Vehicle 5

Setiap kendaraan:
- ✅ Bergerak realistic di area Jakarta
- ✅ Update lokasi setiap 2 detik via MQTT
- ✅ Otomatis terdeteksi saat masuk landmark (Monas, Bundaran HI, dll)

## 🏛️ Jakarta Landmarks (Geofences)

System mendeteksi kendaraan yang masuk radius 50m dari:
- **Monas** (National Monument)
- **Bundaran HI** (Hotel Indonesia Roundabout) 
- **Grand Indonesia** (Shopping Mall)
- **Plaza Indonesia** (Shopping Mall)
- **Sarinah** (Department Store)

## 🔧 Troubleshooting

### Port Conflicts
Jika ada konflik port, edit `docker-compose.yml`:
```yaml
ports:
  - "3001:3000"  # Change API port
  - "5433:5432"  # Change PostgreSQL port
```

### Services Not Starting
```bash
# Check container status
docker compose ps

# View specific service logs
docker compose logs <service-name>

# Restart specific service
docker compose restart <service-name>
```

### Reset Everything
```bash
# Complete cleanup and restart
make clean
make setup
```

## 📈 Expected Behavior

### After 1-2 minutes:
- ✅ All containers running
- ✅ Database populated with vehicle locations
- ✅ API endpoints responding
- ✅ MQTT messages flowing

### After 5-10 minutes:
- ✅ Geofence events detected
- ✅ RabbitMQ processing events
- ✅ Worker logging landmark entries
- ✅ Complete system operational

## 🎯 Success Indicators

✅ **System is working correctly if:**
1. Health check shows all services "connected"
2. Vehicle location APIs return real coordinates
3. Database contains growing location records
4. Geofence events are detected and logged
5. RabbitMQ queue shows message flow

## 📞 Quick Commands Reference

```bash
# Start system
make setup

# Run tests  
make docker-test

# Monitor system
make monitor

# View logs
make monitor-logs

# Stop system
docker compose down

# Complete cleanup
make clean
```

---

## 🎉 Technical Assessment Notes

**Sistem ini mendemonstrasikan:**

✅ **MQTT Integration** - Real-time vehicle tracking  
✅ **PostgreSQL Storage** - Optimized with indexing  
✅ **REST APIs** - Complete vehicle management  
✅ **RabbitMQ Geofencing** - Event-driven architecture  
✅ **Docker Deployment** - Production-ready containers  
✅ **Jakarta Landmarks** - Real geofencing implementation  

**Plus bonus features:**
- 🚀 One-command deployment
- 📊 Real-time monitoring  
- 🎯 Comprehensive testing
- 🏗️ Production-ready architecture

---

*Estimated total review time: 15-20 minutes*