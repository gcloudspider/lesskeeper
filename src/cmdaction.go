
package main

import (
    "fmt"
    "strconv"
    "math/rand"
    "strings"
    "encoding/json"
)

type ActionRequst map[string]interface{}

func CommandDispatchEvent(peer *NetUDP, p *NetPacket) {

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
        ActionWatchEvnet(req, ip)
    }
}

func ActionWatchEvnet(req ActionRequst, addr string) {
    
    if !req.isset("path") {
        return
    }

    Println("WE", req)
    agent.watchmq <- req["path"].(string)

    /*if tag, ok := req["Tag"]; ok {
        if rs, ok := req["status"]; ok {
            if status, err := strconv.Atoi(rs.(string)); err == nil {
                agent.Lock.Lock()
                if c, ok := agent.clients[tag.(string)]; ok {
                    c.Sig <- status
                }
                //fmt.Println("ActionAgentItemPutCb", status)
                agent.Lock.Unlock()
            }
        }
    }*/
}

func ActionNodeCast(req ActionRequst, addr string) {
    
    node, ok := req["node"]
    if !ok {
        return
    }

    set := map[string]string{
        "addr": addr,
        "status": "1",
    }
    db.Hmset("ls:"+ node.(string), set)
    
    db.Setex("on:"+ node.(string), 16, "1")
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

    vnum, _ := db.Get("ctl:voteid")
    vnumi, _ := strconv.Atoi(vnum)
    vval, _ := db.Get("ctl:voteval")
    // ACCEPT!
    if vnumi == 0 || vval == "" || vnumi == pnumi || (vval == "" && vnumi <= pnumi) {
        vnumi = pnumi
        vval = pvals

        db.Set("ctl:voteid", strconv.Itoa(vnumi))
        db.Setex("ctl:voteval", 2, vval)
    }

    //
    rno := pnumi / len(kps)
    lno, _ := db.Incrby("ctl:ltid", 0)
    //lnoi, _ := strconv.Atoi(lno)
    if lno < rno && node.(string) != locNode {
        db.Incrby("ctl:ltid", (rno - lno))
    }

    msg := map[string]string{
        "action": "LedNewCb",
        "node": locNode,
        "kpno": strconv.Itoa(kpsNum),
        "VerNew": strconv.Itoa(vnumi),
        "AcceptContent": vval,
    }
    peer.Send(msg, bcip +":"+ port)
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

    lno, _ := db.Incrby("ctl:ltid", 0)
    rno := anumi / len(kps)
    if lno < rno && node.(string) != locNode {
        db.Incrby("ctl:ltid", (rno - lno))
    }

    tid, _ := db.Get("ctl:tid")
    tidi, _ := strconv.Atoi(tid)
    if tidi == 0 {
        return
    }

    var prok string
    if tidi == anumi && locNode == aval.(string) {
        prok = "px:value:"
    } else if locNode != aval.(string) {
        db.Expire("ctl:tid", rand.Intn(3) + 1)
    } else {
        prok = "px:unvalue:"
    }

    prok = prok + tid +":"+ anum.(string) +":"+ addr
    db.Setex(prok, 7, "1")

    fmt.Println("Checking if valued:", prok)

    vs2, _ := db.Keys("ctl:*")
    fmt.Println(vs2)

    // Valued
    vs, _ := db.Keys("px:value:"+ tid +":*")
    fmt.Println(vs)
    if 2 * len(vs) > len(kps) {
        for _, v := range vs {
            ls := strings.Split(v, ":")
            msg := map[string]string{
                "node": locNode,
                "action": "LedValue",
                "ValueNumber": ls[3],
                "ValueContent": locNode,
            }
            peer.Send(msg, ls[4] +":"+ port)
            //fmt.Println("Value:", msg)
        }
        db.Expire("ctl:tid", rand.Intn(3) + 1)
        
        fmt.Println("Majory Valued")
        return;
    }

    // UnValued
    vs, _ = db.Keys("px:unvalue:"+ tid +":*")
    //fmt.Println(vs)
    if 2 * len(vs) > len(kps) {
        // Prepare?
        lno, _ = db.Incrby("ctl:ltid", 0)
        gno := len(kps) * lno + kpsNum - 1
        if gno > tidi {
            lno, _ = db.Incrby("ctl:ltid", 1)
            gno = len(kps) * lno + kpsNum - 1
            db.Setex("ctl:tid", rand.Intn(3) + 1, strconv.Itoa(gno))

            msg := map[string]string{
                "action": "LedNew",
                "node": locNode,
                "ProposalNumber": strconv.Itoa(gno),
                "ProposalContent": locNode,
            }
            peer.Send(msg, bcip +":"+ port)
        } else {
            db.Expire("ctl:tid", rand.Intn(3) + 1)
        }

        fmt.Println("Majory UnValued")
    }
}

