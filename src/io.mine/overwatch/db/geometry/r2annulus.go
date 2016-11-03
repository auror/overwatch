package geometry

import (
	"github.com/golang/geo/s2"
)

type R2Annulus struct {
	Point 	s2.Point
	Inner	float64
	Outer	float64
}
