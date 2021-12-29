# Distributed Log Store (in Go)

Iterative development of a distributed log store from scratch in Go.

To test, run 

```
make init gencert test
```

Principal struct is the `Agent` struct. High-level:
- `Agent`s replicate logs written to other `Agents` leveraging HashiCorps Serf package (implements gossip protocol). 
- `Agent`s communicate via gRPC.

Current State:
- Simple replication via gossip protocol has been implemented.
- Tested using multiple local instances in testing.

To Do: 
- Implement Raft (Hashicorp implemtation)
- DNS
