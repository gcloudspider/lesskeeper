package main

import (
    "fmt"
	"time"
    //"math"
) 

var kpd Kpdata
var kpn *Kpnet

func main() {
    
    start := time.Now()
    
    kpd.Initialize()

    kpn = NewNet(9528)
    
    // go client-cronjob
    JobTrackerLocal()

    fmt.Println(time.Since(start))

    // go checker
    for {
        //udpRequest();
        time.Sleep(1e9)
    }
}