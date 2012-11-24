
package main

import (
    //"regexp"
    "strings"
    "strconv"
)

const (
    
    NodePathPat     = `[a-zA-Z0-9.\-\/]`

    NodeSepFile     = "x"
    NodeSepDir      = "d"
    NodeSepTmp      = "t"
    
    INodeFile       = "in"+ NodeSepFile +":"
    INodeDir        = "in"+ NodeSepDir +":"

    NodeNil         = uint64(1)
    NodeDir         = uint64(2)    
)

//var pathRe = mustBuildRe(NodePathPat)

type Node struct {
    
    C   string  // Content
    R   uint64  // Revison (100 ~ n)

    // TODO U   uint16  // uid
    // TODO G   uint16  // gid
    // TODO M   uint16  // Mode
}

/*
func mustBuildRe(path string) *regexp.Regexp {
    return regexp.MustCompile(`^/$|^(/` + path + `+)+$`)
}

func checkPath(path string) uint64 {
    if !pathRe.MatchString(path) {
        return NodeNil
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
    in  := strings.Trim(pl.Key, "/")
    item := map[string]string{
        "v": pl.Val,
        "r": strconv.FormatUint(pl.VerSet, 10),
    }
    db.Hmset(INodeFile + in, item)

    // Saving DIRs
    p := split(in, "/")
    for i := len(p) - 1; i >= 0; i-- {
        in = join(p[0:i], "/")
        if i == len(p) - 1 {
            db.Sadd(INodeDir + in, NodeSepFile + p[i])
        } else {
            db.Sadd(INodeDir + in, NodeSepDir + p[i])
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
        if k == "v" {
            node.C = v
        } else if k == "r" {
            if rev, e := strconv.ParseUint(v, 10, 64); e == nil {
                node.R = rev
            }
        }
    }
    return node, nil
}