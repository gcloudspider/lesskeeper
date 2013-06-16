package agent

import (
    "../peer"
    "../utils"
    "strings"
)

type KprInfo struct {
    Leader  string                   `json:"leader"`
    Members map[string]KprInfoMember `json:"members"`
    Local   KprInfoLocal             `json:"local"`
}
type KprInfoMember struct {
    Id         string `json:"id"`
    Addr       string `json:"addr"`
    Status     int    `json:"status"`
    KeeperPort string `json:"keeperport"`
}
type KprInfoLocal struct {
    Id         string `json:"id"`
    Addr       string `json:"addr"`
    Status     int    `json:"status"`
    KeeperPort string `json:"keeperport"`
    AgentPort  string `json:"agentport"`
}

func (this *Agent) apiInfoHandler(method string, body string, rp *peer.Reply) {

    rp.Type = peer.ReplyString

    info := new(KprInfo)
    //info.Members = map[string]KprInfoMember{}

    kprLed, err := this.stor.Get("ctl:led")
    if err == nil || len(kprLed) > 4 {
        info.Leader = kprLed
    }

    ms, err := this.stor.Get("ctl:members")
    if err == nil && len(ms) > 4 {

        msa := strings.Split(ms, ",")
        for _, v := range msa {
            if len(v) < 4 {
                continue
            }

            if info.Members == nil {
                info.Members = map[string]KprInfoMember{}
            }

        }
    }

    loc, err := this.stor.Hgetall("ctl:loc")
    if err == nil {

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

    rp.Body, err = utils.JsonEncode(info)
    if err != nil {
        rp.Type = peer.ReplyError
    }
}
