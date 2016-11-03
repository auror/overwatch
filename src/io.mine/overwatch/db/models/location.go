package models

import (
	"github.com/golang/geo/s2"
)

type Location struct {
    Lat 	float64 `json:"lat"`
    Lng 	float64 `json:"lng"`
    CellId 	uint64	`json:"-"`
}

func (loc Location) Empty() bool {
	return loc.CellId == 0
}

func (loc Location) GetCellId() s2.CellID {
	return s2.CellID(loc.CellId)
}

func (loc *Location) SetCellId(cellId s2.CellID) {
	loc.CellId = uint64(cellId)
}

type Record struct {
	Id	uint64		`json:"id"`
	Loc	*Location	`json:"loc"`
}
