package main

import (
    "fmt"
    "net"
    "net/rpc"
    "strconv"
//    "net/rpc"
 //   "net/http"
    "os"
    "errors"
    //"bytes"
    "strings"
    "time"
    "sync"
)

const MAX_QUERYBUF_LEN = 1024 * 1024    // 1GB max query buffer
const IOBUF_LEN = 8
const INLINE_MAX_SIZE  = 1024 * 64      // Max size of inline reads

type AgentCommand struct {
    Tag string
    Argv map[int][]byte
    Addr string
    Type int
}

type Agent struct {
    ln  *net.Listener
    in  chan *AgentCommand
    out chan *AgentCommand

    flags int
    stat_numconnections int
    maxclients int
    maxidletime int

    Lock sync.Mutex
}

type AgentClient struct {
    status int
    lastinteraction int
    Callback chan int
    Reply *AgentReply
}

var clients = map[string]*AgentClient{}

type ReplyType uint8
const (
    ReplyStatus ReplyType = iota
    ReplyError
    ReplyInteger
    ReplyNil
    ReplyString
    ReplyMulti
)

// Reply holds a Command reply.
type AgentReply struct {
    Type    ReplyType     // Reply type
    Elems   []*AgentReply // Sub-replies
    Err     error         // Reply error
    Val     string
    Ver     uint64
}

// Str returns the reply value as a string or
// an error, if the reply type is not ReplyStatus or ReplyString.
func (r *AgentReply) Str() (string, error) {

    if r.Type == ReplyError {
        return "", r.Err
    }
    if !(r.Type == ReplyStatus || r.Type == ReplyString) {
        return "", errors.New("string value is not available for this reply type")
    }

    return r.Val, nil
}

func NewAgent(port int) *Agent {

    fmt.Println("Getting NewAgent")
    
    agn := new(Agent)
    agn.flags = 0
    agn.stat_numconnections = 0

    agn.in  = make(chan *AgentCommand, 100)
    agn.out = make(chan *AgentCommand, 100)
    
    go agn.handleSending()

    go func() {
        
        ln, err := net.Listen("tcp", ":"+ strconv.Itoa(port))
        agnCheckError(err)
        if err != nil {
            // handle error
        }

        for {            
            conn, err := ln.Accept()
            if err != nil {
                // handle error
                continue
            }
            go agn.agnNetHandle(conn)
        }
    }()

    go func() {
        for k, v := range(clients) {
            fmt.Println("current client:", k, v)
        }
    }()

    return agn
}


func (agn *Agent) handleSending() {

    client, err := rpc.DialHTTP("tcp", "127.0.0.1:9531")
    if err != nil {
        fmt.Println("dialing:", err)
        os.Exit(0)
    }

    for p := range agn.out {

        if p == nil {
            continue
        }
        
        if p.Tag == "" {
            continue
        }

        reply := new(AgentReply)
        //reply.Stat = map[string]string{}
        rs := client.Go("Command.Process", p, &reply, nil)
        //go func() {
        //    time.Sleep(30e9)
        //    status[p.Tag] <- 9
        //}()
        go func() {
            
            if p.Type == 1 {

                time.Sleep(12e9)

                agn.Lock.Lock()
                if c, ok := clients[p.Tag]; ok {
                    c.Reply = reply
                    c.Callback <- 9
                }
                agn.Lock.Unlock()
            
                return
            }

            <- rs.Done

            agn.Lock.Lock()
            if c, ok := clients[p.Tag]; ok {
                c.Reply = reply
                c.Callback <- 1
            }
            agn.Lock.Unlock()
        }()
    }
}

