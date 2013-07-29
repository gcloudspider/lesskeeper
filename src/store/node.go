package store

import (
    "encoding/json"
    //"fmt"
    "regexp"
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

type NodeProposal struct {
    Key string
    Val string
    Ttl int

    VerNow uint64
    VerSet uint64

    Tag  string
    Addr string

    Valued   int
    Unvalued int
}

//var pathRe = mustBuildRe(NodePathPat)

type Node struct {
    P   string // Path
    C   string // Content
    R   uint64 // Revison (100 ~ n)
    // TODO U   uint16  // uid
    // TODO G   uint16  // gid
    // TODO M   uint16  // Mode
    T   uint8 // Type
}

func split(path string, p string) []string {
    if path == p {
        return []string{}
    }
    return strings.Split(path, p)
}

func join(parts []string, p string) string {
    return strings.Join(parts, p)
}

func (this *Store) NodeSet(pl *NodeProposal) uint16 {

    // Saving File
    in := nodePathFilter(pl.Key)
    p := split(in, "/")
    l := len(p)
    if pl.Val == NodeDelFlag {
        this.Hdel(INodeFile+in, "v")
        this.Hdel(INodeFile+in, "r")
        this.Srem(INodeDir+join(p[0:l-1], "/"), NodeSepFile+p[l-1])
        return 0
    }

    item := map[string]string{
        "v": pl.Val,
        "r": strconv.FormatUint(pl.VerSet, 10),
    }
    this.Hmset(INodeFile+in, item)

    // Saving DIRs
    for i := l - 1; i >= 0; i-- {
        in = join(p[0:i], "/")
        if i == len(p)-1 {
            this.Sadd(INodeDir+in, NodeSepFile+p[i])
        } else {
            this.Sadd(INodeDir+in, NodeSepDir+p[i])
        }
    }

    return 0
}

func (this *Store) NodeGet(path string) (*Node, error) {

    in := nodePathFilter(path)

    item, e := this.Hgetall(INodeFile + in)
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

func (this *Store) NodeGets(keys string) (string, error) {

    ks := split(keys, " ")

    list := []Node{}

    for _, path := range ks {

        in := nodePathFilter(path)

        item, e := this.Hgetall(INodeFile + in)
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

func (this *Store) NodeList(path string) (string, error) {

    in := nodePathFilter(path)

    item, e := this.Smembers(INodeDir + in)
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

func nodePathFilter(path string) string {

    reg, _ := regexp.Compile("/+")

    return strings.Trim(reg.ReplaceAllString(path, "/"), "/")
}
