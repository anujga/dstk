### Tasks
1. state mgmt in badger. successful restart
    1. tests for durability
1. make keyT a strongly typed datastructure with proper validation

### Testing
1. automatic integration test
    1. workflow se, dstk 1, 2, 3; verify; stop; cleanup
    1. local perf regression suite
    1. maintain history of test results
1. test cases around rpc failures via mocks

        
### Api
1. etag
1. scan
1. bulk insert api. The semantic is weird. route to the
partition of the first key and insert whatever is possible. 
return the successful items. will only be useful if keys 
are already sorted. 2 use cases would be
    1. spark ingestion, where spark can do sorting
    2. restore where dumped files are already sorted

### Production
1. opentracing @sudip
1. docker and metrics
    1. distributed benchmark
    1. performance baseline, cost and latency
1. enable metrics
    1. dashboards and alerts
    1. elk integration with zap logs
    1. partition level stats
    1. sentry for errors
        

### Features
1. backup, restore - spark job that makes bulk rpc calls
and upload data to blob. Similarly, restore.

1. Metric server
    1. Maintain partition level metrics
    1. approx count of entries, count partitions. These 2 will help spark job
    in sizing for backup restore.

### Client
1. Gateway mode in addition to thick client. gateway vs thick client
    1. Gateway will give us more control than thick clients.
    1. Language agnostic. Mandatory since we dont have a thick client 
    in java

1. Add utility methods to keep the main server remain thin. Not sure about
the use cases for these api except ttl:
    1. Implement TTL cleanup logic
    1. ScanEx which will retry and give you all elements requested
    1. BulkPutEx again which will keep retrying and insert all entries
    1. expose grpc streaming api for these
    1. documentdb with partial get/put

