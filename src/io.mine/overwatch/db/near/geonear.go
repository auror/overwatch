package near

import (
	"math"
	
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
	
	//"io.mine/overwatch/db/log"
	"io.mine/overwatch/db/geometry"
	"io.mine/overwatch/db/googles2"
	"io.mine/overwatch/db/models"
	"io.mine/overwatch/db/utils"
)

const (
	KRadiusOfEarthInMeters 			= (6378.1 * 1000)
	kCircOfEarthInMeters			= 2 * math.Pi * KRadiusOfEarthInMeters
	kMaxEarthDistanceInMeters		= kCircOfEarthInMeters / 2
)

const (
	InternalQueryS2GeoCoarsestLevel 	= 0
	InternalQueryS2GeoFinestLevel 		= 23
	InternalQueryS2GeoMaxCells 			= 20
)

const (
	CellScan		= "1"
	CellListScan	= "list"
	CellUnionScan	= "union"
)

type GeoNearPlan struct {
	_currLevel 			int
	_r2Annulus 			geometry.R2Annulus
	_currAnnulus 		geometry.R2Annulus
	_point				s2.Point
	_ll					s2.LatLng
	_cell				s2.Cell
	_boundsIncrement 	float64
	_scannedCellUnion 	s2.CellUnion
	_stats				int
	scanIndex 			utils.Callback
}

func buildS2Region(r2Annulus geometry.R2Annulus) s2.Region {
	var regions []s2.Region
	if r2Annulus.Inner > 0 {
		innerCap := s2.CapFromCenterAngle(r2Annulus.Point, s1.Angle(r2Annulus.Inner/KRadiusOfEarthInMeters))
		innerCap = innerCap.Complement()
		regions = append(regions, innerCap)
	}
	
	if r2Annulus.Outer < kMaxEarthDistanceInMeters {
		outerCap := s2.CapFromCenterAngle(r2Annulus.Point, s1.Angle(r2Annulus.Outer/KRadiusOfEarthInMeters))
		regions = append(regions, outerCap)
	}
	
	// if annulus is entire world, return a full cap
    if len(regions) == 0 {
    	regions = append(regions, s2.FullCap())
    }
	return googles2.InitRegionIntersection(regions)
}

func get2DSphereCovering(region s2.Region) s2.CellUnion {
	regionCoverer := s2.RegionCoverer{
		MinLevel: InternalQueryS2GeoCoarsestLevel,
		MaxLevel: InternalQueryS2GeoFinestLevel,
		MaxCells: InternalQueryS2GeoMaxCells}
	
	covering := regionCoverer.Covering(region)
	
	return covering
}

func (plan *GeoNearPlan) Initialize(geoNearRequest models.GeoNearRequest, scanIndex utils.Callback) {
	plan._currLevel = InternalQueryS2GeoFinestLevel
	
	plan._ll = s2.LatLngFromDegrees(geoNearRequest.Point.Lat, geoNearRequest.Point.Lng)
	plan._point = s2.PointFromLatLng(plan._ll)
	plan._cell = s2.CellFromPoint(plan._point)
	
	plan._r2Annulus = geometry.R2Annulus{plan._point, 0, geoNearRequest.Distance}
	plan._currAnnulus = geometry.R2Annulus{plan._point, -1, 0}
	plan._stats = -1
	plan._boundsIncrement = 0.0
	plan.scanIndex = scanIndex
}

func (plan *GeoNearPlan) Estimate() {
	cellId := plan._cell.ID()
	ret := 0
	var (
		cellUnion 	s2.CellUnion
	)
	
	for ok := true; ok; ok = ret == 0 {
		initialCellSet := append(cellId.VertexNeighbors(plan._currLevel), cellId.Parent(plan._currLevel))
		cellUnion = s2.CellUnion(initialCellSet)
		
		// Should we normalize it? After all, max 4 cells are returned
		//cellUnion.Normalize()
		ret1 := plan.scanIndex(cellUnion)
		
		googles2.Add(&plan._scannedCellUnion, cellUnion)
		ret = ret1.(int)
		
		if ret == 0 {
			plan._currLevel--
		} else {
			estimatedDistance := s2.AvgEdgeMetric.Value(plan._currLevel) * KRadiusOfEarthInMeters
			plan._boundsIncrement = 3 * estimatedDistance
			break
		}
	}
}

func (plan *GeoNearPlan) Next() bool {
	// We're done with the search
	if plan._currAnnulus.Inner >= 0 && plan._currAnnulus.Outer == plan._r2Annulus.Outer {
        return true
    }
	
	// TODO: Re-look @ what's been done here
	if plan._stats != -1 {
		if plan._stats < 300 {
			plan._boundsIncrement *= 2;
		} else if plan._stats > 600 {
			plan._boundsIncrement /= 2;
		}
	}
	
	plan._currAnnulus.Inner = plan._currAnnulus.Outer
	plan._currAnnulus.Outer = math.Min(plan._currAnnulus.Outer + plan._boundsIncrement, plan._r2Annulus.Outer)
	isLast := (plan._currAnnulus.Outer == plan._r2Annulus.Outer)
	
	region := buildS2Region(plan._currAnnulus)
	coveringCellUnion := get2DSphereCovering(region)
	
	var cover []s2.CellID
	diffUnion := googles2.Difference(coveringCellUnion, plan._scannedCellUnion)
	for _, cellId := range diffUnion {
		if region.IntersectsCell(s2.CellFromCellID(cellId)) {
			cover = append(cover, cellId)
		}
	}
	coveringCellUnion = s2.CellUnion(cover)
	
	ret := plan.scanIndex(coveringCellUnion)
	plan._stats = ret.(int)
	
	googles2.Add(&plan._scannedCellUnion, coveringCellUnion)
	return isLast
}

func (plan *GeoNearPlan) ComputeGeoNearDistance(location models.Location) float64 {
	return plan._ll.Distance(s2.LatLngFromDegrees(location.Lat, location.Lng)).Radians()*KRadiusOfEarthInMeters
}
