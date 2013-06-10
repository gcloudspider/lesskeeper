package store

import (
    "../../deps/radix/redis"
    "../conf"
    "fmt"
    "os/exec"
    "strings"
)

type Store struct {
    c *redis.Client
}

func (this *Store) Initialize(cfg conf.Config) {

    if this.c != nil {
        fmt.Println("this.c exists...")
        return
    }

    pid, err := exec.Command("/bin/pidof", "h5keeper-store").Output()
    if err != nil {
        // TODO
    }
    if string(pid) == "" {
        rdsv := exec.Command(cfg.StoreServer, strings.Fields(cfg.StoreOption)...)
        if err := rdsv.Run(); err != nil {
            fmt.Println(err)
        }
    }

    conf := redis.DefaultConfig()
    conf.PoolCapacity = 10
    conf.Network = cfg.StoreNetwork
    conf.Address = cfg.StoreAddress

    this.c = redis.NewClient(conf)
}

func (this *Store) Set(key string, val string) error {

    r := this.c.Set(key, val)
    if r.Err != nil {
        return r.Err
    }

    return nil
}

func (this *Store) Setex(key string, ttl int, val string) error {

    r := this.c.Setex(key, ttl, val)
    if r.Err != nil {
        return r.Err
    }

    return nil
}

func (this *Store) Get(key string) (string, error) {
    return this.c.Get(key).Str()
}

func (this *Store) Keys(key string) ([]string, error) {
    return this.c.Keys(key).List()
}

func (this *Store) Del(key string) (bool, error) {
    r := this.c.Del(key)
    if r.Err != nil {
        return false, r.Err
    }
    return true, nil
}

func (this *Store) Hget(key string, hkey string) (string, error) {
    r := this.c.Hget(key, hkey)
    if r.Err != nil {
        return "", r.Err
    }
    return r.Str()
}

func (this *Store) Hset(key string, hkey string, hval string) error {
    r := this.c.Hset(key, hkey, hval)
    if r.Err != nil {
        return r.Err
    }
    return nil
}

func (this *Store) Hdel(key string, hkey string) error {
    r := this.c.Hdel(key, hkey)
    if r.Err != nil {
        return r.Err
    }
    return nil
}

func (this *Store) Hmset(key string, val map[string]string) error {
    r := this.c.Hmset(key, val)
    if r.Err != nil {
        return r.Err
    }
    return nil
}

func (this *Store) Hgetall(key string) (map[string]string, error) {
    return this.c.Hgetall(key).Hash()
}

func (this *Store) Smembers(key string) ([]string, error) {
    return this.c.Smembers(key).List()
}

func (this *Store) Sadd(key string, member string) error {
    r := this.c.Sadd(key, member)
    if r.Err != nil {
        return r.Err
    }
    return nil
}
func (this *Store) Srem(key string, member string) error {
    r := this.c.Srem(key, member)
    if r.Err != nil {
        return r.Err
    }
    return nil
}

func (this *Store) Incrby(key string, val int) (int, error) {
    return this.c.Incrby(key, val).Int()
}

func (this *Store) Expire(key string, val int) error {
    r := this.c.Expire(key, val)
    if r.Err != nil {
        return r.Err
    }
    return nil
}

func (this *Store) Exists(key string) error {
    r := this.c.Exists(key)
    if r.Err != nil {
        return r.Err
    }
    return nil
}

func (this *Store) Ttl(key string) (int, error) {
    return this.c.Ttl(key).Int()
}
