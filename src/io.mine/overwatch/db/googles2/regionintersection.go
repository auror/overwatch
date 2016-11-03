package googles2

import (
	"github.com/golang/geo/s2"
)

type RegionIntersection struct {
	regions []s2.Region
}

func InitRegionIntersection(regions []s2.Region) RegionIntersection {
	regionIntersection := RegionIntersection{regions}
	return regionIntersection
}

func (r RegionIntersection) CapBound() s2.Cap {
	return r.RectBound().CapBound()
}

func (r RegionIntersection) RectBound() s2.Rect {
	result := s2.FullRect()
	for _, region := range r.regions {
		result = result.Intersection(region.RectBound())
	}
	return result
}

func (r RegionIntersection) ContainsCell(c s2.Cell) bool {
	for _, region := range r.regions {
		if !region.ContainsCell(c) {
			return false
		}
	}
	
	return true
}

func (r RegionIntersection) IntersectsCell(c s2.Cell) bool {
	for _, region := range r.regions {
		if !region.IntersectsCell(c) {
			return false
		}
	}
	
	return true
}
