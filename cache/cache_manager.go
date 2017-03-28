package cache_manager

import (
	"log"
	"os"
	"time"

	"github.com/boltdb/bolt"
)

const CacheBucket = "MartianBucket"

type cacheDatabase struct {
	DBMap map[string]string
}

var the_cache_database *cacheDatabase

func GetTheCacheDatabase() *cacheDatabase {
	if the_cache_database == nil {
		the_cache_database = &cacheDatabase{DBMap: make(map[string]string)}
	}
	return the_cache_database
}

func (cd *cacheDatabase) SerializeToDisk(loc string) {
	log.Printf("Serializing %v cache database entries to disk %v", len(cd.DBMap), loc)
	db, err := bolt.Open(loc, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(CacheBucket))
		if err != nil {
			return err
		}
		for k, v := range cd.DBMap {
			b.Put([]byte(k), []byte(v))
		}
		return nil
	})
}

func (cd *cacheDatabase) LoadFromDisk(loc string) {
	cd.DBMap = make(map[string]string)
	if _, err := os.Stat(loc); os.IsNotExist(err) {
		return
	}

	db, err := bolt.Open(loc, 0600, &bolt.Options{ReadOnly: true})
	if err != nil {
		log.Printf("Error opening DB %v. Will act like it is empty.", err)
		return
	}
	defer db.Close()

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(CacheBucket))
		c := b.Cursor()

		lc := 0

		for k, v := c.First(); k != nil; k, v = c.Next() {
			cd.DBMap[string(k)] = string(v)
			lc += 1
		}

		log.Printf("Loaded %v items from the cache file.", lc)

		return nil
	})
}
