# Worker lifecycle

1. Acquire the lease to become the authoritative worker for my workerId. Might
not be required for stateful sets.
1. Fetch all partitions for this worker
1. Load partition map / [actorStore](/pkg/ss/partmgr/state.go) from disk
1. Reconcile missing or deleted partitions in background1

1. Initialize all partitions. for each partition
  1. check the current state as per PE and what exists in reality
  1. warn on divergence
  1. converge towards the current state as per PE.

1. wait for all partitions to converge towards pe.current_state
1. start serving traffic
1. start a background daemon to reconcile desired state on worker level
as well as per partition. 

# Integration Testing

1. Randomly `kill -9` dc process. If testing in local is getting messy, skip it.

1. Add support for https://github.com/chaos-mesh/chaos-mesh

1. Since rpc will start failing, verification of legitimate state transition cannot happen. 
The current algorithm for verify will break. Might need something like jespen. clients can
blob upload the transition functions. a spark job can verify the validity something like 
[knossos](https://github.com/jepsen-io/knossos#concepts)

1. 
