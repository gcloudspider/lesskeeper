package store

type Host struct {
    Id     string `json:"id"`
    Addr   string `json:"addr"`
    Port   string `json:"port"`
    Status int    `json:"status"`
}
