package conf

import (
    "errors"
    "os"
    "io/ioutil"
    "encoding/json"
    "fmt"
)


type Config struct {
    AgentPort   string
    KeeperPort  string
    RedisServer string
    RedisOption string
}


func NewConfig(prefix string) (*Config, error) {

    file := prefix + "/etc/h5keeper.json"
    
    fmt.Println("Loading config ("+ file +")")

    if _, err := os.Stat(file); err != nil && os.IsNotExist(err) {
        return nil, errors.New("Error: The config file is not exists!")
    }
    
    fp, err := os.Open(file)
    if err != nil {
        return nil, errors.New("Error: Can not open")
    }
    defer fp.Close()

    cfgjson, err := ioutil.ReadAll(fp)
    if err != nil {
        return nil, errors.New("Error: Can not read")
    }
    
    var cfg Config
    if err = json.Unmarshal(cfgjson, &cfg); err != nil {
        return nil, errors.New(fmt.Sprintf("Error: the config file is invalid. (%s)", err.Error()))
    }
    
    redis_server := prefix +"/bin/redis-server"
    if _, err := os.Stat(redis_server); err != nil && os.IsNotExist(err) {
        return nil, errors.New(fmt.Sprintf("Error: The redis-server (%s) is not exists", redis_server))
    }
    cfg.RedisServer = redis_server
    
    redis_option := ""
    //redis_option += "--port 5500 "
    redis_option += "--unixsocket /tmp/h5keeper.rdsock "
    redis_option += "--dir "+ prefix +"/data/ "
    redis_option += "--dbfilename main.rds "
    redis_option += "--daemonize yes"
    
    cfg.RedisOption = redis_option
    
    return &cfg, nil
}
