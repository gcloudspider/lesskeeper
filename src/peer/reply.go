package peer

const (
    ReplyOK    uint8 = 1
    ReplyError uint8 = 2

    ReplyInteger uint8 = 3
    ReplyNil     uint8 = 4
    ReplyString  uint8 = 5
    ReplyMulti   uint8 = 6
    ReplyWatch   uint8 = 7
)

//#define REDIS_REPLY_STRING 1
//#define REDIS_REPLY_ARRAY 2
//#define REDIS_REPLY_INTEGER 3
//#define REDIS_REPLY_NIL 4
//#define REDIS_REPLY_STATUS 5
//#define REDIS_REPLY_ERROR 6

// Reply holds a Command reply.
type Reply struct {
    Status uint8 // Reply status
    Err    error // Reply error

    Type uint8 // Reply type
    Val  string
    Ver  uint64

    Elems []*Reply // Sub-replies
}
