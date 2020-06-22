# Dstk on K8s

Requirements/Assumptions:
* Partitions are closed-open ranges
* Length of keys is predefined
* A worker (pod) can handle thousands of partitions
* We can create as many CRD objects as the number of partitions

`Partition` would be modelled as a `CustomResourceDefinition` and the following is how a newly created one would look.
```yaml
metadata:
  name: part1
spec:
  range:
    start: a
    end: o
status:
```

# Assigner
Assigner would monitor metrics of the workers and also watches `Partition` objects to assing pods for partitions. It populates `assignments` fields in the spec of partition as shown:

```yaml
metadata:
  name: part1
spec:
  range:
    start: a
    end: o
  assignments:
    - name: pod1
      type: primary
status:
```

# Assignment Reconciler
This would be a kubernetes custom controller that watches `Partition` objects for changes in `assignments` field of `Spec`. On a change, it tries to make the pod respect the specified assignment. It updates the `Status` field of the `Partition` object based on the status of assignment. I made `status` show pod level status so that we can make it more clear when we add replication of data plane etc.
```yaml
metadata:
  name: part1
spec:
  range:
    start: a
    end: o
  assignments:
    - name: pod1
      type: primary
status:
  pod1: primary
```
