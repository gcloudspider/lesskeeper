package peer

import (
    "encoding/json"
    "fmt"
    "net"
    "net/rpc"
    "time"
)

const UDPMessageSize = 512

type NetPacket struct {
    Addr string
    Body []byte
}

type NetUDPEventHandler func(*NetUDP, *NetPacket)

type NetUDP struct {
    sock *net.UDPConn

    in  chan *NetPacket
    out chan *NetPacket

    // Handle incoming packets read from the socket
    handlers []NetUDPEventHandler
}

func NewUDPInstance() *NetUDP {

    p := new(NetUDP)

    p.in = make(chan *NetPacket, 100000)
    p.out = make(chan *NetPacket, 100000)

    p.handlers = make([]NetUDPEventHandler, 0, 4)

    return p
}

// Registers an event handler which is invoked on incoming packets.
func (this *NetUDP) AddHandler(f NetUDPEventHandler) {
    this.handlers = append(this.handlers, f)
}

func (this *NetUDP) ListenAndServe(port string, f NetUDPEventHandler) (err error) {

    var addr *net.UDPAddr

    if addr, err = net.ResolveUDPAddr("udp4", ":"+port); err != nil {
        fmt.Println("error: ListenUDP() laddr: ", err)
        return err
    }

    if this.sock, err = net.ListenUDP("udp4", addr); err != nil {
        fmt.Println("error: ListenUDP() ", err)
        return err
    }

    this.handlers = append(this.handlers, f)

    go this.handleReceiving()
    go this.handleSending()
    go this.handleDispatching()

    return nil
}

func (this *NetUDP) Send(msg interface{}, addr string) {

    switch v := msg.(type) {
    case []byte:
        this.out <- &NetPacket{addr, v}
    case string:
        this.out <- &NetPacket{addr, []byte(v)}
    case map[string]string:
        mb, _ := json.Marshal(v)
        this.out <- &NetPacket{addr, mb}
    default:
    }
}

func (this *NetUDP) handleDispatching() {
    for p := range this.in {
        for _, f := range this.handlers {
            go f(this, p)
        }
    }
}

func (this *NetUDP) handleReceiving() {
    for {
        var buf [UDPMessageSize]byte
        n, addr, err := this.sock.ReadFromUDP(buf[0:])
        if err != nil {
            fmt.Println("error receiving(): ", err)
        }

        msg := make([]byte, n)
        copy(msg, buf[0:n])

        this.in <- &NetPacket{addr.String(), msg}
    }
}

func (this *NetUDP) handleSending() {

    for p := range this.out {

        go func() {

            if p == nil {
                return
            }

            if p.Addr == "" {
                return
            }

            addr, err := net.ResolveUDPAddr("udp4", p.Addr)
            if err != nil {
                fmt.Println("error: handleSending() invalid p.Addr")
                return
            }

            if _, err = this.sock.WriteTo(p.Body, addr); err != nil {
                fmt.Println("error: handleSending() ", addr.String(), err)
            }
        }()
    }
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
type NetTCP struct {
    Ln  net.Listener

    out chan *NetCall

    pool     chan *rpc.Client
    poolsize int
}

type NetCall struct {
    Method string
    Addr   string

    Args  interface{}
    Reply interface{}

    Status  chan uint8
    Timeout time.Duration
}

func NewNetCall() *NetCall {

    c := new(NetCall)
    c.Status = make(chan uint8, 4)
    c.Timeout = 20e9

    return c
}

func NewTCPInstance() *NetTCP {

    this := new(NetTCP)

    this.out = make(chan *NetCall, 30000)

    this.poolsize = 10
    this.pool = make(chan *rpc.Client, this.poolsize)
    for i := 0; i < this.poolsize; i++ {
        this.pool <- nil
    }

    go this.sending()

    return this
}

func (this *NetTCP) Listen(port string) (err error) {

    this.Ln, err = net.Listen("tcp", ":"+port)
    if err != nil {
        fmt.Println("listen error:", err)
    }

    return nil
}

func (this *NetTCP) Call(call *NetCall) {
    this.out <- call
}

func (this *NetTCP) sending() {

    var err error

    for p := range this.out {

        conn := <-this.pool

        if conn == nil {
            if conn, err = rpc.DialHTTP("tcp", p.Addr); err != nil {
                time.Sleep(1e9)
                this.pool <- nil
                this.out <- p
                continue
            }
        }

        go func(this *NetTCP, conn *rpc.Client, p *NetCall) {

            rs := conn.Go(p.Method, p.Args, p.Reply, nil)

            select {
            case <-rs.Done:
                p.Status <- 1
            case <-time.After(p.Timeout):
                p.Status <- 9
            }

            if rs.Error != nil {
                this.pool <- nil
                this.out <- p
                return
            }

            this.pool <- conn

        }(this, conn, p)

        /*
           var sock *rpc.Client

           if sock = this.pool[p.Addr]; sock == nil {
               if sock, err = rpc.DialHTTP("tcp", p.Addr); err != nil {
                   return
               } else {
                   this.pool[p.Addr] = sock
               }
           }*/
    }
}
