package main

import (
    pr "./peer"
    "./utils"
    "strconv"
    "strings"
    "sync"
    "time"
)

type Proposer int

type Proposal struct {
    Key string
    Val string

    VerNow uint64
    VerSet uint64

    Tag  string
    Addr string

    Valued   int
    Unvalued int
}

type Request pr.Request

type ProposalPromise struct {
    VerNow, VerSet uint64
}

var proposals = map[uint64]*Proposal{}
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

func (p *Proposer) Cmd(rq *Request, rp *Reply) error {

    switch string(rq.Method) {
    case "GET":
        CmdGet(rq, rp)
    case "GETS":
        CmdGets(rq, rp)
    case "LIST":
        CmdList(rq, rp)
    case "SET":
        CmdSet(rq, rp)
    case "DEL":
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

    if node, e := NodeGet(rqbody.Path); e == nil {
        rp.Type = ReplyString
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

    if rs, e := NodeGets(rqbody.Path); e == nil {
        rp.Type = ReplyString
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

    if rs, e := NodeList(rqbody.Path); e == nil {
        rp.Type = ReplyString
        rp.Body = rs
    }
}

func CmdSet(rq *Request, rp *Reply) {

    var rqbody struct {
        Path string
        Val  string
    }
    e := utils.JsonDecode(rq.Body, &rqbody)
    if e != nil {
        return
    }

    nodeEvent := EventNone

    /* if ok, _ := regexp.MatchString("^([0-9a-zA-Z ._-]{1,64})$", rqbody.Path); !ok {
        rp.Type = ReplyError
        return
    } */

    node, _ := NodeGet(rqbody.Path)
    if node.R > 0 {
        if node.C == rqbody.Val {
            //Println("same node", rqbody.Path)
            rp.Type = ReplyOK
            return
        }
        nodeEvent = EventNodeDataChanged
    } else {
        nodeEvent = EventNodeCreated
    }

    n, _ := db.Incrby("ctl:ltid", 1)
    vernewi := len(kps)*n + kpsNum - 1
    verset := uint64(vernewi)
    //vernews := strconv.Itoa(vernewi)

    pl := new(Proposal)
    pl.Key = rqbody.Path
    pl.Val = rqbody.Val
    pl.VerNow = node.R
    pl.VerSet = verset
    pl.Valued = 0
    pl.Unvalued = 0

    proposals[verset] = pl
    //fmt.Println("PUT Acceptor.Prepare", pl)

    promised := make(chan uint8, len(kp))
    go func() {
        time.Sleep(30e9)
        promised <- 9
    }()

    // Acceptor.Prepare
    for _, v := range kp {

        go func() {

            call := NewNetCall()

            call.Method = "Acceptor.Prepare"
            call.Args = pl
            call.Reply = new(ProposalPromise)
            call.Addr = v + ":" + cfg.KeeperPort

            gnet.Call(call)

            _ = <-call.Status

            rs := call.Reply.(*ProposalPromise)
            if rs.VerNow == pl.VerNow {
                promised <- 1
            } else {
                promised <- 0
            }
        }()

        //fmt.Println(k, v)
    }

    valued := 0
    unvalued := 0

L:
    for {
        select {
        case s := <-promised:
            if s == 1 {
                valued++
                if 2*valued > len(kp) {
                    //fmt.Println("Valued")
                    break L
                }
            } else if s == 0 {
                unvalued++
                if 2*unvalued > len(kp) {
                    rp.Type = ReplyError
                    rp.Body = "UnValued"
                    return
                }
            } else {
                rp.Type = ReplyError
                return
            }
        }
    }

    // Acceptor.Accept
    accepted := make(chan uint8, len(kp))
    go func() {
        time.Sleep(30e9)
        accepted <- 9
    }()

    for _, v := range kp {

        go func() {

            call := NewNetCall()

            call.Method = "Acceptor.Accept"
            call.Args = pl
            call.Reply = new(Reply)
            call.Addr = v + ":" + cfg.KeeperPort

            gnet.Call(call)

            _ = <-call.Status

            rs := call.Reply.(*Reply)

            //fmt.Println("Acceptor.Accept", rs)
            if rs.Type == ReplyOK {
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
                if 2*valued > len(kp) {
                    rp.Type = ReplyOK
                    watchmq <- &WatcherQueue{strings.Trim(rqbody.Path, "/"), nodeEvent, 0}
                    break A
                }
            } else if s == 0 {
                unvalued++
                if 2*unvalued > len(kp) {
                    rp.Type = ReplyError
                    rp.Body = "UnValued"
                    return
                }
            } else {
                rp.Type = ReplyError
                return
            }
        }
    }
}

////////////////////////////////////////////////////////////////////////////////

func WatcherInitialize() {

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
                        peer.Send(msg, ip+":"+cfg.KeeperPort)
                    }
                }
            }
        }
    }()
}

func (p *Proposer) Process(args map[int][]byte, rep *Reply) error {

    if len(args) == 0 {
        return nil
    }

    //Println(string(args[0]))

    switch string(args[0]) {
    case "GET":
        ProposerGet(args, rep)
    case "GETS":
        ProposerGets(args, rep)
    case "SET":
        ProposerSet(args, rep)
    case "DEL":
        args[2] = []byte(NodeDelFlag)
        ProposerSet(args, rep)
    case "LIST":
        ProposerList(args, rep)
    case "WATCH":
        ProposerWatch(args, rep)
    case "SELECT":
        rep.Type = ReplyOK
    }

    //Println(rep)
    return nil
}

