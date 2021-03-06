![Github Actions](https://github.com/0xmercurial/go-distributed-log-store/actions/workflows/ci.yml/badge.svg)
[![codecov](https://codecov.io/gh/0xmercurial/go-distributed-log-store/branch/main/graph/badge.svg)](https://codecov.io/gh/0xmercurial/go-distributed-log-store)



# Distributed Log Store (in Go)

Iterative development of a distributed log store from scratch in Go.

To test* entire codebase, run 

```
make init gencert test
```
*test takes approx. 30-40 seconds


Principal struct is the `Agent` struct. High-level:
- `Agent`s can be read from/ written to, and replicate/propgate logs to other `Agents` leveraging HashiCorp's Serf package (implements gossip protocol). 
- `Agent`s communicate via gRPC.
- `Agent` code + tests found here: `logstore/internal/agent`

Current State:
- Simple replication via gossip protocol has been implemented.
- Tested using multiple local instances in testing.

To Do: 
- Implement Raft (Hashicorp implemtation)
- DNS
