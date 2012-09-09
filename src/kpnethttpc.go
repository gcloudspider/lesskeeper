package main 

import (
    "net/http"
    //"net/url"
    "time"
    "fmt"
    //"io"
    "bytes"
    "strconv"
    "math/rand"
    "os"
)


func main() {

    rand.Seed(time.Now().UnixNano())

    for i := 0; i < 1; i++ {
        ItemBench(10000)
    }
    os.Exit(0)

    //time.Sleep(300 * time.Second)

    //

     for {

        time.Sleep(1e9)

        vi := rand.Intn(99999999)
        vs := strconv.Itoa(vi)
        body := bytes.NewBufferString("put key1\r\nValue"+ vs)
        //
        rs, e := http.Post("http://localhost:9529/h5keeper/api/item", "text/html", body)
        if e != nil {
            fmt.Println("Error", e)
            //os.Exit(1)
            continue
        } else {
            fmt.Println("OK")
        }
        rs.Body.Close()

        if false {
            fmt.Println(body)
        }
        //_, _ = http.PostForm("http://localhost:9529/h5keeper/api/item", url.Values{"title": {"AAAAAAAAA"}, "content": {"BBBBB"}})
 
        //time.Sleep(1e6)
    }
}

func ItemBench(n int) {
    //fmt.Println("OK")
    for i := 0; i < n; i++ {
        vi := rand.Intn(99999999)
        vs := strconv.Itoa(vi)
        body := bytes.NewBufferString("put "+ vs +"\r\nValue"+ vs)
        
        rs, e := http.Post("http://localhost:9529/h5keeper/api/item", "text/html", body)
        if e != nil {
            fmt.Println("Error", e)
        }

        rs.Body.Close()
    }

}

func ItemSet(key, val string) error {

    return nil
}
