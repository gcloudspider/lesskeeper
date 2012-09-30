
package main

import (
    //"fmt"
    //"strconv"
    //"math/rand"
    //"strings"
)

type Command struct {
    Tag     string
    Addr    string
    Type    uint8
    
    Argv    map[int][]byte
    Reply   Reply
}

const (
    CmdSync     uint8 = 0
    CmdAsync    uint8 = 1
)

