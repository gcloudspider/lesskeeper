
package main

import (
    "fmt"
    "strconv"
    "math/rand"
    "strings"
)

type ActionRequst map[string]interface{}

func checkCommonParams(req ActionRequst) bool {
    if _, ok := req["node"]; !ok {
        return false
    }
    return true
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
    kpd.Hmset("ls:"+ node.(string), set)
    
    kpd.Setex("on:"+ node.(string), 16, "1")
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

    vnum, _ := kpd.Get("ct:voteid")
    vnumi, _ := strconv.Atoi(vnum)
    vval, _ := kpd.Get("ct:voteval")
    // ACCEPT!
    if vnumi == 0 || vval == "" || vnumi == pnumi || (vval == "" && vnumi <= pnumi) {
        vnumi = pnumi
        vval = pvals

        kpd.Set("ct:voteid", strconv.Itoa(vnumi))
        kpd.Setex("ct:voteval", 2, vval)
    }

    //
    rno := pnumi / len(kps)
    lno, _ := kpd.Incrby("ct:ltid", 0)
    //lnoi, _ := strconv.Atoi(lno)
    if lno < rno && node.(string) != locNode {
        kpd.Incrby("ct:ltid", (rno - lno))
    }

    msg := map[string]string{
        "action": "LedNewCb",
        "node": locNode,
        "kpno": strconv.Itoa(kpsNum),
        "AcceptNumber": strconv.Itoa(vnumi),
        "AcceptContent": vval,
    }
    kpn.Send(msg, "255.255.255.255:9528")    
}

func ActionLedNewCb(req ActionRequst, addr string) {
    
    if !req.isset("node") || !req.isset("kpno") || !req.isset("AcceptNumber") || !req.isset("AcceptContent") {
        return
    }

    if kpsNum == 0 {
        return
    }

    node, _ := req["node"]
    anum, _ := req["AcceptNumber"]
    anumi, _ := strconv.Atoi(anum.(string))

    aval, _ := req["AcceptContent"]

    lno, _ := kpd.Incrby("ct:ltid", 0)
    rno := anumi / len(kps)
    if lno < rno && node.(string) != locNode {
        kpd.Incrby("ct:ltid", (rno - lno))
    }

    tid, _ := kpd.Get("ct:tid")
    tidi, _ := strconv.Atoi(tid)
    if tidi == 0 {
        return
    }

    var prok string
    if tidi == anumi && locNode == aval.(string) {
        prok = "px:value:"
    } else if locNode != aval.(string) {
        kpd.Expire("ct:tid", rand.Intn(3) + 1)
    } else {
        prok = "px:unvalue:"
    }

    prok = prok + tid +":"+ anum.(string) +":"+ addr
    kpd.Setex(prok, 7, "1")

    fmt.Println("Checking if valued:", prok)

    vs2, _ := kpd.Keys("ct:*")
    fmt.Println(vs2)

    // Valued
    vs, _ := kpd.Keys("px:value:"+ tid +":*")
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
            kpn.Send(msg, ls[4] +":9528")
            //fmt.Println("Value:", k, v, ls)
        }
        kpd.Expire("ct:tid", rand.Intn(3) + 1)
        
        fmt.Println("Majory Valued")
        return;
    }

    // UnValued
    vs, _ = kpd.Keys("px:unvalue:"+ tid +":*")
    //fmt.Println(vs)
    if 2 * len(vs) > len(kps) {
        // Prepare?
        lno, _ = kpd.Incrby("ct:ltid", 0)
        gno := len(kps) * lno + kpsNum - 1
        if gno > tidi {
            lno, _ = kpd.Incrby("ct:ltid", 1)
            gno = len(kps) * lno + kpsNum - 1
            kpd.Setex("ct:tid", rand.Intn(3) + 1, strconv.Itoa(gno))

            msg := map[string]string{
                "action": "LedNew",
                "node": locNode,
                "ProposalNumber": strconv.Itoa(gno),
                "ProposalContent": locNode,
            }
            kpn.Send(msg, "255.255.255.255:9528")
        } else {
            kpd.Expire("ct:tid", rand.Intn(3) + 1)
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

    anum, _ := kpd.Get("ct:voteid")
    anumi, _ := strconv.Atoi(anum)
    aval, _ := kpd.Get("ct:voteval")
    if anumi == 0 {
        return 
    }

    if anumi == valnumi && valnode.(string) == aval {
        kpd.Setex("ct:led", 12, aval)
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
        kpd.Del("ct:led")
        return
    }

    kpd.Setex("ct:led", 12, node.(string))

    if node.(string) == locNode {
        // TODO
        //return
    }

    ///
    ltid, _ := kpd.Incrby("ct:ltid", 0)
    vnum, _ := req["ValueNumber"]
    vnumi, _ := strconv.Atoi(vnum.(string))
    rtid := vnumi / len(kpls)
    if ltid < rtid && node.(string) != locNode {
        kpd.Incrby("ct:ltid", rtid - ltid)
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
            kpd.Hdel("kps", k)
        } else {
            sp = strings.Split(str, ",")
            if sp[3] == "1" {
                kpd.Setex("on:"+ v, 16, "1")
            }
            kpd.Hset("ls:"+ v, "addr", sp[4])
            delete(kplsn, k)
        }
    }

    for k, v := range kplsn {
        sp = strings.Split(v, ",")
        if sp[3] == "1" {
            kpd.Setex("on:"+ sp[2], 16, "1")
        }
        kpd.Hset("ls:"+ sp[2], "addr", sp[4])
        kpd.Hset("kps", k, sp[2])
    }
}

