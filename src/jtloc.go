package main

import (
    "./store"
    "./utils"
    "encoding/json"
    "fmt"
    "math/rand"
    "net"
    "regexp"
    "strconv"
    "strings"
    "time"
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

        kpsLed, err = stor.Get("ctl:led")
        if err == nil && kpsLed != "" {
            if addr, err := stor.Hget("ls:"+kpsLed, "addr"); err == nil {
                prbc.Send(msg, addr+":"+cfg.KeeperPort)
            }
        } else if rand.Intn(8) == 0 {
            prbc.Send(msg, bcip+":"+cfg.KeeperPort)
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
            if tid, _ := stor.Get("ctl:tid"); len(tid) == 0 {
                n, _ := stor.Incrby("ctl:ltid", 1)
                //kpnoi, _ := strconv.Atoi(kpno)
                kpnoi := len(kps)*n + kpsNum - 1

                // One Proposal alive in rand seconds
                stor.Setex("ctl:tid", 3, strconv.Itoa(kpnoi))
                msg = map[string]string{
                    "action":          "LedNew",
                    "node":            locNode,
                    "ProposalNumber":  strconv.Itoa(kpnoi),
                    "ProposalContent": locNode,
                }
                prbc.Send(msg, bcip+":"+cfg.KeeperPort)
                //fmt.Println(n, len(kps), kpno, n)
            }
        }

        // Leader Cast
        if kpsLed != "" && kpsLed == locNode {

            n, _ := stor.Incrby("ctl:ltid", 0)
            kpnoi, _ := strconv.Atoi(kpno)
            kpnoi = len(kps)*n + kpnoi - 1

            kpsm := []string{}

            //fmt.Println("kpslad", kpsLed, locNode, kpls)

            for k, v := range kpls {
                addr, _ := stor.Hget("ls:"+v, "addr")
                if err := stor.Exists("on:" + v); err == nil {
                    kpsm = append(kpsm, "1,"+k+","+v+",1,"+addr)
                } else {
                    kpsm = append(kpsm, "1,"+k+","+v+",0,"+addr)
                }
            }

            //fmt.Println(kpsm)

            if len(kpsm) > 0 {
                msg := map[string]string{
                    "action":      "LedCast",
                    "node":        locNode,
                    "ValueNumber": strconv.Itoa(kpnoi),
                    "kpls":        strings.Join(kpsm, ";"),
                }
                //fmt.Println(msg)
                prbc.Send(msg, bcip+":"+cfg.KeeperPort)
            }
        }

        //fmt.Println("JobTrackerLocal Checking")
        time.Sleep(2e9)
    }
}

func jobTrackerLocalRefresh() {

    loc, _ = stor.Hgetall("ctl:loc")

    // if new node then ID setting
    if _, ok := loc["node"]; !ok {
        loc["node"] = utils.NewRandString(10)
        stor.Hset("ctl:loc", "node", loc["node"])
    }
    locNode, _ = loc["node"]

    // Lesse time setting of Keeper's leader
    if _, ok := loc["tick"]; !ok {
        loc["tick"] = "2000"
        stor.Hset("ctl:loc", "tick", loc["tick"])
    }

    // Fetch local ip address
    addrs, _ := net.InterfaceAddrs()
    reg, _ := regexp.Compile(`^(.*)\.(.*)\.(.*)\.(.*)\/(.*)$`)
    for _, addr := range addrs {
        ips := reg.FindStringSubmatch(addr.String())
        if len(ips) != 6 || (ips[1] == "127" && ips[2] == "0") {
            continue
        }
        loc["addr"] = fmt.Sprintf("%s.%s.%s.%s", ips[1], ips[2], ips[3], ips[4])
        stor.Hset("ctl:loc", "addr", loc["addr"])
        stor.Hset("ctl:loc", "port", cfg.KeeperPort)
    }

    req["node"] = loc["node"]
    req["addr"] = loc["addr"]
    locNodeAddr = loc["addr"]

    kpls, _ = stor.Hgetall("ctl:members")
    for k, v := range kpls {

        kps[v] = k

        if addr, e := stor.Hget("ls:"+v, "addr"); e == nil {
            kp[v] = addr
        }
    }

    // TODO
    host := store.Host{
        Id:     locNode,
        Addr:   locNodeAddr,
        Status: 1,
    }
    if b, err := json.Marshal(host); err == nil {
        pl := &store.NodeProposal{
            Key: "/kpr/local",
            Val: string(b),
        }
        stor.NodeSet(pl)
    }

    //fmt.Println(kps)
}
