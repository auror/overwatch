package models

type GeoNearRequest struct {
	Point 		Location	`json:"point"`
	Distance 	float64		`json:"distance"`
	Bucket		string
}
