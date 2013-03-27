package agent

import (
    //"fmt"
    //"net"
    //"strconv"
    //"errors"
    //"io"
    //"strings"
    "sync"
    "time"
    "net/http"
    "../peer"
)

var pr         *peer.NetTCP
var locker      sync.Mutex

type Agent struct {
    Locker      sync.Mutex
    net         *peer.NetTCP
}

// API V2
func (this *Agent) Serve(port string) {

    //this.net = new(peer.NetTCP)
    //this.net = peer.NewTCPInstance()

    pr = peer.NewTCPInstance()

    go func() {

        http.HandleFunc("/h5keeper/api/list", ApiList)
        http.HandleFunc("/h5keeper/api/debug", ApiDebug)

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