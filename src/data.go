
package main

import (
    "fmt"
    "github.com/fzzbt/radix/redis"    
)

type Kpdata struct {
    c   *redis.Client
}

func (db *Kpdata) Initialize() {

    if db.c != nil {
        fmt.Println("db.c exists...")
        return
    }

    conf := redis.DefaultConfig()
    db.c = redis.NewClient(conf)
}

func (db *Kpdata) Set(key string, val string) error {
    
    r := db.c.Set(key, val)
    if r.Err != nil {
        return r.Err
    }

    return nil
}

func (db *Kpdata) Setex(key string, ttl int, val string) error {

    r := db.c.Setex(key, ttl, val)
    if r.Err != nil {
        return r.Err
    }

    return nil
}

func (db *Kpdata) Get(key string) (string, error) {
    return db.c.Get(key).Str()      
}

func (db *Kpdata) Keys(key string) ([]string, error) {
    return db.c.Keys(key).List()
}

func (db *Kpdata) Del(key string) (bool, error) {
    r := db.c.Del(key)
    if r.Err != nil {
        return false, r.Err
    }
    return true, nil
}

func (db *Kpdata) Hget(key string, hkey string) (string, error) {
    r := db.c.Hget(key, hkey)
    if r.Err != nil {
        return "", r.Err
    }
    return r.Str()
}

func (db *Kpdata) Hset(key string, hkey string, hval string) error {
    r := db.c.Hset(key, hkey, hval)
    if r.Err != nil {
        return r.Err
    }
    return nil
}

func (db *Kpdata) Hdel(key string, hkey string) error {
    r := db.c.Hdel(key, hkey)
    if r.Err != nil {
        return r.Err
    }
    return nil
}

func (db *Kpdata) Hmset(key string, val map[string]string) error {
    r := db.c.Hmset(key, val)
    if r.Err != nil {
        return r.Err
    }
    return nil
}

func (db *Kpdata) Hgetall(key string) (map[string]string, error) {
    return db.c.Hgetall(key).Hash()
}

func (db *Kpdata) Incrby(key string, val int) (int, error) {
    return db.c.Incrby(key, val).Int()
}

func (db *Kpdata) Expire(key string, val int) error {
    r := db.c.Expire(key, val)
    if r.Err != nil {
        return r.Err
    }
    return nil
}

func (db *Kpdata) Exists(key string) error {
    r := db.c.Exists(key)
    if r.Err != nil {
        return r.Err
    }
    return nil
}

func (db *Kpdata) Ttl(key string) (int, error) {
    return db.c.Ttl(key).Int()
}
