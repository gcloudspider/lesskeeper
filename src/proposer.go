package main

import (
    "strconv"
    "sync"
    "time"
)

type Proposer int

type Proposal struct {
    Key     string
    Val     string

    Ver     string
    VerNew  string

    VerNow  uint64
    VerSet  uint64
    
    Tag     string
    Addr    string
    Yes     int

    Valued  int
    Unvalued int
}

type ProposalVersion uint64

type Proposal2 struct {
    Key     string
    Val     string
    VerNow  uint64
    VerSet  uint64
}
var pls map[uint64]*Proposal2

type ProposalPromise struct {
    VerNow, VerSet uint64
}

var proposals map[string]*Proposal
var proposal_servlock sync.Mutex


func (p *Proposer) Process(args map[int][]byte, rep *Reply) error {

    if len(args) == 0 {
        return nil
    }

    switch string(args[0]) {
    case "GET":
        ProposerGet(args, rep)
    case "SET":
        ProposerPut(args, rep)
    }
    
    return nil
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

func ProposerPut(args map[int][]byte, rep *Reply) {

    if len(args) < 3 {
        rep.Type = ReplyError
        return
    }
    
    path := string(args[1])
    /* if ok, _ := regexp.MatchString("^([0-9a-zA-Z ._-]{1,64})$", path); !ok {
        rep.Type = ReplyError
        return
    } */

    vers := "0"
    item, err := db.Hgetall("c:def:"+ path)
    if err == nil {

        if val, ok := item["n"]; ok {
            if val != vers {
                vers = val
            }
        }

        if val, ok := item["v"]; ok {
            if val == string(args[2]) {
                rep.Type = ReplyOK
                return
            }
        }
    }

    n , _ := db.Incrby("ctl:ltid", 1)
    vernewi := len(kps) * n + kpsNum - 1

    vernews := strconv.Itoa(vernewi)

    pl := new(Proposal)
    pl.Key = path
    pl.Val = string(args[2])
    pl.VerNew = vernews
    pl.Ver    = vers
    pl.Valued    = 0
    pl.Unvalued  = 0

    if v, e := strconv.ParseUint(vers, 10, 64); e == nil {
        pl.VerNow = v
    }
    if v, e := strconv.ParseUint(vernews, 10, 64); e == nil {
        pl.VerSet = v
    }

    proposals[vernews] = pl

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