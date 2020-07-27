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
    - `task se`
    - postgres
    - ```
    helm repo add bitnami https://charts.bitnami.com/bitnami
    helm install pq bitnami/postgresql
    ```

    - 
- dev tools
    - k9s
    - stern
    - kubectl
    - helm3(not helm 2)
    - kube_ps1
    - krew
    - kubectx

## Reference Architectures
- [Stateful Services](pkg/ss/README.md)
    - [Memory based counters](examples/mem_counters/memcountes_cmd.go)
    - [MKV](pkg/ss/README.md)
    


PostgreSQL can be accessed via port 5432 on the following DNS name from within your cluster:

    pq-postgresql.default.svc.cluster.local - Read/Write connection

To get the password for "postgres" run:

    export POSTGRES_PASSWORD=$(kubectl get secret --namespace default pq-postgresql -o jsonpath="{.data.postgresql-password}" | base64 --decode)

To connect to your database run the following command:

    kubectl run pq-postgresql-client --rm --tty -i --restart='Never' --namespace default --image docker.io/bitnami/postgresql:11.8.0-debian-10-r61 --env="PGPASSWORD=$POSTGRES_PASSWORD" --command -- psql --host pq-postgresql -U postgres -d postgres -p 5432



To connect to your database from outside the cluster execute the following commands:

    kubectl port-forward --namespace default svc/pq-postgresql 5432:5432 &
    PGPASSWORD="$POSTGRES_PASSWORD" psql --host 127.0.0.1 -U postgres -d postgres -p 5432
>kgpoowide                                                              [integ↑1|●8✚9…2] /(⎈ |kind-kind:default)
NAME                    READY   STATUS             RESTARTS   AGE     IP            NODE                 NOMINATED NODE   READINESS GATES
dstk-6bb44bb6c9-n49lh   0/1     CrashLoopBackOff   6          7m16s   10.244.0.9    kind-control-plane   <none>           <none>
dstk-76f76fb66d-vlgp5   0/1     CrashLoopBackOff   7          12m     10.244.0.8    kind-control-plane   <none>           <none>
pq-postgresql-0         1/1     Running            0          46s     10.244.0.11   kind-control-plane   <none>           <none>
>POSTGRES_PASSWORD=$(kubectl get secret --namespace default pq-postgresql -o jsonpath="{.data.postgresql-password}" | base64 --decode)
>echo $POSTGRES_PASSWORD                                                               [integ↑1|●8✚9…2] /(⎈ |kind-kind:default)
9iTsWIW7l7
>telnet 10.244.0.11 5432                                                               [integ↑1|●8✚9…2] /(⎈ |kind-kind:default)
Trying 10.244.0.11...
^C
>kubectl port-forward --namespace default svc/pq-postgresql 5432:5432                  [integ↑1|●8✚9…2] /(⎈ |kind-kind:default)
Forwarding from 127.0.0.1:5432 -> 5432
Forwarding from [::1]:5432 -> 5432
^Z
zsh: suspended  kubectl port-forward --namespace default svc/pq-postgresql 5432:5432
>bg                                                                                    [integ↑1|●8✚9…2] /(⎈ |kind-kind:default)
[1]  + continued  kubectl port-forward --namespace default svc/pq-postgresql 5432:5432
>PGPASSWORD="$POSTGRES_PASSWORD" psql --host 127.0.0.1 -U postgres -d postgres -p 5432 [integ↑1|●8✚9…2] /(⎈ |kind-kind:default)
zsh: command not found: psql
>Handling connection for 5432                                                          [integ↑1|●8✚9…2] /(⎈ |kind-kind:default)
Handling connection for 5432
Handling connection for 5432
Handling connection for 5432

>                                                                                     [integ↑1|●8✚10…2] /(⎈ |kind-kind:default)
>./scripts/config.sh configs/small-local                                              [integ↑1|●8✚10…2] /(⎈ |kind-kind:default)
configmap/dstk-config configured
>Handling connection for 5432                                                         [integ↑1|●8✚10…2] /(⎈ |kind-kind:default)
Handling connection for 5432
