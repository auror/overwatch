package googles2

import (
	"math"
	"github.com/golang/geo/r1"
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

const M_PI_2 = math.Pi / 2

type LatLngRect struct {
	lat r1.Interval
	lng s1.Interval
}

func empty() LatLngRect {
	return LatLngRect{}
}

func FullLat() r1.Interval {
	return r1.Interval{-M_PI_2, M_PI_2}
}

func FullLng() s1.Interval {
	return s1.FullInterval()
}

func FullLatLngRect() LatLngRect {
	return LatLngRect{FullLat(), FullLng()}
}

func (self *LatLngRect) isEmpty() bool {
	return self.lat.IsEmpty()
}

func (self *LatLngRect) Intersection(other LatLngRect) LatLngRect {
	latitude 	:= self.lat.Intersection(other.lat)
	longitude	:= self.lng.Intersection(other.lng)
	
	if latitude.IsEmpty() || longitude.IsEmpty() {
		return empty()
	}
	
	return LatLngRect{latitude, longitude}
}


// TODO: Implement this later
func (self *LatLngRect) CapBound() s2.Cap {
	return s2.Cap{}
}
