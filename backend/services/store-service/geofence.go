package main

import (
	"math"
	"time"
)

const earthRadiusMeters = 6371000.0

// HaversineDistance returns the distance between two coordinates in meters.
func HaversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLng := (lng2 - lng1) * math.Pi / 180.0

	rLat1 := lat1 * math.Pi / 180.0
	rLat2 := lat2 * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(rLat1)*math.Cos(rLat2)*
			math.Sin(dLng/2)*math.Sin(dLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusMeters * c
}

// PointInPolygon tests whether point p is inside the polygon using Ray-Casting algorithm.
func PointInPolygon(p Point, polygon []Point) bool {
	n := len(polygon)
	if n < 3 {
		return false
	}

	inside := false
	j := n - 1
	for i := 0; i < n; i++ {
		pi := polygon[i]
		pj := polygon[j]

		// Check ray intersection
		if (pi.Lng > p.Lng) != (pj.Lng > p.Lng) &&
			(p.Lat < (pj.Lat-pi.Lat)*(p.Lng-pi.Lng)/(pj.Lng-pi.Lng)+pi.Lat) {
			inside = !inside
		}
		j = i
	}
	return inside
}

// IsWithinGeofence validates if (userLat, userLng) is inside polygon or within radius meters.
func IsWithinGeofence(userLat, userLng float64, store *Store) bool {
	p := Point{Lat: userLat, Lng: userLng}

	if len(store.GeofencePolygon) >= 3 {
		return PointInPolygon(p, store.GeofencePolygon)
	}

	// Fallback to Haversine distance vs radius
	dist := HaversineDistance(userLat, userLng, store.Lat, store.Lng)
	radius := float64(store.GeofenceRadiusMeters)
	if radius <= 0 {
		radius = 100.0 // Default 100 meters
	}
	return dist <= radius
}

// IsStoreOpenNow checks if current time in store's timezone falls within opening_time and closing_time.
func IsStoreOpenNow(store *Store, now time.Time) bool {
	if store.OpeningTime == "" || store.ClosingTime == "" {
		return true // Default to open if no hours specified
	}

	tzLocation, err := time.LoadLocation(store.Timezone)
	if err != nil {
		tzLocation = time.UTC
	}

	localNow := now.In(tzLocation)
	currentTimeStr := localNow.Format("15:04:05")

	if store.OpeningTime <= store.ClosingTime {
		return currentTimeStr >= store.OpeningTime && currentTimeStr <= store.ClosingTime
	} else {
		// Overnight store hours (e.g. 22:00:00 to 06:00:00)
		return currentTimeStr >= store.OpeningTime || currentTimeStr <= store.ClosingTime
	}
}
