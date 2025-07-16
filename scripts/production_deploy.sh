#!/bin/bash

echo "🚀 Production Deployment - TransJakarta Fleet Management"
echo "========================================================"

# Production deployment script
if [ "$1" != "--confirm" ]; then
    echo "This script will deploy the system in production mode."
    echo "Make sure you have:"
    echo "  • Updated environment variables for production"
    echo "  • Configured secure passwords"
    echo "  • Set up SSL certificates (if needed)"
    echo "  • Configured monitoring and logging"
    echo ""
    echo "To proceed, run: $0 --confirm"
    exit 1
fi

echo "🔧 Starting production deployment..."

# Check if production environment file exists
if [ ! -f .env.production ]; then
    echo "❌ Production environment file (.env.production) not found!"
    echo "Please create .env.production with secure configuration."
    exit 1
fi

# Backup current deployment
if docker compose ps | grep -q "Up"; then
    echo "📦 Creating backup of current deployment..."
    docker compose exec postgres pg_dump -U postgres transjakarta_fleet > backup-$(date +%Y%m%d_%H%M%S).sql
fi

# Pull latest images
echo "📥 Pulling latest images..."
docker compose pull

# Deploy with production configuration
echo "🚀 Deploying production system..."
docker compose -f docker-compose.yml -f docker/docker-compose.prod.yml --env-file .env.production up -d --build

# Wait for services
echo "⏳ Waiting for services to start..."
sleep 30

# Verify deployment
echo "🔍 Verifying production deployment..."
HEALTH_CHECK=$(curl -s http://localhost:3000/health 2>/dev/null)

if echo "$HEALTH_CHECK" | grep -q "healthy"; then
    echo "✅ Production deployment successful!"
    echo ""
    echo "📊 System Status:"
    echo "$HEALTH_CHECK" | jq . 2>/dev/null || echo "$HEALTH_CHECK"
else
    echo "❌ Production deployment verification failed!"
    echo "Health check response: $HEALTH_CHECK"
    exit 1
fi

echo ""
echo "✅ Production deployment complete!"
echo ""
echo "🔍 Production checklist:"
echo "  • Update passwords in environment variables"
echo "  • Configure SSL/TLS certificates"
echo "  • Set up log aggregation"
echo "  • Configure monitoring alerts"
echo "  • Schedule database backups"
echo "  • Test failover procedures"
echo ""
echo "📊 Monitoring:"
echo "  • Health: http://localhost:3000/health"
echo "  • RabbitMQ: http://localhost:15672"
echo "  • Logs: docker compose logs -f"