# Stateful Service

## API
- DataPlane:
    - Upsert : Batch size upto 1MB
    - Get
    - Remove
- Control Plane:
    - AddPartition(id?, range)
    - RmPart(range/id ?)
    - ListPart
    - Snapshot(partition, local path) # sidecar uploads it to remote storage

They are exposed as 2 different grpc service because for control plane
- RateLimits are very low
- timeout is very high
- privileged operations

Service mesh will provide the above functionality


## Components

### Partition (id, range)
- Implements the data plane api in a single goroutine and synchronously. 
    - Significantly hurt performance for simple operations like kv
    lookup, spark shuffle due to context switching overheads. These
    operations are supported with atomic guarantees by the underlying
    datastructures. So thread jumping is simply wasteful.
    
    - Simplify guaranteeing [Linearizability](
    https://en.wikipedia.org/wiki/Linearizability) semantics even for simple
    things like stream join using read then write. Updating the record with non
    commutative operations.
    
    - Increase modularity for operations like binLog, raft.  
    
- Scale target is around 100 r/s. In case of managed disk, it may be lower.
- Snapshot operation is async

### PartitionMgr
Implements control api and the Lifecycle of partition

1. CurrentState = On startup read the current state from disk if present, else
initialize a new state store. 

2. NewState = Connects to SE and gets the snapshot of all partitions that
should be present on this node.

3. Delta = Calculate the change from CurrentState -> NewState. This will give
a list<Op> where Op = Add | Del | Update

4. Opcodes are sent to individual partitions  


### Router
- Uses the registry to route the traffic to correct channel

### Pipeline
Basic reusable components which `transform` the request. It can either be a
filter, enricher, sanitizer (remove sensitive fields). Most of them are not
mandatory. User can mix and match these. Eg: if ur using managed disk for
state store, then WAL and snapshotting may not be required.

Some useful items could be

1. A monotonic clock for this partition

1. WAL: need something for replay. Although the state store will also
contain a similar component, we will still keep it as an decoupled
component for providing flexibility in statestore implementations.
Something simple like mem-counters will not have WAL but it will need a
reliable backup restore mechanism. Possible implementations:
    
    1. Kafka: Extremely reliable and this solution can easily double as a
    flume collector.
    1. Managed Disk: Use a badger data structure and rely on durability

1. Replicate: choice of protocol like Two Phase Commit (2PC), raft, paxos.
For readonly use cases 2PC with multiple slaves might be a better fit. Even
for not so critical data like user history, we may choose something easy
like 2PC gaining automated recovery at the cost of some data loss or
corruption.

1. Dispatcher: Default one guarantees linear serializability. Potential
optimization could be to allow 2 independent queues for reads and writes. This
will result in eventual consistency. We can be smart about this and ensure
requests from same session is queued behind write to give session consistency

1. Line Monitor: Some of these pipeline components are async as they do batching
, rpc, disk io. However, they are still expected to maintain ordering of events. 
Since msgs are already timestamped, this will provide an algorithm similar to
the moving window in tcp protocol to guarantee ordered flow of messages
upstream.


### StateStore
Mostly defined and controlled by application but some OOB KV stores will
snapshot, restore features. 

Caching stores: apart from traditional use cases a local nvme can be used as 
write through cache for data persisted on a managed disk attached to the same vm
. Extra IOPS for this local cache will drastically help read heavy workloads. 
    

## Lifecycle of a request
You can render this diagram in vscode using 
https://marketplace.visualstudio.com/items?itemName=bierner.markdown-mermaid

```mermaid

graph TD

    subgraph Client
        A["find node using key"] --> 
        B["send request to node"]
    end

    subgraph Node
        B --> C{"find partition using key"} 
        C -- "status=BadNode" --> Response
        C -- "found" --> D("enqueue msg to partition")
        D --> E("State Machine")
        E --> F("Generate BinLog")
        F -- "status=Success" --> Response
    end


```


    
## TODO:
- 
