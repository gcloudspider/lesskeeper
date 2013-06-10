package main

import (
    pr "./peer"
    "./store"
    "sync"
    "time"
)

type Acceptor int

var promiselock sync.Mutex
var promises = map[string]*ProposalPromise{}

func (p *Acceptor) Prepare(args *store.NodeProposal, rep *ProposalPromise) error {

    //Println("Acceptor/Prepare", args)

    if locNode == "" || kpsNum == 0 || kpsLed == "" {
        return nil
    }

    promiselock.Lock()
    if pl, ok := promises[args.Key]; ok {
        rep = pl
    }
    promiselock.Unlock()

    if rep.VerNow > 0 {
        return nil
    }

    rep.VerNow = args.VerNow
    rep.VerSet = args.VerSet

    n, _ := stor.NodeGet(args.Key)
    if rep.VerNow != n.R {
        rep.VerNow = n.R
    }

    promiselock.Lock()
    promises[args.Key] = rep
    promiselock.Unlock()

    go func() {
        time.Sleep(3e9)
        promiselock.Lock()
        if _, ok := promises[args.Key]; ok {
            delete(promises, args.Key)
        }
        promiselock.Unlock()
    }()

    return nil
}

func (p *Acceptor) Accept(args *store.NodeProposal, rep *Reply) error {

    rep.Type = pr.ReplyError

    promiselock.Lock()
    pl, _ := promises[args.Key]
    promiselock.Unlock()

    if pl == nil {
        return nil
    }

    if args.VerNow == pl.VerNow && args.VerSet == pl.VerSet {

        // TODO
        // Method Dispatch
        _ = stor.NodeSet(args)

        rep.Type = pr.ReplyOK

        promiselock.Lock()
        if _, ok := promises[args.Key]; ok {
            delete(promises, args.Key)
        }
        promiselock.Unlock()
    }

    return nil
}