func ActionItemPut(req ActionRequst, addr string) {
    if !req.isset("node") || !req.isset("ItemNumber") || !req.isset("ItemKey") || !req.isset("ItemContent") || !req.isset("ItemNumberNext") {
        return
    }

    if locNode == "" || kpsNum == 0 || kpsLed == "" {
        return
    }

    node, _ := req["node"]

    linum := "0"

    prenum, _ := req["ItemNumber"]
    key, _ := req["ItemKey"]

    lset, err := kpd.Hgetall("c:def:"+ key.(string))
    if err == nil {
        lsetn, _ := lset["n"]
        if lsetn != "" && lsetn != linum  {
            linum = lsetn
        }
    }

    msg := map[string]string{
        "action": "ItemPutCb",
        "node": node.(string),
        "ctlid": strconv.Itoa(kpsNum),
        "AcceptKey": key.(string),
    }

    postnum, _ := req["ItemNumberNext"]

    if linum == "0" || linum == prenum.(string) {
        
        postval, _ := req["ItemContent"]
        it := map[string]string{
            "n": postnum.(string),
            "v": postval.(string),
        }
        kpd.Hmset("c:def:"+ key.(string), it)
        msg["AcceptNumber"] = postnum.(string)
    } else {
        msg["AcceptNumber"] = prenum.(string)
    }

    // Ensure ctl:loctid to Max
    postnumi, _ := strconv.Atoi(postnum.(string));
    nnum := postnumi / len(kpls)
    ltid, _ := kpd.Incrby("ct:ltid", 0)
    if ltid < nnum && locNode != node.(string) {
        kpd.Incrby("ct:ltid", nnum - ltid)
    }

    kpn.Send(msg, addr +":9528")
}

func ActionItemPutCb(req ActionRequst, addr string) {
    if !req.isset("node") || !req.isset("ctlid") || !req.isset("AcceptNumber") || !req.isset("AcceptKey") {
        return
    }

    if locNode == "" || kpsNum == 0 {
        return
    }

    node, _ := req["node"]
    anum, _ := req["AcceptNumber"]
    anumi, _ := strconv.Atoi(anum.(string))
    ltid, _ := kpd.Incrby("ct:ltid", 0)
    rnum := anumi / len(kpls)
    if ltid < rnum && node.(string) != locNode {
        kpd.Incrby("ct:ltid", rnum - ltid)
    }

    akey, _ := req["AcceptKey"]
    if err := kpd.Exists("qk:"+ akey.(string) + anum.(string)); err != nil {
        return
    }

    apt, _ := kpd.Incrby("qk:"+ akey.(string) + anum.(string), 1)
    kpd.Expire("qk:"+ akey.(string) + anum.(string), 3)

    if 2 * apt > len(kpls) {
        //fmt.Println("QQQQQQQQQQQQQQQQQQ", "qv:"+ akey.(string) + anum.(string))
        bdy, _ := kpd.Get("qv:"+ akey.(string) + anum.(string))
        //fmt.Println(bdy)
        bdy2 := strings.Split(bdy, "\r\n")
        fmt.Println("ItemPutCbClient", bdy2)
        
        // TODO isset[3]
        it := map[string]string{
            "n": anum.(string),
            //"v": bdy2[2],
        }
        fmt.Println(it)
        fmt.Println("c:def:"+ akey.(string))
        //return
        _ = kpd.Hmset("c:def:"+ akey.(string), it)
        //return

        if node.(string) != locNode {
            it["k"] = akey.(string)
            it["node"] = locNode
            it["action"] = "ItemPutCbClient"
            kpn.Send(it, bdy2[0] +":9528")
        }
    }
}

func ActionItemPutCbClient(req ActionRequst, addr string) {
    if !req.isset("node") || !req.isset("n") || !req.isset("v") || !req.isset("k") {
        return
    }

    node, _ := req["node"]
    if node.(string) == locNode {
        return
    }


    n, _ := req["n"]
    v, _ := req["v"]
    k, _ := req["k"]
    it := map[string]string{
        "n": n.(string),
        "v": v.(string),
    }

    kpd.Hmset("c:def:"+ k.(string), it)

    fmt.Println("ActionItemPutCbClient OK")
}


func (req ActionRequst) isset(key string) bool {    
    if _, ok := req[key]; ok {
        return true
    }
    return false
}