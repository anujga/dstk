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