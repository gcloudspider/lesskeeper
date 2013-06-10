package main

import (
    pr "./peer"
    "./store"
    "encoding/json"
    "fmt"
    "math/rand"
    "strconv"
    "strings"
    "time"
)

type ActionRequst map[string]interface{}

func CommandDispatchEvent(prbc *pr.NetUDP, p *pr.NetPacket) {

    var f interface{}
    err := json.Unmarshal(p.Body, &f)
    if err != nil {
        return
    }

    req := f.(map[string]interface{})
    action, ok := req["action"]
    if !ok {
        return
    }

    //fmt.Println("v2 dispatchEvent -> ", action.(string), "\n\t", req)

    ip := strings.Split(p.Addr, ":")[0]

    switch action.(string) {

    case "NodeCast":
        ActionNodeCast(req, ip)

    case "LedNew":
        ActionLedNew(req, ip)
    case "LedNewCb":
        ActionLedNewCb(req, ip)
    case "LedValue":
        ActionLedValue(req, ip)
    case "LedCast":
        ActionLedCast(req, ip)

    case "WatchEvent":
        ActionWatchEvent(req, ip)
    case "WatchLease":
        ActionWatchLease(req, ip)
    }
}

func ActionWatchLease(req ActionRequst, addr string) {

    if !req.isset("path") || !req.isset("host") || !req.isset("ttl") {
        return
    }

    path := strings.Trim(req["path"].(string), "/")
    ttlen, _ := strconv.Atoi(req["ttl"].(string))
    ttl := time.Now().Unix() + int64(ttlen)

    watcherlock.Lock()
    if w, ok := watches[path]; ok {
        if i, ok := w[req["host"].(string)]; ok {
            if ttl > i {
                w[req["host"].(string)] = ttl
            }
        }
    }
    watcherlock.Unlock()

    return
}

func ActionWatchEvent(req ActionRequst, addr string) {

    if !req.isset("path") || !req.isset("event") {
        return
    }

    //Println("WE", req)
    // ###TODO###
    //agent.watchmq <- &WatcherQueue{req["path"].(string), req["event"].(string), 0}
}

func ActionNodeCast(req ActionRequst, addr string) {

    node, ok := req["node"]
    if !ok {
        return
    }

    set := map[string]string{
        "addr":   addr,
        "status": "1",
    }
    stor.Hmset("ls:"+node.(string), set)
    stor.Setex("on:"+node.(string), 16, "1")

    // TODO
    host := KprHost{
        Id: node.(string),
        Ip: addr,
        St: "1",
    }
    if b, err := json.Marshal(host); err == nil {
        pl := &store.NodeProposal{
            Key: "/kpr/ls/" + node.(string),
            Val: string(b),
        }
        stor.NodeSet(pl)
    }
}

func ActionLedNew(req ActionRequst, addr string) {
    if !req.isset("node") || !req.isset("ProposalNumber") || !req.isset("ProposalContent") {
        return
    }
    if kpsNum == 0 || kpsLed != "" {
        fmt.Println(req)
        return
    }

    node, _ := req["node"]

    pnum, _ := req["ProposalNumber"]
    pnumi, _ := strconv.Atoi(pnum.(string))
    pval, _ := req["ProposalContent"]
    pvals := pval.(string)

    vnum, _ := stor.Get("ctl:voteid")
    vnumi, _ := strconv.Atoi(vnum)
    vval, _ := stor.Get("ctl:voteval")
    // ACCEPT!
    if vnumi == 0 || vval == "" || vnumi == pnumi || (vval == "" && vnumi <= pnumi) {
        vnumi = pnumi
        vval = pvals

        stor.Set("ctl:voteid", strconv.Itoa(vnumi))
        stor.Setex("ctl:voteval", 2, vval)
    }

    //
    rno := pnumi / len(kps)
    lno, _ := stor.Incrby("ctl:ltid", 0)
    //lnoi, _ := strconv.Atoi(lno)
    if lno < rno && node.(string) != locNode {
        stor.Incrby("ctl:ltid", (rno - lno))
    }

    msg := map[string]string{
        "action":        "LedNewCb",
        "node":          locNode,
        "kpno":          strconv.Itoa(kpsNum),
        "VerNew":        strconv.Itoa(vnumi),
        "AcceptContent": vval,
    }
    prbc.Send(msg, bcip+":"+cfg.KeeperPort)
}

