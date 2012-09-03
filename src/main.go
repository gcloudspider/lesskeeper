package main

import (
    "fmt"
	"time"
    "math/rand"
    "runtime"
) 

var kpd Kpdata
var kpn *Kpnet

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