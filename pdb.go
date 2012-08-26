
package main

import (
    "fmt"
    "github.com/fzzbt/radix/redis"    
)

type PdbConn struct {
    c   *redis.Client
}

func (pdb *PdbConn) init() {

    if pdb.c != nil {
        fmt.Println("pdb.c exists...")
        return
    }

    conf := redis.DefaultConfig()
    pdb.c = redis.NewClient(conf)
}

func (pdb *PdbConn) Set(key string, val string) error {
    
    r := pdb.c.Set(key, val)
    if r.Err != nil {
        return r.Err
    }

    return nil
}

func (pdb *PdbConn) Setex(key string, ttl int, val string) error {
    r := pdb.c.Setex(key, ttl, val)
    if r.Err != nil {
        return r.Err
    }

    return nil
}

func (pdb *PdbConn) Get(key string) (string, error) {
    return pdb.c.Get(key).Str()      
}

func (pdb *PdbConn) Keys(key string) (map[string]string, error) {
    return pdb.c.Keys(key).Hash()
}

func (pdb *PdbConn) Del(key string) (bool, error) {
    r := pdb.c.Del(key)
    if r.Err != nil {
        return false, r.Err
    }
    return true, nil
}

func (pdb *PdbConn) Hget(key string, hkey string) (string, error) {
    r := pdb.c.Hget(key, hkey)
    if r.Err != nil {
        return "", r.Err
    }
    return r.Str()
}

func (pdb *PdbConn) Hset(key string, hkey string, hval string) error {
    r := pdb.c.Hset(key, hkey, hval)
    if r.Err != nil {
        return r.Err
    }
    return nil
}

func (pdb *PdbConn) Hdel(key string, hkey string) error {
    r := pdb.c.Hdel(key, hkey)
    if r.Err != nil {
        return r.Err
    }
    return nil
}

func (pdb *PdbConn) Hmset(key string, val map[string]string) error {
    r := pdb.c.Hmset(key, val)
    if r.Err != nil {
        return r.Err
    }
    return nil
}

func (pdb *PdbConn) Hgetall(key string) (map[string]string, error) {
    return pdb.c.Hgetall(key).Hash()
}

func (pdb *PdbConn) Incrby(key string, val int) (int, error) {
    return pdb.c.Incrby(key, val).Int()
}

func (pdb *PdbConn) Expire(key string, val int) error {
    r := pdb.c.Expire(key, val)
    if r.Err != nil {
        return r.Err
    }
    return nil
}

func (pdb *PdbConn) Ttl(key string) (int, error) {
    return pdb.c.Ttl(key).Int()
}