func ProposerWatch(args map[int][]byte, rep *Reply) {

    rep.Type = ReplyWatch

    if len(args) < 4 {
        rep.Type = ReplyError
        return
    }
    path := strings.Trim(string(args[1]), "/")

    var w ProposalWatcher
    var ok bool
    watcherlock.Lock()
    if w, ok = watches[path]; !ok {
        w = map[string]int64{}
        watches[path] = w
    }
    ttlen, _ := strconv.Atoi(string(args[2]))
    w[string(args[3])] = time.Now().Unix() + int64(ttlen)
    watcherlock.Unlock()

    //Println("args==", args, w)
    return
}

func ProposerList(args map[int][]byte, rep *Reply) {

    if len(args) < 2 {
        rep.Type = ReplyError
        return
    }

    path := string(args[1])
    /* if ok, _ := regexp.MatchString("^([0-9a-zA-Z ._-]{1,64})$", path); !ok {
        rep.Type = ReplyError
        return
    } */

    if rs, e := NodeList(path); e == nil {
        rep.Type = ReplyString
        rep.Body = rs
    }

    return
}

func ProposerGet(args map[int][]byte, rep *Reply) {

    if len(args) < 2 {
        rep.Type = ReplyError
        return
    }

    path := string(args[1])
    /* if ok, _ := regexp.MatchString("^([0-9a-zA-Z ._-]{1,64})$", path); !ok {
        rep.Type = ReplyError
        return
    } */

    if node, e := NodeGet(path); e == nil {
        rep.Type = ReplyString
        rep.Body = node.C
    }

    return
}

func ProposerGets(args map[int][]byte, rep *Reply) {

    if len(args) < 2 {
        rep.Type = ReplyError
        return
    }

    keys := string(args[1])
    if rs, e := NodeGets(keys); e == nil {
        rep.Type = ReplyString
        rep.Body = rs
    }

    return
}

func ProposerSet(args map[int][]byte, rep *Reply) {

    if len(args) < 3 {
        rep.Type = ReplyError
        return
    }

    nodeEvent := EventNone

    path := string(args[1])
    /* if ok, _ := regexp.MatchString("^([0-9a-zA-Z ._-]{1,64})$", path); !ok {
        rep.Type = ReplyError
        return
    } */

    node, _ := NodeGet(path)
    if node.R > 0 {
        if node.C == string(args[2]) {
            //Println("same node", path)
            rep.Type = ReplyOK
            return
        }
        nodeEvent = EventNodeDataChanged
    } else {
        nodeEvent = EventNodeCreated
    }

    n, _ := db.Incrby("ctl:ltid", 1)
    vernewi := len(kps)*n + kpsNum - 1
    verset := uint64(vernewi)
    //vernews := strconv.Itoa(vernewi)

    pl := new(Proposal)
    pl.Key = path
    pl.Val = string(args[2])
    pl.VerNow = node.R
    pl.VerSet = verset
    pl.Valued = 0
    pl.Unvalued = 0

    proposals[verset] = pl
    //fmt.Println("PUT Acceptor.Prepare", pl)

    promised := make(chan uint8, len(kp))
    go func() {
        time.Sleep(30e9)
        promised <- 9
    }()

    // Acceptor.Prepare
    for _, v := range kp {

        go func() {

            call := NewNetCall()

            call.Method = "Acceptor.Prepare"
            call.Args = pl
            call.Reply = new(ProposalPromise)
            call.Addr = v + ":" + gport

            gnet.Call(call)

            _ = <-call.Status

            rs := call.Reply.(*ProposalPromise)
            if rs.VerNow == pl.VerNow {
                promised <- 1
            } else {
                promised <- 0
            }
        }()

        //fmt.Println(k, v)
    }

    valued := 0
    unvalued := 0

L:
    for {
        select {
        case s := <-promised:
            if s == 1 {
                valued++
                if 2*valued > len(kp) {
                    //fmt.Println("Valued")
                    break L
                }
            } else if s == 0 {
                unvalued++
                if 2*unvalued > len(kp) {
                    rep.Type = ReplyError
                    rep.Body = "UnValued"
                    return
                }
            } else {
                rep.Type = ReplyError
                return
            }
        }
    }

    // Acceptor.Accept
    accepted := make(chan uint8, len(kp))
    go func() {
        time.Sleep(30e9)
        accepted <- 9
    }()

    for _, v := range kp {

        go func() {

            call := NewNetCall()

            call.Method = "Acceptor.Accept"
            call.Args = pl
            call.Reply = new(Reply)
            call.Addr = v + ":" + gport

            gnet.Call(call)

            _ = <-call.Status

            rs := call.Reply.(*Reply)

            //fmt.Println("Acceptor.Accept", rs)
            if rs.Type == ReplyOK {
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
                if 2*valued > len(kp) {
                    rep.Type = ReplyOK
                    watchmq <- &WatcherQueue{strings.Trim(path, "/"), nodeEvent, 0}
                    break A
                }
            } else if s == 0 {
                unvalued++
                if 2*unvalued > len(kp) {
                    rep.Type = ReplyError
                    rep.Body = "UnValued"
                    return
                }
            } else {
                rep.Type = ReplyError
                return
            }
        }
    }

    return
}
