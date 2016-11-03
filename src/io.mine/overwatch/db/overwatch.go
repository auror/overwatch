package main

import (
	"errors"
	"strconv"
	
	"github.com/boltdb/bolt"
	"github.com/kataras/iris"
	"github.com/golang/geo/s2"
	
	"io.mine/overwatch/db/models"
	"io.mine/overwatch/db/log"
	"io.mine/overwatch/db/utils"
)

func createBucket(bucket string, callback func(error, *bolt.Bucket)) {
	DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			log.Debugf("create bucket: %s", err)
			callback(err, nil)
			return err
		}
		
		callback(nil, b)
		return nil
	})
}

func main() {
	defer DB.Close()
	defer log.Close()

	iris.Get("/", func(ctx *iris.Context) {
			log.Debug("In INDEX")
		ctx.Text(iris.StatusOK, "Overwatch is running")
	})
	
	iris.Get("/buckets", func(ctx *iris.Context) {
		ctx.JSON(iris.StatusOK, Buckets())
	})

	iris.Post("/buckets", func(ctx *iris.Context) {
		bucket := new(models.BucketRequest)
		if err := ctx.ReadJSON(bucket); err != nil {
			ctx.Text(iris.StatusInternalServerError, err.Error())
			return
		}

		createBucket(bucket.Name, func(err error, bucket *bolt.Bucket) {
			ctx.JSON(iris.StatusCreated, bucket)
		})
	})
	
	iris.Get("/buckets/:bucket", func(ctx *iris.Context) {
		bucket := ctx.Param("bucket")
		var keys []uint64
		
		err := DB.View(func(tx *bolt.Tx) error {
			DBBucket := tx.Bucket([]byte(bucket))
			
			if DBBucket == nil {
				log.Debug("Bucket Not found: " + bucket)
				ctx.Text(iris.StatusInternalServerError, "Bucket Not found")
				return nil
			}
		    c := DBBucket.Cursor()
		
		    for k, _ := c.First(); k != nil; k, _ = c.Next() {
		        key := DecodeId(k)
		        keys = append(keys, key)
		    }
		
		    return nil
		})
		
		if err != nil {
			log.Debug(err)
			ctx.Text(iris.StatusInternalServerError, err.Error())
		} else {
			ctx.JSON(iris.StatusOK, keys)
		}
	})

	iris.Post("/buckets/:bucket", func(ctx *iris.Context) {
		bucket := ctx.Param("bucket")
		location := new(models.Location)

		if err := ctx.ReadJSON(location); err != nil {
			log.Debug(err)
			ctx.Text(iris.StatusInternalServerError, err.Error())
			return
		}

		location.SetCellId(s2.CellIDFromLatLng(s2.LatLngFromDegrees(location.Lat, location.Lng)))		

		err := DB.Update(func(tx *bolt.Tx) error {
			DBBucket := tx.Bucket([]byte(bucket))
			
			if DBBucket == nil {
				log.Debug("Bucket Not found: " + bucket)
				ctx.Text(iris.StatusInternalServerError, "Bucket Not found")
				return nil
			}

			id, _ := DBBucket.NextSequence()

			bytes := Encode(location)
			if bytes == nil {
				log.Debugf("Cannot be encoded: %v", location)
				return errors.New("Encoding Problem")
			}

			err := DBBucket.Put(EncodeId(id), bytes)
			if err == nil {
				log.Info(DB)
				DB.InsertIntoIndex(bucket, id, *location)
				ctx.Text(iris.StatusCreated, strconv.FormatUint(id, 10))
			}
			return err
		})
		
		if err != nil {
			log.Debug(err)
			ctx.Text(iris.StatusInternalServerError, err.Error())
		}
	})
	
	iris.Get("/buckets/:bucket/:id", func(ctx *iris.Context) {
		bucket := ctx.Param("bucket")
		idString := ctx.Param("id")
		
		id, errp := ParseUint64(idString)
		if errp != nil {
			log.Debug(errp)
			ctx.Text(iris.StatusInternalServerError, errp.Error())
			return
		}

		log.Info(DB)
		err := DB.View(func(tx *bolt.Tx) error {
			DBBucket := tx.Bucket([]byte(bucket))
			
			if DBBucket == nil {
				log.Debug("Bucket Not found: " + bucket)
				ctx.Text(iris.StatusNoContent, "")
				return nil
			}
			
			record := DBBucket.Get(EncodeId(id))
			if len(record) <= 0 {
				log.Debug("No record found with id: " + idString)
				ctx.Text(iris.StatusNoContent, "")
				return nil
			}
			
			location := new(models.Location)
	        err := Decode(record, location)
	        if err != nil {
	        	return err
	        }
	        
			log.Debugf("The record is: %v\n", location)
			ctx.JSON(iris.StatusOK, location)
			return nil
		})

		if err != nil {
			log.Error(err)
			ctx.Text(iris.StatusInternalServerError, err.Error())
		}
	})

	iris.Put("/buckets/:bucket/:id", func(ctx *iris.Context) {
		bucket := ctx.Param("bucket")
		idString := ctx.Param("id")
		
		id, errp := ParseUint64(idString)
		if errp != nil {
			log.Debug(errp)
			ctx.Text(iris.StatusInternalServerError, errp.Error())
			return
		}
		
		location := new(models.Location)

		if err := ctx.ReadJSON(location); err != nil {
			log.Debug(err)
			ctx.Text(iris.StatusInternalServerError, err.Error())
			return
		}
		
		cellId := s2.CellIDFromLatLng(s2.LatLngFromDegrees(location.Lat, location.Lng))
		location.SetCellId(cellId)

		err := DB.Update(func(tx *bolt.Tx) error {
			DBBucket := tx.Bucket([]byte(bucket))
			
			if DBBucket == nil {
				log.Debug("Bucket Not found: " + bucket)
				ctx.Text(iris.StatusNoContent, "")
				return nil
			}
			
			encodedId := EncodeId(id)
			bytes := Encode(location)
			if bytes == nil {
				log.Debugf("Cannot be encoded: %v", location)
				return errors.New("Encoding Problem")
			}

			err := DBBucket.Put(encodedId, bytes)
			if err != nil {
				return err
			}
			
			DB.UpdateIndex(bucket, id, *location)
			ctx.Text(iris.StatusNoContent, "")
			return nil
		})
		
		if err != nil {
			log.Error(err)
			ctx.Text(iris.StatusInternalServerError, err.Error())
		}
	})
	
	iris.Put("/buckets/:bucket", func(ctx *iris.Context) {
		start := utils.Now()
		bucket := ctx.Param("bucket")
		records := new([]models.Record)

		if err := ctx.ReadJSON(records); err != nil {
			log.Debug(err)
			ctx.Text(iris.StatusInternalServerError, err.Error())
			return
		}
		
		err := DB.Batch(func(tx *bolt.Tx) error {
			DBBucket := tx.Bucket([]byte(bucket))
			
			if DBBucket == nil {
				log.Debug("Bucket Not found: " + bucket)
				ctx.Text(iris.StatusNoContent, "")
				return nil
			}
			
			for _, record := range *records {
				encodedId := EncodeId(record.Id)
				location := *record.Loc
				bytes := Encode(location)
				if bytes == nil {
					log.Debugf("Cannot be encoded: %v", location)
					return errors.New("Encoding Problem")
				}
	
				err := DBBucket.Put(encodedId, bytes)
				if err != nil {
					return err
				}
				
				DB.UpdateIndex(bucket, record.Id, location)
			}
			ctx.Text(iris.StatusNoContent, "")
			return nil
		})
		end := utils.Now()
		log.Debugf("Elapsed Time is: %v", end-start)
		if err != nil {
			log.Error(err)
			ctx.Text(iris.StatusInternalServerError, err.Error())
		}
	})
	
	iris.Post("/buckets/:bucket/geoNear", func(ctx *iris.Context) {
		start := utils.Now()
		bucket := ctx.Param("bucket")
		geoNearRequest := new(models.GeoNearRequest)
		geoNearRequest.Bucket = bucket

		if err := ctx.ReadJSON(geoNearRequest); err != nil {
			log.Debug(err)
			ctx.Text(iris.StatusInternalServerError, err.Error())
			return
		}
		
		guys, err := ExecuteGeoNear(*geoNearRequest)
		end := utils.Now()
		log.Debugf("Elapsed Time is: %v", end-start)
		
		if err != nil {
			log.Error(err)
			ctx.Text(iris.StatusInternalServerError, err.Error())
		} else {
			ctx.JSON(iris.StatusOK, guys)
		}
	})

	iris.Listen(":8080")
}
