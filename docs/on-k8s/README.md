# Dstk on K8s

Requirements/Assumptions:
* Partitions are closed-open ranges
* Length of keys is predefined
* A worker (pod) can handle thousands of partitions
* We can create as many CRD objects as the number of partitions

`Partition` would be modelled as a `CustomResourceDefinition` and the following is how a newly created one would look.
```yaml
kind: Partition
metadata:
  name: part1
spec:
  range:
    start: a
    end: o
status:
```

# Assigner
`Assigner` would monitor metrics of the workers(pods) and also watches `Partition` objects to assign pods for partitions. It generates an `Assignment` object as shown:

```yaml
kind: Assignment
metadata:
  name: a1
spec:
  type: New
  newParams:
    partition: part
    primary: pod1
status:
```

# Assignment Reconciler
This would be a kubernetes custom controller that watches `Assignment` objects and tries to reconcile it. In this case, it interacts with pod1 and makes it handle part1.
```yaml
kind: Assignment
metadata:
  name: a1
spec:
  type: New
  newParams:
    partition: part
    primary: pod1
status:
  new: 
    status: completed
```

```yaml
kind: Partition
metadata:
  name: part1
spec:
  range:
    start: a
    end: o
  primary: pod1
status:
  ready
```

Note that this is similar to how a pod runs on a node in k8s. `kube-scheduler` populates `node` field in `PodSpec` which is watched by `kubelet`. `kubelet` makes sure the assigned pods are run on the corresponding node. In our case, `Assigner` plays the role of `kube-scheduler` and `AssignmentReconciler` plays the role of `kubelet`, but not at a node level.

Also, we can think about running an `AssignmentReconciler` in every worker pod as a side car and make it responsible for just that pod, but I haven't worked out pros and cons in that case yet.


## Partition split, merge, move

Following has links to how those operations are modelled:

Note that the following operations have a common theme:
1. Make new partitions follow (load the state) the existing partitions
1. Make the old ones proxy the requests to new ones so that the service is available to stale clients as well.
1. Flip after a while

 
[Splitting a partition in the same worker](split.md)

[Merging partitions in the same worker](merge.md)

[Moving a partition from one worker to another](move.md)
