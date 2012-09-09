package main

import (
    "fmt"
	"time"
    "math/rand"
    "runtime"
)

type ClientWatcher struct {
    status chan int
}

var kpd Kpdata
var kpn *Kpnet
var kpcw = map[string]ClientWatcher{}

func main() {
    
    fmt.Println("Starting NumCPU:", runtime.NumCPU())
    
    runtime.GOMAXPROCS(runtime.NumCPU())

    rand.Seed(time.Now().UnixNano())

    start := time.Now()
    
    kpd.Initialize()

    kpn = NewNet(9528)
    
    // go client-cronjob
    go JobTrackerLocal()

    // go http servicing
    kpnhListenAndServe()

    fmt.Println(time.Since(start))

    // go checker
    for {
        //udpRequest();
        time.Sleep(1e9)
    }
}