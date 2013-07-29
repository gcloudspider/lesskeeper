package store

import (
    "encoding/json"
)

const (
    ILocalNodeFile = "lo" + NodeSepFile + ":"
    ILocalNodeDir  = "lo" + NodeSepDir + ":"
)

func (this *Store) LocalNodeSet(pl *NodeProposal) uint16 {

    // Saving File
    in := nodePathFilter(pl.Key)
    p := split(in, "/")
    l := len(p)
    if pl.Val == NodeDelFlag {
        this.Del(ILocalNodeFile + in)
        this.Srem(ILocalNodeDir+join(p[0:l-1], "/"), NodeSepFile+p[l-1])
        return 0
    }

    if pl.Ttl < 1 {
        pl.Ttl = 86400
    } else if pl.Ttl > 86400*30 {
        pl.Ttl = 86400 * 30
    }

    this.Setex(ILocalNodeFile+in, pl.Ttl, pl.Val)

    // Saving DIRs
    for i := l - 1; i >= 0; i-- {
        in = join(p[0:i], "/")
        if i == len(p)-1 {
            this.Sadd(ILocalNodeDir+in, NodeSepFile+p[i])
        } else {
            this.Sadd(ILocalNodeDir+in, NodeSepDir+p[i])
        }
    }

    return 0
}

func (this *Store) LocalNodeGet(path string) (*Node, error) {

    in := nodePathFilter(path)

    v, e := this.Get(ILocalNodeFile + in)
    if e != nil {
        return nil, e
    }

    node := new(Node)
    node.C = v

    return node, nil
}

func (this *Store) LocalNodeList(path string) (string, error) {

    in := nodePathFilter(path)

    item, e := this.Smembers(ILocalNodeDir + in)
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

    if rs, e := json.Marshal(list); e == nil {
        return string(rs), nil
    }

    return "", nil
}
