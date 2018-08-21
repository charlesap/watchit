package main

import (
//    "crypto/tls"
    "net/http"
    "time"
    "strings"
    "fmt"
    "os"
    "io/ioutil"
    "os/exec"
)

const (
	ApacheFormatPattern = "%s - - [%s] \"%s %d %d\" %f\n"
)


type ApacheLogRecord struct {
	http.ResponseWriter

	ip                    string
	time                  time.Time
	method, uri, protocol string
	status                int
	responseBytes         int64
	elapsedTime           time.Duration
}

func (r *ApacheLogRecord) Log(out *os.File) {
	timeFormatted := r.time.Format("02/Jan/2006 03:04:05")
	requestLine := fmt.Sprintf("%s %s %s", r.method, r.uri, r.protocol)
	fmt.Fprintf(out, ApacheFormatPattern, r.ip, timeFormatted, requestLine, r.status, r.responseBytes,
		r.elapsedTime.Seconds())
}

func (r *ApacheLogRecord) Write(p []byte) (int, error) {
	written, err := r.ResponseWriter.Write(p)
	r.responseBytes += int64(written)
	return written, err
}

func (r *ApacheLogRecord) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}



func myHandler(w http.ResponseWriter, r *http.Request) {
	clientIP := r.RemoteAddr
	if colon := strings.LastIndex(clientIP, ":"); colon != -1 {
		clientIP = clientIP[:colon]
	}

	record := &ApacheLogRecord{
		ResponseWriter: w,
		ip:             clientIP,
		time:           time.Time{},
		method:         r.Method,
		uri:            r.RequestURI,
		protocol:       r.Proto,
		status:         http.StatusOK,
		elapsedTime:    time.Duration(0),
	}

	startTime := time.Now()

	lp := len(r.URL.Path)
	if lp > 10 && r.URL.Path[1:5] == "kzzp" {
		http.ServeFile(w, r, "/aidata/"+r.URL.Path[1:])
	} else if lp > 10 && r.URL.Path[1:11] == "robots.txt" {

	} else if lp > 6 && (r.URL.Path[lp-4:] == ".png" ) {
             dataname:=r.URL.Path[1:lp-6]
             span:=r.URL.Path[lp-5:lp-4]
             st:="now-1h"
             if span=="h" {
             }else if span=="d"{
               st="now-24h"
             }else if span=="w"{
               st="now-168h"
             }else if span=="m"{
               st="now-744h"
             }else{
               span="h"
             }
             if _, err := os.Stat(dataname+".rrd"); os.IsNotExist(err) {  
                w.Header().Set("Content-Type", "text/html")
                fmt.Fprintf(w, "no dataset %s\n", r.URL.Path)
             }else{
                out, err := exec.Command("/usr/bin/rrdtool", 
                                         "graph",
                                         dataname+"."+span+".png", 
                                         "--start", st,
                                         "--end", "now",
                                         "DEF:"+dataname+"="+dataname+".rrd:"+dataname+":AVERAGE",
                                         "LINE2:"+dataname+"#FF0000").Output()
                if err != nil {
                  w.Header().Set("Content-Type", "text/html")
                  fmt.Fprintf(w, "rrd creation error %s<br>%s", err, string(out))
		}else{
		  http.ServeFile(w, r, "/aidata/"+dataname+"."+span+".png")
		}
             }
	} else {
                w.Header().Set("Content-Type", "text/html")
                files, err := ioutil.ReadDir(".")
                if err != nil {
                   //    log.Fatal(err)
                }
                fmt.Fprintf(w, "Streams:<br><br>")
                for _, file := range files {
                   lfn:=len(file.Name())
                   if file.Name()[lfn-4:] == ".rrd" {
                       fmt.Fprintf(w,"<img src=\"%s.h.png\"> %s ", file.Name()[:lfn-4],file.Name()[:lfn-4])
                       fmt.Fprintf(w,"<a href=\"%s.h.png\">hour</a> ", file.Name()[:lfn-4])
                       fmt.Fprintf(w,"<a href=\"%s.d.png\">day</a> ", file.Name()[:lfn-4])
                       fmt.Fprintf(w,"<a href=\"%s.w.png\">week</a> ", file.Name()[:lfn-4])
                       fmt.Fprintf(w,"<a href=\"%s.m.png\">month</a><br>", file.Name()[:lfn-4])
                   }
                }

                //fmt.Fprintf(w, "hello, you've hit %s\n", r.URL.Path)
                //http.ServeFile(w, r, "/aidata/keep/index.html")
	}

	finishTime := time.Now()
	record.time = finishTime.UTC()
	record.elapsedTime = finishTime.Sub(startTime)
        log, err := os.OpenFile("/var/log/serveit.log", os.O_RDWR|os.O_APPEND, 0666)
	record.Log(log)
	if err != nil {
	   fmt.Println(err)
	}
	
	log.Close()
}

func main() {

    http.ListenAndServe(":8888", http.HandlerFunc(myHandler))

}
