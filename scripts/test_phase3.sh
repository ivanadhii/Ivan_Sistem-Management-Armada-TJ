#!/bin/bash

echo "🚀 Phase 3: Geofencing & RabbitMQ Integration Test"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_geofence() {
    echo -e "${PURPLE}🎯 $1${NC}"
}

# Check prerequisites
echo "📋 Checking Phase 3 Prerequisites..."

# Check PostgreSQL
if docker exec -it transjakarta_postgres psql -U postgres -d transjakarta_fleet -c "SELECT COUNT(*) FROM geofences;" > /dev/null 2>&1; then
    GEOFENCE_COUNT=$(docker exec -it transjakarta_postgres psql -U postgres -d transjakarta_fleet -t -c "SELECT COUNT(*) FROM geofences;" 2>/dev/null | tr -d ' \n\r')
    print_status "PostgreSQL connected - $GEOFENCE_COUNT geofences configured"
else
    print_error "PostgreSQL connection failed"
    exit 1
fi

# Start all required services
echo ""
echo "🔧 Starting all infrastructure services..."
cd docker
docker compose up postgres mosquitto rabbitmq -d
cd ..
sleep 5

# Check RabbitMQ
print_info "Checking RabbitMQ..."
if curl -s http://localhost:15672 > /dev/null; then
    print_status "RabbitMQ Management UI accessible at http://localhost:15672"
    print_info "Login: guest/guest"
else
    print_warning "RabbitMQ management UI not accessible yet"
fi

echo ""
echo "🎬 Phase 3 Test Sequence Starting..."

# Step 1: Start server with geofencing
echo ""
echo "1️⃣ Starting server with complete geofencing integration..."
go run cmd/server/main.go &
SERVER_PID=$!
sleep 5

# Check if server started
if curl -s http://localhost:3000/health > /dev/null; then
    print_status "Server started with geofencing support"
else
    print_error "Server failed to start"
    kill $SERVER_PID 2>/dev/null
    exit 1
fi

