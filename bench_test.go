package sharding_test

import (
	"testing"
	"time"

	"gopkg.in/go-pg/sharding.v4"
	"gopkg.in/pg.v4"
)

func benchmarkDB() *pg.DB {
	return pg.Connect(&pg.Options{
		User:         "postgres",
		Database:     "postgres",
		DialTimeout:  30 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
	})
}

func BenchmarkGopg(b *testing.B) {
	db := benchmarkDB()
	defer db.Close()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := db.Exec("SELECT 1")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkCluster(b *testing.B) {
	db := benchmarkDB()
	defer db.Close()

	cluster := sharding.NewCluster([]*pg.DB{db}, 1)
	defer cluster.Close()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := cluster.Shard(0).Exec("SELECT 1")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}