func (agn *Agent) agnNetHandle(conn net.Conn) {

    tag := NewRandString(16)

    defer func() {
        conn.Close()

        agn.Lock.Lock()
        delete(clients, tag)
        agn.Lock.Unlock()
    }()

    //conn.SetReadTimeout(1e8)

    qbuf := []byte{}

    multiBulkLen := 0
    bulkLen := -1

    pos := 0

    argc := 0
    argv := map[int][]byte{}
    
    client := new(AgentClient)
    client.Callback = make(chan int, 1)

    clients[tag] = client

    for {

        var buf [IOBUF_LEN]byte
        n, err := conn.Read(buf[0:])
        if err != nil {
            //if err.Timeout() {
            //    break
            //}
            return
        }
        if n > 0 {
            qbuf = append(qbuf, buf[0:n]...)
        }

        n = len(qbuf)
        //fmt.Println("Query Buffer", len(qbuf), "[", string(qbuf), "]")

        if multiBulkLen == 0 {

            // Multi bulk length cannot be read without a \r\n
            li := strings.SplitN(string(qbuf[0:n]), "\r", 2)
            if len(li) == 1 {
                // TODO "Protocol error: too big mbulk count string"
                if len(li[0]) > INLINE_MAX_SIZE {
                    _, _ = conn.Write([]byte("-ERROR"))
                }
                return // TODO
            }

            // Buffer should also contain \n
            if len(li[1]) < 1 || li[1][0] != 10 {
                return  // TODO
            }

            // We know for sure there is a whole line since newline != NULL,
            // so go ahead and find out the multi bulk length.
            if qbuf[0] != []byte("*")[0] {
                return // TODO
            }
            // multi bulk length can not be empty
            if len(li[0]) < 2 {
                return // TODO
            }
            //
            mblen, err := strconv.Atoi(li[0][1:])
            if err != nil || mblen > 1024 * 1024 {
                return // TODO "Protocol error: invalid multibulk length"
            }

            multiBulkLen = mblen
            pos = len(li[0]) + 2

            argc = 0
            argv = map[int][]byte{}
        }

        for {
            // Read bulk length if unknown
            if bulkLen == -1 {
                
                li := strings.SplitN(string(qbuf[pos:]), "\r", 2)
                if len(li) == 1 {
                    if len(li[0]) > INLINE_MAX_SIZE {
                        // "Protocol error: too big bulk count string"
                        _, _ = conn.Write([]byte("-ERROR"))
                    }
                    break // TODO
                }

                // Buffer should also contain \n
                if len(li[1]) < 1 || li[1][0] != 10 {
                    break  // TODO
                }

                if qbuf[pos] != []byte("$")[0] {
                    return // TODO
                }

                lis, err := strconv.Atoi(li[0][1:])
                if err != nil || lis < 0 || lis > 512 * 1024 * 1024 {
                    return // TODO "Protocol error: invalid bulk length"
                }

                pos += len(li[0]) + 2
                bulkLen = lis
            }

            /* Read bulk argument */
            if n - pos < bulkLen + 2 {
                // Not enough data (+2 == trailing \r\n)
                break
            } else {
                
                argv[argc] = qbuf[pos:pos+bulkLen]
                argc++

                pos += bulkLen + 2
                bulkLen = -1
                multiBulkLen--
            }

            if multiBulkLen <= 0 {
                //fmt.Println("multi bulk len END", len(argv))
                break
            }
        }

        // ProcessCommand()
        if multiBulkLen == 0 && argc > 0 {

            qbuf = qbuf[pos:n]

            //fmt.Println("DONE Buffer", pos, len(qbuf), string(qbuf[0:pos]), string(qbuf[pos:]))
            agn.out <- &AgentCommand{Tag: tag, Argv: argv, Addr: locNodeAddr}

            select {
            case rs := <- client.Callback:
                var rsp string
                if rs == 9 {
                    rsp = "-ERR timeout"
                } else if rs == 1 {
                    rsp = fmt.Sprintf(client.Reply.Val +"\r\n")
                }
                fmt.Println("Reply", client.Reply)
                //ret := fmt.Sprintf("-ERROR %d\r\n", st)
                //rsp := fmt.Sprintf(client.Reply.Val +"\r\n")
                _, _ = conn.Write([]byte(rsp))
                //fmt.Println("Call Back", st)
                if false {
                    time.Sleep(2e9)
                }
            }
        }
    }

    return
}

func agnCheckError(err error) {
    if err != nil {
        fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
    }
}