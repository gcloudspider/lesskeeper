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

    var conf Config

    if prefix == "" {
        prefix = "/opt/prefix"
    }
    reg, _ := regexp.Compile("/+")
    prefix = "/" + strings.Trim(reg.ReplaceAllString(prefix, "/"), "/")

    file := prefix + "/etc/h5keeper.json"

    if _, err := os.Stat(file); err != nil && os.IsNotExist(err) {
        return conf, errors.New("Error: config file is not exists")
    }

    fp, err := os.Open(file)
    if err != nil {
        return conf, errors.New(fmt.Sprintf("Error: Can not open (%s)", file))
    }
    defer fp.Close()

    confstr, err := ioutil.ReadAll(fp)
    if err != nil {
        return conf, errors.New(fmt.Sprintf("Error: Can not read (%s)", file))
    }

    if err = json.Unmarshal(confstr, &conf); err != nil {
        return conf, errors.New(fmt.Sprintf("Error: "+
            "config file invalid. (%s)", err.Error()))
    }

    store_server := prefix + "/bin/h5keeper-store"
    if _, err := os.Stat(store_server); err != nil && os.IsNotExist(err) {
        return conf, errors.New(fmt.Sprintf("Error: "+
            "h5keeper-store (%s) is not exists", store_server))
    }
    conf.StoreServer = store_server
    conf.StoreNetwork = "unix"
    conf.StoreAddress = prefix + "/var/h5keeper.sock"

    store_option := "--daemonize yes"
    store_option += " --port 9526"
    store_option += " --unixsocket " + conf.StoreAddress
    store_option += " --dir " + prefix + "/var/"
    store_option += " --dbfilename main.rdb"
    conf.StoreOption = store_option

    conf.Prefix = prefix

    return conf, nil
}
