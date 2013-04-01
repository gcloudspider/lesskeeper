package main

type ReplyType uint8

const (
    ReplyOK      ReplyType = 0
    ReplyError   ReplyType = 1
    ReplyTimeout ReplyType = 2

    ReplyNil     ReplyType = 10
    ReplyInteger ReplyType = 11
    ReplyString  ReplyType = 12
    ReplyJson    ReplyType = 13
    ReplyMulti   ReplyType = 14
    ReplyWatch   ReplyType = 15
)

type Reply struct {
    Err   error
    Type  ReplyType
    Body  string
    Elems []*Reply
}

type ReplyNode struct {
    P string
    C string
    R uint64
}
