package models

type GeoNearResponse struct {
	Id			uint64
	Point 		Location
	Distance 	float64
}
