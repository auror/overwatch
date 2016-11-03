package indexes

import (
	"github.com/google/btree"
	"github.com/golang/geo/s2"
	rbt "github.com/emirpasic/gods/trees/redblacktree"
	"github.com/emirpasic/gods/utils"
	
	"io.mine/overwatch/db/models"
)

// Int implements the Item interface for uint64 integers.
type UInt64 uint64

// Less returns true if uint64(a) < uint64(b).
func (a UInt64) Less(b btree.Item) bool {
	return a < b.(UInt64)
}

func getBTreeKey(cellId s2.CellID) (key UInt64) {
	key = UInt64(uint64(cellId))
	return
}

func getDataKey(cellId s2.CellID) (key uint64) {
	key = uint64(cellId)
	return
}

type BTree struct {
	root 	*btree.BTree
	data	map[uint64]*rbt.Tree
}

func (bTree *BTree) Get(cellId s2.CellID, iterator IndexIterator) {
	tree, ok := bTree.data[getDataKey(cellId)]
	if !ok {
		return
	} else {
		var ret bool
		it := tree.Iterator()
		for it.Next() {
			ret = iterator(it.Value().(models.Record))
			if !ret {
				return
			}
		}
	}
}

func (bTree *BTree) Scan(start s2.CellID, end s2.CellID, iterator IndexIterator) {
	bTree.root.AscendRange(getBTreeKey(start), getBTreeKey(end+1), func(a btree.Item) bool {
		key := uint64(a.(UInt64))
		tree, ok := bTree.data[key]
		if ok {
			var ret bool
			it := tree.Iterator()
			for it.Next() {
				ret = iterator(it.Key().(models.Record))
				if !ret {
					return false
				}
			}
		}
		return true
	})
}

func (bTree *BTree) Insert(record models.Record) {
	cellId := record.Loc.GetCellId()
	bTree.root.ReplaceOrInsert(getBTreeKey(cellId))
	dataKey := getDataKey(cellId)
	value, ok := bTree.data[dataKey]
	
	if !ok {
		value = rbt.NewWith(func(a, b interface{}) int {
			g1 := a.(models.Record)
			g2 := b.(models.Record)
	        return utils.UInt64Comparator(g1.Id, g2.Id)
	    })
		bTree.data[dataKey] = value
	}
	
	value.Put(record, true)
}

func (bTree *BTree) Delete(record models.Record) {
	cellId := record.Loc.GetCellId()
	
	dataKey  := getDataKey(cellId)
	value, ok := bTree.data[dataKey]
	
	if ok {
		value.Remove(record)
		size := value.Size()
		
		if size < 1 {
			bTree.root.Delete(getBTreeKey(cellId))
			delete(bTree.data, dataKey)
		}
	}
}

func (bTree *BTree) Update(oldRecord models.Record, newRecord models.Record) {
	bTree.Delete(oldRecord)
	bTree.Insert(newRecord)
}

func GetBTreeIndex() *BTree {
	return &BTree{btree.New(10), make(map[uint64]*rbt.Tree)}
}
