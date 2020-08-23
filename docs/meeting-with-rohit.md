# Meeting with Rohit on dstk

Etcd: concensus is OE heavy
Spark RDD for OLTP: HTAP is the modern approach


1. TreeMap - inmemory - btree, rbtree, avltree, skiplist. research ds for OLAP
  1. vertical scalabilty: 
    1. partition the keyspace 
    
    2. ram problem
    3. ssd based datastructure - LSM with btree, skiplist.

    - Horizontal
    - 1.1.1 move to service parition engine
    - chokepoint for qps; thick client 
  
2. Hot key
   1. divide the keyspace in to 10K partitions
   2. 100K parts and 100Nodes (10TB / node) packing 1K partitions. 0.01% of traffic cannot be hot enough to kill a machine
   3. 1GB / partition
   4. Rebalancing

3. Durability
   1. Disk based: 1.1.3 should have durability guarantees
   2. Remote Disk cloud native: - low on iops
   3. Replication :- master slave, paxos, raft, geo replication (epaxos) 
   4. Point in time recovery: 3.3 has this WAL with remote snapshots

4. Transaction
   1. Tunable semantic within partition:
      1. session consistency, eventual consistency, linear serializable
   2. 2PC: Multi phase commit across partitions:

5. Distibution
   1. Policy for cache eviction of LSM tree. index organized table. User reinforcement learning to come up with correct policy
   2. pin users 