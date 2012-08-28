
package main

import (
    "fmt"
    "net"
    "strconv"
    //"os"
    //"reflect"
    "encoding/json"
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
    
    kpn.in  = make(chan *Packet, 10000)
    kpn.out = make(chan *Packet, 10000)

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
            //fmt.Println("handleSending() OK")
        }
    }
}

func (kpn *Kpnet) handleReceiving() {
    
    //fmt.Println("enter handleReceiving()")

    /* udpAddr, err := net.ResolveUDPAddr("ip4", ":9528")
    if err != nil {
        fmt.Println("Err, ResolveUDPAddr, %s", err)
        os.Exit(1)
    }
    conn, err := net.ListenUDP("udp", udpAddr)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Fatal error ", err.Error())
        os.Exit(1)
    }*/

    for {

        var buf [MessageSize]byte
        n, addr, err := kpn.sock.ReadFromUDP(buf[0:])
        if err != nil {
            fmt.Println("error receiving(): ", err)
        }

        msg := make([]byte, n)
        copy(msg, buf[0:n])
        
        kpn.in <- &Packet{addr.String(), msg}

        fmt.Println("Receive(", string(buf[0:n]), ") len(", n, ")")

        //daytime := time.Now().String()
        //conn.WriteToUDP([]byte(daytime), addr)
        //fmt.Println("sever out: ", daytime)
    }
}

func (kpn *Kpnet) dispatching() {
    for p := range kpn.in {
        go dispatchEvent(kpn, p)
    }
}

func dispatchEvent(kpn *Kpnet, p *Packet) {
    
    //fmt.Println("dispatchEvent from: ",p.Addr, " body: ", string(p.Msg))

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

    switch action.(string) {
    case "NodeCast":
        ActionNodeCast(req, p.Addr)
    case "LedNew":
        ActionLedNew(req, p.Addr)
    case "LedNewCb":
        ActionLedNew(req, p.Addr)
    case "LedValue":
        ActionLedNew(req, p.Addr)
    case "LedCast":
        ActionLedNew(req, p.Addr)
    case "ItemPut":
        ActionLedNew(req, p.Addr)
    case "ItemPutCb":
        ActionLedNew(req, p.Addr)
    case "ItemPutCbClient":
        ActionLedNew(req, p.Addr)
    case "LockLease":
        ActionLedNew(req, p.Addr)
    case "GroupLease":
        ActionLedNew(req, p.Addr)
    case "WatchReg":
        ActionLedNew(req, p.Addr)
    case "WatchNotify":
        ActionLedNew(req, p.Addr)
    }

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