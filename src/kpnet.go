
package main

import (
    "fmt"
    "net"
    "strconv"
    //"os"
    //"reflect"
    "encoding/json"
    "strings"
)

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

        if p == nil {
            continue
        }
        
        if p.Addr == "" {
            continue
        }

        addr, err := net.ResolveUDPAddr("ip4", p.Addr)
        if err != nil {
            fmt.Println("error: handleSending() invalid p.Addr")
            continue
        }

        if _, err = kpn.sock.WriteTo(p.Msg, addr); err != nil {
            fmt.Println("error: handleSending() ", addr.String(), err)
        } else {
            //fmt.Println("handleSending() to", p.Addr)
        }
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

    //fmt.Println("Handling -> ", action.(string), "\n\t", req)

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

    case "ItemPut":
        ActionItemPut(req, ip)
    case "ItemPutCb":
        ActionItemPutCb(req, ip)
    case "AgentItemPutCb":
        ActionAgentItemPutCb(req, ip)
    
    case "LockLease":
        ActionLedNew(req, ip)
    case "GroupLease":
        ActionLedNew(req, ip)
    case "WatchReg":
        ActionLedNew(req, ip)
    case "WatchNotify":
        ActionLedNew(req, ip)
    }

    return
}


/*
func (kpn *Kpnet) udpRequest() {
    udpAddr, err := net.ResolveUDPAddr("up4", "127.0.0.1:5000")
    checkError(err)

    conn, err := net.DialUDP("udp", nil, udpAddr)
    checkError(err)
    
    _, err = conn.Write([]byte("hi echo"))
    checkError(err)
    fmt.Println("client send: hi echo")
    
    var buf [512]byte
    n, err := conn.Read(buf[0:])
    checkError(err)
    fmt.Println("clien response: ", string(buf[0:n]))   
}

func (kpn *Kpnet) checkError(err error) {
    if err != nil {
        fmt.Fprintf(os.Stderr, "Fatal error ", err.Error())
        os.Exit(1)
    }
}
**/