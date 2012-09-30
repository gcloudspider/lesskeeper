package main

import (
    "fmt"
    "net"
    "net/rpc"
    "strconv"
    "os"
    //"errors"
    "strings"
    "time"
    "sync"
)

//const MAX_QUERYBUF_LEN = 1024 * 1024    // 1GB max query buffer
const IOBUF_LEN         = 32
const INLINE_MAX_SIZE   = 1024 * 64      // Max size of inline reads
const AGENT_TIMEOUT     = 3e9

type Agent struct {

    in  chan *Command
    out chan *Command

    //flags int
    //stat_numconnections int
    //maxclients int
    //maxidletime int

    clients map[string]*AgentClient

    Lock sync.Mutex
}

type AgentClient struct {
    // lastinteraction int
    Sig chan int
    Rep *Reply
}

func NewAgent(port int) *Agent {
    
    a    := new(Agent)
    a.in  = make(chan *Command, 100)
    a.out = make(chan *Command, 100)
    a.clients = map[string]*AgentClient{}

    go a.sending()

    go func() {
        
        ln, err := net.Listen("tcp", ":"+ strconv.Itoa(port))
        if err != nil {
            // handle error
        }

        for {
            conn, err := ln.Accept()
            if err != nil {
                // handle error
                continue
            }
            go a.handler(conn)
        }
    }()

    return a
}


func (a *Agent) sending() {

    sock, err := rpc.DialHTTP("tcp", "127.0.0.1:9531")
    if err != nil {
        fmt.Println("dialing:", err)
        os.Exit(0)
    }

    for cmd := range a.out {

        go func() {

            rep := new(Reply)

            // Asynchronous callback
            if cmd.Type == CmdAsync {
                _ = sock.Go("Server.Process", cmd, &rep, nil)
                return
            }
        
            // Synchronous Reply
            err = sock.Call("Server.Process", cmd, &rep)
        
            a.Lock.Lock()
            if c, ok := a.clients[cmd.Tag]; ok {
                c.Rep = rep
                c.Sig <- 1
            }
            a.Lock.Unlock()
        }()
    }
}

func (a *Agent) handler(conn net.Conn) {

    sid := NewRandString(16)
    
    c := new(AgentClient)
    c.Sig = make(chan int, 1)
    
    a.clients[sid] = c

    defer func() {
        conn.Close()
        a.Lock.Lock()
        delete(a.clients, sid)
        a.Lock.Unlock()
    }()

    //conn.SetReadTimeout(1e8)    

    qbuf            := []byte{}
    multiBulkLen    := 0
    bulkLen         := -1
    pos             := 0
  
    cmd := new(Command)
    cmd.Addr = locNodeAddr
    cmd.Tag  = sid

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

        // Process Multibulk Buffer
        if multiBulkLen == 0 {
            // Multi bulk length cannot be read without a \r\n
            li := strings.SplitN(string(qbuf[0:n]), "\r", 2)
            if len(li) == 1 {
                // TODO "Protocol error: too big mbulk count string"
                if len(li[0]) > INLINE_MAX_SIZE {
                    _, _ = conn.Write([]byte("-ERR\r\n"))
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

            cmd.Argv = map[int][]byte{}
            cmd.Type = CmdSync
            c.Rep = new(Reply)
        }

        for {
            // Read bulk length if unknown
            if bulkLen == -1 {
                
                li := strings.SplitN(string(qbuf[pos:]), "\r", 2)
                if len(li) == 1 {
                    if len(li[0]) > INLINE_MAX_SIZE {
                        // "Protocol error: too big bulk count string"
                        _, _ = conn.Write([]byte("-ERR\r\n"))
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
                
                cmd.Argv[len(cmd.Argv)] = qbuf[pos:pos+bulkLen]

                pos     += bulkLen + 2
                bulkLen = -1
                multiBulkLen--
            }

            if multiBulkLen <= 0 {
                //fmt.Println("multi bulk len END", len(cmd.Argv))
                break
            }
        }

        // RPC: Process Command
        if multiBulkLen == 0 && len(cmd.Argv) > 0 {

            qbuf = qbuf[pos:n]

            //fmt.Println("Agent DONE Buffer", sid, pos, len(qbuf), string(qbuf[0:pos]), string(qbuf[pos:]))
            switch string(cmd.Argv[0]) {
            case "PUT", "SET":
                cmd.Type = CmdAsync
            default:
                cmd.Type = CmdSync
            }

            // Append command object to RPC Queue
            a.out <- cmd

            // Waiting the reply, or socket timeout
            select {

            case st := <- c.Sig:

                //fmt.Println("c.Sig", st)
                var rsp string
                
                if c.Rep.Err != nil {
                
                    rsp = "-ERR\r\n"
                } else if st == 10 {
                    rsp = "+OK\r\n" // TODO
                } else if st == 9 {
                    rsp = "-ERR timeout\r\n"
                } else {

                    switch c.Rep.Type {
                    case ReplyOK:
                        rsp = "+OK\r\n" // TODO
                    case ReplyError:
                        rsp = "-ERR\r\n" // TODO
                    case ReplyString:
                        rsp = fmt.Sprintf("$%d\r\n%s\r\n",  len(c.Rep.Val), c.Rep.Val)
                    case ReplyMulti:
                        rsp = "+OK\r\n" // TODO
                    case ReplyInteger:
                        rsp = "+OK\r\n" // TODO
                    case ReplyNil:
                        rsp = "+OK\r\n" // TODO
                    default:
                        rsp = "-ERR\r\n" // TODO
                    }

                }

                _, _ = conn.Write([]byte(rsp))

            case <- time.After(AGENT_TIMEOUT):    // RPC timeout
                _, _ = conn.Write([]byte("-ERR timeout\r\n"))
                fmt.Println("Time out", sid)
            }
        }

    }

    return
}