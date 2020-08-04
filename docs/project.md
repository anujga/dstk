
## Milestones:

### v0.1
Theme: Production Hardening. Basic kv store


1. SE: @subhash
    
    3. [PE](names.md) - earlier called assigned simple python script
    triggered manually that reads from yaml file and invokes SE

2. DC: @harsha

    2. Code: get and put with etag support. TTL. 
    
3. SS: @gowri
    
    1. Durable [Partition Mgr](/pkg/ss/partition_mgr.go). Help maintain state
    after a restart.
    

### v0.3
1. Backup and restore
1. Partition Move (manual trigger)
1. Metric server for partition data.

    
### v0.5
1. Leader election
3. PE: feedback loop for load balancing decisions
2. Move partition lifecycle





