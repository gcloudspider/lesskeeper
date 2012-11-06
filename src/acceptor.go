package main

import (
    "fmt"
    "strconv"
    "sync"
    "time"
)

type Acceptor int

var proposal_promiselock sync.Mutex
var proposal_promises = map[string]*ProposalPromise{}

func (p *Acceptor) Prepare(args *Proposal, rep *ProposalPromise) error {
    
    fmt.Println("Acceptor/Prepare", args)

    if locNode == "" || kpsNum == 0 || kpsLed == "" {
        return nil
    }

    proposal_promiselock.Lock()
    if pl, ok := proposal_promises[args.Key]; ok {
        rep = pl
    }
    proposal_promiselock.Unlock()

    if rep.VerNow > 0 {
        return nil
    }

    rep.VerNow = args.VerNow
    rep.VerSet = args.VerSet
    
    // 
    item, _ := db.Hgetall("c:def:"+ args.Key)   
    if val, ok := item["n"]; ok {
        if ver, e := strconv.ParseUint(val, 10, 64); e == nil && rep.VerNow != ver {
            rep.VerNow = ver
        }
    }

    proposal_promiselock.Lock()
    proposal_promises[args.Key] = rep
    proposal_promiselock.Unlock()

    go func() {
        time.Sleep(3e9)
        proposal_promiselock.Lock()
        if _, ok := proposal_promises[args.Key]; ok {
            delete(proposal_promises, args.Key)
        }
        proposal_promiselock.Unlock()
    }()

    return nil
}

func (p *Acceptor) Accept(args *Proposal, rep *Reply) error {
    
    rep.Status = ReplyError

    proposal_promiselock.Lock()
    pl, _ := proposal_promises[args.Key]
    proposal_promiselock.Unlock()

    if pl == nil {
        return nil
    }

    if args.VerNow == pl.VerNow && args.VerSet == pl.VerSet {
        
        item := map[string]string{
            "v": args.Val,
            "n": strconv.FormatUint(args.VerSet, 10),
        }

        //fmt.Println("database", item)
        db.Hmset("c:def:"+ args.Key, item)
        rep.Status = ReplyOK

        proposal_promiselock.Lock()
        if _, ok := proposal_promises[args.Key]; ok {
            delete(proposal_promises, args.Key)
        }
        proposal_promiselock.Unlock()
    }

    return nil
}
