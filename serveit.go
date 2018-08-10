package main

import (
//    "crypto/tls"
    "net/http"
    "time"
    "strings"
    "fmt"
    "os"
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

	} else if lp > 6 && (r.URL.Path[lp-4:] == ".png" || r.URL.Path[lp-5:] == ".jpeg" || r.URL.Path[lp-5:] == ".json") {
		http.ServeFile(w, r, "/aidata/keep/"+r.URL.Path[1:])
	} else {
                http.ServeFile(w, r, "/aidata/keep/index.html")
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
