package main

import (
    "fmt"
    //"net"
    "net/rpc"
    "net/http"
    "sync"
    "strconv"
    //"regexp"
    //"errors"
    "time"
)

type Server int

type Proposal struct {
    Key     string
    Val     string
    Ver     string
    VerNew  string
    Tag     string
    Addr    string
    Yes     int
}

var proposals map[string]*Proposal
var proposal_servlock sync.Mutex

func NewServer() {

    fmt.Println("Proposal Server Start")

    proposals = map[string]*Proposal{}

    serv := new(Server)
    rpc.Register(serv)
    rpc.HandleHTTP()

    /* l, e := net.Listen("tcp", ":"+ strconv.Itoa(port))
    if e != nil {
        fmt.Println("listen error:", e)
    } */
    
    go http.Serve(peer.tl, nil)
}

func (s *Server) Process(cmd *Command, rep *Reply) error {

    //fmt.Println("Server Process", cmd)
    //rep.Val = "TEST"
    //return nil

    if len(cmd.Argv) == 0 {
        return nil
    }

    switch string(cmd.Argv[0]) {
    case "GET":
        CmdGet(cmd.Argv, rep)
    case "PUT", "SET":
        //CmdPut(cmd, rep)
    }

    /* for i := 0; i < len(cmd.Argv); i++ {
        fmt.Println("CMD", i, string(cmd.Argv[i]))
    } */
    
    return nil
}

func CmdGet(argv map[int][]byte, rep *Reply) {

    if len(argv) < 2 {
        rep.Type = ReplyError
        return
    }

    key := string(argv[1])
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

func CmdPut(cmd *Command, rep *Reply) {

    if len(cmd.Argv) < 3 {
        rep.Type = ReplyError
        return
    }
    
    key := string(cmd.Argv[1])
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
            if val == string(cmd.Argv[2]) {
                it := map[string]string{
                    "action": "AgentItemPutCb",
                    "Tag":  cmd.Tag,
                    "status": "10",
                }
                peer.Send(it, cmd.Addr +":"+ port)
                //rep.Err = errors.New("400")
                return
            }
        }
    }

    n , _ := db.Incrby("ctl:ltid", 1)
    vernewi := len(kps) * n + kpsNum - 1

    vernews := strconv.Itoa(vernewi)

    req := map[string]string{
        "node": locNode,
        "action": "ItemPut",
        "ItemKey": key,
        "ItemContent": string(cmd.Argv[2]),
        "VerNew": vernews,
    }

    p := new(Proposal)
    p.Key   = key
    p.Val   = string(cmd.Argv[2])
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
}