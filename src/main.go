package main

import (
    "fmt"
	"time"
    "math/rand"
    "runtime"
)

var kpd Kpdata
var kpn *Kpnet
var agent *Agent

func main() {

    start := time.Now()

    runtime.GOMAXPROCS(runtime.NumCPU())

    rand.Seed(time.Now().UnixNano())    
    
    NewServer(9531)

    agent = NewAgent(9530)

    kpd.Initialize()

    kpn = NewNet(9528)
    
    // go client-cronjob
    go JobTrackerLocal()

    // go http servicing
    // go kpnhListenAndServe()

    fmt.Println("Started in", time.Since(start))

    // go checker
    for {
        //udpRequest();
        time.Sleep(3e9)
        //runtime.GC()
    }
}