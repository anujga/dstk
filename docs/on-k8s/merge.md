# Merging multiple partitions in a worker into one partition

Consider the following partitions that are primarily served from `pod1`
```yaml
kind: Partition
metadata:
  name: part1
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
  name: part2
spec:
  range:
    start: j
    end: o
  primary: pod1
status:
  ready
```

Let's say `Assigner` decided to merge the above partitions and generated an `Assignment` as follows:
```yaml
kind: Assignment
metadata:
  name: a1
spec:
  type: InWorkerMerge
  mergeParams:
    partitions:
      - part1
      - part2
status:
```


`AssignmentReconciler` is responsible for merging the partitions and getting the new merged partition into ready state and terminating the old partitions. Following is how it achieves the same:

It creates a super partition (with lowest key and biggest key of given ranges?) and loads its state from sub partitions.
```yaml
kind: Partition
metadata:
  name: superPart
spec:
  range:
    start: a
    end: o
status:
```

```yaml
kind: Assignment
metadata:
  name: a1
spec:
  type: InWorkerMerge
  mergeParams:
    partitions:
      - part1
      - part2
status:
  merge:
    status: loading 
    mergedPart: loading
    subParts:
      part1: primary
      part2: primary
```

```yaml
status:
  merge:
    status: following 
    mergedPart: follower
    subParts:
      part1: primary
      part2: primary
```

Once the merged partition has become a follower, `AssignmentReconciler` would make the sub partitions proxy requests to merged partition so that the service is available for stale clients as well.

```yaml
kind: Partition
metadata:
  name: superPart
spec:
  range:
    start: a
    end: o
  primary: pod1
status:
  ready
```

```yaml
kind: Partition
metadata:
  name: part1
spec:
  range:
    start: a
    end: e
  proxy: pod1
status:
  ready
```

```yaml
kind: Partition
metadata:
  name: part2
spec:
  range:
    start: j
    end: o
  proxy: pod1
status:
  ready
```

Marks the assignment status accordingly
```yaml
status:
  merge:
    status: proxying
    mergedPart: primary
    subParts:
      part1: proxy
      part2: proxy
```

Eventually, `AssignmentReconciler` would mark the assignment as merged and deletes the old partition objects.
```yaml
kind: Assignment
metadata:
  name: a1
spec:
  type: InWorkerMerge
  mergeParams:
    partitions:
      - part1
      - part2
status:
  merge:
    status: completed
    mergedPart: primary
    subParts:
      part1: disconnected
      part2: disconnected
```
