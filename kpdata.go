
package main

import (
    "fmt"
    "github.com/fzzbt/radix/redis"    
)

type Kpdata struct {
    c   *redis.Client
}

func (kpd *Kpdata) Initialize() {

    if kpd.c != nil {
        fmt.Println("kpd.c exists...")
        return
    }

    conf := redis.DefaultConfig()
    kpd.c = redis.NewClient(conf)
}

func (kpd *Kpdata) Set(key string, val string) error {
    
    r := kpd.c.Set(key, val)
    if r.Err != nil {
        return r.Err
    }

    return nil
}

func (kpd *Kpdata) Setex(key string, ttl int, val string) error {
    r := kpd.c.Setex(key, ttl, val)
    if r.Err != nil {
        return r.Err
    }

    return nil
}

func (kpd *Kpdata) Get(key string) (string, error) {
    return kpd.c.Get(key).Str()      
}

func (kpd *Kpdata) Keys(key string) (map[string]string, error) {
    return kpd.c.Keys(key).Hash()
}

func (kpd *Kpdata) Del(key string) (bool, error) {
    r := kpd.c.Del(key)
    if r.Err != nil {
        return false, r.Err
    }
    return true, nil
}

func (kpd *Kpdata) Hget(key string, hkey string) (string, error) {
    r := kpd.c.Hget(key, hkey)
    if r.Err != nil {
        return "", r.Err
    }
    return r.Str()
}

func (kpd *Kpdata) Hset(key string, hkey string, hval string) error {
    r := kpd.c.Hset(key, hkey, hval)
    if r.Err != nil {
        return r.Err
    }
    return nil
}

func (kpd *Kpdata) Hdel(key string, hkey string) error {
    r := kpd.c.Hdel(key, hkey)
    if r.Err != nil {
        return r.Err
    }
    return nil
}

func (kpd *Kpdata) Hmset(key string, val map[string]string) error {
    r := kpd.c.Hmset(key, val)
    if r.Err != nil {
        return r.Err
    }
    return nil
}

func (kpd *Kpdata) Hgetall(key string) (map[string]string, error) {
    return kpd.c.Hgetall(key).Hash()
}

func (kpd *Kpdata) Incrby(key string, val int) (int, error) {
    return kpd.c.Incrby(key, val).Int()
}

func (kpd *Kpdata) Expire(key string, val int) error {
    r := kpd.c.Expire(key, val)
    if r.Err != nil {
        return r.Err
    }
    return nil
}

func (kpd *Kpdata) Ttl(key string) (int, error) {
    return kpd.c.Ttl(key).Int()
}
