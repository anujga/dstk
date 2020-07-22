# DSTK:
Distributed System Toolkit. Collection of modules and reference designs to
 implement stateful services.

# Dev Center
- Build: install using
    ```shell script
    apt  install protobuf-compiler golang-1.14
  
    # Perfer this over makefile for new changes
    https://taskfile.dev/#/installation
    ```
- [Dev Guidlines](docs/dev.md)
- Start SE on port 6001
    `task se`

## Reference Architectures
- [Stateful Services](pkg/ss/README.md)
    - [Memory based counters](examples/mem_counters/memcountes_cmd.go)
    - [MKV](pkg/ss/README.md)
    

## tasks:
1. opentracing @sudip
1. automatic integration test
    1. local perf regression suite
    1. maintain history of test results
1. etag in api
    1. distributed benchmark
1. docker and metrics
    1. performance baseline, cost and latency    
1. state mgmt in badger. successful restart
1. backup, restore

1. test cases around rpc failures via mocks

## todo:
    
- low
    1. make keyT a strongly typed datastructure with proper validation