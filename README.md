
**THIS IS A DEVELOPMENT PREVIEW - DO NOT USE IT IN PRODUCTION!**

## What is Hooto Keeper ?
 * High reliable distributed coordination service
 * Similar to Google Chubby, Apache ZooKeeper
 * Open Source, lightweight implementation in Go

## Architecture
<pre><code>
/---------Client---------\                       /---------Server----------\

APIs <=> http/json <=> Agent <--- PRC/UDP ---> Proposer <== RPC/UDP ==> Acceptor
                         ^                        ^                        ^
                         |                        |                        |
                         v                        v                        v
                       Redis                    Redis                    Redis
</code></pre>

## Similar or Reference Projects
 * Google Chubby <http://research.google.com/archive/chubby.html>
 * Apache Zookeeper <http://zookeeper.apache.org/>
 * Doozer <https://github.com/ha/doozerd>

