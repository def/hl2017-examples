package main

import (
	"time"
	"net/http"
	"log"
	"os"
	"github.com/quipo/statsd"
	"strconv"
)

var (
	statsdClient *statsd.StatsdClient
)

func factorial(n uint64) (result uint64) {
	if (n > 0) {
		result = n * factorial(n-1)
		return result
	}
	return 1
}


func handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var n uint64 = 500000
	nStr := r.URL.Query().Get("n")
	if nStr != "" {
		var err error
		n, err = strconv.ParseUint(nStr, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	factorial(n)
	err := statsdClient.PrecisionTiming("request_time", time.Since(start))
	if err != nil {
		log.Println("failed to send statsd metrics", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("OK\n"))
	return
}


func main() {
	port := os.Getenv("HTTP_PORT")
	http.HandleFunc("/", handler)
	if port == "" {
		log.Println("empty env HTTP_PORT, using default 8080")
		port = "8080"
	}
	statsdClient = statsd.NewStatsdClient("127.0.0.1:8125", "httpservice.port=" + port + ".")
	err := statsdClient.CreateSocket()
	if err != nil {
		panic(err)
	}
	log.Println("listening", port)
	http.ListenAndServe(":" + port, nil)
}
