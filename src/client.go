package main

import (
    "fmt"
    "net"
    "net/rpc"
    "strconv"
    "os"
    //"errors"
    "strings"
    "time"
    "sync"
)

//const MAX_QUERYBUF_LEN = 1024 * 1024    // 1GB max query buffer
const IOBUF_LEN         = 32
const INLINE_MAX_SIZE   = 1024 * 64      // Max size of inline reads
const AGENT_TIMEOUT     = 3e9

type Agent struct {

    in  chan *Command
    out chan *Command

    //flags int
    //stat_numconnections int
    //maxclients int
    //maxidletime int

    clients map[string]*AgentClient

    Lock sync.Mutex
}


