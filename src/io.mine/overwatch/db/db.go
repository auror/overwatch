package main

import (
	"github.com/boltdb/bolt"
	"github.com/golang/geo/s2"
	
	"io.mine/overwatch/db/log"
	"io.mine/overwatch/db/models"
	"io.mine/overwatch/db/indexes"
)

type DBProxy struct {
	*bolt.DB
}

var DB DBProxy
var s2Indexes = make(map[string]Index)
var idIndexes = make(map[string]interface{})

func init() {
	var err error
	db, err := bolt.Open("overwatch.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	DB = DBProxy{db}
	
	log.Debug("Indexing.....")
	
	err = DB.View(func(tx *bolt.Tx) error {
        return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
    		bucketName := string(name)
    		s2Indexes[bucketName] = GetSpaceTimeEfficientIndex()
    		idIndexes[bucketName] = make(map[uint64]models.Location)
            
            err := DB.View(func(tx *bolt.Tx) error {
			    c := b.Cursor()
			
			    for k, v := c.First(); k != nil; k, v = c.Next() {
			        key := DecodeId(k)
			        location := new(models.Location)
			        err := Decode(v, location)
			        if err != nil {
			        	return err
			        }
			        
			        //log.Debugf("Id, CellId is %v, %v", key, location.CellId)
			        DB.InsertIntoIndex(bucketName, key, *location)
			    }
			
			    return nil
			})
            
            return err
        })
    })
	
	if err != nil {
    	log.Fatal(err)
	} else {
		log.Debugf("Indexing done: %v", s2Indexes)
	}
}

func Buckets() []string{
	buckets := make([]string, 0, len(s2Indexes))
	for k := range s2Indexes {
        buckets = append(buckets, k)
    }
	return buckets
}

func getRecord(id uint64, location models.Location) models.Record {
	return models.Record{id, &location}
}

func (db DBProxy) InsertIntoIndex(bucket string, id uint64, location models.Location) {
	s2Indexes[bucket].Insert(getRecord(id, location))
	idIndexes[bucket].(map[uint64]models.Location)[id] = location
}

func (db DBProxy) UpdateIndex(bucket string, id uint64, location models.Location) {
	oldLocation, ok := db.SearchId(bucket, id)
	if ok {
		s2Indexes[bucket].Update(getRecord(id, oldLocation), getRecord(id, location))
		idIndexes[bucket].(map[uint64]models.Location)[id] = location
	} else {
		db.InsertIntoIndex(bucket, id, location)
	}
}

func (db DBProxy) SearchId(bucket string, id uint64) (models.Location, bool) {
	loc, ok := idIndexes[bucket].(map[uint64]models.Location)[id]
	return loc, ok
}

func (db DBProxy) ScanS2Index(bucket string, start s2.CellID, end s2.CellID, iterator indexes.IndexIterator) {
	s2Indexes[bucket].Scan(start, end, iterator)
}
