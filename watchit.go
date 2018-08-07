package main

import (
    "fmt"
    "log"
    "time"

    "github.com/mileusna/crontab"
    "github.com/go-redis/redis"
)

func KeeperClient() * redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:36379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
        return client
}

func GetSettings(client * redis.Client) {
	err := client.Set("key", "value", 0).Err()
	if err != nil {
		panic(err)
	}

	val, err := client.Get("hostname").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("hostname", val)

	val2, err := client.Get("watchlist").Result()
	if err == redis.Nil {
		fmt.Println("key2 does not exist")
	} else if err != nil {
		panic(err)
	} else {
		fmt.Println("watchlist", val2)
	}
	// Output: key value
	// key2 does not exist
}

func main() {

    ctab := crontab.New() // create cron table

    // AddJob and test the errors
    err := ctab.AddJob("0 12 1 * *", myFunc) // on 1st day of month
    if err != nil {
        log.Println(err)
        return
    }    

    // MustAddJob is like AddJob but panics on wrong syntax or problems with func/args
    // This aproach is similar to regexp.Compile and regexp.MustCompile from go's standard library,  used for easier initialization on startup
    ctab.MustAddJob("* * * * *", myFunc) // every minute
    ctab.MustAddJob("0 12 * * *", myFunc3) // noon lauch

    // fn with args
    ctab.MustAddJob("0 0 * * 1,2", myFunc2, "Monday and Tuesday midnight", 123) 
    ctab.MustAddJob("*/5 * * * *", myFunc2, "every five min", 0)

    // all your other app code as usual, or put sleep timer for demo

    keeper := KeeperClient()
    GetSettings(keeper)

    time.Sleep(11 * time.Minute)
}

func myFunc() {
    fmt.Println("Helo, world")
}

func myFunc3() {
    fmt.Println("Noon!")
}

func myFunc2(s string, n int) {
    fmt.Println("We have params here, string", s, "and number", n)
}
