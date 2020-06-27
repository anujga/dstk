## Moving a partition from one pod to another

Moving would be a multi-phase operation with the following phases:
1. Make the new pod a follower of existing partition
1. Once the new pod is caught up, we make the new pod primary and old pod as a proxy to new pod so that the service is available to stale clients as well.
1. After a threshold time, old pod is removed from assignments of partition.

Consider a case where `Assigner` decided to move the partition `part1` from pod `srcPod` to `dstPod`.

Current `Partition` definition:
```yaml
kind: Partition
metadata:
  name: part1
spec:
  range:
    start: a
    end: o
  primary: srcPod
status:
  ready
```

`Assigner` would generate an `Assignment` object as shown:
```yaml
kind: Assignment
metadata:
  name: a1
spec:
  # Move could be simpler it makes sense only inter-worker, but picking an explicit name
  type: InterWorkerMove
  moveParams:
    partition: part1
    from: srcPod
    to: dstPod
status:
```

`AssignmentReconciler` will make a request to dstPod to start following part1 and will mark its status accordingly.
```yaml
kind: Assignment
metadata:
  name: a1
spec:
  type: InterWorkerMove
  moveParams:
    partition: part1
    from: srcPod
    to: dstPod
status:
  move:
    status: loading
    fromPod: primary
    toPod: loading
```

Once the state is loaded, Assignment Reconciler would update the status of new pod as follower.
```yaml
status:
  move:
    status: following
    fromPod: primary
    toPod: following
```

`AssignmentReconciler` would make srcPod proxy requests to dstPod for a while.
```yaml
kind: Partition
metadata:
  name: part1
spec:
  range:
    start: a
    end: o
  primary: dstPod
  proxy: srcPod
status:
  ready
```

```yaml
kind: Assignment
metadata:
  name: a1
spec:
  type: InterWorkerMove
  moveParams:
    partition: part1
    from: srcPod
    to: dstPod
status:
  move:
    status: proxying
    fromPod: proxy
    toPod: primary
```

Eventually it makes the dstPod primary
```yaml
kind: Assignment
metadata:
  name: a1
spec:
  type: InterWorkerMove
  moveParams:
    partition: part1
    from: srcPod
    to: dstPod
status:
  move:
    status: completed
    # is there a better word to say srcPod doesn't handle part1 anymore? decommissioned?
    fromPod: disconnected
    toPod: primary
```

```yaml
kind: Partition
metadata:
  name: part1
spec:
  range:
    start: a
    end: o
  primary: dstPod
status:
  ready
```
