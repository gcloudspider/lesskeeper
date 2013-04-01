package main

import (
    //"regexp"
    "encoding/json"
    "strconv"
    "strings"
)

const (
    NodePathPat = `[a-zA-Z0-9.\-\/]`
    NodeDelFlag = "ukqmv4jfxyapbeqo"

    NodeSepFile = "x"
    NodeSepDir  = "d"
    NodeSepTmp  = "t"
    NodeSepRev  = "n"
    NodeSepVal  = "v"

    INodeFile = "in" + NodeSepFile + ":"
    INodeDir  = "in" + NodeSepDir + ":"

    NodeTypeNil  = uint8(0)
    NodeTypeDir  = uint8(1)
    NodeTypeFile = uint8(2)

    EventNone                = "10"
    EventNodeCreated         = "11"
    EventNodeDeleted         = "12"
    EventNodeDataChanged     = "13"
    EventNodeChildrenChanged = "14"
)

//var pathRe = mustBuildRe(NodePathPat)

type Node struct {
    P string // Path
    C string // Content
    R uint64 // Revison (100 ~ n)
    // TODO U   uint16  // uid
    // TODO G   uint16  // gid
    // TODO M   uint16  // Mode
    T uint8 // Type
}

/*
func mustBuildRe(path string) *regexp.Regexp {
    return regexp.MustCompile(`^/$|^(/` + path + `+)+$`)
}

func checkPath(path string) uint64 {
    if !pathRe.MatchString(path) {
        return NodeTypeNil
    }

    return 0
}
*/

func split(path string, p string) []string {
    if path == p {
        return []string{}
    }
    return strings.Split(path, p)
}

func join(parts []string, p string) string {
    return strings.Join(parts, p)
}

func NodeSet(pl *Proposal) uint16 {

    // Saving File
    in := strings.Trim(pl.Key, "/")
    if pl.Val == NodeDelFlag {
        db.Hdel(INodeFile+in, "v")
        db.Hdel(INodeFile+in, "r")
        // TODO clean inodex
        return 0
    }

    item := map[string]string{
        "v": pl.Val,
        "r": strconv.FormatUint(pl.VerSet, 10),
    }
    db.Hmset(INodeFile+in, item)

    // Saving DIRs
    p := split(in, "/")
    for i := len(p) - 1; i >= 0; i-- {
        in = join(p[0:i], "/")
        if i == len(p)-1 {
            db.Sadd(INodeDir+in, NodeSepFile+p[i])
        } else {
            db.Sadd(INodeDir+in, NodeSepDir+p[i])
        }
    }

    return 0
}

func NodeGet(path string) (*Node, error) {

    in := strings.Trim(path, "/")

    item, e := db.Hgetall(INodeFile + in)
    if e != nil {
        return nil, e
    }

    node := new(Node)
    for k, v := range item {
        switch k {
        case "v":
            node.C = v
        case "r":
            if rev, e := strconv.ParseUint(v, 10, 64); e == nil {
                node.R = rev
            }
        }
    }
    return node, nil
}

func NodeGets(keys string) (string, error) {

    ks := split(keys, " ")

    list := []Node{}

    for _, path := range ks {

        in := strings.Trim(path, "/")

        item, e := db.Hgetall(INodeFile + in)
        if e != nil {
            return "", e
        }

        n := Node{}
        for k, v := range item {

            switch k {
            case NodeSepRev:
                //n.R = uint64(v)
            case NodeSepVal:
                n.C = v
            }
            n.T = NodeTypeFile
            n.P = path
        }

        list = append(list, n)
    }

    if rs, e := json.Marshal(list); e == nil {
        return string(rs), nil
    }
    return "", nil
}

func NodeList(path string) (string, error) {

    in := strings.Trim(path, "/")

    item, e := db.Smembers(INodeDir + in)
    if e != nil {
        return "", e
    }

    list := []Node{}
    for _, v := range item {

        n := Node{}
        switch v[0:1] {
        case NodeSepDir:
            n.T = NodeTypeDir
        case NodeSepFile:
            n.T = NodeTypeFile
        }
        n.P = v[1:]
        list = append(list, n)
    }
    //return node, nil
    //Println(list)
    if rs, e := json.Marshal(list); e == nil {
        return string(rs), nil
    }
    return "", nil
}