func ActionLedValue(req ActionRequst, addr string) {
    
    if !req.isset("node") || !req.isset("ValueNumber") || !req.isset("ValueContent") {
        return
    }

    if kpsNum == 0 {
        return;
    }

    valnum, _ := req["ValueNumber"]
    valnumi, _ := strconv.Atoi(valnum.(string))

    valnode, _ := req["ValueContent"]
    kpss, ok := kps[valnode.(string)]
    if !ok || kpss == "" {
        return
    }

    anum, _ := db.Get("ctl:voteid")
    anumi, _ := strconv.Atoi(anum)
    aval, _ := db.Get("ctl:voteval")
    if anumi == 0 {
        return 
    }

    if anumi == valnumi && valnode.(string) == aval {
        db.Setex("ctl:led", 12, aval)
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
        db.Del("ctl:led")
        return
    }

    db.Setex("ctl:led", 12, node.(string))

    if node.(string) == locNode {
        // TODO
        //return
    }

    ///
    ltid, _ := db.Incrby("ctl:ltid", 0)
    vnum, _ := req["ValueNumber"]
    vnumi, _ := strconv.Atoi(vnum.(string))
    rtid := vnumi / len(kpls)
    if ltid < rtid && node.(string) != locNode {
        db.Incrby("ctl:ltid", rtid - ltid)
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
            db.Hdel("kps", k)
        } else {
            sp = strings.Split(str, ",")
            if sp[3] == "1" {
                db.Setex("on:"+ v, 16, "1")
            }
            db.Hset("ls:"+ v, "addr", sp[4])
            delete(kplsn, k)
        }
    }

    for k, v := range kplsn {
        sp = strings.Split(v, ",")
        if sp[3] == "1" {
            db.Setex("on:"+ sp[2], 16, "1")
        }
        db.Hset("ls:"+ sp[2], "addr", sp[4])
        db.Hset("kps", k, sp[2])
    }
}

/*
func ActionItemPut(req ActionRequst, addr string) {
    
    if !req.isset("node") || !req.isset("ItemKey") || !req.isset("ItemContent") || !req.isset("VerNew") {
        return
    }

    if locNode == "" || kpsNum == 0 || kpsLed == "" {
        return
    }

    node, _ := req["node"]
    key, _ := req["ItemKey"]

    vernew, _ := req["VerNew"]
    msg := map[string]string{
        "node": locNode,
        "action":   "ItemPutCb",
        "VerNew":   vernew.(string),
        //"ctlid":    locNode,
    }

    valnew, _ := req["ItemContent"]
    it := map[string]string{
        "n": vernew.(string),
        "v": valnew.(string),
    }

    db.Hmset("c:def:"+ key.(string), it) // TODO, waiting valued

    // Ensure ctl:loctid to Max
    vernewi, _ := strconv.Atoi(vernew.(string));
    nnum := vernewi / len(kpls)
    ltid, _ := db.Incrby("ctl:ltid", 0)
    if ltid < nnum && locNode != node.(string) {
        db.Incrby("ctl:ltid", nnum - ltid)
    }

    //fmt.Println("ActionItemPut", msg)
    peer.Send(msg, addr +":"+ port)
}

func ActionItemPutCb(req ActionRequst, addr string) {

    if !req.isset("node") || !req.isset("VerNew") {
        return
    }

    if locNode == "" || kpsNum == 0 {
        return
    }

    node, _ := req["node"]
    vernew, _ := req["VerNew"]
    
    anumi, _ := strconv.Atoi(vernew.(string))
    ltid, _ := db.Incrby("ctl:ltid", 0)
    rnum := anumi / len(kpls)
    if ltid < rnum && node.(string) != locNode {
        db.Incrby("ctl:ltid", rnum - ltid)
    }

    key  := ""
    tag  := ""
    ipcb := ""
    it := map[string]string{}

    proposal_servlock.Lock()
    if p, ok := proposals[vernew.(string)]; ok {
        
        proposals[vernew.(string)].Yes++

        if 2 * p.Yes > len(kpls) {
            
            key = p.Key
            tag = p.Tag
            ipcb = p.Addr
            
            it = map[string]string{
                "n": vernew.(string),
                "v": p.Val,
            }
        }
    }
    proposal_servlock.Unlock()

    
    // SUCCESS, Callback status
    if key != "" {
        //fmt.Println("OK")
        _ = db.Hmset("c:def:"+ key, it)

        msg := map[string]string{
            "action": "AgentItemPutCb",
            "Tag": tag,
            "status": "10",// TODO strconv.Itoa(ReplyOK),
        }
        //fmt.Println("ActionItemPutCb", msg)
        peer.Send(msg,  ipcb+":"+ port)
    }
}

func ActionAgentItemPutCb(req ActionRequst, addr string) { 

    if tag, ok := req["Tag"]; ok {
        if rs, ok := req["status"]; ok {
            if status, err := strconv.Atoi(rs.(string)); err == nil {
                agent.Lock.Lock()
                if c, ok := agent.clients[tag.(string)]; ok {
                    c.Sig <- status
                }
                //fmt.Println("ActionAgentItemPutCb", status)
                agent.Lock.Unlock()
            }
        }
    }
}
*/

func (req ActionRequst) isset(key string) bool {    
    if _, ok := req[key]; ok {
        return true
    }
    return false
}