package main

import (
	"testing"
)

func TestGeofence_RadiusFallback(t *testing.T) {
	store := &Store{
		ID:                   "store-1",
		Lat:                  12.9716, // Bangalore center
		Lng:                  77.5946,
		GeofenceRadiusMeters: 100, // 100 meters
	}

	// 1. Coordinates ~30 meters away (inside 100m)
	userInsideLat := 12.9718
	userInsideLng := 77.5948
	if !IsWithinGeofence(userInsideLat, userInsideLng, store) {
		t.Errorf("Expected user at (%f, %f) to be inside geofence radius", userInsideLat, userInsideLng)
	}

	// 2. Coordinates ~500 meters away (outside 100m)
	userOutsideLat := 12.9760
	userOutsideLng := 77.5990
	if IsWithinGeofence(userOutsideLat, userOutsideLng, store) {
		t.Errorf("Expected user at (%f, %f) to be outside geofence radius", userOutsideLat, userOutsideLng)
	}
}

func TestGeofence_PointInPolygon(t *testing.T) {
	// Polygon around Bangalore store (4 corners)
	polygon := []Point{
		{Lat: 12.9700, Lng: 77.5900},
		{Lat: 12.9700, Lng: 77.6000},
		{Lat: 12.9800, Lng: 77.6000},
		{Lat: 12.9800, Lng: 77.5900},
	}

	store := &Store{
		ID:              "store-polygon",
		Lat:             12.9750,
		Lng:             77.5950,
		GeofencePolygon: polygon,
	}

	// 1. Point strictly inside polygon
	insidePointLat := 12.9750
	insidePointLng := 77.5950
	if !IsWithinGeofence(insidePointLat, insidePointLng, store) {
		t.Errorf("Expected point (%f, %f) to be inside polygon", insidePointLat, insidePointLng)
	}

	// 2. Point strictly outside polygon
	outsidePointLat := 12.9900
	outsidePointLng := 77.6100
	if IsWithinGeofence(outsidePointLat, outsidePointLng, store) {
		t.Errorf("Expected point (%f, %f) to be outside polygon", outsidePointLat, outsidePointLng)
	}
}
