package main 

import (
    "net/http"
    //"net/url"
    "time"
    "fmt"
    //"io"
    "bytes"
    //"os"
    //"net/http/httputil"
)


func main() {

    for {

        time.Sleep(1e9)

        body := bytes.NewBufferString("put tests\r\nValueAAAA")
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
