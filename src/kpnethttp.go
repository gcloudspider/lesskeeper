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
)

func kpnhListenAndServe() { 

    fmt.Println("Starting HTTP Server")

    http.HandleFunc("/", kpnhDefault)
    http.HandleFunc("/h5keeper/api/item", kpnhApiItem)
    http.HandleFunc("/h5keeper/api/item2", kpnhApiItem2)

    s := &http.Server{
        Addr:           ":9529",
        ReadTimeout:    30 * time.Second,
        WriteTimeout:   30 * time.Second,
        //MaxHeaderBytes: 1 << 20,
    }

    s.ListenAndServe()
}


func kpnhDefault(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "Hello, world")
}

func kpnhApiItem(w http.ResponseWriter, r *http.Request) {

    //buf := new(bytes.Buffer)
    //io.Copy(buf, r.Body)
    //body := buf.String()

    defer r.Body.Close()

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
    if cmd[0] == "put" {
        lset, err2 := kpd.Hgetall("c:def:"+ key)
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
    }

    n , _ := kpd.Incrby("ct:ltid", 1)
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

    cq := ClientWatcher{make(chan int, 1)}
    kpcw[kpnos] = cq
    go func() {
        time.Sleep(3e9)
        cq.status <- 9
    }()

    kpn.Send(req, "255.255.255.255:9528")

    select {
    case st := <- cq.status:
        //fmt.Println("RET", st)
        if st == 1 {
            fmt.Fprint(w, "OK")
        } else {
            fmt.Fprint(w, "ER")
        }
        delete(kpcw, kpnos)
    }
   
    //fmt.Println("http.Method:", r.Method, "From:", r.RemoteAddr, "Len", r.ContentLength, string(body))

    return
}

func kpnhApiItem2(w http.ResponseWriter, r *http.Request) {
    
    fmt.Println("http.Method:", r.Method, "From:", r.RemoteAddr, "Len", r.ContentLength)

    fmt.Fprint(w, "Hello, world")
    
    //w.Body.Close()
}