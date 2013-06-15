package agent

import (
    "../peer"
    "../store"
    "../utils"
)

func (this *Agent) apiLocalHandler(method string, body string, rp *peer.Reply) {

    switch string(method) {
    case "LOCGET":
        this.apiLocalGet(body, rp)
    case "LOCLIST":
        this.apiLocalList(body, rp)
    case "LOCSET":
        this.apiLocalSet(body, rp, false)
    case "LOCDEL":
        this.apiLocalSet(body, rp, true)
    }
}

func (this *Agent) apiLocalGet(body string, rp *peer.Reply) {

    var rqbody struct {
        Path string
    }
    e := utils.JsonDecode(body, &rqbody)
    if e != nil {
        return
    }

    if node, e := this.stor.LocalNodeGet(rqbody.Path); e == nil {
        rp.Type = peer.ReplyString
        rp.Body = node.C
    }
}

func (this *Agent) apiLocalSet(body string, rp *peer.Reply, del bool) {

    var rqbody struct {
        Path string
        Val  string
        Ttl  int
    }
    e := utils.JsonDecode(body, &rqbody)
    if e != nil {
        return
    }
    if del {
        rqbody.Val = store.NodeDelFlag
    }

    pl := new(store.NodeProposal)
    pl.Key = rqbody.Path
    pl.Val = rqbody.Val
    pl.Ttl = rqbody.Ttl

    // TODO
    _ = this.stor.LocalNodeSet(pl)

    rp = new(peer.Reply)
    rp.Type = peer.ReplyString
}

func (this *Agent) apiLocalList(body string, rp *peer.Reply) {

    var rqbody struct {
        Path string
    }
    e := utils.JsonDecode(body, &rqbody)
    if e != nil {
        return
    }

    if rs, e := this.stor.LocalNodeList(rqbody.Path); e == nil {
        rp.Type = peer.ReplyString
        rp.Body = rs
    }
}
