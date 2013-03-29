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
    //"time"
    "strings"
)

func ApiDebug(w http.ResponseWriter, r *http.Request) {
    
    io.WriteString(w, "{\"Status\": \"ERR\"}")

    /* hj, ok := w.(http.Hijacker)
    if !ok {
        http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
        return
    }
    conn, bufrw, err := hj.Hijack()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    // Don't forget to close the connection:
    defer conn.Close()
    bufrw.WriteString("Now we're speaking raw TCP. Say hi: ")
    bufrw.Flush()
    */


    //s, err := bufrw.ReadString('\n')
    //if err != nil {
    //    fmt.Println("error reading string: %v", err)
    //    return
    //}
    //fmt.Fprintf(bufrw, "You said: %q\nBye.\n", s)
    //bufrw.Flush()
}

func ApiGen(w http.ResponseWriter, r *http.Request) {

    defer func() {
        //fmt.Println("defer close")
        r.Body.Close()
    }()

    args   := map[int][]byte{}
    args[0] = []byte(strings.ToUpper(r.FormValue("func")))
    args[1] = []byte(r.FormValue("path"))

    call           := peer.NewNetCall()
    call.Method     = "Proposer.Process"
    call.Addr       = "127.0.0.1:9538"
    call.Args       = args
    call.Reply      = new(peer.Reply)

    pr.Call(call)    

    st := <-call.Status
    close(call.Status)

    rs := call.Reply.(*peer.Reply)

    if st == 9 {
        rs.Type = peer.ReplyTimeout
    }

    if rs.Err != nil {
        rs.Type = peer.ReplyError
    }

    if rs.Type == peer.ReplyWatch {
        /**
        for {
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
        }
        */
    }

//RSP:
    w.Header().Add("Connection", "close")

    if rsjson, err := json.Marshal(rs); err == nil {
        //w.Header().Add("Content-Length", string(len(rsjson)))
        io.WriteString(w, string(rsjson))
    } else {
        io.WriteString(w, "{\"Status\": \"ERR\"}")
    }

    if false {
        fmt.Println("hi return")
    }
    return
}