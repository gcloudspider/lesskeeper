package main

import (
    "net/http"
    //"net/url"
    "fmt"
    "time"
    //"io"
    "bytes"
    "io/ioutil"
    "math/rand"
    "os"
    "runtime"
    "strconv"
    "sync"
)

var threadnum = 100
var threadflag = 0
var l sync.Mutex
var uriItem = "http://localhost:9529/lesskeeper/api/item"
var maxnum = 10000

var counter = 0
var counter_error = 0

func main() {

    runtime.GOMAXPROCS(runtime.NumCPU())

    rand.Seed(time.Now().UnixNano())

    go func() {
        runtime.GC()
        time.Sleep(1e9)
    }()

    for i := 0; i < 10000000; i++ {
        go func() {
            time.Sleep(1e9)
        }()
        if i%10000 == 0 {
            fmt.Println("go", i)
        }
    }
    os.Exit(0)

    for i := 0; i < threadnum; i++ {
        go ItemBench2(i, 1000)
    }

    //os.Exit(0)

    //time.Sleep(30e9)

    //time.Sleep(300 * time.Second)

    //
    /*
       client := &http.Client{}
       for i := 0; i < maxnum; i++ {

           body := bytes.NewBufferString("get key1\r\n")
           req, err := http.NewRequest("PUT", "http://localhost:9529/lesskeeper/api/item", body)
           if err != nil {
               fmt.Println("Error", err)
           }
           //req.Body = ioutil.NopCloser(bytes.NewBufferString("get key1\r\n"))

           //req.Header.Add("User-Agent", "Our Custom User-Agent")
           //req.Header.Add("If-None-Match", `W/"TheFileEtag"`)
           //req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
           resp, err2 := client.Do(req)
           if err2 != nil {
               fmt.Println("Error", err2)
           }

           if ret, err := ioutil.ReadAll(resp.Body); err == nil {
               if true {
                   fmt.Println(string(ret))
               }
           }

           resp.Body.Close()
       }
       os.Exit(0)
    */

    for {

        time.Sleep(1e9)
        continue

        vi := rand.Intn(99999999)
        vs := strconv.Itoa(vi)
        //body := bytes.NewBufferString("put key1\r\nValue"+ vs)
        body := bytes.NewBufferString("get key1\r\n")
        //
        resp, e := http.Post(uriItem, "application/x-www-form-urlencoded", body)
        if e != nil {
            fmt.Println("Error", e)
            //os.Exit(1)
            continue
        } else {
            respbody, _ := ioutil.ReadAll(resp.Body)
            if string(respbody) != "OK" {
                fmt.Println("Error2", string(respbody))
            }
            //fmt.Println("OK", string(respbody))
        }
        resp.Body.Close()

        if false {
            fmt.Println(body, vs)
        }
        //_, _ = http.PostForm("http://localhost:9529/lesskeeper/api/item", url.Values{"title": {"AAAAAAAAA"}, "content": {"BBBBB"}})

        //time.Sleep(1e6)
    }

    //os.Exit(0)
}

func ItemBench2(t, n int) {

    ts := strconv.Itoa(t)

    for i := 0; i < n; i++ {

        is := strconv.Itoa(i)

        body := bytes.NewBufferString("put " + ts + "." + is + "\r\nValue" + is)
        resp, err := http.Post(uriItem, "text/plain", body)
        if err != nil {
            //fmt.Println("Error", err)
            l.Lock()
            counter_error++
            l.Unlock()
            continue
        }

        ret, err3 := ioutil.ReadAll(resp.Body)
        if err3 != nil {
            //fmt.Println("Error3", err3)
            l.Lock()
            counter_error++
            l.Unlock()
            continue
        }

        if string(ret) != "OK" {
            //fmt.Println("Error body", string(ret))
            l.Lock()
            counter_error++
            l.Unlock()
        } else {
            l.Lock()
            counter++
            l.Unlock()
        }

        resp.Body.Close()

        if counter%1000 == 0 {
            fmt.Println("OK", counter, "ERROR", counter_error)
        }
    }

    l.Lock()
    threadflag++
    l.Unlock()

    fmt.Println(threadflag)

    if threadflag == threadnum {
        fmt.Println("OK", counter, "ERROR", counter_error)
        os.Exit(0)
    }
}

func ItemBench(n int) {
    //fmt.Println("OK")

    tr := &http.Transport{
        DisableKeepAlives: true,
        //MaxIdleConnsPerHost: 100,
    }
    client := &http.Client{Transport: tr}

    for i := 0; i < n; i++ {

        vi := rand.Intn(99999999)
        vs := strconv.Itoa(vi)

        body := bytes.NewBufferString("put " + vs + "\r\nValue" + vs)
        //body := bytes.NewBufferString("get key1\r\n")

        req, err := http.NewRequest("PUT", uriItem, body)
        if err != nil {
            fmt.Println("Error1", err)
            continue
        }
        // Disable keep-alive
        req.Close = true

        resp, err2 := client.Do(req)
        if err2 != nil {
            fmt.Println("Error2", err2)
            continue
        }

        /* ret, err3 := ioutil.ReadAll(resp.Body)
           if err != nil {
               fmt.Println("Error3", err3)
               continue
           }

           if false {
               fmt.Println(string(ret))
           }*/

        resp.Body.Close()

        if false {
            fmt.Println(vs)
        }
    }

    l.Lock()
    threadflag++
    l.Unlock()

    fmt.Println(threadflag)

    if threadflag == threadnum {
        os.Exit(0)
    }
}

func ItemSet(key, val string) error {

    return nil
}
