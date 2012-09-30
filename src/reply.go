package main


const (
    ReplyStatus     uint8 = 1
    ReplyOK         uint8 = 10
    ReplyError      uint8 = 2

    ReplyInteger    uint8 = 3
    ReplyNil        uint8 = 4
    ReplyString     uint8 = 5
    ReplyMulti      uint8 = 6
)

// Reply holds a Command reply.
type Reply struct {
    Type    uint8       // Reply type
    Val     string
    Ver     uint64
    Elems   []*Reply    // Sub-replies
    Err     error       // Reply error
}

