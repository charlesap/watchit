package main

import (
    "fmt"
//    "log"
    "time"
//    "strconv"
    "strings"
//    "net"
//    "os"
//    "os/exec"

//    "github.com/mileusna/crontab"
    "github.com/go-redis/redis"
)

func KeeperClient() * redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
        return client
}

func RedisStringArray(client * redis.Client, key string) []string {

	return strings.Split(RedisString(client,key),"\n")
}

func RedisString(client * redis.Client, key string) string {

        rv, err := client.Get( key ).Result()
        if err == redis.Nil {
                panic("redis key does not exist")
        } else if err != nil {
                panic(err)
        }

        return rv
}

func main() {

    keeper := KeeperClient()

//    watchlist := RedisStringArray(keeper,"watchlist")

    mid:="0"
    for {
      res, err := keeper.XRead(&redis.XReadArgs{
         Streams: []string{"injest",mid},
         Count: 65535,
         //Block: 0,
      }).Result()
      if err != nil {
          fmt.Println(err)
      }else{
         ic:=len(res[0].Messages)
        if ic > 0 {
          fmt.Println("got ",ic," items.")
          mid=res[0].Messages[ic-1].ID
          for _,e := range res[0].Messages{
            fmt.Println(e.ID)
            fmt.Println(e.Values)
          }
        }
      }
      time.Sleep(1 * time.Second)
    }

}
