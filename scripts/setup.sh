#!/bin/bash

echo "🚀 TransJakarta Fleet Management - Complete Setup"
echo "=================================================="

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

print_success() {
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

# Check prerequisites
echo "📋 Checking prerequisites..."

# Check Docker
if command -v docker &> /dev/null; then
    DOCKER_VERSION=$(docker --version | cut -d' ' -f3 | cut -d',' -f1)
    print_success "Docker $DOCKER_VERSION is installed"
else
    print_error "Docker is not installed. Please install Docker first."
    echo "Visit: https://docs.docker.com/get-docker/"
    exit 1
fi

# Check Docker Compose
if command -v docker compose &> /dev/null; then
    print_success "Docker Compose is installed"
else
    print_error "Docker Compose is not installed"
    exit 1
fi

# Check if Docker is running
if docker info &> /dev/null; then
    print_success "Docker daemon is running"
else
    print_error "Docker daemon is not running. Please start Docker."
    exit 1
fi

# Check available ports
echo ""
echo "🔍 Checking port availability..."

check_port() {
    if lsof -i:$1 &> /dev/null; then
        print_warning "Port $1 is already in use"
        return 1
    else
        print_success "Port $1 is available"
        return 0
    fi
}

PORTS_OK=true
check_port 3000 || PORTS_OK=false  # API Server
check_port 5432 || PORTS_OK=false  # PostgreSQL
check_port 1883 || PORTS_OK=false  # MQTT
check_port 5672 || PORTS_OK=false  # RabbitMQ
check_port 15672 || PORTS_OK=false # RabbitMQ Management

if [ "$PORTS_OK" = false ]; then
    print_warning "Some ports are in use. The setup will continue, but you may encounter conflicts."
    echo "You can stop conflicting services or change ports in docker-compose.yml"
    echo ""
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Setup project structure
echo ""
echo "📁 Setting up project structure..."

# Create necessary directories
mkdir -p {bin,logs,backups}
mkdir -p postman
mkdir -p docs

print_success "Project directories created"

# Create environment file if it doesn't exist
if [ ! -f .env ]; then
    echo ""
    echo "⚙️ Creating environment configuration..."
    cp .env.example .env
    print_success "Environment configuration created (.env)"
else
    print_info "Environment file already exists"
fi

# Build and start the system
echo ""
echo "🔨 Building and starting the system..."

# Clean up any existing containers
docker compose down -v 2>/dev/null

# Build all images
print_info "Building Docker images (this may take a few minutes)..."
if docker compose build; then
    print_success "Docker images built successfully"
else
    print_error "Failed to build Docker images"
    exit 1
fi

# Start all services
print_info "Starting all services..."
if docker compose up -d; then
    print_success "All services started"
else
    print_error "Failed to start services"
    exit 1
fi

# Wait for services to be ready
echo ""
echo "⏳ Waiting for services to initialize..."

print_info "This may take 30-60 seconds for all services to be ready..."

# Wait for database
echo -n "Waiting for PostgreSQL"
for i in {1..30}; do
    if docker compose exec -T postgres pg_isready -U postgres &> /dev/null; then
        echo ""
        print_success "PostgreSQL is ready"
        break
    fi
    echo -n "."
    sleep 2
done

# Wait for API server
echo -n "Waiting for API server"
for i in {1..30}; do
    if curl -s http://localhost:3000/health &> /dev/null; then
        echo ""
        print_success "API server is ready"
        break
    fi
    echo -n "."
    sleep 2
done

# Verify system health
echo ""
echo "🔍 Verifying system health..."

HEALTH_RESPONSE=$(curl -s http://localhost:3000/health 2>/dev/null)
if echo "$HEALTH_RESPONSE" | grep -q "healthy"; then
    print_success "System health check passed"
    
    # Parse health status
    DB_STATUS=$(echo "$HEALTH_RESPONSE" | grep -o '"database":"[^"]*"' | cut -d'"' -f4)
    MQTT_STATUS=$(echo "$HEALTH_RESPONSE" | grep -o '"mqtt":"[^"]*"' | cut -d'"' -f4)
    RABBITMQ_STATUS=$(echo "$HEALTH_RESPONSE" | grep -o '"rabbitmq":"[^"]*"' | cut -d'"' -f4)
    
    echo "  • Database: $DB_STATUS"
    echo "  • MQTT: $MQTT_STATUS"
    echo "  • RabbitMQ: $RABBITMQ_STATUS"
else
    print_warning "System health check inconclusive"
    echo "Response: $HEALTH_RESPONSE"
fi

# Test vehicle tracking
echo ""
echo "🚌 Testing vehicle tracking..."
sleep 5

VEHICLE_RESPONSE=$(curl -s http://localhost:3000/api/v1/vehicles/B1234XYZ/location 2>/dev/null)
if echo "$VEHICLE_RESPONSE" | grep -q "latitude"; then
    print_success "Vehicle tracking is working"
    LAT=$(echo "$VEHICLE_RESPONSE" | grep -o '"latitude":[^,]*' | cut -d':' -f2)
    LNG=$(echo "$VEHICLE_RESPONSE" | grep -o '"longitude":[^,]*' | cut -d':' -f2)
    echo "  • Vehicle B1234XYZ is at: ($LAT, $LNG)"
else
    print_info "Vehicle data not available yet (vehicles are still starting up)"
fi

# Show container status
echo ""
echo "📊 Container status:"
docker compose ps

echo ""
echo "🎉 Setup Complete!"
echo "=================="
echo ""
echo "✅ TransJakarta Fleet Management System is now running!"
echo ""
echo "📱 Access Points:"
echo "  • API Documentation: http://localhost:3000"
echo "  • Health Check: http://localhost:3000/health"
echo "  • RabbitMQ Management: http://localhost:15672 (guest/guest)"
echo ""
echo "🚌 Vehicle Fleet:"
echo "  • 5 vehicles are simulating movement around Jakarta"
echo "  • Real-time location updates every 2 seconds"
echo "  • Automatic geofence detection at landmarks"
echo ""
echo "🎯 Test Commands:"
echo "  • make docker-test    # Complete integration test"
echo "  • make postman-test   # API testing with Postman"
echo "  • make monitor        # Real-time system monitoring"
echo ""
echo "📊 Useful Commands:"
echo "  • docker compose logs -f              # View all logs"
echo "  • docker compose restart <service>    # Restart specific service"
echo "  • docker compose down                 # Stop all services"
echo "  • make help                           # Show all available commands"
echo ""
echo "🏛️ Jakarta Landmarks (Geofences):"
echo "  • Monas (National Monument)"
echo "  • Bundaran HI (Hotel Indonesia Roundabout)"
echo "  • Grand Indonesia (Shopping Mall)"
echo "  • Plaza Indonesia (Shopping Mall)"
echo "  • Sarinah (Department Store)"
echo ""

if echo "$HEALTH_RESPONSE" | grep -q "healthy"; then
    print_success "SETUP SUCCESSFUL! Your TransJakarta Fleet Management System is ready! 🚀"
else
    print_warning "Setup completed but some services may need more time to initialize."
    echo "Run 'curl http://localhost:3000/health' in a few minutes to verify."
fi
