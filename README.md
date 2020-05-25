# DSTK:
Distributed System Toolkit. Collection of modules and reference designs to
 implement stateful services.

# Dev Center
- Build: install using
    ```shell script
    apt  install protobuf-compiler golang-1.14
  
    # Perfer this over makefile for new changes
    go install -v github.com/go-task/task/cmd/task
    ```
- [Dev Guidlines](docs/dev.md)

## Reference Architectures
- [Stateful Services](docs/stateful-service.md)
    - [Memory based counters](examples/mem_counters/memcountes_cmd.go)
    - [MKV](pkg/mkv/README.md)