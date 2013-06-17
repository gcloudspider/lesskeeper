package agent

import (
    "../peer"
    "../utils"
    "strconv"
)

type KprInfo struct {
    Leader  string          `json:"leader"`
    Vote    uint64          `json:"vote"`
    Members []KprInfoMember `json:"members"`
    Local   KprInfoLocal    `json:"local"`
}
type KprInfoMember struct {
    Id     string `json:"id"`
    Seat   int    `json:"seat"`
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

    kprID, e := this.stor.Get("ctl:voteid")
    if e == nil {
        if rev, e := strconv.ParseUint(kprID, 10, 64); e == nil {
            info.Vote = rev
        }
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

            kprSeat, _ := strconv.Atoi(k)
            member.Seat = kprSeat

            info.Members = append(info.Members, member)
        }
    }

    loc, e := this.stor.Hgetall("ctl:loc")
    if e == nil {

        if val, ok := loc["node"]; ok {
            info.Local.Id = val
        }
        if val, ok := loc["addr"]; ok {
            info.Local.Addr = val
        }
        if val, ok := loc["port"]; ok {
            info.Local.KeeperPort = val
        }
    }

    rp.Body, e = utils.JsonEncode(info)
    if e != nil {
        rp.Type = peer.ReplyError
    }
}
