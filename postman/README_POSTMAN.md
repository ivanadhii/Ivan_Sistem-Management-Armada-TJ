# 📮 TransJakarta Fleet Management - Postman Collection

## 🎯 Overview

Postman Collection lengkap untuk testing **TransJakarta Fleet Management System** - Technical Assessment.

### 📦 File Collection:
- `transjakarta-fleet.postman_collection.json` - Main collection
- `transjakarta-fleet.postman_environment.json` - Environment variables

## 🚀 Quick Setup

### 1. Import Collection
```bash
# Option A: Import files
1. Open Postman
2. Click "Import"
3. Drag & drop both JSON files
4. Select "TransJakarta Fleet Environment"

# Option B: Import from URL (if hosted)
1. File → Import → Link
2. Paste collection URL
```

### 2. Start System
```bash
# Make sure system is running
make setup
# or
docker compose up -d

# Wait 1-2 minutes for initialization
```

### 3. Run Tests
```bash
# Option A: Run entire collection
1. Click "TransJakarta Fleet Management API"
2. Click "Run collection" 
3. Select all folders
4. Click "Run TransJakarta Fleet Management API"

# Option B: Run individual folders
- 🏥 System Health & Status
- 🚌 Vehicle Location Tracking  
- 🎯 Geofencing & Events
- ❌ Error Handling Tests
- 🎭 Demo & Integration Tests
```

## 📁 Collection Structure

### 🏥 **System Health & Status**
- **System Health Check** - Verify all services (Database, MQTT, RabbitMQ)
- **System Statistics** - Get total locations and geofence events
- **MQTT Status** - Check MQTT broker connection
- **RabbitMQ Status** - Check RabbitMQ connection

### 🚌 **Vehicle Location Tracking**
- **Get Latest Location** - Get current position for vehicles B1234XYZ, B5678ABC, B9012DEF
- **Get Vehicle History** - Location history for last 1 hour and 5 minutes
- **Dynamic Time Calculation** - Auto-calculates timestamps

### 🎯 **Geofencing & Events**
- **Get Geofence Events** - Check when vehicles enter Jakarta landmarks
- **Multiple Vehicles** - Test all active vehicles
- **Landmark Detection** - Monas, Bundaran HI, Grand Indonesia, Plaza Indonesia, Sarinah

### ❌ **Error Handling Tests**
- **Invalid Vehicle ID** - Test 404 responses
- **Invalid Time Range** - Test 400 responses  
- **Missing Parameters** - Test validation

### 🎭 **Demo & Integration Tests**
- **Complete System Test** - Comprehensive test of all components
- **Fleet Overview Test** - Status of all 5 vehicles

## 🧪 Test Features

### ✅ **Automated Tests**
- Status code validation
- Response structure validation
- Data type validation
- Business logic validation

### 📊 **Console Logging**
- Real-time test results
- Vehicle coordinates
- Geofence events
- System statistics
- Error messages

### ⏰ **Dynamic Variables**
- Auto-calculated timestamps
- Flexible time ranges
- Environment-specific URLs

## 🎯 Expected Results

### **After System Startup (1-2 minutes):**
✅ Health check shows all services "connected"  
✅ Statistics show growing location counts  
✅ Vehicle locations return real coordinates  
✅ MQTT and RabbitMQ status "true"  

### **After 5-10 minutes:**
✅ Vehicle history shows movement patterns  
✅ Geofence events detected (vehicles entering landmarks)  
✅ All 5 vehicles showing active locations  
✅ Error handling working correctly  

## 🔧 Configuration

### **Environment Variables**
```json
{
  "baseUrl": "http://localhost:3000",
  "start_time": "auto-calculated",
  "end_time": "auto-calculated"
}
```

### **Custom Configuration**
```javascript
// Change base URL if different port
pm.environment.set('baseUrl', 'http://localhost:3001');

// Test different time ranges
const customStart = Math.floor(Date.now() / 1000) - 7200; // 2 hours ago
pm.environment.set('start_time', customStart);
```

## 📱 Key Test Scenarios

### **Scenario 1: Basic Functionality**
1. Run "System Health Check"
2. Run "System Statistics" 
3. Run "Get Latest Location" for any vehicle
4. ✅ Verify: 200 responses, valid data structure

### **Scenario 2: Real-time Tracking**
1. Run "Get Latest Location" for B1234XYZ
2. Wait 30 seconds
3. Run again
4. ✅ Verify: Coordinates changed (vehicle moved)

### **Scenario 3: Geofencing**
1. Run "Get Geofence Events" for all vehicles
2. Check console for landmark entries
3. ✅ Verify: Events show vehicles entering Monas, Bundaran HI, etc.

### **Scenario 4: Complete Integration**
1. Run "Complete System Test"
2. Check console output
3. ✅ Verify: All components working together

## 🚨 Troubleshooting

### **No Location Data**
```bash
# System may still be starting
curl http://localhost:3000/health

# Check if publisher is running
docker compose ps

# Wait 2-3 minutes and try again
```

### **No Geofence Events**
```bash
# This is normal - vehicles need time to move to landmarks
# Check worker logs for real-time detection
docker compose logs -f worker
```

### **Connection Errors**
```bash
# Verify system is running
docker compose ps

# Check specific service
docker compose logs server
```

## 📊 Performance Benchmarks

### **Expected Response Times:**
- Health Check: < 100ms
- Location API: < 50ms  
- History API: < 200ms
- Statistics: < 100ms

### **Data Growth Rates:**
- Locations: ~5 records/second (5 vehicles × 1 update/2s)
- Geofence Events: ~1-3 events/minute (when vehicles enter landmarks)

## 🎉 Success Indicators

✅ **All tests pass with 200/404/400 status codes**  
✅ **Vehicle coordinates are within Jakarta bounds (-6.35 to -6.05 lat, 106.65 to 107.05 lng)**  
✅ **Location data grows over time**  
✅ **Geofence events detected for landmark entries**  
✅ **Error handling works correctly**  
✅ **All 5 vehicles show active status**  

---

## 💡 **Tips for Technical Assessment**

1. **Start with "Complete System Test"** - gives overall status
2. **Check console output** - shows detailed results and explanations
3. **Run tests multiple times** - verify real-time functionality
4. **Monitor geofence events** - demonstrates advanced features
5. **Test error scenarios** - shows robust error handling

**Total test time: ~5-10 minutes**  
**Expected success rate: 100% when system is properly initialized**