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