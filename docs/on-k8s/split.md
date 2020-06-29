Consider a partition `part` that's primarily served from `pod1`
```yaml
kind: Partition
metadata:
  name: part
spec:
  range:
    start: a
    end: o
  primary: pod1
status:
  ready
```

Let's say `part` is getting hot and `Assigner` decided to split it at the key `e` and generated an `Assignment` as follows:
```yaml
kind: Assignment
metadata:
  name: a1
spec:
  type: InWorkerSplit
  splitParams:
    partition: part
    # should we think about splitting at multiple points?
    splitKey: e
status:
```


`AssignmentReconciler` is responsible for getting the split partitions into ready state and terminating the parent partition. Following is how it achieves the same:
 
It makes the sub-partitions load the state from parent partition.
```yaml
kind: Assignment
metadata:
  name: a1
spec:
  type: InWorkerSplit
  splitParams:
    partition: part
    splitKey: e
status:
  split:
    status: loading
    fromPart: primary
    splits:
      # loading means the state is being streamed from part to subPart1
      subPart1: loading
      # following means subPart2 is caught up with part
      subPart2: following      
```

```yaml
status:
  split:
    status: following
    fromPart: primary
    splits:
      subPart1: following
      subPart2: following
```

It would have been ideal to make the parent partition proxy requests to split partition for sometime, but that requires us to keep the multiplexing logic in parent partition. Dropping the idea for now and assuming a single flip:
```yaml
kind: Assignment
metadata:
  name: a1
spec:
  type: InWorkerSplit
  splitParams:
    partition: part
    splitKey: e
status:
  split:
    status: completed
    fromPart: disconnected
    splits:
      subPart1: primary
      subPart2: primary
```

Partition object for `part` would be deleted and created for splits as shown:
```yaml
kind: Partition
metadata:
  name: subPart1
spec:
  range:
    start: a
    end: e
  primary: pod1
status:
  ready
```

```yaml
kind: Partition
metadata:
  name: subPart2
spec:
  range:
    start: e
    end: o
  primary: pod1
status:
  ready
```
