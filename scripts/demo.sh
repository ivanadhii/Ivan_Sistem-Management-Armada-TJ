#!/bin/bash

echo "🎬 TransJakarta Fleet Management - Live Demo"
echo "=============================================="

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

demo_step() {
    echo -e "\n${BLUE}🎯 $1${NC}"
    echo "Press Enter to continue..."
    read
}

demo_step "Step 1: Starting Infrastructure (PostgreSQL + MQTT)"
cd docker
docker compose up postgres mosquitto -d
cd ..

demo_step "Step 2: Starting Backend Server with MQTT Integration"
go run cmd/server/main.go &
SERVER_PID=$!
sleep 3

demo_step "Step 3: Testing Health Check"
echo -e "${YELLOW}Health Check Response:${NC}"
curl -s http://localhost:3000/health | jq .

demo_step "Step 4: Testing Empty Vehicle Data (Should be 404)"
echo -e "${YELLOW}Latest Location (should be 404):${NC}"
curl -s http://localhost:3000/api/v1/vehicles/B1234XYZ/location

demo_step "Step 5: Starting Vehicle Publisher (Real-time Location Streaming)"
echo -e "${GREEN}Starting 5 vehicles publishing location every 2 seconds...${NC}"
go run cmd/publisher/main.go &
PUBLISHER_PID=$!

demo_step "Step 6: Watching Real-time Data (30 seconds)"
echo -e "${GREEN}Watching vehicle B1234XYZ location updates...${NC}"
for i in {1..15}; do
    echo "Update $i:"
    curl -s http://localhost:3000/api/v1/vehicles/B1234XYZ/location | jq .
    sleep 2
done

demo_step "Step 7: Checking Database Growth"
echo -e "${YELLOW}Database Records Count:${NC}"
docker exec -it transjakarta_postgres psql -U postgres -d transjakarta_fleet -c "SELECT COUNT(*) as total_records FROM vehicle_locations;"

demo_step "Step 8: Vehicle History Data"
START_TIME=$(($(date +%s) - 300))
END_TIME=$(date +%s)
echo -e "${YELLOW}Last 5 minutes of B1234XYZ history:${NC}"
curl -s "http://localhost:3000/api/v1/vehicles/B1234XYZ/history?start=$START_TIME&end=$END_TIME" | jq .

demo_step "Step 9: All Active Vehicles"
echo -e "${YELLOW}Testing all vehicles:${NC}"
for vehicle in "B1234XYZ" "B5678ABC" "B9012DEF" "B3456GHI" "B7890JKL"; do
    echo "Vehicle $vehicle:"
    curl -s http://localhost:3000/api/v1/vehicles/$vehicle/location | jq .
    echo ""
done

demo_step "Step 10: Real-time Database Monitoring"
echo -e "${GREEN}Live database monitoring (Ctrl+C to stop):${NC}"
echo "Watching database record count in real-time..."
while true; do
    COUNT=$(docker exec transjakarta_postgres psql -U postgres -d transjakarta_fleet -t -c "SELECT COUNT(*) FROM vehicle_locations;" 2>/dev/null | tr -d ' \n\r')
    echo "$(date): $COUNT total location records"
    sleep 3
done &
MONITOR_PID=$!

echo "Press Enter to stop demo..."
read

# Cleanup
echo -e "\n${BLUE}🧹 Cleaning up demo...${NC}"
kill $PUBLISHER_PID 2>/dev/null
kill $SERVER_PID 2>/dev/null
kill $MONITOR_PID 2>/dev/null

echo -e "${GREEN}✅ Demo completed successfully!${NC}"
echo ""
echo "📊 What we demonstrated:"
echo "  • Real-time MQTT location streaming"
echo "  • Database integration and storage"
echo "  • REST API serving live data"
echo "  • Multiple vehicle simulation"
echo "  • Historical data queries"
echo ""
echo "🎉 Phase 2 MQTT Integration is fully functional!"