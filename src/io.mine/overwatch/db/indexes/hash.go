package indexes

type Value struct {
	Map map[uint64]bool
}

type Hash struct {
	Map map[uint64]Value
}

func (h *Hash) Get(cellId uint64) []uint64 {
	var keys []uint64
	for key, _ := range h.Map[cellId].Map {
		keys = append(keys, key)
	}
	return keys
}

// Incomplete
func (h *Hash) Scan(start uint64, end uint64) []uint64 {
	var keys []uint64
	return keys
}

func (h *Hash) Insert(cellId uint64, id uint64) {
	value, ok := h.Map[cellId]
	
	if !ok {
		value = Value{make(map[uint64]bool)}
	}
	value.Map[id] = true
	h.Map[cellId] = value
}

func (h *Hash) Delete(cellId uint64, id uint64) {
	if cellId == 0 {
		return
	}
	
	value, ok := h.Map[cellId]
	
	if ok {
		delete(value.Map, id)
		size := len(value.Map)
		
		if size < 1 {
			delete(h.Map, cellId)
		}
	}
}

func (h *Hash) Update(newCellId uint64, currentCellId uint64, id uint64) {
	h.Delete(currentCellId, id)
	h.Insert(newCellId, id)
}

func GetHashIndex() *Hash {
	return &Hash{make(map[uint64]Value)}
}
