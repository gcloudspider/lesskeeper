package main

import (
    agt "./agent"
    "./conf"
    "flag"
    "fmt"
    "math/rand"
    "net/http"
    "net/rpc"
    "os"
    "runtime"
    "time"
)

var db Kpdata

var peer *NetUDP

//var port = "9628"

var agent *Agent
var agentPort = "9530"

var agn *agt.Agent

var gport = "9538"
var gnet *NetTCP

var bcip = "127.0.0.1"

var kp = map[string]string{}

var cfg conf.Config

var flag_prefix = flag.String("prefix", "", "the prefix folder path")

func main() {

    start := time.Now()

    flag.Parse()
    var err error

    // Environment variable initialization
    runtime.GOMAXPROCS(runtime.NumCPU())
    rand.Seed(time.Now().UnixNano())

    //
    if cfg, err = conf.NewConfig(*flag_prefix); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    /** v2/ */
    db.Initialize(cfg)

    //peer = NewPeer(port)
    //peer.AddHandler(UDPdispatchEvent)

    peer = NewUDPInstance()
    peer.ListenAndServe(cfg.KeeperPort, CommandDispatchEvent)

    agn = new(agt.Agent)
    agn.Serve(cfg.AgentPort)
    /** /v2 */

    agent = NewAgent(agentPort)

    WatcherInitialize()
    //
    gnet = NewTCPInstance()
    if err := gnet.Listen(cfg.KeeperPort); err != nil {
        // TODO
    }

    pp := new(Proposer)
    rpc.Register(pp)

    at := new(Acceptor)
    rpc.Register(at)

    rpc.HandleHTTP()

    go http.Serve(gnet.ln, nil)

    // go client-cronjob
    go JobTrackerLocal()

    fmt.Println("Started in", time.Since(start))

    // go checker
    for {
        time.Sleep(3e9)
        //runtime.GC()
    }
}
