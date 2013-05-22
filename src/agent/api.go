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
    //"../data"
)

func ApiGen(w http.ResponseWriter, r *http.Request) {
   
    defer func() {
        //fmt.Println("defer close")
        r.Body.Close()
    }()

    var rs *peer.Reply
    args := map[int][]byte{}

    method := strings.ToUpper(r.FormValue("func"))
    /**
      if method == "GETLOCAL" {
      //fmt.Println(method)
          if rn, err := data.NodeGet(r.FormValue("path")); err == nil {
              rs.Body = rn.C
              rs.Type = peer.ReplyString
          } else {
              rs.Type = peer.ReplyError
          }

          return
      } */

    args[0] = []byte(method)
    args[1] = []byte(r.FormValue("path"))
    if body := r.FormValue("body"); body != "" {
        // TODO case ""
        args[2] = []byte(body)
    }

    call := peer.NewNetCall()
    call.Method = "Proposer.Process"
    call.Addr = "127.0.0.1:9538"
    call.Args = args
    call.Reply = new(peer.Reply)

    pr.Call(call)

    st := <-call.Status
    close(call.Status)

    rs = call.Reply.(*peer.Reply)

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
    goto RSP

RSP:
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
