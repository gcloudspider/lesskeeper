package conf

import (
    "encoding/json"
    "errors"
    "fmt"
    "io/ioutil"
    "os"
    "regexp"
    "strings"
)

type Config struct {
    Prefix       string
    AgentPort    string
    KeeperPort   string
    StoreServer  string
    StoreOption  string
    StoreNetwork string
    StoreAddress string
}

func NewConfig(prefix string) (Config, error) {

    var cfg Config

    if prefix == "" {
        prefix = "/opt/less/keeper"
    }
    reg, _ := regexp.Compile("/+")
    cfg.Prefix = "/" + strings.Trim(reg.ReplaceAllString(prefix, "/"), "/")

    file := cfg.Prefix + "/etc/keeper.json"
    if _, err := os.Stat(file); err != nil && os.IsNotExist(err) {
        return cfg, errors.New("Error: config file is not exists")
    }

    fp, err := os.Open(file)
    if err != nil {
        return cfg, errors.New(fmt.Sprintf("Error: Can not open (%s)", file))
    }
    defer fp.Close()

    cfgstr, err := ioutil.ReadAll(fp)
    if err != nil {
        return cfg, errors.New(fmt.Sprintf("Error: Can not read (%s)", file))
    }

    if err = json.Unmarshal(cfgstr, &cfg); err != nil {
        return cfg, errors.New(fmt.Sprintf("Error: "+
            "config file invalid. (%s)", err.Error()))
    }

    cfg.StoreServer = "less-keeper-store"
    store_server := cfg.Prefix + "/bin/" + cfg.StoreServer
    if _, err := os.Stat(store_server); err != nil && os.IsNotExist(err) {
        return cfg, errors.New(fmt.Sprintf("Error: "+
            "less-keeper-store (%s) is not exists", store_server))
    }

    cfg.StoreNetwork = "unix"
    cfg.StoreAddress = cfg.Prefix + "/var/keeper.sock"

    store_option := cfg.Prefix + "/etc/redis.conf"
    store_option += " --daemonize yes"
    store_option += " --port 9526"
    store_option += " --unixsocket " + cfg.StoreAddress
    store_option += " --dir " + cfg.Prefix + "/var/"
    store_option += " --dbfilename main.rdb"

    cfg.StoreOption = store_option

    return cfg, nil
}
