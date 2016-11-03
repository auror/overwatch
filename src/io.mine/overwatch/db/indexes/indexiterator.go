package indexes

import (
	"io.mine/overwatch/db/models"
)

type IndexIterator func(record models.Record) bool
