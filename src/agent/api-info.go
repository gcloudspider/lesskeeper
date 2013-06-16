package agent

import (
    "../peer"
    "../utils"
)

type KprInfo struct {
    Leader  string                   `json:"leader"`
    Members map[string]KprInfoMember `json:"members"`
    Local   KprInfoLocal             `json:"local"`
}
type KprInfoMember struct {
    Host       string `json:"host"`
    Addr       string `json:"addr"`
    Status     int    `json:"status"`
    KeeperPort string `json:"keeperport"`
}
type KprInfoLocal struct {
    Host       string `json:"host"`
    Addr       string `json:"addr"`
    Status     int    `json:"status"`
    KeeperPort string `json:"keeperport"`
    AgentPort  string `json:"agentport"`
}

func (this *Agent) apiInfoHandler(method string, body string, rp *peer.Reply) {

    rp.Type = peer.ReplyString

    info := new(KprInfo)
    info.Members = map[string]KprInfoMember{}

    kprLed, err := this.stor.Get("ctl:led")
    if err == nil || len(kprLed) > 4 {
        info.Leader = kprLed
    }

    loc, err := this.stor.Hgetall("ctl:loc")
    if err == nil {

        if _, ok := loc["node"]; ok {
            info.Local.Host = loc["node"]
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
