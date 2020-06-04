
## Milestones:

### v0.1
Theme: Demo using Disk cache (DC)

1. SE: @subhash
    
    1. Partition Lifecycle. Only add/ remove partition and bootstrap.
    2. Productionize SE server, ram based
    3. [PE](names.md) - earlier called assigned simple python script
    triggered manually that reads from yaml file and invokes SE

2. DC: @harsha

    1. Smart Client for DC @subhash
    1. Grpc Gateway for thin clients. deployable as sidecar or embedded inside
    [SSP](names.md) 
    2. Code: get and put with etag support. TTL. 
    3. docker, prometheus, stateful set, restart workflow
    5. ?? smart client in java
    
3. SS: @gowri
    
    1. Durable [Partition Mgr](/pkg/ss/partition_mgr.go). Help maintain state
    after a restart.
    
    2. [Pipeline](/pkg/ss/README.md#Pipeline) design and plumbing
    
    3. General reusable components for sysbench, simulation, verification
    and demonstration on DC .


### v0.2
Theme: Durability

1. SE:

    1. Partition meta like move, remote url, master, slave
    1. etcd backend

1. DC:

    1. Snapshot and restore
    1. Fault testing

1. SS:

    1. Timestamp
    1. WAL
    1. Lifecycle for backup/restore
    

### v0.3
Theme: SSP



### v0.4
Theme: KV

    
### v0.5
Theme: Load Balance

1. SE:
    
    1. Leader election
    3. PE: feedback loop for load balancing decisions
    2. Move partition lifecycle

1. DC:

    1. Master slave(s) using 2PC. Ensure high success for graceful restarts.
    1. Move Partition

1. SS: @gowri

    1. [2PC]((/pkg/ss/pipeline/two_pc.go)    
    1. Move partition 
    1. Some chaos tools optionally embedded in all apps.


### v0.4
Theme: Chaos


