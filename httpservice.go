package main

import (
	"math/rand"
	"time"
	"net/http"
	"log"
	"os"
	"strconv"
)

var (
	cpuTimeMs = 50
)

func loadCpu(t time.Duration) {
	timer := time.NewTimer(t)
	for {
		select {
		case <-timer.C:
			return
		default:
		}
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	firstCpuPartMs := rand.Intn(cpuTimeMs)
	secondCpuPartMs := cpuTimeMs - firstCpuPartMs
	log.Println("req", firstCpuPartMs, secondCpuPartMs)
	loadCpu(time.Duration(firstCpuPartMs)*time.Millisecond)
	time.Sleep(time.Duration(cpuTimeMs)*time.Millisecond)
	loadCpu(time.Duration(secondCpuPartMs)*time.Millisecond)
	w.Write([]byte("OK\n"))
	return
}


func main() {
	port := os.Getenv("HTTP_PORT")
	cpuTimeMsFromEnv, err := strconv.Atoi(os.Getenv("CPU_MS_PER_REQ"))
	if err != nil {
		log.Println("can't read env CPU_MS_PER_REQ", err, "using default")
	} else {
		cpuTimeMs = cpuTimeMsFromEnv
	}
	http.HandleFunc("/", handler)
	if port == "" {
		log.Println("empty env HTTP_PORT, using default 8080")
		port = "8080"
	}
	log.Println("listening", port)
	http.ListenAndServe(":" + port, nil)
}