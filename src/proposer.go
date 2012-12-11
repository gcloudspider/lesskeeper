package main

import (
    "strconv"
    "sync"
    "time"
    "strings"
)

type Proposer int

type Proposal struct {
    Key     string
    Val     string

    VerNow  uint64
    VerSet  uint64
    
    Tag     string
    Addr    string

    Valued  int
    Unvalued int
}

type ProposalPromise struct {
    VerNow, VerSet uint64
}

var proposals = map[uint64]*Proposal{}
var proposal_servlock sync.Mutex

type ProposalWatcher map[string]int

var watcherlock sync.Mutex
var watches     = map[string]ProposalWatcher{}
var watchmq     = make(chan *WatcherQueue, 100000)
type WatcherQueue struct {
    Path    string
    Event   string
    Rev     uint64
}

func WatcherInitialize() {

    go func() {

        for q := range watchmq {

            if w, ok := watches[q.Path]; ok {
                for hostid, _ := range w {
                    if ip, ok := kp[hostid]; ok {
                        // Println("Send to", ip +":"+ port)
                        msg := map[string]string{
                            "action": "WatchEvent",
                            "path": q.Path,
                            "event": q.Event,
                        }
                        peer.Send(msg, ip +":"+ port)
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

    switch string(args[0]) {
    case "GET":
        ProposerGet(args, rep)
    case "SET":
        ProposerSet(args, rep)
    case "WATCH":
        ProposerWatch(args, rep)
    }
    
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
        w = map[string]int{}
        watches[path] = w
    }
    ttl, _ := strconv.Atoi(string(args[2]))
    w[string(args[3])] = ttl
    watcherlock.Unlock()

    //Println("args==", args, w)
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
        rep.Val  = node.C
    }

    return
}

func ProposerSet(args map[int][]byte, rep *Reply) {

    if len(args) < 3 {
        rep.Type = ReplyError
        return
    }

    nodeEvent := NodeEventNone

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
        nodeEvent = NodeEventDataChanged
    } else {
        nodeEvent = NodeEventCreated
    }

    n , _ := db.Incrby("ctl:ltid", 1)
    vernewi := len(kps) * n + kpsNum - 1
    verset := uint64(vernewi)
    //vernews := strconv.Itoa(vernewi)

    pl := new(Proposal)
    pl.Key = path
    pl.Val = string(args[2])
    pl.VerNow = node.R
    pl.VerSet = verset
    pl.Valued    = 0
    pl.Unvalued  = 0   

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
            call.Addr = v +":"+ gport
            
            gnet.Call(call)

            _ = <- call.Status
            
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

    L: for {
        select {
        case s := <- promised:
            if s == 1 {
                valued++
                if 2 * valued > len(kp) {
                    //fmt.Println("Valued")
                    break L
                }
            } else if s == 0 {
                unvalued++
                if 2 * unvalued > len(kp) {
                    rep.Type = ReplyError
                    rep.Val = "UnValued"
                    return                }
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
            call.Addr = v +":"+ gport
            
            gnet.Call(call)

            _ = <- call.Status
            
            rs := call.Reply.(*Reply)

            //fmt.Println("Acceptor.Accept", rs)
            if rs.Status == ReplyOK {
                accepted <- 1
            } else {
                accepted <- 0
            }
        }()
    }

    valued = 0
    unvalued = 0

    A: for {
        select {
        case s := <- accepted:
            if s == 1 {
                valued++
                if 2 * valued > len(kp) {
                    rep.Type = ReplyOK
                    watchmq <- &WatcherQueue{strings.Trim(path, "/"), nodeEvent, 0}
                    break A
                }
            } else if s == 0 {
                unvalued++
                if 2 * unvalued > len(kp) {
                    rep.Type = ReplyError
                    rep.Val = "UnValued"
                    return                }
            } else {
                rep.Type = ReplyError
                return
            }
        }
    }


    return
}

