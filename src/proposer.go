package main

import (
    "../deps/lessgo/utils"
    "./peer"
    "./store"
    "strings"
    "sync"
    "time"
)

type Reply peer.Reply

type Proposer int

type Request peer.Request

type ProposalPromise struct {
    VerNow, VerSet uint64
}

var proposals = map[uint64]*store.NodeProposal{}
var proposal_servlock sync.Mutex

type ProposalWatcher map[string]int64 // map[host]ttl

var watcherlock sync.Mutex
var watches = map[string]ProposalWatcher{} // map[path]*
var watchmq = make(chan *WatcherQueue, 100000)

type WatcherQueue struct {
    Path  string
    Event string
    Rev   uint64
}

/** TODO func WatcherInitialize() {

    go func() {

        for q := range watchmq {

            if w, ok := watches[q.Path]; ok {
                for hostid, ttl := range w {

                    // delay clean watcher queue
                    if ttl < time.Now().Unix() {
                        watcherlock.Lock()
                        delete(w, hostid)
                        watcherlock.Unlock()
                        continue
                    }

                    if ip, ok := kp[hostid]; ok {
                        // Println("Send to", ip +":"+ port)
                        msg := map[string]string{
                            "action": "WatchEvent",
                            "path":   q.Path,
                            "event":  q.Event,
                        }
                        prbc.Send(msg, ip+":"+cfg.KeeperPort)
                    }
                }
            }
        }
    }()
}
*/

func (p *Proposer) Cmd(rq *Request, rp *Reply) error {

    switch string(rq.Method) {
    case "GET":
        CmdGet(rq, rp)
    case "GETS":
        CmdGets(rq, rp)
    case "LIST":
        CmdList(rq, rp)
    case "SET":
        CmdSet(rq, rp, false)
    case "DEL":
        CmdSet(rq, rp, true)
    case "KPRMEMSET":
        CmdKprMemSet(rq, rp)
    }

    return nil
}

func CmdGet(rq *Request, rp *Reply) {

    var rqbody struct {
        Path string
    }
    e := utils.JsonDecode(rq.Body, &rqbody)
    if e != nil {
        return
    }

    if node, e := stor.NodeGet(rqbody.Path); e == nil {
        rp.Type = peer.ReplyString
        rp.Body = node.C
    }
}

func CmdGets(rq *Request, rp *Reply) {

    var rqbody struct {
        Path string
    }
    e := utils.JsonDecode(rq.Body, &rqbody)
    if e != nil {
        return
    }

    if rs, e := stor.NodeGets(rqbody.Path); e == nil {
        rp.Type = peer.ReplyString
        rp.Body = rs
    }

    return
}

func CmdList(rq *Request, rp *Reply) {

    var rqbody struct {
        Path string
    }
    e := utils.JsonDecode(rq.Body, &rqbody)
    if e != nil {
        return
    }

    if rs, e := stor.NodeList(rqbody.Path); e == nil {
        rp.Type = peer.ReplyString
        rp.Body = rs
    }
}

func CmdSet(rq *Request, rp *Reply, del bool) {

    var rqbody struct {
        Path string
        Val  string
    }
    e := utils.JsonDecode(rq.Body, &rqbody)
    if e != nil {
        return
    }
    if del {
        rqbody.Val = store.NodeDelFlag
    }

    nodeEvent := store.EventNone

    node, _ := stor.NodeGet(rqbody.Path)
    if node.R > 0 {
        if node.C == rqbody.Val {
            //Println("same node", rqbody.Path)
            rp.Type = peer.ReplyOK
            return
        }
        nodeEvent = store.EventNodeDataChanged
    } else {
        nodeEvent = store.EventNodeCreated
    }

    n, _ := stor.Incrby("ctl:ltid", 1)
    vernewi := len(kprGrp)*n + kprSef.KprNum - 1
    verset := uint64(vernewi)
    //vernews := strconv.Itoa(vernewi)

    pl := new(store.NodeProposal)
    pl.Key = rqbody.Path
    pl.Val = rqbody.Val
    pl.VerNow = node.R
    pl.VerSet = verset
    pl.Valued = 0
    pl.Unvalued = 0

    proposals[verset] = pl

    promised := make(chan uint8, len(kprGrp))
    go func() {
        time.Sleep(30e9)
        promised <- 9
    }()

    // Acceptor.Prepare
    for _, v := range kprGrp {

        go func() {

            call := peer.NewNetCall()

            call.Method = "Acceptor.Prepare"
            call.Args = pl
            call.Reply = new(ProposalPromise)
            call.Addr = v.Addr + ":" + cfg.KeeperPort

            prkp.Call(call)

            _ = <-call.Status

            rs := call.Reply.(*ProposalPromise)
            if rs.VerNow == pl.VerNow {
                promised <- 1
            } else {
                promised <- 0
            }
        }()
    }

    valued := 0
    unvalued := 0

L:
    for {
        select {
        case s := <-promised:
            if s == 1 {
                valued++
                if 2*valued > len(kprGrp) {
                    //fmt.Println("Valued")
                    break L
                }
            } else if s == 0 {
                unvalued++
                if 2*unvalued > len(kprGrp) {
                    rp.Type = peer.ReplyError
                    rp.Body = "UnValued"
                    return
                }
            } else {
                rp.Type = peer.ReplyError
                return
            }
        }
    }

    // Acceptor.Accept
    accepted := make(chan uint8, len(kprGrp))
    go func() {
        time.Sleep(30e9)
        accepted <- 9
    }()

    for _, v := range kprGrp {

        go func() {

            call := peer.NewNetCall()

            call.Method = "Acceptor.Accept"
            call.Args = pl
            call.Reply = new(Reply)
            call.Addr = v.Addr + ":" + cfg.KeeperPort

            prkp.Call(call)

            _ = <-call.Status

            rs := call.Reply.(*Reply)

            //fmt.Println("Acceptor.Accept", rs)
            if rs.Type == peer.ReplyOK {
                accepted <- 1
            } else {
                accepted <- 0
            }
        }()
    }

    valued = 0
    unvalued = 0

A:
    for {
        select {
        case s := <-accepted:
            if s == 1 {
                valued++
                if 2*valued > len(kprGrp) {
                    rp.Type = peer.ReplyOK
                    watchmq <- &WatcherQueue{strings.Trim(rqbody.Path, "/"), nodeEvent, 0}
                    break A
                }
            } else if s == 0 {
                unvalued++
                if 2*unvalued > len(kprGrp) {
                    rp.Type = peer.ReplyError
                    rp.Body = "UnValued"
                    return
                }
            } else {
                rp.Type = peer.ReplyError
                return
            }
        }
    }
}

func CmdKprMemSet(rq *Request, rp *Reply) {

    rp.Type = peer.ReplyError

    var rqbody struct {
        Addr string
        Port string
    }
    e := utils.JsonDecode(rq.Body, &rqbody)
    if e != nil {
        return
    }

    if len(kprLed) > 4 || rqbody.Addr != kprSef.Addr {
        return
    }

    msa, _ := stor.Hgetall("ctl:members")
    if len(msa) > 0 {
        return
    }

    if e := stor.Hset("ctl:members", "1", kprSef.Id); e != nil {
        return
    }

    rp.Type = peer.ReplyOK
}
