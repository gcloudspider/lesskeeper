package agent

import (
    "../peer"
    "../utils"
    "strconv"
)

type KprInfo struct {
    Leader  string                   `json:"leader"`
    Members map[string]KprInfoMember `json:"members"`
    Local   KprInfoLocal             `json:"local"`
}
type KprInfoMember struct {
    Id     string `json:"id"`
    Number int    `json:"number"`
    Addr   string `json:"addr"`
    Port   string `json:"port"`
    Status int    `json:"status"`
}
type KprInfoLocal struct {
    Id         string `json:"id"`
    Addr       string `json:"addr"`
    KeeperPort string `json:"keeperport"`
    AgentPort  string `json:"agentport"`
    Status     int    `json:"status"`
}

func (this *Agent) apiInfoHandler(method string, body string, rp *peer.Reply) {

    rp.Type = peer.ReplyString

    info := new(KprInfo)

    kprLed, e := this.stor.Get("ctl:led")
    if e == nil || len(kprLed) > 4 {
        info.Leader = kprLed
    }

    msa, _ := this.stor.Hgetall("ctl:members")
    if len(msa) > 0 {

        for k, v := range msa {

            if len(v) < 4 {
                continue
            }

            rs, e := this.stor.NodeGet("/kpr/ls/" + v)
            if e != nil {
                continue
            }

            var member KprInfoMember
            e = utils.JsonDecode(rs.C, &member)
            if e != nil {
                continue
            }

            if info.Members == nil {
                info.Members = map[string]KprInfoMember{}
            }
            kprNumber, _ := strconv.Atoi(k)
            member.Number = kprNumber

            info.Members[v] = member
        }

    }

    loc, e := this.stor.Hgetall("ctl:loc")
    if e == nil {

        if _, ok := loc["node"]; ok {
            info.Local.Id = loc["node"]
        }
        if _, ok := loc["addr"]; ok {
            info.Local.Addr = loc["addr"]
        }
        if _, ok := loc["port"]; ok {
            info.Local.KeeperPort = loc["port"]
        }
    }

    rp.Body, e = utils.JsonEncode(info)
    if e != nil {
        rp.Type = peer.ReplyError
    }
}
