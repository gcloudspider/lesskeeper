package agent

import (
    "../conf"
    "../peer"
    "../store"
    "net/http"
    "sync"
    "time"
)

var pr *peer.NetTCP
var locker sync.Mutex

type Agent struct {
    Locker sync.Mutex
    net    *peer.NetTCP
    cfg    conf.Config
    stor   store.Store
}

func NewAgentInstance(cfg conf.Config, stor store.Store) *Agent {

    this := new(Agent)

    this.cfg = cfg
    this.stor = stor

    return this
}

// API
func (this *Agent) Serve(port string) {

    //this.net = new(peer.NetTCP)
    //this.net = peer.NewTCPInstance()

    pr = peer.NewTCPInstance()

    go func() {

        http.HandleFunc("/lesskeeper/api", this.ApiHandler)

        s := &http.Server{
            Addr:           ":" + port,
            Handler:        nil,
            ReadTimeout:    30 * time.Second,
            WriteTimeout:   30 * time.Second,
            MaxHeaderBytes: 1 << 20,
        }
        s.ListenAndServe()

        return
    }()
}
