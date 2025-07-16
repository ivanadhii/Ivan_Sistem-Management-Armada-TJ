echo "🎬 TransJakarta Fleet Management - Geofencing Demo"
echo "=================================================="

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
PURPLE='\033[0;35m'
NC='\033[0m'

demo_step() {
    echo -e "\n${BLUE}🎯 $1${NC}"
    echo "Press Enter to continue..."
    read
}

geofence_alert() {
    echo -e "${PURPLE}🚨 GEOFENCE ALERT: $1${NC}"
}

demo_step "Step 1: Starting Complete Infrastructure"
echo "Starting PostgreSQL, MQTT Broker, and RabbitMQ..."
cd docker
docker compose up postgres mosquitto rabbitmq -d
cd ..
sleep 5

demo_step "Step 2: Starting Server with Geofencing"
go run cmd/server/main.go &
SERVER_PID=$!
sleep 3

demo_step "Step 3: Starting Geofence Event Worker"
echo -e "${GREEN}This worker will process and log all geofence events...${NC}"
go run cmd/worker/main.go &
WORKER_PID=$!
sleep 2

demo_step "Step 4: Checking System Health"
echo -e "${YELLOW}Enhanced health check with all services:${NC}"
curl -s http://localhost:3000/health | jq .

demo_step "Step 5: Viewing Jakarta Landmarks (Geofence Areas)"
echo -e "${YELLOW}Our geofence areas (50m radius each):${NC}"
docker exec -it transjakarta_postgres psql -U postgres -d transjakarta_fleet -c "SELECT name, latitude, longitude, radius FROM geofences ORDER BY name;"

demo_step "Step 6: Starting Vehicle Fleet (Real-time Simulation)"
echo -e "${GREEN}Starting 5 vehicles moving around Jakarta...${NC}"
go run cmd/publisher/main.go &
PUBLISHER_PID=$!

