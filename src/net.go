
package main

import (
    "fmt"
    "net"
    "net/rpc"
    "strconv"
    "encoding/json"
    "strings"
    "sync"
    "time"
)

const UDPMessageSize = 512

type NetRequest struct {
    Addr string
    Body []byte
}
type NetResponse struct {

}

type NetEventHandler func(... interface{})

type NetUDP struct {
    sock    *net.UDPConn
    in      chan *NetRequest
    out     chan *NetRequest
}

func NewUDPInstance() *NetUDP {
    
    p := new(NetUDP)

    p.in    = make(chan *NetRequest, 100000)
    p.out   = make(chan *NetRequest, 100000)

    go p.handleSending()

    return p
}

func (nc *NetUDP) ListenAndServe(port string, f NetEventHandler) (err error) {
    
    var addr *net.UDPAddr
    
    if addr, err = net.ResolveUDPAddr("ud4", ":"+ port); err != nil {
        fmt.Println("error: ListenUDP() laddr: ", err)
        return err
    }

    if nc.sock, err = net.ListenUDP("udp4", addr); err != nil {
        fmt.Println("error: ListenUDP() ", err)
        return err
    }

    go func() {
        for {
            var buf [UDPMessageSize]byte
            n, addr, err := nc.sock.ReadFromUDP(buf[0:])
            if err != nil {
                fmt.Println("error receiving(): ", err)
            }

            msg := make([]byte, n)
            copy(msg, buf[0:n])
        
            nc.in <- &NetRequest{addr.String(), msg}
        }
    }()

    go func() {
        for p := range nc.in {
            go f(nc, p)
        }
    }()

    return nil
}

func (nc *NetUDP) handleSending() {

    for p := range nc.out {

        go func() {
            
            if p == nil {
                return
            }
        
            if p.Addr == "" {
                return
            }

            addr, err := net.ResolveUDPAddr("ip4", p.Addr)
            if err != nil {
                fmt.Println("error: handleSending() invalid p.Addr")
                return
            }

            if _, err = nc.sock.WriteTo(p.Body, addr); err != nil {
                fmt.Println("error: handleSending() ", addr.String(), err)
            }
        }()
    }
}


func (nc *NetUDP) Send(body interface{}, addr string) {

    switch v := body.(type) {
    case []byte:
        nc.out <- &NetRequest{addr, v}
    case string:
        nc.out <- &NetRequest{addr, []byte(v)}
    case map[string]string:
        mb, _ := json.Marshal(v)
        nc.out <- &NetRequest{addr, mb}
    default:
    }
}

type NetTCP struct {

    ln      net.Listener

    out     chan *NetCall

    pool    map[string]*rpc.Client

    Lock    sync.Mutex    
}

type NetCall struct {
    
    Method string
    Addr   string
    
    Args interface{}
    Reply interface{}
    
    Status chan uint8
    Timeout time.Duration
}

func NewNetCall() *NetCall {
    
    c := new(NetCall)
    c.Status = make(chan uint8, 2)
    c.Timeout = 30e9

    return c
}

func NewTCPInstance() *NetTCP {
    
    nc := new(NetTCP)

    nc.out   = make(chan *NetCall, 100000)

    nc.pool  = map[string]*rpc.Client{}

    go nc.sending()

    return nc
}

func (nc *NetTCP) Listen(port string) (err error) {

    nc.ln, err = net.Listen("tcp", ":"+ port)
    if err != nil {
        fmt.Println("listen error:", err)
    }

    return nil
}

func (nc *NetTCP) Call(call *NetCall) {
    nc.out <- call
}

