package main

import (
    "fmt"
//    "log"
//    "time"
    "strings"
    "net"

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


func Enroll(client * redis.Client, ctab *crontab.Crontab, item string, pfmap map[string]string){
    sa:=strings.Split(item,":")
    fmt.Println(strings.Trim(sa[1]," "))
    wh:=strings.Trim(sa[0]," ")
    wa:=strings.Split( strings.Trim(sa[1]," ")," ")
    _,prs := pfmap[wa[1]]
    if !prs {
      //fmt.Println("need ",wa[1])

      pfmap[wa[1]]=RedisString(client, wa[1])
    }
    cf,_ := pfmap[wa[1]]

    ctab.MustAddJob(wh, checkHost, wa[0], cf)
}

func doCheck(wi string, c string){

    ca:=strings.Split(c," ")
    if ca[0] == "ping" {
       fmt.Println("pinging ",wi)
    }else if ca[0] == "snmp"{
       fmt.Println("querying ",wi)
    }else{
       fmt.Println("don't know how to use ",c," on ",wi)
    }

}

func checkHost(wi string,pf string) {
    
    _ , err := net.LookupHost(wi)
    if err != nil {
      fmt.Println("Can't find ", wi, " (DNS resolution failed)" )
    }else{
      for _,s:= range strings.Split( pf, "\n"){
        if s[0]!='#'{
          doCheck( wi, s )
        }
      }
    }
}

func main() {

    ctab := crontab.New() // create cron table

    keeper := KeeperClient()

    pfmap := make(map[string]string)

    watchlist := RedisStringArray(keeper,"watchlist")

    for _,s:= range watchlist{
      if s[0] != '#' {
        Enroll(keeper,ctab,s,pfmap)
      }
    }
    select{ }
}
