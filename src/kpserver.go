package main

import (
    "fmt"
    "net"
    "net/rpc"
    "net/http"
    "sync"
    "strconv"
    "regexp"
)

var lock sync.Mutex

type Command int

type Coordinator struct {
    Status int
    Number uint64
    Content string 
}

var copool = map[string]*Coordinator{}

func NewServer(port int) {

    cmd := new(Command)
    rpc.Register(cmd)
    rpc.HandleHTTP()

    l, e := net.Listen("tcp", ":"+ strconv.Itoa(port))
    if e != nil {
        fmt.Println("listen error:", e)
    }
    
    go http.Serve(l, nil)
}

func (cmd *Command) Process(args *AgentCommand, reply *AgentReply) error {

    if len(args.Argv) == 0 {
        return nil
    }

    switch string(args.Argv[0]) {
    case "GET":
        cmd.CmdGet(args.Argv, reply)
    case "PUT":
        cmd.CmdPut(args, reply)
    }

    for i := 0; i < len(args.Argv); i++ {
        fmt.Println("CMD", i, string(args.Argv[i]))
    }

    // fmt.Println("Server Command.Process", args.Tag)
    
    return nil
}

func (cmd *Command) CmdGet(argv map[int][]byte, reply *AgentReply) {
    
    if len(argv) < 2 {
        return
    }

    key := string(argv[1])

    if ok, _ := regexp.MatchString("^([0-9a-zA-Z ._-]{1,64})$", key); !ok {
        return
    }

    item, _ := kpd.Hgetall("c:def:"+ key)

    if val, ok := item["v"]; ok {
        reply.Val = val
    }

    if val, ok := item["n"]; ok {
        if num, e := strconv.ParseUint(val, 10, 64); e == nil {
            reply.Ver = num
        }
    }
}

func (cmd *Command) CmdPut(args *AgentCommand, reply *AgentReply) {
    
    if len(args.Argv) < 3 {
        return
    }

    key := string(args.Argv[1])
    if ok, _ := regexp.MatchString("^([0-9a-zA-Z ._-]{1,64})$", key); !ok {
        return
    }

    num := "0"
    item, err := kpd.Hgetall("c:def:"+ key)
    if err == nil {

        if val, ok := item["n"]; ok {
            if val != num {
                num = val
            }
        }

        if val, ok := item["v"]; ok {
            if val == string(args.Argv[2]) {
                reply.Val = "+OK"
                return
            }
        }
    }

    n , _ := kpd.Incrby("ctl:ltid", 1)
    kpnoi := len(kps) * n + kpsNum - 1

    kpnos := strconv.Itoa(kpnoi)

    req := map[string]string{
        "node": locNode,
        "action": "ItemPut",
        "ItemNumber": num,
        "ItemKey": key,
        "ItemContent": string(args.Argv[2]),
        "ItemNumberNext": kpnos,
    }

    bdy := args.Addr +"#"+ args.Tag +"#"+ key +"#"+ string(args.Argv[2])
    
    kpd.Setex("qk:"+ key + num, 3, "0")
    kpd.Setex("qv:"+ key + num, 3, bdy)
    
    kpn.Send(req, "255.255.255.255:9528")

    reply.Val = "+OK"

    return
}