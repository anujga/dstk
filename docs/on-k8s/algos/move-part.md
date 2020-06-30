
| Legend  | |
| -----------------| --- |
| AC | Assignment Cordinator  | 
| W1 | Worker1 | 
| W2 | Worker2 | 
| P | Partition to be moved | 

## Workflow

1. AC

    1. Update  State
    ```yaml
    workers:
        W1: Primary
        W2: follower
    status:
        W1: Primary
    ```    

2. W2
    1. Create a Partition
       1. Accept Writes. 
       2. Dont apply any writes, just buffer them
       3. Reject Reads
       4. Split to disk would mostly be required. Cap the on disk size
       5. If timeout or disk runs out, `status.W2: abort`
    
    ```yaml
    workers:
        W1: primary + repeater
        W2: follower
    status:
        W1: Primary
        W2: follower

    ```

3. W1
   1. Replicate  all writes to `W2`
      1. Ordering is decided by `W1`
      2. Take snapshot and send to `W1`

   2. State
    ```yaml
    workers:
        W1: primary + repeater
        W2: CaughtUp
    status:
        W1: primary + repeater
        W2: follower
    ```

4. W2
   1. Apply Snapshot
   2. Apply writes
   3. Caughtup
   4. Start serving reads

    ```yaml
    workers:
        W1: NoticePeriod
        W2: Primary
    status:
        W1: Primary
        W2: Read4Primary
    ```
 
5. W1
   1. Fwd all reads and writes to `W2` . Use server side redirects but inform the client about rerouting in the response. TODO: should we use client side redirect 302 ?
   2. Writes are not applied locally
   3. Delete the partition data

    ```yaml
    workers:
        W1: Die
        W2: Primary
    status:
        W1: Retired
        W2: Primary
    ```

6. AC
   1. 1 min after `5`


    ```yaml
    workers:
        W2: Primary
    status:
        W1: Die
        W2: Primary
    ```

6. W1
   1. unregister partition

    ```yaml
    workers:
        W2: Primary
    status:
        W2: Primary
    ```
