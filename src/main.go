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
var agn *Agent
var kpcw = map[string]ClientWatcher{}

func main() {
    
    fmt.Println("ENV NumCPU:", runtime.NumCPU())
    
    runtime.GOMAXPROCS(runtime.NumCPU())

    rand.Seed(time.Now().UnixNano())

    start := time.Now()
    
    NewServer(9531)

    agn = NewAgent(9530)

    kpd.Initialize()

    //time.Sleep(1800e9)

    kpn = NewNet(9528)
    
    // go client-cronjob
    go JobTrackerLocal()

    // go http servicing
    go kpnhListenAndServe()

    fmt.Println("Started in", time.Since(start))

    // go checker
    for {
        //udpRequest();
        time.Sleep(3e9)
        //runtime.GC()
    }
}