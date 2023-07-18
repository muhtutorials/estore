package database

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

var Cache *redis.Client
var CacheChannel chan string

func ConnectToCache() {
	Cache = redis.NewClient(&redis.Options{
		Addr:     "cache:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}

func SetupCacheChannel() {
	CacheChannel = make(chan string)

	go func(ch chan string) {
		for {
			time.Sleep(5 * time.Second)
			key := <-ch
			Cache.Del(context.Background(), key)
			fmt.Println("Cache cleared", key)
		}
	}(CacheChannel)
}

func ClearCache(keys ...string) {
	for _, key := range keys {
		CacheChannel <- key
	}
}
