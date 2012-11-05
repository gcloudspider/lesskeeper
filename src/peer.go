
package main

import (
    "fmt"
    "net"
    "encoding/json"
    "strings"
)

const PeerMessageSize = 512

type PeerPacket struct {
    Addr string
    Msg  []byte
}

// Closure interface to handle incoming packets
type PeerEventHandler func(*Peer, *PeerPacket)

type Peer struct {
    in  chan *PeerPacket
    out chan *PeerPacket
    uc *net.UDPConn

    // Handle incoming packets read from the socket
    handlers []PeerEventHandler

    tl net.Listener
}

func NewPeer(port string) *Peer {

    p := new(Peer)
    
    p.in  = make(chan *PeerPacket, 100000)
    p.out = make(chan *PeerPacket, 100000)

    p.handlers = make([]PeerEventHandler, 0, 4)

    p.UDPListen(port)

    p.TCPListen(port)
    
    return p
}

// Registers an event handler which is invoked on incoming packets.
func (peer *Peer) AddHandler(f PeerEventHandler) {
    peer.handlers = append(peer.handlers, f)
}

func (peer *Peer) UDPListen(port string) (err error) {
    
    var laddr *net.UDPAddr
    if laddr, err = net.ResolveUDPAddr("ud4", ":"+ port); err != nil {
        fmt.Println("error: ListenUDP() laddr: ", err)
        return err
    }
    if peer.uc, err = net.ListenUDP("udp4", laddr); err != nil {
        fmt.Println("error: ListenUDP() ", err)
        return err
    }

    go peer.UDPhandleReceiving()
    go peer.UDPhandleSending()
    go peer.UDPhandleDispatching()

    return nil
}

func (peer *Peer) TCPListen(port string) (err error) {

    peer.tl, err = net.Listen("tcp", ":"+ port)
    if err != nil {
        fmt.Println("listen error:", err)
    }

    return nil
}

func (peer *Peer) Send(msg interface{}, addr string) {

    //fmt.Println("v2 Send -> ", addr, msg)

    switch v := msg.(type) {
    case []byte:
        peer.out <- &PeerPacket{addr, v}
    case string:
        peer.out <- &PeerPacket{addr, []byte(v)}
    case map[string]string:
        mb, _ := json.Marshal(v)
        peer.out <- &PeerPacket{addr, mb}
    default:
    }
}

func (peer *Peer) UDPhandleSending() {

    for p := range peer.out {

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

            if _, err = peer.uc.WriteTo(p.Msg, addr); err != nil {
                fmt.Println("error: handleSending() ", addr.String(), err)
            }

        }()
    }
}


func (peer *Peer) UDPhandleReceiving() {

    for {
        var buf [PeerMessageSize]byte
        n, addr, err := peer.uc.ReadFromUDP(buf[0:])
        if err != nil {
            fmt.Println("error receiving(): ", err)
        }

        msg := make([]byte, n)
        copy(msg, buf[0:n])
        
        peer.in <- &PeerPacket{addr.String(), msg}

        //fmt.Println("v2 receiving", addr, msg)
    }
}

func (peer *Peer) UDPhandleDispatching() {
    for p := range peer.in {
        for _, f := range peer.handlers {
            go f(peer, p)
        }
    }
}

func UDPdispatchEvent(peer *Peer, p *PeerPacket) {

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

    //fmt.Println("v2 dispatchEvent -> ", action.(string), "\n\t", req)

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
    }

    return
}
