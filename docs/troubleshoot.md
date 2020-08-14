## se
#### panic: pq: password authentication failed for user "postgres"
helm chart upgrades can accidentally update the secret without changing the
underlying storage. so readers of the secret will find wrong value. helm remove
also does not work because it does not remove the pvc. need to delete helm chart
and the pvc followed by reinstallation.