demo_step "Step 7: Monitoring Vehicle Positions"
echo -e "${YELLOW}Watching vehicles approach landmarks...${NC}"
for i in {1..10}; do
    echo "--- Update $i ---"
    for vehicle in "B1234XYZ" "B5678ABC" "B9012DEF"; do
        LOCATION=$(curl -s http://localhost:3000/api/v1/vehicles/$vehicle/location 2>/dev/null)
        if echo $LOCATION | grep -q "latitude"; then
            LAT=$(echo $LOCATION | jq -r '.latitude' 2>/dev/null)
            LNG=$(echo $LOCATION | jq -r '.longitude' 2>/dev/null)
            echo "🚌 $vehicle: ($LAT, $LNG)"
        fi
    done
    sleep 3
done

demo_step "Step 8: Checking for Geofence Events"
echo -e "${PURPLE}Looking for geofence entries...${NC}"

EVENT_COUNT=$(docker exec transjakarta_postgres psql -U postgres -d transjakarta_fleet -t -c "SELECT COUNT(*) FROM geofence_events;" 2>/dev/null | tr -d ' \n\r')

if [ ! -z "$EVENT_COUNT" ] && [ "$EVENT_COUNT" -gt 0 ]; then
    geofence_alert "$EVENT_COUNT geofence events detected!"
    
    echo ""
    echo "Recent geofence events:"
    docker exec -it transjakarta_postgres psql -U postgres -d transjakarta_fleet -c "
    SELECT 
        ge.vehicle_id,
        g.name as landmark_entered,
        to_timestamp(ge.timestamp) as entry_time
    FROM geofence_events ge 
    JOIN geofences g ON ge.geofence_id = g.id 
    ORDER BY ge.timestamp DESC 
    LIMIT 10;"
else
    echo "No geofence events yet - vehicles still approaching landmarks..."
fi

demo_step "Step 9: Real-time Geofence Monitoring"
echo -e "${GREEN}Monitoring geofence events in real-time (watch the worker logs!)${NC}"
echo "This will run for 60 seconds - watch for landmark entries..."

START_TIME=$(date +%s)
END_TIME=$((START_TIME + 60))

while [ $(date +%s) -lt $END_TIME ]; do
    CURRENT_EVENTS=$(docker exec transjakarta_postgres psql -U postgres -d transjakarta_fleet -t -c "SELECT COUNT(*) FROM geofence_events;" 2>/dev/null | tr -d ' \n\r')
    
    if [ ! -z "$CURRENT_EVENTS" ] && [ "$CURRENT_EVENTS" -gt "${EVENT_COUNT:-0}" ]; then
        NEW_EVENTS=$((CURRENT_EVENTS - ${EVENT_COUNT:-0}))
        geofence_alert "NEW EVENT! Total events: $CURRENT_EVENTS (+$NEW_EVENTS)"
        
        # Show the latest event
        docker exec transjakarta_postgres psql -U postgres -d transjakarta_fleet -t -c "
        SELECT 
            '🚌 ' || ge.vehicle_id || ' entered ' || g.name || ' at ' || to_timestamp(ge.timestamp)
        FROM geofence_events ge 
        JOIN geofences g ON ge.geofence_id = g.id 
        ORDER BY ge.timestamp DESC 
        LIMIT 1;" 2>/dev/null
        
        EVENT_COUNT=$CURRENT_EVENTS
    fi
    
    echo "Monitoring... (${CURRENT_EVENTS:-0} total events)"
    sleep 5
done

demo_step "Step 10: API Testing - Geofence Events"
echo -e "${YELLOW}Testing geofence events API for each vehicle:${NC}"

for vehicle in "B1234XYZ" "B5678ABC" "B9012DEF"; do
    echo ""
    echo "=== $vehicle Geofence History ==="
    curl -s "http://localhost:3000/api/v1/vehicles/$vehicle/geofence-events?limit=3" | jq .
done

demo_step "Step 11: System Statistics"
echo -e "${YELLOW}Final system statistics:${NC}"
curl -s http://localhost:3000/api/v1/stats | jq .

demo_step "Step 12: RabbitMQ Queue Status"
echo -e "${YELLOW}RabbitMQ geofence event queue:${NC}"
curl -s -u guest:guest http://localhost:15672/api/queues/%2F/geofence_alerts | jq '{messages: .messages, consumers: .consumers}' 2>/dev/null || echo "Queue status unavailable"

demo_step "Step 13: Live Dashboard Simulation"
echo -e "${GREEN}Simulating a real-time fleet dashboard...${NC}"

for i in {1..5}; do
    echo ""
    echo "=== FLEET DASHBOARD UPDATE $i ==="
    echo "$(date)"
    
    # Active vehicles
    echo ""
    echo "🚌 ACTIVE VEHICLES:"
    for vehicle in "B1234XYZ" "B5678ABC" "B9012DEF" "B3456GHI" "B7890JKL"; do
        LOCATION=$(curl -s http://localhost:3000/api/v1/vehicles/$vehicle/location 2>/dev/null)
        if echo $LOCATION | grep -q "latitude"; then
            echo "   ✅ $vehicle - Active"
        else
            echo "   ❌ $vehicle - No data"
        fi
    done
    
    # Recent landmark entries
    echo ""
    echo "🏛️ RECENT LANDMARK ENTRIES:"
    docker exec transjakarta_postgres psql -U postgres -d transjakarta_fleet -t -c "
    SELECT 
        '   🎯 ' || ge.vehicle_id || ' → ' || g.name || ' (' || 
        EXTRACT(EPOCH FROM (NOW() - to_timestamp(ge.timestamp)))::int || 's ago)'
    FROM geofence_events ge 
    JOIN geofences g ON ge.geofence_id = g.id 
    WHERE ge.timestamp > EXTRACT(EPOCH FROM NOW() - INTERVAL '5 minutes')
    ORDER BY ge.timestamp DESC 
    LIMIT 3;" 2>/dev/null || echo "   No recent entries"
    
    # System stats
    STATS=$(curl -s http://localhost:3000/api/v1/stats 2>/dev/null)
    LOCATIONS=$(echo $STATS | jq -r '.total_locations' 2>/dev/null)
    EVENTS=$(echo $STATS | jq -r '.total_geofence_events' 2>/dev/null)
    
    echo ""
    echo "📊 SYSTEM METRICS:"
    echo "   Location Updates: ${LOCATIONS:-0}"
    echo "   Geofence Events: ${EVENTS:-0}"
    
    sleep 8
done

# Cleanup
echo ""
echo -e "\n${BLUE}🧹 Demo cleanup...${NC}"
kill $PUBLISHER_PID 2>/dev/null
kill $WORKER_PID 2>/dev/null  
kill $SERVER_PID 2>/dev/null

echo ""
echo -e "${GREEN}✅ Geofencing Demo Complete!${NC}"
echo ""
echo "🎉 What you just saw:"
echo "  • Real-time vehicle tracking"
echo "  • Automatic landmark detection" 
echo "  • Event-driven architecture"
echo "  • Multi-service integration"
echo "  • Production-ready APIs"
echo ""
echo "🚀 This system can now:"
echo "  • Track TransJakarta buses in real-time"
echo "  • Detect arrivals at important locations"
echo "  • Send notifications for passenger apps"
echo "  • Power real-time dashboards"
echo "  • Scale to thousands of vehicles"
