package main

import (
    "bytes"
    "fmt"
    "math/rand"
    "os/exec"
    "time"
    //"net"
    "encoding/json"
    "strconv"
    "strings"
)

var (
    ref         = 0
    loc         = map[string]string{}
    kps         = map[string]string{}
    kpls        = map[string]string{}
    req         = map[string]string{}
    locNode     = ""
    locNodeAddr = ""
    kpsNum      = 0
    kpsLed      = ""
)

type KprHost struct {
    Id string
    Ip string
    St string
}

func JobTrackerLocal() {

    var err error

    for {

        if (int(time.Now().Unix()) - ref) > 3 {
            jobTrackerLocalRefresh()
            ref = int(time.Now().Unix())
        }

        // broadcast self-info
        msg := map[string]string{
            "action": "NodeCast",
            "node":   locNode,
        }

        kpsLed, err = db.Get("ctl:led")
        if err == nil && kpsLed != "" {
            if addr, err := db.Hget("ls:"+kpsLed, "addr"); err == nil {
                peer.Send(msg, addr+":9628")
            }
        } else if rand.Intn(8) == 0 {
            peer.Send(msg, bcip+":9628")
        }

        // Paxos::P1a
        // try to become a leader
        kpno, _ := kps[locNode]
        if len(kpno) > 0 {
            kpsNum, _ = strconv.Atoi(kpno)
        }
        if len(kpsLed) == 0 && kpsNum > 0 {

            fmt.Println("try to become new leader")

            // Paxos::P2c
            if tid, _ := db.Get("ctl:tid"); len(tid) == 0 {
                n, _ := db.Incrby("ctl:ltid", 1)
                //kpnoi, _ := strconv.Atoi(kpno)
                kpnoi := len(kps)*n + kpsNum - 1

                // One Proposal alive in rand seconds
                db.Setex("ctl:tid", 3, strconv.Itoa(kpnoi))
                msg = map[string]string{
                    "action":          "LedNew",
                    "node":            locNode,
                    "ProposalNumber":  strconv.Itoa(kpnoi),
                    "ProposalContent": locNode,
                }
                peer.Send(msg, bcip+":9628")
                //fmt.Println(n, len(kps), kpno, n)
            }
        }

        // Leader Cast
        if kpsLed != "" && kpsLed == locNode {

            n, _ := db.Incrby("ctl:ltid", 0)
            kpnoi, _ := strconv.Atoi(kpno)
            kpnoi = len(kps)*n + kpnoi - 1

            kpsm := []string{}

            //fmt.Println("kpslad", kpsLed, locNode, kpls)

            for k, v := range kpls {
                addr, _ := db.Hget("ls:"+v, "addr")
                if err := db.Exists("on:" + v); err == nil {
                    kpsm = append(kpsm, "1,"+k+","+v+",1,"+addr)
                } else {
                    kpsm = append(kpsm, "1,"+k+","+v+",0,"+addr)
                }
            }

            //fmt.Println(kpsm)

            msg := map[string]string{
                "action":      "LedCast",
                "node":        locNode,
                "ValueNumber": strconv.Itoa(kpnoi),
                "kpls":        strings.Join(kpsm, ";"),
            }
            //fmt.Println(msg)
            peer.Send(msg, bcip+":9628")
        }

        //fmt.Println("JobTrackerLocal Checking")        
        time.Sleep(2e9)
    }
}

func jobTrackerLocalRefresh() {

    loc, _ = db.Hgetall("ctl:loc")

    // if new node then ID setting 
    if _, ok := loc["node"]; !ok {
        loc["node"] = NewRandString(10)
        db.Hset("ctl:loc", "node", loc["node"])
    }
    locNode, _ = loc["node"]

    // Lesse time setting of Keeper's leader
    if _, ok := loc["tick"]; !ok {
        loc["tick"] = "2000"
        db.Hset("ctl:loc", "tick", loc["tick"])
    }

    // Fetch local ip address
    var out bytes.Buffer
    cmd := "ip addr|grep inet|grep -v inet6|grep -v 127.0.|head -n1" +
        "|awk ' {print $2}'|awk -F \"/\" '{print $1}'"
    ec := exec.Command("sh", "-c", cmd)
    ec.Stdout = &out
    if err := ec.Run(); err == nil {
        loc["addr"] = strings.TrimSpace(out.String())
        db.Hset("ctl:loc", "addr", loc["addr"])
        db.Hset("ctl:loc", "port", "9628")
    }

    req["node"] = loc["node"]
    req["addr"] = loc["addr"]
    locNodeAddr = loc["addr"]

    kpls, _ = db.Hgetall("kps")
    for k, v := range kpls {

        kps[v] = k

        if addr, e := db.Hget("ls:"+v, "addr"); e == nil {
            kp[v] = addr
        }
    }

    // TODO
    host := KprHost{
        Id: locNode,
        Ip: locNodeAddr,
        St: "1",
    }
    if b, err := json.Marshal(host); err == nil {
        pl := &Proposal{
            Key: "/kpr/local",
            Val: string(b),
        }
        NodeSet(pl)
    }

    //fmt.Println(kps)
}

func jobTrackLocalNodeID(length int) string {

    chars := []byte("0123456789abcdefghijklmnopqrstuvwxyz")

    rs := make([]byte, length)

    rs[0] = chars[rand.Intn(len(chars)-10)+10]
    for i := 1; i < length; i++ {
        rs[i] = chars[rand.Intn(len(chars))]
    }

    //fmt.Println(jobTrackLocalNodeID, string(rs))
    return string(rs)
}
