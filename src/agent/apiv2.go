package agent

import (
    "fmt"
    "io"
    "net/http"
    "io/ioutil"
    "../peer"
    "../utils"
    "strings"
)

func ApiV2(w http.ResponseWriter, r *http.Request) {
   
    var rsp *peer.Reply

    defer func() {
        w.Header().Add("Connection", "close")
        
        if rspj, err := utils.JsonEncode(rsp); err == nil {
            io.WriteString(w, rspj)
        }

        r.Body.Close()
    }()

    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        return
    }

    var req peer.Request
    err = utils.JsonDecode(string(body), &req)
    if err != nil {
        return
    }
    
    req.Method = strings.ToUpper(req.Method)
    req.Body = string(body)

    call := peer.NewNetCall()
    call.Method = "Proposer.Cmd"
    call.Addr = "127.0.0.1:9529"
    call.Args = req
    call.Reply = new(peer.Reply)

    pr.Call(call)

    st := <-call.Status
    close(call.Status)

    rsp = call.Reply.(*peer.Reply)

    if st == 9 {
        rsp.Type = peer.ReplyTimeout
    }

    if rsp.Err != nil {
        rsp.Type = peer.ReplyError
    }

    if rsp.Type == peer.ReplyWatch {
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
    goto RSP

RSP:
    

    /* if rsjson, err := json.Marshal(rsp); err == nil {
        //w.Header().Add("Content-Length", string(len(rsjson)))
        //io.WriteString(w, string(rsjson))
    } else {
        rsp.Err = errors.New("ERR")
        //io.WriteString(w, "{\"Status\": \"ERR\"}")
    } */

    if false {
        fmt.Println("hi return")
    }
    return
}
