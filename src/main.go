package main

import (
    "./agent"
    "./conf"
    "./peer"
    "./store"
    "flag"
    "fmt"
    "log"
    "math/rand"
    "net/http"
    "net/rpc"
    "os"
    "runtime"
    "time"
)

var stor store.Store

var prbc *peer.NetUDP
var prkp *peer.NetTCP

var agt *agent.Agent

var bcip = "127.0.0.1"

var cfg conf.Config

var flag_prefix = flag.String("prefix", "", "the prefix folder path")

var err error

var lgr *log.Logger

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

    logfile := cfg.Prefix + "/var/keeper.log"
    if f, e := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); e == nil {
        lgr = log.New(f, "", log.Ldate|log.Ltime)
    } else {
        os.Exit(1)
    }

    stor.Initialize(cfg)

    prbc = peer.NewUDPInstance()
    prbc.ListenAndServe(cfg.KeeperPort, CommandDispatchEvent)

    //agt = new(agent.Agent)
    agt = agent.NewAgentInstance(cfg, stor)
    agt.Serve(cfg.AgentPort)

    //WatcherInitialize()

    //
    pp := new(Proposer)
    rpc.Register(pp)

    at := new(Acceptor)
    rpc.Register(at)

    rpc.HandleHTTP()

    prkp = peer.NewTCPInstance()
    if err := prkp.Listen(cfg.KeeperPort); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    go http.Serve(prkp.Ln, nil)

    // go client-cronjob
    go JobTrackerLocal()

    fmt.Println("Started in", time.Since(start))

    defer func() {
        if err := recover(); err != nil {
            lgr.Printf("main panic:", err)
        }
    }()

    // go checker
    for {
        time.Sleep(3e9)
        //runtime.GC()
    }
}
