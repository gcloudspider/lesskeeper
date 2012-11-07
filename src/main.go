package main

import (
    "fmt"
	"time"
    "math/rand"
    "runtime"
    "net/rpc"
    "net/http"
)

// TODO read
// http://www.ituring.com.cn/article/14931
// https://github.com/astaxie/build-web-application-with-golang
//

var db Kpdata

var peer *NetUDP
var port = "9628"

var agent *Agent
var agentPort = "9530"

var gport = "9538"
var gnet *NetTCP

var bcip = "127.0.0.1"

var kp = map[string]string{}

func main() {

    start := time.Now()
    
    // Environment variable initialization
    runtime.GOMAXPROCS(runtime.NumCPU())
    rand.Seed(time.Now().UnixNano())    

    /** v2/ */
    db.Initialize()

    //peer = NewPeer(port)
    //peer.AddHandler(UDPdispatchEvent)

    peer = NewUDPInstance()
    peer.ListenAndServe(port, CommandDispatchEvent)


    /** /v2 */

    agent = NewAgent(agentPort)


    // 
    gnet = NewTCPInstance()
    if err := gnet.Listen(gport); err != nil {
            // TODO
    }

    proposals = map[string]*Proposal{}
    
    pp := new(Proposer)
    rpc.Register(pp)
    //rpc.HandleHTTP()

    at := new(Acceptor)
    rpc.Register(at)
    
    rpc.HandleHTTP()
   
    go http.Serve(gnet.ln, nil)

    // go client-cronjob
    go JobTrackerLocal()

    fmt.Println("Started in", time.Since(start))

    // go checker
    for {
        //udpRequest();
        //fmt.Println(kp, kps, kpls)
        time.Sleep(3e9)
        //runtime.GC()
    }
}