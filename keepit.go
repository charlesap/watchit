package main

import (
    "fmt"
//    "log"
    "time"
    "strconv"
    "strings"
//    "net"
    "os"
    "os/exec"

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
    rrdmap := make(map[string]string)

//    watchlist := RedisStringArray(keeper,"watchlist")

    mid:="0"
    for {
      res, err := keeper.XRead(&redis.XReadArgs{
         Streams: []string{"injest",mid},
         Count: 65535,
         //Block: 0,
      }).Result()
      if err != nil {
//          fmt.Println(err)
      }else{
         ic:=len(res[0].Messages)
        if ic > 0 {
          fmt.Println("got ",ic," items.")
          mid=res[0].Messages[ic-1].ID
          for _,e := range res[0].Messages{
//            fmt.Println(e.ID) // p: pov, m: method, e: element, v: value, w: watcher, f: frequency, t: type, s: timestamp
           
           rrdname:=e.Values["p"].(string)+"-"+e.Values["m"].(string)+"-"+e.Values["e"].(string)
           filename := rrdname+".rrd"
           if e.Values["t"].(string)=="r" {
            _,prs := rrdmap[rrdname]
            if !prs {
               
               
               if _, err := os.Stat(filename); os.IsNotExist(err) {  
                   n,_:=strconv.ParseInt(e.Values["s"].(string),10,0)
                   startat:=fmt.Sprint(n-60)
                   fmt.Println("/usr/bin/rrdtool", "create", filename, "--step", "1", "--start", startat, "DS:"+rrdname+":GAUGE:600:0:U", "RRA:AVERAGE:0.5:1:525600")
                   out, err := exec.Command("/usr/bin/rrdtool", "create", filename, "--step", "1", "--start", startat, "DS:"+rrdname+":GAUGE:600:0:U", "RRA:AVERAGE:0.5:1:525600").Output()
                   if err != nil {
                      fmt.Println(err)
                   }else{
                     fmt.Println(rrdname+".rrd"+" does not exist, created:",string(out))
                   }
               } else {
                   fmt.Println(rrdname+".rrd"+" exists")
               }
            }
                   fmt.Println("updating",rrdname,"via","/usr/bin/rrdtool", "update", filename, e.Values["s"].(string)+":"+e.Values["v"].(string) )
                   out, err := exec.Command("/usr/bin/rrdtool", "update", filename, e.Values["s"].(string)+":"+e.Values["v"].(string) ).Output()
                   if err != nil {
                      fmt.Println(err)
                   }else{
                     fmt.Println(rrdname+".rrd"+" updated:",string(out))
                   }

           }
           rrdmap[rrdname]=e.Values["s"].(string)
           fmt.Println(e.Values["p"],e.Values["m"],e.Values["e"],e.Values["v"],e.Values["w"],e.Values["f"],e.Values["t"],e.Values["s"])
           
          }
        }
      }
      time.Sleep(1 * time.Second)
    }

}
