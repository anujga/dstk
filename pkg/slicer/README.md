## Use cases:

- envoy plugin
- memory+disk cache
  see https://github.com/golang/groupcache
  for list of issues in caching. like
  this avoids the thundering herd naturally.
- mkv: disk store
    
## Assigner
### Move Algorithm
- TBD

## Project plan:
- [x] Client that polls for changes
- etcd integration + persistence
- Client with server push
- assigner that can with 
    - add/rm partition
    - Move partition