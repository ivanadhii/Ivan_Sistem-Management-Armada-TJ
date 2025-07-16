#!/bin/bash

echo "📮 Postman Collection Test Runner"
echo "================================="

# Check if Newman is installed
if ! command -v newman &> /dev/null; then
    echo "Newman not found. Installing..."
    if command -v npm &> /dev/null; then
        npm install -g newman
    else
        echo "❌ npm not found. Please install Node.js and npm first."
        exit 1
    fi
fi

# Check if Docker services are running
if ! curl -s http://localhost:3000/health > /dev/null; then
    echo "❌ Server is not running. Please start with: docker compose up -d"
    exit 1
fi

# Create postman directory if it doesn't exist
mkdir -p postman

echo "🚀 Running Postman collection tests..."

# Run the collection
newman run postman/transjakarta-fleet.postman_collection.json \
    --reporters cli,html \
    --reporter-html-export postman/test-results.html \
    --delay-request 1000 \
    --timeout-request 10000 \
    --bail

# Check if test results were generated
if [ -f "postman/test-results.html" ]; then
    echo ""
    echo "📊 Test results saved to: postman/test-results.html"
    echo "🌐 Open in browser to view detailed results"
else
    echo "⚠️ Test results file not generated"
fi

echo ""
echo "✅ Postman tests completed"