func ActionLedNewCb(req ActionRequst, addr string) {

    if !req.isset("node") || !req.isset("kpno") || !req.isset("VerNew") || !req.isset("AcceptContent") {
        return
    }

    if kpsNum == 0 {
        return
    }

    node, _ := req["node"]
    anum, _ := req["VerNew"]
    anumi, _ := strconv.Atoi(anum.(string))

    aval, _ := req["AcceptContent"]

    lno, _ := stor.Incrby("ctl:ltid", 0)
    rno := anumi / len(kps)
    if lno < rno && node.(string) != locNode {
        stor.Incrby("ctl:ltid", (rno - lno))
    }

    tid, _ := stor.Get("ctl:tid")
    tidi, _ := strconv.Atoi(tid)
    if tidi == 0 {
        return
    }

    var prok string
    if tidi == anumi && locNode == aval.(string) {
        prok = "px:value:"
    } else if locNode != aval.(string) {
        stor.Expire("ctl:tid", rand.Intn(3)+1)
    } else {
        prok = "px:unvalue:"
    }

    prok = prok + tid + ":" + anum.(string) + ":" + addr
    stor.Setex(prok, 7, "1")

    fmt.Println("Checking if valued:", prok)

    vs2, _ := stor.Keys("ctl:*")
    fmt.Println(vs2)

    // Valued
    vs, _ := stor.Keys("px:value:" + tid + ":*")
    fmt.Println(vs)
    if 2*len(vs) > len(kps) {
        for _, v := range vs {
            ls := strings.Split(v, ":")
            msg := map[string]string{
                "node":         locNode,
                "action":       "LedValue",
                "ValueNumber":  ls[3],
                "ValueContent": locNode,
            }
            prbc.Send(msg, ls[4]+":"+cfg.KeeperPort)
            //fmt.Println("Value:", msg)
        }
        stor.Expire("ctl:tid", rand.Intn(3)+1)

        fmt.Println("Majory Valued")
        return
    }

    // UnValued
    vs, _ = stor.Keys("px:unvalue:" + tid + ":*")
    //fmt.Println(vs)
    if 2*len(vs) > len(kps) {
        // Prepare?
        lno, _ = stor.Incrby("ctl:ltid", 0)
        gno := len(kps)*lno + kpsNum - 1
        if gno > tidi {
            lno, _ = stor.Incrby("ctl:ltid", 1)
            gno = len(kps)*lno + kpsNum - 1
            stor.Setex("ctl:tid", rand.Intn(3)+1, strconv.Itoa(gno))

            msg := map[string]string{
                "action":          "LedNew",
                "node":            locNode,
                "ProposalNumber":  strconv.Itoa(gno),
                "ProposalContent": locNode,
            }
            prbc.Send(msg, bcip+":"+cfg.KeeperPort)
        } else {
            stor.Expire("ctl:tid", rand.Intn(3)+1)
        }

        fmt.Println("Majory UnValued")
    }
}

func ActionLedValue(req ActionRequst, addr string) {

    if !req.isset("node") || !req.isset("ValueNumber") || !req.isset("ValueContent") {
        return
    }

    if kpsNum == 0 {
        return
    }

    valnum, _ := req["ValueNumber"]
    valnumi, _ := strconv.Atoi(valnum.(string))

    valnode, _ := req["ValueContent"]
    kpss, ok := kps[valnode.(string)]
    if !ok || kpss == "" {
        return
    }

    anum, _ := stor.Get("ctl:voteid")
    anumi, _ := strconv.Atoi(anum)
    aval, _ := stor.Get("ctl:voteval")
    if anumi == 0 {
        return
    }

    if anumi == valnumi && valnode.(string) == aval {
        stor.Setex("ctl:led", 12, aval)
    }

    fmt.Println("Value OK", anum, aval)
}

func ActionLedCast(req ActionRequst, addr string) {

    if !req.isset("node") || !req.isset("kpls") || !req.isset("ValueNumber") {
        return
    }

    node, ok := req["node"]
    if !ok {
        return
    }

    if kpsLed != "" && kpsLed != node.(string) {
        stor.Del("ctl:led")
        return
    }

    stor.Setex("ctl:led", 12, node.(string))

    if node.(string) == locNode {
        // TODO
        //return
    }

    ///
    ltid, _ := stor.Incrby("ctl:ltid", 0)
    vnum, _ := req["ValueNumber"]
    vnumi, _ := strconv.Atoi(vnum.(string))
    rtid := vnumi / len(kpls)
    if ltid < rtid && node.(string) != locNode {
        stor.Incrby("ctl:ltid", rtid-ltid)
    }

    //
    kplsn := map[string]string{}

    rs, _ := req["kpls"]
    sp := strings.Split(rs.(string), ";")
    for _, v := range sp {
        rs2 := strings.Split(v, ",")
        kplsn[rs2[1]] = v
    }

    for k, v := range kpls {
        if str, ok := kplsn[k]; !ok {
            stor.Hdel("kps", k)
        } else {
            sp = strings.Split(str, ",")
            if sp[3] == "1" {
                stor.Setex("on:"+v, 16, "1")
            }
            stor.Hset("ls:"+v, "addr", sp[4])
            delete(kplsn, k)
        }
    }

    for k, v := range kplsn {
        sp = strings.Split(v, ",")
        if sp[3] == "1" {
            stor.Setex("on:"+sp[2], 16, "1")
        }
        stor.Hset("ls:"+sp[2], "addr", sp[4])
        stor.Hset("kps", k, sp[2])
    }
}

func (req ActionRequst) isset(key string) bool {
    if _, ok := req[key]; ok {
        return true
    }
    return false
}
