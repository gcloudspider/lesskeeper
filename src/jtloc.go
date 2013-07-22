package main

import (
    "../deps/lessgo/utils"
    "./store"
    "fmt"
    "math/rand"
    "net"
    "regexp"
    "strconv"
    "strings"
    "time"
)

type KeeperHost struct {
    store.Host
    KprNum int
}

var (
    kprLed = ""
    kprGrp = map[string]KeeperHost{}
    kprSef = KeeperHost{}
)

func JobTrackerLocal() {

    var err error
    ref := 0

    for {

        if (int(time.Now().Unix()) - ref) > 3 {
            jobTrackerLocalRefresh()
            ref = int(time.Now().Unix())
        }

        // broadcast self-info
        msg := map[string]string{
            "action": "NodeCast",
            "node":   kprSef.Id,
            "port":   kprSef.Port,
        }

        kprLed, err = stor.Get("ctl:led")
        if err == nil && kprLed != "" {
            if addr, err := stor.Hget("ls:"+kprLed, "addr"); err == nil {
                prbc.Send(msg, addr+":"+cfg.KeeperPort)
            }
        } else if rand.Intn(8) == 0 {
            prbc.Send(msg, bcip+":"+cfg.KeeperPort)
        }

        // Paxos::P1a
        // try to become a leader
        if len(kprLed) == 0 && kprSef.KprNum > 0 {

            fmt.Println("try to become new leader")

            // Paxos::P2c
            if tid, _ := stor.Get("ctl:tid"); len(tid) == 0 {
                n, _ := stor.Incrby("ctl:ltid", 1)
                kpnoi := len(kprGrp)*n + kprSef.KprNum - 1

                // One Proposal alive in rand seconds
                stor.Setex("ctl:tid", 3, strconv.Itoa(kpnoi))
                msg = map[string]string{
                    "action":          "LedNew",
                    "node":            kprSef.Id,
                    "ProposalNumber":  strconv.Itoa(kpnoi),
                    "ProposalContent": kprSef.Id,
                }
                prbc.Send(msg, bcip+":"+cfg.KeeperPort)
                //fmt.Println(n, len(kps), kpno, n)
            }
        }

        // Leader Cast
        if kprLed != "" && kprLed == kprSef.Id {

            n, _ := stor.Incrby("ctl:ltid", 0)
            kpnoi := len(kprGrp)*n + kprSef.KprNum - 1
            kpsm := []string{}

            //fmt.Println("kpslad", kprLed, kprSef.Id, kpls)

            for _, v := range kprGrp {
                //addr, _ := stor.Hget("ls:"+v, "addr")
                if err := stor.Exists("on:" + v.Id); err == nil {
                    kpsm = append(kpsm, "1,"+strconv.Itoa(v.KprNum)+","+v.Id+",1,"+v.Addr)
                } else {
                    kpsm = append(kpsm, "1,"+strconv.Itoa(v.KprNum)+","+v.Id+",0,"+v.Addr)
                }
            }

            //fmt.Println(kpsm)

            if len(kpsm) > 0 {
                msg := map[string]string{
                    "action":      "LedCast",
                    "node":        kprSef.Id,
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

    loc, _ := stor.Hgetall("ctl:loc")

    // if new node then ID setting
    if id, ok := loc["node"]; !ok {
        kprSef.Id = utils.StringNewRand(10)
        stor.Hset("ctl:loc", "node", kprSef.Id)
    } else {
        kprSef.Id = id
    }

    // Lesse time setting of Keeper's leader
    /** TODO if _, ok := loc["tick"]; !ok {
        loc["tick"] = "2000"
        stor.Hset("ctl:loc", "tick", loc["tick"])
    } */

    // Fetch local ip address
    addrs, _ := net.InterfaceAddrs()
    reg, _ := regexp.Compile(`^(.*)\.(.*)\.(.*)\.(.*)\/(.*)$`)
    for _, addr := range addrs {
        ips := reg.FindStringSubmatch(addr.String())
        if len(ips) != 6 || (ips[1] == "127" && ips[2] == "0") {
            continue
        }
        kprSef.Addr = fmt.Sprintf("%s.%s.%s.%s", ips[1], ips[2], ips[3], ips[4])

        stor.Hset("ctl:loc", "addr", kprSef.Addr)
        stor.Hset("ctl:loc", "port", cfg.KeeperPort)
    }
    kprSef.Port = cfg.KeeperPort

    ms, _ := stor.Hgetall("ctl:members")
    for k, _ := range kprGrp {

        for _, v2 := range ms {

            if k == v2 {
                break
            }
        }

        delete(kprGrp, k)
    }

    for k, v := range ms {

        if _, ok := kprGrp[v]; !ok {
            rs, _ := stor.NodeGet("/kpr/ls/" + v)
            var node KeeperHost
            err := utils.JsonDecode(rs.C, &node)
            if err == nil {
                node.KprNum, _ = strconv.Atoi(k)
                kprGrp[v] = node
            }
        }

        if v2, ok := kprGrp[kprSef.Id]; ok && kprSef.KprNum != v2.KprNum {
            kprSef.KprNum = v2.KprNum
        }
    }

    // TODO
    host := store.Host{
        Id:     kprSef.Id,
        Addr:   kprSef.Addr,
        Port:   cfg.KeeperPort,
        Status: 1,
    }
    if b, err := utils.JsonEncode(host); err == nil {
        pl := &store.NodeProposal{
            Key: "/kpr/local",
            Val: b,
        }
        stor.NodeSet(pl)
    }

    // Every Keeper Instance, Default as Single-Node Type
    if len(kprGrp) == 0 {

        stor.Hset("ctl:members", "1", kprSef.Id)
    }
}
