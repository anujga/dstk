## Moving a partition from one pod to another

Moving would be a multi-phase operation with the following phases:
1. Make the new pod a follower of existing partition
1. Once the new pod is caught up, we make the new pod primary and old pod as a proxy to new pod so that the service is available to stale clients as well.
1. After a threshold time, old pod is removed from assignments of partition.

Consider a case where `Assigner` decided to move the partition `part1` from pod `srcPod` to `dstPod`.

Current `Partition` definition:
```yaml
metadata:
  name: part1
spec:
  range:
    start: a
    end: o
  assignedPods:
    - name: srcPod
      status: primary
status:
  srcPod: primary
```

`Assigner` would update the `assignments` field as follows:
```yaml
metadata:
  name: part1
spec:
  range:
    start: a
    end: o
  assignedPods:
    - name: srcPod
      type: primary
    - name: dstPod
      type: follower
status:
  srcPod: primary
```

Assignment Reconciler will make a request to dstPod to start following part1 and will mark its status as starting.
```yaml
status:
  srcPod: primary
  dstPod: starting
```

Once the state is loaded, Assignment Reconciler would update the status of new pod as follower.
```yaml
status:
  srcPod: primary
  dstPod: follower
```

Assigner would make dstPod primary when it sees that the new pod is following and will make the old pod a proxy.
```yaml
metadata:
  name: part1
spec:
  range:
    start: a
    end: o
  assignedPods:
    - name: srcPod
      type: proxy
    - name: dstPod
      type: primary
status:
  srcPod: primary
  dstPod: follower
```

Assignment Reconciler would interact with corresponding pods and reconcile them.
```yaml
status:
  srcPod: proxy
  dstPod: primary
```

After a specified time, old pod would be removed from assignments.
```yaml
metadata:
  name: part1
spec:
  range:
    start: a
    end: o
  assignedPods:
    - name: dstPod
      type: primary
status:
  srcPod: proxy
  dstPod: primary
```

Assignment Reconciler would make the old pod not handle the partition.
```yaml
status:
  srcPod: unassigning
  dstPod: primary
```

Eventually:
```yaml
metadata:
  name: part1
spec:
  range:
    start: a
    end: o
  assignedPods:
    - name: dstPod
      type: primary
status:
  dstPod: primary
```