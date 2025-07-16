# Postman Collection - TransJakarta Fleet Management API

Comprehensive API testing collection for TransJakarta Fleet Management System technical assessment.

## Overview

This Postman collection provides complete API testing coverage for the fleet management system, including real-time vehicle tracking, geofencing events, and system monitoring capabilities.

## Files

- `transjakarta-fleet.postman_collection.json` - Main API collection
- `transjakarta-fleet.postman_environment.json` - Environment variables (optional)

## Prerequisites

- Postman Desktop App or Web Version
- TransJakarta Fleet Management System running locally
- Docker services started (`docker compose up -d`)

## Import Instructions

### Method 1: File Import
1. Open Postman
2. Click **Import** button
3. Drag and drop `transjakarta-fleet.postman_collection.json`
4. Optionally import `transjakarta-fleet.postman_environment.json`

### Method 2: Direct URL Import
1. Click **Import** → **Link**
2. Paste collection URL (if hosted)
3. Click **Continue**

## Collection Structure

### 1. System Health & Status
- **System Health Check** - Verify all service connectivity
- **System Statistics** - Total locations and geofence events
- **MQTT Status** - MQTT broker connection status
- **RabbitMQ Status** - Message queue connection status

### 2. Vehicle Location Tracking
- **Latest Location** - Current position for all 5 vehicles
- **Location History** - Historical data with time range queries
- **Dynamic Timestamps** - Auto-calculated time parameters

### 3. Geofencing Events
- **Geofence Events** - Jakarta landmark entry detection
- **Multi-Vehicle Support** - Events for all active vehicles
- **Landmark Monitoring** - Monas, Bundaran HI, Grand Indonesia, etc.

### 4. Error Handling Tests
- **Invalid Vehicle ID** - 404 error validation
- **Invalid Parameters** - 400 error validation
- **Missing Parameters** - Input validation testing

### 5. Integration Tests
- **Complete System Test** - End-to-end workflow validation
- **Performance Overview** - Response time and system metrics

## Test Execution

### Running Complete Collection
1. Click collection name → **Run collection**
2. Select all folders or specific test groups
3. Configure iterations and delay (optional)
4. Click **Run TransJakarta Fleet Management**

### Running Individual Tests
1. Select specific request or folder
2. Click **Send** or **Run**
3. View results in Test Results tab

## Environment Variables

### Built-in Variables
- `baseUrl`: `http://localhost:3000` (default)
- `start_time`: Auto-calculated timestamp
- `end_time`: Auto-calculated timestamp

### Custom Configuration
```javascript
// Change base URL if needed
pm.environment.set('baseUrl', 'http://localhost:3001');

// Custom time ranges
const customStart = Math.floor(Date.now() / 1000) - 7200; // 2 hours ago
pm.environment.set('start_time', customStart);
```

## Test Scenarios

### Basic Functionality Test
1. Run **System Health Check**
2. Verify all services show "connected"
3. Test **Vehicle Location** endpoints
4. Validate response structure and data types

### Real-time Tracking Test
1. Get location for vehicle B1234XYZ
2. Wait 30 seconds
3. Request same vehicle location again
4. Verify coordinates have changed (vehicle movement)

### Geofencing Test
1. Run **Geofence Events** for all vehicles
2. Monitor console output for landmark entries
3. Verify events show vehicles entering Jakarta landmarks

### Error Handling Test
1. Test invalid vehicle IDs
2. Test malformed parameters
3. Verify appropriate HTTP status codes

## Expected Results

### System Startup (2-3 minutes)
- Health check: All services "connected"
- Statistics: Growing location counts
- Vehicle locations: Real Jakarta coordinates
- Service status: All "true"

### After 5-10 minutes
- Vehicle history: Movement patterns visible
- Geofence events: Landmark entries detected
- All vehicles: Active location data
- Error responses: Proper HTTP status codes

## Performance Benchmarks

### Response Time Expectations
- Health Check: < 100ms
- Location API: < 50ms
- History API: < 200ms
- Statistics: < 100ms

### Data Validation
- Coordinates: Within Jakarta bounds (-6.35° to -6.05° lat, 106.65° to 107.05° lng)
- Timestamps: Valid Unix epoch format
- Vehicle IDs: Match expected format (B1234XYZ pattern)

## Console Output

The collection includes comprehensive logging:
- Real-time test results
- Vehicle coordinates and movement
- Geofence event notifications
- System performance metrics
- Error details and debugging info

## Troubleshooting

### Collection Issues
- **No response**: Verify system is running (`docker compose ps`)
- **Connection errors**: Check port 3000 is available
- **No location data**: Wait 2-3 minutes for system initialization

### Environment Issues
- **Missing variables**: Collection works without environment file
- **Wrong base URL**: Update `baseUrl` variable in collection settings
- **Timestamp errors**: Pre-request scripts handle automatic calculation

## Test Data

### Active Vehicles
- B1234XYZ (Primary route)
- B5678ABC (Secondary route)
- B9012DEF (Express route)
- B3456GHI (Feeder route)
- B7890JKL (Secondary route)

### Geofence Locations
- **Monas**: -6.1754, 106.8272
- **Bundaran HI**: -6.1944, 106.8229
- **Grand Indonesia**: -6.1944, 106.8229
- **Plaza Indonesia**: -6.1928, 106.8218
- **Sarinah**: -6.1922, 106.8219

## Validation Criteria

### Success Indicators
- All tests pass with expected status codes (200/404/400)
- Vehicle coordinates within Jakarta metropolitan area
- Location record count increases over time
- Geofence events detected for landmark entries
- Error handling responds appropriately
- System performance meets benchmarks

### Test Coverage
- **Functional Testing**: All API endpoints
- **Integration Testing**: Service interactions
- **Error Testing**: Invalid inputs and edge cases
- **Performance Testing**: Response time validation
- **Data Validation**: Response structure and content

## Additional Features

### Automated Testing
- Pre-request scripts for dynamic data
- Test scripts for response validation
- Conditional logic for different scenarios
- Error handling and retry mechanisms

### Monitoring Integration
- Real-time system metrics
- Fleet status dashboard simulation
- Performance tracking
- Business logic validation

---

## Technical Assessment Notes

This collection demonstrates:
- **Complete API Coverage**: All endpoints tested
- **Real-world Scenarios**: Practical use cases
- **Error Handling**: Comprehensive validation
- **Performance Testing**: Response time monitoring
- **Integration Testing**: Multi-service workflows

**Estimated Testing Time**: 5-10 minutes  
**Success Rate**: 100% when system is properly initialized

*Postman Collection developed for TransJakarta Fleet Management Technical Assessment*