package main

import (
    "fmt"
    "net"
    "strconv"
    //"errors"
    "strings"
    "sync"
    "time"
)

//const MAX_QUERYBUF_LEN = 1024 * 1024    // 1GB max query buffer
const AGENT_IOBUF_LEN       = 32
const AGENT_INLINE_MAX_SIZE = 1024 * 64      // Max size of inline reads
const AGENT_TIMEOUT         = 3e9

//var watches map[string]ProposalWatcher

type Agent struct {

    //flags int
    //stat_numconnections int
    //maxclients int
    //maxidletime int

    clients map[string]*AgentClient

    watchmq chan string

    Lock sync.Mutex

    net *NetTCP
}

type AgentClient struct {
    // lastinteraction int
    Sig         chan int
    Rep        *Reply
    WatchPath   string
}

func NewAgent(port string) *Agent {
    
    this    := new(Agent)
    this.clients = map[string]*AgentClient{}
    this.watchmq = make(chan string, 100000)

    go func() {
        
        this.net = NewTCPInstance()
        
        if err := this.net.Listen(port); err != nil {
            // TODO
        }

        for {
            conn, err := this.net.ln.Accept()
            if err != nil {
                // handle error
                continue
            }
            go this.Handler(conn)
        }
        
    }()

    go func() {
        for path := range this.watchmq {
            // Println("Agent Watch Event", path)
            this.Lock.Lock()
            for _, c := range this.clients {
                if c.WatchPath == path {
                    c.Sig <- 1
                }
            }
            this.Lock.Unlock()            
        }
    }()

    return this
}

func (this *Agent) Handler(conn net.Conn) {

    sid := NewRandString(16)
    
    c := new(AgentClient)
    c.Sig = make(chan int, 1)
    
    this.clients[sid] = c

    defer func() {

        Println("connection close()")

        conn.Close()
        
        this.Lock.Lock()
        delete(this.clients, sid)
        this.Lock.Unlock()
    }()

    conn.SetDeadline(time.Now().Add(5 * time.Second))
    //return

    qbuf            := []byte{}
    multiBulkLen    := 0
    bulkLen         := -1
    pos             := 0
  
    call := NewNetCall()
    call.Method = "Proposer.Process"
    call.Addr = "127.0.0.1:"+gport

    argc := 0
    argv := map[int][]byte{}

    for {

        var buf [AGENT_IOBUF_LEN]byte
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
                if len(li[0]) > AGENT_INLINE_MAX_SIZE {
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
       
            // Reset all
            argc = 0
            argv = map[int][]byte{}
            c.WatchPath = ""
        }

        for {
            // Read bulk length if unknown
            if bulkLen == -1 {
                
                li := strings.SplitN(string(qbuf[pos:]), "\r", 2)
                if len(li) == 1 {
                    if len(li[0]) > AGENT_INLINE_MAX_SIZE {
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
                
                argv[argc] = qbuf[pos:pos+bulkLen]
                argc++

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
        if multiBulkLen == 0 && argc > 0 {

            qbuf = qbuf[pos:n]
            var rsp string

            // fmt.Println("Agent DONE Buffer", sid, pos, len(qbuf), string(qbuf[0:pos]), string(qbuf[pos:]))

            // Watch(path, ttl, +sid)
            if string(argv[0]) == "WATCH" {
                //Println("argv", argv)
                if len(argv) == 3 {
                    argv[3] = []byte(locNode)
                    c.WatchPath = strings.Trim(string(argv[1]), "/")
                }
            }

            // Append command object to RPC Queue
            call.Args = argv
            call.Reply = new(Reply)
            
            this.net.Call(call)

            //fmt.Println("req", call)

            st := <- call.Status            
            rs := call.Reply.(*Reply)

            //fmt.Println("call.reply", st, rs)

            if st == 9 {
                rsp = "-ERR timeout\r\n"
            } else if rs.Err != nil {
                rsp = "-ERR\r\n"
            } else {

                switch rs.Type {
                case ReplyOK:
                    rsp = "+OK\r\n" // TODO
                case ReplyError:
                    rsp = "-ERR\r\n" // TODO
                case ReplyString:
                    rsp = fmt.Sprintf("$%d\r\n%s\r\n",  len(rs.Val), rs.Val)
                case ReplyMulti:
                    rsp = "+OK\r\n" // TODO
                case ReplyInteger:
                    rsp = "+OK\r\n" // TODO
                case ReplyNil:
                    rsp = "+OK\r\n" // TODO
                case ReplyWatch:
                    for {
                        t := time.Now()
                        ut := t.Unix()
                        select {
                        case <- c.Sig:
                            Println("Agent Watch Sig", c.Sig)
                            rsp = "+OK\r\n"
                            goto RSP
                        case <- time.After(3e9):
                            Println("Agent Watch Loop", ut)
                        }
                    }
                    rsp = "+OK\r\n"
                default:
                    rsp = "-ERR\r\n" // TODO
                }

            }

RSP:
            //Println("rsp", rsp)
            _, _ = conn.Write([]byte(rsp))
            
        }

    }

    return
}