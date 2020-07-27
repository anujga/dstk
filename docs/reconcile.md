# Partitions Reconcile
Following is how a worker node reconciles partitions

1. Worker node makes a call periodically to SE to fetch the partitions it is responsible for.
1. An actor is created for each newly fetched partition objects. Note that because of this, there could be multiple actors for a key range, so, I've added logic to look for actors in either primary or proxy state to serve a client request.
1. In case of node restarts, we need to start the partitions in the state they were in before the node went down, which will also be taken care by passing an appropriate message to Run()
1. If the desired state of a partition is different from current state of an actor, a message is written to partition mailboxes to tell them to become a different actor. There is a transition table that governs these state transitions.

# Split algo
1. Realise partition A is getting hot which would be in primary state.
1. Create partitions B, C by splitting key range of A and set leader of B, C as A.
1. Wait till both B and C become followers
1. Move A to a proxy state proxying requests to B and C
1. Move B and C to primary state
1. Retire partition A