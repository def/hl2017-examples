package main

import (
	"time"
	"net/http"
	"log"
	"os"
	"github.com/quipo/statsd"
	"strconv"
	"math/rand"
)

var (
	statsdClient *statsd.StatsdClient
	datafilePath string
	datafileSize int64
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
	nStr := r.URL.Query().Get("n")
	bStr := r.URL.Query().Get("b")
	switch {
	case nStr != "":
		n, err := strconv.ParseUint(nStr, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		factorial(n)
		time.Sleep(20 * time.Millisecond)
		resp := make([]byte, 50000)
		rand.Read(resp)
		w.Write(resp)
	case bStr != "":
		b, err := strconv.ParseInt(bStr, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		f, err := os.Open(datafilePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		randLimit := datafileSize - b
		if randLimit < 0 {
			http.Error(w, "invalid size", http.StatusBadRequest)
			return
		}
		data := make([]byte, int(b))
		if _, err := f.ReadAt(data, rand.Int63n(randLimit)); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Write(data)
	}
	err := statsdClient.PrecisionTiming("request_time", time.Since(start))
	if err != nil {
		log.Println("failed to send statsd metrics", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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
	datafilePath = os.Getenv("DATAFILE_PATH")
	datafile, err := os.Open(datafilePath)
	if err != nil {
		panic(err)
	}
	st, err := datafile.Stat()
	if err != nil {
		panic(err)
	}
	datafileSize = st.Size()
	log.Println("datafile size", datafileSize)
	datafile.Close()
	log.Println("listening", port)
	http.ListenAndServe(":" + port, nil)
}
