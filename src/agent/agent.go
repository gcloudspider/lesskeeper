package agent

import (
    "../peer"
    "net/http"
    "sync"
    "time"
    "../conf"
)

var pr *peer.NetTCP
var locker sync.Mutex

type Agent struct {
    Locker sync.Mutex
    net    *peer.NetTCP
    cfg    conf.Config
}

// API V2
func (this *Agent) Serve(port string) {

    //this.net = new(peer.NetTCP)
    //this.net = peer.NewTCPInstance()

    pr = peer.NewTCPInstance()

    go func() {

        http.HandleFunc("/h5keeper/api", ApiGen)
        http.HandleFunc("/h5keeper/apiv2", ApiV2)

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
