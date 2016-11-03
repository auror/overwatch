package googles2

import (
	"github.com/golang/geo/s2"
)

func differenceInternal(cellId s2.CellID, to s2.CellUnion, cellIds *[]s2.CellID) {
	if !to.IntersectsCellID(cellId) {
		*cellIds = append(*cellIds, cellId)
	} else if !to.ContainsCellID(cellId) {
		child := cellId.ChildBegin()
		for i:=0; ;i++ {
			differenceInternal(child, to, cellIds)
			if i == 3 {
				break // Avoid unnecessary next() computation.
			}
			child = child.Next()
		}
	}
}

func Difference(from s2.CellUnion, to s2.CellUnion) s2.CellUnion {
	// TODO: this is approximately O(N*log(N)), but could probably use similar
	// techniques as GetIntersection() to be more efficient.
	var cellIds []s2.CellID
	for _, cellId := range from {
		differenceInternal(cellId, to, &cellIds);
	}
	union := s2.CellUnion(cellIds)
	union.Normalize()
	return union
}

func Add(cellUnion *s2.CellUnion, cellIds s2.CellUnion) {
	*cellUnion = append(*cellUnion, cellIds...)
	cellUnion.Normalize()
}
