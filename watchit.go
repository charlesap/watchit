package main

import (
    "fmt"
//    "log"
    "time"
    "strconv"
    "strings"
    "net"
    "os"
    "os/exec"

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

    ctab.MustAddJob(wh, checkHost, wa[0], cf, client)
}

func doCheck(wi string, c string, client * redis.Client){

    ca:=strings.Split(c," ")
    if ca[0] == "ping" || ca[0] == "pingvia" {
       var out []byte
       chktime:=strconv.FormatInt(time.Now().UTC().UnixNano(),10)
       if ca[0] == "ping" {
         out, _ = exec.Command("/bin/ping", "-c", "1",wi).Output()
       }else{
         out, _ = exec.Command("/usr/bin/ssh", ca[1], "/bin/ping", "-c", "1",wi).Output()
       }
       la:=strings.Split(strings.Split(string(out),"\n")[1]," ")
       ok:="N"
       if len(la)>7 && la[2]=="from" {
         ok="Y"
       }
       n,_:=os.Hostname()
       rtt:=0.0
       if ok=="Y" {
          rtt, _ = strconv.ParseFloat(strings.Split(la[7],"=")[1],64)
       }
       //fmt.Println("XADD injest MAXLEN ~ 65536 *",wi,"ping","ok",ok,chktime)
       client.XAdd(&redis.XAddArgs{
          Stream: "injest",
          ID: "*",
          MaxLenApprox: 65536,
          Values: map[string]interface{}{ "p":wi , "m": "ping" , "e": "ok", "v": ok, "w": n, "t": chktime},
       }).Result()
       //fmt.Println("XADD injest *",wi,"ping","ms",rtt,chktime)       
       client.XAdd(&redis.XAddArgs{
          Stream: "injest",
          ID: "*",
          MaxLenApprox: 65536,
          Values: map[string]interface{}{ "p": wi , "m": "ping" , "e": "ms", "v": rtt, "w": n, "t": chktime},
       }).Result()
    }else if ca[0] == "snmp"{
       fmt.Println("querying ",wi)
    }else{
       fmt.Println("don't know how to use ",c," on ",wi)
    }

}

func checkHost(wi string,pf string, client * redis.Client) {
    
    _ , err := net.LookupHost(wi)
    if err != nil {
      fmt.Println("Can't find ", wi, " (DNS resolution failed)" )
    }else{
      for _,s:= range strings.Split( pf, "\n"){
        if s[0]!='#'{
          doCheck( wi, s, client )
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
