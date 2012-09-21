package main 

import (
    "net/http"
    //"net/url"
    "time"
    "fmt"
    //"io"
    "io/ioutil"
    //"bytes"
    //"net/http/httputil"
    "strings"
    "regexp"
    "strconv"
    "encoding/json"
)

func kpnhListenAndServe() { 

    fmt.Println("Starting HTTP Server")

    http.HandleFunc("/", kpnhDefault)
    http.HandleFunc("/h5keeper/api/item", kpnhApiItem)
    http.HandleFunc("/h5keeper/api/item/", kpnhApiItem)
    http.HandleFunc("/h5keeper/api/item2", kpnhApiItem2)

    s := &http.Server{
        Addr:           ":9529",
        ReadTimeout:    10 * time.Second,
        WriteTimeout:   10 * time.Second,
        MaxHeaderBytes: 1 << 20,
    }
    //go kpnMonitoring()
    
    s.ListenAndServe()

    fmt.Println("http down")
}


func kpnMonitoring() {
    for {
        time.Sleep(1e9)
        fmt.Println("ClientWatcher", len(kpcw))
    }
}

func kpnhDefault(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "Hello, world")
}

func kpnhApiItem(w http.ResponseWriter, r *http.Request) {

    //buf := new(bytes.Buffer)
    //io.Copy(buf, r.Body)
    //body := buf.String()
    //fmt.Println("TEST")

    defer func() {
        r.Body.Close()
    }()

    //w.WriteHeader(200)
    //w.Header().Add("Connection", "close")
    //w.Write([]byte("OK"))
    //fmt.Fprint(w, "OK")
    //return

    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        return
    }

    ips := strings.Split(r.RemoteAddr, ":")
    if len(ips) < 2 {
        return
    }
    
    body2 := strings.Split(string(body), "\r\n")
    if len(body2) < 2 {
        return
    }

    cmd := strings.Split(body2[0], " ")
    if len(cmd) < 2 {
        return
    }
    
    if cmd[0] == "put" && (kpsLed == "" || kpsLed != locNode) {
        return
    }

    key := strings.TrimSpace(cmd[1])
    if len(key) < 1 || len(key) > 2000 {
        return
    }
    //fmt.Println("key", key)

    mat, _ := regexp.MatchString("^([0-9a-zA-Z ._-]{1,64})$", key)
    if !mat {
        //fmt.Println("match failed 1")
        return
    }

    linum := "0"
    ival := body2[1]

    //fmt.Println("val", body2[1])
    lset, err2 := kpd.Hgetall("c:def:"+ key)
    if cmd[0] == "put" {
        
        if err2 == nil {
            
            if lsetn, ok := lset["n"]; ok {
                if lsetn != linum {
                    linum = lsetn
                }
            }

            if lsetv, ok := lset["v"]; ok {
                if lsetv == ival {
                    fmt.Fprint(w, "OK")
                    return
                }
            }
        }

    } else if cmd[0] == "get" {

        if err2 != nil {
            fmt.Fprint(w, "ER")
        }

        msg := map[string]string{
            "node": locNode,
        }

        if lsetn, ok := lset["n"]; ok {
            msg["n"] = lsetn
            
        }
        if lsetv, ok := lset["v"]; ok {
            msg["v"] = lsetv
        }

        //fmt.Println("Send JSON")
        mb, _ := json.Marshal(msg)
        fmt.Fprint(w, string(mb))
        
        return
    } else {
        fmt.Println("CURRENT")
        fmt.Fprint(w, "ER")
        return
    }

    n , _ := kpd.Incrby("ctl:ltid", 1)
    kpnoi := len(kps) * n + kpsNum - 1

    kpnos := strconv.Itoa(kpnoi)

    req := map[string]string{
        "node": locNode,
        "action": "ItemPut",
        "ItemNumber": linum,
        "ItemKey": key,
        "ItemContent": ival,
        "ItemNumberNext": kpnos,
    }

    bdy := ips[0] +"\r\n"+ key +"\r\n"+ ival
    
    kpd.Setex("qk:"+ key + linum, 3, "0")
    kpd.Setex("qv:"+ key + linum, 3, bdy)

    
    /* kpcw[kpnos] = ClientWatcher{make(chan int, 2)}
    go func() {
        time.Sleep(3e9)
        kpcw[kpnos].status <- 9
        //delete(kpcw, kpnos)
    }() */

    kpn.Send(req, "255.255.255.255:9528")

    /* select {
    case st := <- kpcw[kpnos].status:
        //fmt.Println("RET", st)
        if st == 1 {
            fmt.Fprint(w, "OK")
        } else {
            fmt.Fprint(w, "ER")
        }
        delete(kpcw, kpnos)
    }*/

    fmt.Fprint(w, "OK")
   
    //fmt.Println("http.Method:", r.Method, "From:", r.RemoteAddr, "Len", r.ContentLength, string(body))

    return
}

func kpnhApiItem2(w http.ResponseWriter, r *http.Request) {
    
    fmt.Println("http.Method:", r.Method, "From:", r.RemoteAddr, "Len", r.ContentLength)

    fmt.Fprint(w, "Hello, world")
    
    //w.Body.Close()
}