func (nc *NetTCP) sending() {

    var err error

    for p := range nc.out {
        
        var sock *rpc.Client

        if sock = nc.pool[p.Addr]; sock == nil {
             if sock, err = rpc.DialHTTP("tcp", p.Addr); err != nil {
                return
            } else {
                nc.pool[p.Addr] = sock
            }
        }

        go func() {
          
            rs := sock.Go(p.Method, p.Args, p.Reply, nil)

            select {

            case <- rs.Done:
                p.Status <- 1
            case <- time.After(p.Timeout):
                p.Status <- 9
            }

        }()
    }
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

const MessageSize = 512

type Packet struct {
    Addr string
    Msg  []byte
}

type Kpnet struct {
    sock *net.UDPConn
    in  chan *Packet
    out chan *Packet
}

type EventHandler func(... interface{})

func NewNet(port int) *Kpnet {
    
    kpn := new(Kpnet)
    
    kpn.in  = make(chan *Packet, 100000)
    kpn.out = make(chan *Packet, 100000)

    kpn.Listen(port)
    
    return kpn
}

func (kpn *Kpnet) Listen(port int) (err error) {
    
    var laddr *net.UDPAddr
    if laddr, err = net.ResolveUDPAddr("ud4", ":" + strconv.Itoa(port)); err != nil {
        fmt.Println("error: ListenUDP() laddr: ", err)
        return err
    }
    if kpn.sock, err = net.ListenUDP("udp4", laddr); err != nil {
        fmt.Println("error: ListenUDP() ", err)
        return err
    }

    go kpn.handleReceiving()
    go kpn.handleSending()
    go kpn.dispatching()

    return nil
}

func (kpn *Kpnet) Send(msg interface{}, addr string) {

    switch v := msg.(type) {
    case []byte:
        kpn.out <- &Packet{addr, v}
    case string:
        kpn.out <- &Packet{addr, []byte(v)}
    case map[string]string:
        mb, _ := json.Marshal(v)
        kpn.out <- &Packet{addr, mb}
        //fmt.Println("Send (type) map")
    default:
    }
}

func (kpn *Kpnet) handleSending() {

    for p := range kpn.out {

        go func() {
            
            if p == nil {
                return
            }
        
            if p.Addr == "" {
                return
            }

            addr, err := net.ResolveUDPAddr("ip4", p.Addr)
            if err != nil {
                fmt.Println("error: handleSending() invalid p.Addr")
                return
            }

            if _, err = kpn.sock.WriteTo(p.Msg, addr); err != nil {
                fmt.Println("error: handleSending() ", addr.String(), err)
            }
        }()
    }
}

func (kpn *Kpnet) handleReceiving() {

    for {
        var buf [MessageSize]byte
        n, addr, err := kpn.sock.ReadFromUDP(buf[0:])
        if err != nil {
            fmt.Println("error receiving(): ", err)
        }

        msg := make([]byte, n)
        copy(msg, buf[0:n])
        
        kpn.in <- &Packet{addr.String(), msg}
    }
}

func (kpn *Kpnet) dispatching() {
    for p := range kpn.in {
        go dispatchEvent(kpn, p)
    }
}

func dispatchEvent(kpn *Kpnet, p *Packet) {

    var f interface{}
    err := json.Unmarshal(p.Msg, &f)
    if err != nil {
        return
    }
    
    req := f.(map[string]interface{})
    action, ok := req["action"] 
    if !ok {
        return
    }

    // fmt.Println("dispatchEvent -> ", action.(string), "\n\t", req)

    ip := strings.Split(p.Addr, ":")[0]

    switch action.(string) {
    
    case "NodeCast":
        ActionNodeCast(req, ip)
    
    case "LedNew":
        ActionLedNew(req, ip)
    case "LedNewCb":
        ActionLedNewCb(req, ip)
    case "LedValue":
        ActionLedValue(req, ip)
    case "LedCast":
        ActionLedCast(req, ip)

    /* case "ItemPut":
        ActionItemPut(req, ip)
    case "ItemPutCb":
        ActionItemPutCb(req, ip)
    case "AgentItemPutCb":
        ActionAgentItemPutCb(req, ip) */
    
    /* case "LockLease":
        ActionLedNew(req, ip)
    case "GroupLease":
        ActionLedNew(req, ip)
    case "WatchReg":
        ActionLedNew(req, ip)
    case "WatchNotify":
        ActionLedNew(req, ip) */
    }

    return
}