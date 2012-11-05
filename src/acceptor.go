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
    
    // 
    item, _ := db.Hgetall("c:def:"+ args.Key)   
    if val, ok := item["n"]; ok {
        if ver, e := strconv.ParseUint(val, 10, 64); e == nil {
            rep.VerNow = ver
        }
    }

    //if args.VerSet > rep.VerNow {
        rep.VerSet = args.VerSet
    //}

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

func (p *Acceptor) Accept(args *Proposal, rep *ProposalPromise) error {

    return nil
}
