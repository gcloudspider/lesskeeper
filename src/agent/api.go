package agent


import (
    "fmt"
    "io"
    "net/http"
    //"io/ioutil"
    //"net/rpc"
    //"net"
    "../peer"
    "encoding/json"
)

func ApiList(w http.ResponseWriter, r *http.Request) {

    defer func() {
        r.Body.Close()
    }()

    call := peer.NewNetCall()
    call.Method = "Proposer.Process"
    call.Addr = "127.0.0.1:9538"
    call.Reply = new(peer.Reply)
   
    args := map[int][]byte{}
    args[0] = []byte("LIST")
    args[1] = []byte(r.FormValue("path"))

    call.Args = args

    pr.Call(call)

    st := <-call.Status
    rs := call.Reply.(*peer.Reply)

    var rsp string

    if st == 9 {
        rsp = "-ERR timeout\r\n"
    } else if rs.Err != nil {
        rsp = "-ERR\r\n"
    } else {

        switch rs.Type {
        case peer.ReplyOK:
            rsp = "+OK\r\n" // TODO
        case peer.ReplyError:
            rsp = "-ERR\r\n" // TODO
        case peer.ReplyString:
            rsp = fmt.Sprintf("$%d\r\n%s\r\n", len(rs.Val), rs.Val)
        case peer.ReplyMulti:
            rsp = "+OK\r\n" // TODO
        case peer.ReplyInteger:
            rsp = "+OK\r\n" // TODO
        case peer.ReplyNil:
            rsp = "+OK\r\n" // TODO
        case peer.ReplyWatch:

            /* for {
                t := time.Now()
                ut := t.Unix()
                select {
                case <-c.Sig:
                    Println("Agent Watch Sig", c.Sig, "Event", c.Rep.Val)
                    rsp = fmt.Sprintf("+%s\r\n", c.Rep.Val)
                    goto RSP
                case <-time.After(3e9):
                    // if the client closed
                    conn.SetDeadline(time.Now())
                    var buf [AGENT_IOBUF_LEN]byte
                    if _, err := conn.Read(buf[0:]); err == io.EOF {
                        rsp = fmt.Sprintf("-%s\r\n", "ERR")
                        goto RSP
                    }

                    // update ttl to proposer
                    msg := map[string]string{
                        "action": "WatchLease",
                        "host":   locNode,
                        "path":   c.WatchPath,
                        "ttl":    "6",
                    }
                    //Println(msg)
                    if ip, ok := kp[kpsLed]; ok {
                        peer.Send(msg, ip+":"+port)
                        //Println("Send", msg)
                    }

                    Println("Agent Watch Loop", ut)
                }
            } */

            //rsp = "+OK\r\n"

        default:
            rsp = "-ERR\r\n" // TODO
        }

    }

//RSP:
    //Println("rsp", rsp)
    //_, _ = conn.Write([]byte(rsp))

    if rsjson, err := json.Marshal(rs); err == nil {
        io.WriteString(w, string(rsjson))
    } else {
        io.WriteString(w, "{\"Status\": \"ERR\"}")
    }
    
    //fmt.Println(string(ret))

    //call.Args[1] = 
    if false {
        fmt.Println("hi", rsp)
    }

    //io.WriteString(w, "{\"status\": \"OK\"}")

    return
}