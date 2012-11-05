package main

import (
    "fmt"
    "strconv"
    "sync"
    //"time"
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

    key := string(args[1])
    /* if ok, _ := regexp.MatchString("^([0-9a-zA-Z ._-]{1,64})$", key); !ok {
        rep.Type = ReplyError
        return
    } */

    item, _ := db.Hgetall("c:def:"+ key)

    if val, ok := item["v"]; ok {
        rep.Val  = val
        rep.Type = ReplyString
    }
    
    if val, ok := item["n"]; ok {
        if ver, e := strconv.ParseUint(val, 10, 64); e == nil {
            rep.Ver = ver
        }
    }
}

func ProposerPut(args map[int][]byte, rep *Reply) {

    if len(args) < 3 {
        rep.Type = ReplyError
        return
    }
    
    key := string(args[1])
    /* if ok, _ := regexp.MatchString("^([0-9a-zA-Z ._-]{1,64})$", key); !ok {
        rep.Type = ReplyError
        return
    } */

    vers := "0"
    item, err := db.Hgetall("c:def:"+ key)
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
    pl.Key = key
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

    //

    call := NewNetCall()
    call.Method = "Acceptor.Prepare"
    call.Addr = "127.0.0.1:"+gport
    call.Args = pl
    call.Reply = new(ProposalPromise)

    fmt.Println("PUT Acceptor.Prepare", pl)

    gnet.Call(call)

    st := <- call.Status

    //var rsp string
    rs := call.Reply.(*ProposalPromise)

    fmt.Println("PUT Acceptor.Prepare", pl, st, rs)


    rep.Type = ReplyOK
    /*req := map[string]string{
        "node": locNode,
        "action": "ItemPut",
        "ItemKey": key,
        "ItemContent": string(args[2]),
        "VerNew": vernews,
    }

    p := new(Proposal)
    p.Key   = key
    p.Val   = string(args[2])
    p.Ver   = vers
    p.VerNew= vernews
    p.Addr  = cmd.Addr
    p.Tag   = cmd.Tag
    
    proposals[vernews] = p

    //peer.Send(req, "255.255.255.255:"+ port) // TODO
    peer.Send(req, "127.0.0.1:"+ port) // TODO
    
    // timeout and free
    go func() {
        time.Sleep(4e9)
        proposal_servlock.Lock()
        if _, ok := proposals[vernews]; ok {
            delete(proposals, vernews)
        }
        proposal_servlock.Unlock()
    }()
    */
}
