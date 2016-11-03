package main

import (
	"github.com/golang/geo/s2"
	"github.com/emirpasic/gods/trees/binaryheap"
	"github.com/emirpasic/gods/utils"
	
	"io.mine/overwatch/db/near"
	"io.mine/overwatch/db/models"
	"io.mine/overwatch/db/log"
)

func ExecuteGeoNear(geoNearRequest models.GeoNearRequest) ([]models.GeoNearResponse, error){
	nearHeap := binaryheap.NewWith(func(a, b interface{}) int {
		g1 := a.(models.GeoNearResponse)
		g2 := b.(models.GeoNearResponse)
        return utils.Float64Comparator(g1.Distance, g2.Distance)
    })
	
	plan := near.GeoNearPlan{}
	scanIndex := func(args ...interface{}) interface{} {
		cellUnion := args[0].(s2.CellUnion)
		
		var records []models.Record
		for _, cellId := range cellUnion {
			DB.ScanS2Index(geoNearRequest.Bucket, cellId.RangeMin(), cellId.RangeMax(), func(record models.Record) bool {
				records = append(records, record)
				return true
			})
		}
		
		length := len(records)
		
		if length > 0 {
			for _, record := range records {
				distance := plan.ComputeGeoNearDistance(*record.Loc)
		        if distance <= geoNearRequest.Distance {
		        	nearHeap.Push(models.GeoNearResponse{record.Id, *record.Loc, distance})
		        }
			}
		}
		
		return length
	}
	
	plan.Initialize(geoNearRequest, scanIndex)
	
	plan.Estimate()
	
	for !plan.Next() {}
	
	var guys []models.GeoNearResponse
	itr := nearHeap.Iterator()
	for itr.Next() {
		guys = append(guys, itr.Value().(models.GeoNearResponse))
	}
	log.Debugf("Content is : %v", guys)
	return guys, nil
}
