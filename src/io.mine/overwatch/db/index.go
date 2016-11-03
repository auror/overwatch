package main

import (
	"github.com/golang/geo/s2"
	
	"io.mine/overwatch/db/indexes"
	"io.mine/overwatch/db/models"
)

type Index interface {
	Get(cellId s2.CellID, iterator indexes.IndexIterator)
	Scan(start s2.CellID, end s2.CellID, iterator indexes.IndexIterator)
	Insert(record models.Record)
	Delete(record models.Record)
	Update(oldRecord models.Record, newRecord models.Record)
}

// Deprecated for now
//func GetTimeEfficientIndex() Index {
//	return indexes.GetHashIndex()
//}

func GetSpaceTimeEfficientIndex() Index {
	return indexes.GetBTreeIndex()
}