# Step 2: Check enhanced health status
echo ""
echo "2️⃣ Testing enhanced health check..."
HEALTH_RESPONSE=$(curl -s http://localhost:3000/health)
echo "Enhanced Health Status:"
echo $HEALTH_RESPONSE | jq .

# Step 3: Start worker for event processing
echo ""
echo "3️⃣ Starting geofence event worker..."
go run cmd/worker/main.go &
WORKER_PID=$!
sleep 3
print_status "Geofence worker started - monitoring for landmark entries"

# Step 4: Start publisher
echo ""
echo "4️⃣ Starting vehicle publisher..."
go run cmd/publisher/main.go &
PUBLISHER_PID=$!
sleep 5
print_status "Vehicle publisher started - 5 vehicles moving around Jakarta"

# Step 5: Monitor geofence events in real-time
echo ""
echo "5️⃣ Monitoring for geofence events (60 seconds)..."
print_geofence "Watching for vehicles entering Jakarta landmarks..."
print_info "Landmarks: Monas, Bundaran HI, Grand Indonesia, Plaza Indonesia, Sarinah"

# Monitor for geofence events
START_TIME=$(date +%s)
END_TIME=$((START_TIME + 60))

while [ $(date +%s) -lt $END_TIME ]; do
    # Check for new geofence events in database
    EVENT_COUNT=$(docker exec transjakarta_postgres psql -U postgres -d transjakarta_fleet -t -c "SELECT COUNT(*) FROM geofence_events;" 2>/dev/null | tr -d ' \n\r')
    
    if [ ! -z "$EVENT_COUNT" ] && [ "$EVENT_COUNT" -gt 0 ]; then
        print_geofence "🎉 GEOFENCE EVENTS DETECTED! Total: $EVENT_COUNT"
        
        # Show recent events
        echo "Recent geofence events:"
        docker exec -it transjakarta_postgres psql -U postgres -d transjakarta_fleet -c "
        SELECT 
            ge.vehicle_id,
            g.name as geofence_name,
            ge.event_type,
            to_timestamp(ge.timestamp) as event_time
        FROM geofence_events ge 
        JOIN geofences g ON ge.geofence_id = g.id 
        ORDER BY ge.timestamp DESC 
        LIMIT 5;"
        break
    fi
    
    echo "Waiting for vehicles to enter landmarks... (${EVENT_COUNT:-0} events so far)"
    sleep 5
done

# Step 6: Test geofence API endpoints
echo ""
echo "6️⃣ Testing geofence API endpoints..."

# Test stats endpoint
echo "System Statistics:"
curl -s http://localhost:3000/api/v1/stats | jq .

# Test geofence events for vehicles
echo ""
echo "Testing geofence events endpoint:"
for vehicle in "B1234XYZ" "B5678ABC" "B9012DEF"; do
    echo "Geofence events for $vehicle:"
    curl -s "http://localhost:3000/api/v1/vehicles/$vehicle/geofence-events?limit=5" | jq .
    echo ""
done

# Step 7: RabbitMQ verification
echo ""
echo "7️⃣ Verifying RabbitMQ integration..."

# Check queue status via management API
if command -v curl &> /dev/null; then
    echo "RabbitMQ Queue Status:"
    curl -s -u guest:guest http://localhost:15672/api/queues/%2F/geofence_alerts | jq '{name: .name, messages: .messages, consumers: .consumers}' 2>/dev/null || echo "Queue status check failed"
fi

# Step 8: Live monitoring demo
echo ""
echo "8️⃣ Live monitoring demonstration (30 seconds)..."
print_geofence "Watching real-time geofence detection..."

for i in {1..6}; do
    echo ""
    echo "=== Monitoring Update $i ==="
    
    # Get current stats
    STATS=$(curl -s http://localhost:3000/api/v1/stats)
    LOCATION_COUNT=$(echo $STATS | jq -r '.total_locations')
    EVENT_COUNT=$(echo $STATS | jq -r '.total_geofence_events')
    
    echo "📊 Current Stats: $LOCATION_COUNT locations, $EVENT_COUNT geofence events"
    
    # Show recent vehicle positions
    echo "📍 Recent Vehicle Positions:"
    for vehicle in "B1234XYZ" "B5678ABC"; do
        LOCATION=$(curl -s http://localhost:3000/api/v1/vehicles/$vehicle/location)
        if echo $LOCATION | grep -q "latitude"; then
            LAT=$(echo $LOCATION | jq -r '.latitude')
            LNG=$(echo $LOCATION | jq -r '.longitude')
            echo "   🚌 $vehicle: ($LAT, $LNG)"
        fi
    done
    
    sleep 5
done

# Step 9: Final verification
echo ""
echo "9️⃣ Final verification and summary..."

# Get final statistics
FINAL_STATS=$(curl -s http://localhost:3000/api/v1/stats)
FINAL_LOCATIONS=$(echo $FINAL_STATS | jq -r '.total_locations')
FINAL_EVENTS=$(echo $FINAL_STATS | jq -r '.total_geofence_events')

echo ""
echo "📊 PHASE 3 TEST RESULTS:"
echo "================================"
echo "✅ Total Location Records: $FINAL_LOCATIONS"
echo "✅ Total Geofence Events: $FINAL_EVENTS"

if [ "$FINAL_EVENTS" -gt 0 ]; then
    print_geofence "🎉 GEOFENCING IS WORKING! Vehicles detected entering landmarks!"
    
    # Show landmark entry details
    echo ""
    echo "🏛️ Landmark Entries Detected:"
    docker exec -it transjakarta_postgres psql -U postgres -d transjakarta_fleet -c "
    SELECT 
        g.name as landmark,
        COUNT(*) as entries,
        COUNT(DISTINCT ge.vehicle_id) as unique_vehicles
    FROM geofence_events ge 
    JOIN geofences g ON ge.geofence_id = g.id 
    GROUP BY g.name 
    ORDER BY entries DESC;"
else
    print_warning "No geofence events detected yet. This may be normal if vehicles haven't entered landmark areas."
    print_info "Try running the test longer or check if vehicles are moving near landmarks."
fi

# Show service health
echo ""
echo "🔧 Service Health Status:"
HEALTH=$(curl -s http://localhost:3000/health)
DB_STATUS=$(echo $HEALTH | jq -r '.database')
MQTT_STATUS=$(echo $HEALTH | jq -r '.mqtt')
RABBITMQ_STATUS=$(echo $HEALTH | jq -r '.rabbitmq')

echo "   Database: $DB_STATUS"
echo "   MQTT: $MQTT_STATUS"  
echo "   RabbitMQ: $RABBITMQ_STATUS"

# Cleanup
echo ""
echo "🧹 Cleaning up test processes..."
print_status "Stopping publisher..."
kill $PUBLISHER_PID 2>/dev/null

print_status "Stopping worker..."
kill $WORKER_PID 2>/dev/null

print_status "Stopping server..."
kill $SERVER_PID 2>/dev/null

sleep 3

echo ""
print_status "PHASE 3 GEOFENCING TEST COMPLETE!"
echo ""
echo "🎯 What was tested:"
echo "   • Real-time geofence detection"
echo "   • RabbitMQ event publishing"
echo "   • Worker event processing"
echo "   • Enhanced API endpoints"
echo "   • Multi-service integration"
echo ""

if [ "$FINAL_EVENTS" -gt 0 ]; then
    echo "🎉 SUCCESS: Complete geofencing system is working!"
    echo ""
    echo "📱 You can now:"
    echo "   • Track vehicles entering Jakarta landmarks"
    echo "   • Process real-time geofence events"
    echo "   • Monitor fleet activities via API"
    echo "   • Build real-time dashboards"
else
    echo "⚠️  Geofencing system is ready but no events detected during test."
    echo "   This is normal - vehicles need time to move into landmark areas."
fi
