
package main

import (
    "fmt"
    "time"
    "math/rand"
    "os/exec"
    "bytes"
    //"net"
    "strings"
    "strconv"
)

var (
    ref     = 0
    loc     = map[string]string{}
    kps     = map[string]string{}
    kpls    = map[string]string{}
    req     = map[string]string{}
    locNode = ""
    kpsNum  = 0
    kpsLed  = ""
)

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
            "node": locNode,
        }
        
        kpsLed, err = kpd.Get("ct:led")
        if err == nil && kpsLed != "" {
            if addr, err := kpd.Hget("ls:"+ kpsLed, "addr"); err != nil {
                kpn.Send(msg, addr +":9528")
            }
        } else if rand.Intn(8) == 0 {
            kpn.Send(msg, "255.255.255.255:9528")
        }

        // Paxos::P1a
        // try to become a leader
        kpno, _ := kps[locNode];
        if len(kpno) > 0 {
            kpsNum, _ = strconv.Atoi(kpno)
        }
        if len(kpsLed) == 0 && kpsNum > 0 {
            
            fmt.Println("try to become new leader")
            
            // Paxos::P2c
            if tid, _ := kpd.Get("ct:tid"); len(tid) == 0 {
                n, _ := kpd.Incrby("ct:ltid", 1)
                //kpnoi, _ := strconv.Atoi(kpno)
                kpnoi := len(kps) * n + kpsNum - 1

                // One Proposal alive in rand seconds
                kpd.Setex("ct:tid", 3, strconv.Itoa(kpnoi))
                msg = map[string]string{
                    "action": "LedNew",
                    "node": locNode,
                    "ProposalNumber": strconv.Itoa(kpnoi),
                    "ProposalContent": locNode,
                }
                kpn.Send(msg, "255.255.255.255:9528")
                //fmt.Println(n, len(kps), kpno, n)
            }
        }

        // Leader Cast
        if kpsLed != "" && kpsLed == locNode {

            n , _ := kpd.Incrby("ct:ltid", 0)
            kpnoi, _ := strconv.Atoi(kpno)
            kpnoi = len(kps) * n + kpnoi - 1
            
            kpsm := []string{}
            
            //fmt.Println("kpslad", kpsLed, locNode, kpls)

            for k, v := range kpls {
                addr, _ := kpd.Hget("ls:"+ v, "addr")
                if err := kpd.Exists("on:"+ v); err == nil {
                    kpsm = append(kpsm, "1,"+ k +","+ v +",1,"+ addr)
                } else {
                    kpsm = append(kpsm, "1,"+ k +","+ v +",0,"+ addr)
                }
            }

            //fmt.Println(kpsm)

            msg := map[string]string{
                "action": "LedCast",
                "node": locNode,
                "ValueNumber": strconv.Itoa(kpnoi),
                "kpls": strings.Join(kpsm, ";"),
            }
            //fmt.Println(msg)
            kpn.Send(msg, "255.255.255.255:9528")
        }

        //fmt.Println("JobTrackerLocal Checking")        
        time.Sleep(2e9)
    }
}

func jobTrackerLocalRefresh() {

    loc, _ = kpd.Hgetall("ct:loc")

    // if new node then ID setting 
    if _, ok := loc["node"]; !ok {
        loc["node"] = jobTrackLocalNodeID(8)
        kpd.Hset("ct:loc", "node", loc["node"])
    }
    locNode, _ = loc["node"]

    // Lesse time setting of Keeper's leader
    if _, ok := loc["tick"]; !ok {
        loc["tick"] = "2000"
        kpd.Hset("ct:loc", "tick", loc["tick"])
    }

    // Fetch local ip address
    var out bytes.Buffer
    cmd := "ip addr|grep inet|grep -v inet6|grep -v 127.0.|head -n1" +
        "|awk ' {print $2}'|awk -F \"/\" '{print $1}'"
    ec := exec.Command("sh", "-c", cmd)
    ec.Stdout = &out
    if err := ec.Run(); err == nil {
        loc["addr"] = strings.TrimSpace(out.String())
        kpd.Hset("ct:loc", "addr", loc["addr"])
        kpd.Hset("ct:loc", "port", "9528")
    }

    req["node"] = loc["node"]
    req["addr"] = loc["addr"]

    kpls, _ = kpd.Hgetall("kps")
    for k, v := range kpls {
        kps[v] = k
    }

    //fmt.Println(kps)
}

func jobTrackLocalNodeID(length int) string {

    chars := []byte("0123456789abcdefghijklmnopqrstuvwxyz")

    rs := make([]byte, length)

    rs[0] = chars[rand.Intn(len(chars) - 10) + 10]
    for i := 1; i < length; i++ { 
        rs[i] = chars[rand.Intn(len(chars))]
    }

    //fmt.Println(jobTrackLocalNodeID, string(rs))
    return string(rs)
}