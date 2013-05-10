package main

import (
    agt "./agent"
    "flag"
    "fmt"
    "math/rand"
    "net/http"
    "net/rpc"
    "os"
    //"os/exec"
    "runtime"
    "runtime/pprof"
    "time"
    "./conf"
    //"strings"
)
import _ "net/http/pprof"

// TODO read
// http://www.ituring.com.cn/article/14931
// http://talks.golang.org/2012/chat.slide#33
// http://select.yeeyan.org/view/94114/329073/author
//

var db Kpdata

var peer *NetUDP
var port = "9628"

//var agent *Agent
//var agentPort = "9530"

var agn *agt.Agent
var agnPort = "9531"

var gport = "9538"
var gnet *NetTCP

var bcip = "127.0.0.1"

var kp = map[string]string{}

var flag_prof   = flag.String("prof", "", "write cpu profile to file")
var flag_prefix = flag.String("prefix", "", "the prefix folder path")

func main() {

    flag.Parse()
    if *flag_prof != "" {
        f, err := os.Create(*flag_prof)
        if err != nil {
            Println(err)
        }
        pprof.StartCPUProfile(f)
        defer pprof.StopCPUProfile()
    }
    
    //
    conf, err := conf.NewConfig(*flag_prefix)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    
    start := time.Now()

    // Environment variable initialization
    runtime.GOMAXPROCS(runtime.NumCPU())
    rand.Seed(time.Now().UnixNano())

    /** v2/ */
    db.Initialize(conf)

    //peer = NewPeer(port)
    //peer.AddHandler(UDPdispatchEvent)

    peer = NewUDPInstance()
    peer.ListenAndServe(port, CommandDispatchEvent)

    agn = new(agt.Agent)
    agn.Serve(agnPort)
    /** /v2 */

    //agent = NewAgent(agentPort)

    WatcherInitialize()
    //
    gnet = NewTCPInstance()
    if err := gnet.Listen(gport); err != nil {
        // TODO
    }

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
        //pprof.StopCPUProfile()
        //runtime.GC()
    }
}
