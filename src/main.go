package main

import (
    "./agent"
    "./conf"
    pr "./peer"
    "./store"
    "flag"
    "fmt"
    "math/rand"
    "net/http"
    "net/rpc"
    "os"
    "runtime"
    "time"
)

var stor store.Store

var prbc *pr.NetUDP
var prkp *pr.NetTCP

var agt *agent.Agent

var bcip = "127.0.0.1"

var kp = map[string]string{}

var cfg conf.Config

var flag_prefix = flag.String("prefix", "", "the prefix folder path")

var err error

func main() {

    start := time.Now()

    // Environment variable initialization
    runtime.GOMAXPROCS(runtime.NumCPU())
    rand.Seed(time.Now().UnixNano())

    //
    flag.Parse()
    if cfg, err = conf.NewConfig(*flag_prefix); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    stor.Initialize(cfg)

    prbc = pr.NewUDPInstance()
    prbc.ListenAndServe(cfg.KeeperPort, CommandDispatchEvent)

    agt = new(agent.Agent)
    agt.Serve(cfg.AgentPort)

    //WatcherInitialize()

    //
    pp := new(Proposer)
    rpc.Register(pp)

    at := new(Acceptor)
    rpc.Register(at)

    rpc.HandleHTTP()

    prkp = pr.NewTCPInstance()
    if err := prkp.Listen(cfg.KeeperPort); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    go http.Serve(prkp.Ln, nil)

    // go client-cronjob
    go JobTrackerLocal()

    fmt.Println("Started in", time.Since(start))

    // go checker
    for {
        time.Sleep(3e9)
        //runtime.GC()
    }
}
