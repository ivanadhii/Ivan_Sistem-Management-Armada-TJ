package geofence

import (
	"math"
)

// HaversineDistance calculates the distance between two points on Earth
// using the Haversine formula. Returns distance in meters.
func HaversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusMeters = 6371000.0 // Earth radius in meters
	
	// Convert degrees to radians
	lat1Rad := lat1 * math.Pi / 180
	lng1Rad := lng1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lng2Rad := lng2 * math.Pi / 180
	
	// Calculate differences
	dlat := lat2Rad - lat1Rad
	dlng := lng2Rad - lng1Rad
	
	// Haversine formula
	a := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
		math.Sin(dlng/2)*math.Sin(dlng/2)
	
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	
	return earthRadiusMeters * c
}

// IsWithinRadius checks if a point is within a specified radius of a center point
func IsWithinRadius(centerLat, centerLng, pointLat, pointLng float64, radiusMeters float64) bool {
	distance := HaversineDistance(centerLat, centerLng, pointLat, pointLng)
	return distance <= radiusMeters
}

// CalculateBearing calculates the bearing (direction) from one point to another
// Returns bearing in degrees (0-360)
func CalculateBearing(lat1, lng1, lat2, lng2 float64) float64 {
	lat1Rad := lat1 * math.Pi / 180
	lng1Rad := lng1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lng2Rad := lng2 * math.Pi / 180
	
	dlng := lng2Rad - lng1Rad
	
	y := math.Sin(dlng) * math.Cos(lat2Rad)
	x := math.Cos(lat1Rad)*math.Sin(lat2Rad) - 
		math.Sin(lat1Rad)*math.Cos(lat2Rad)*math.Cos(dlng)
	
	bearing := math.Atan2(y, x) * 180 / math.Pi
	
	// Normalize to 0-360 degrees
	for bearing < 0 {
		bearing += 360
	}
	for bearing >= 360 {
		bearing -= 360
	}
	
	return bearing
}