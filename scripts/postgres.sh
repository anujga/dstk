#!/bin/sh
set -o errexit

#helm repo add bitnami https://charts.bitnami.com/bitnami
#helm upgrade -i pq bitnami/postgresql
kubectl apply -f charts/pq.yaml


# if you internet is slow, increase timeout but it more likely to be some error
# if it did not succeed in 2 min
kubectl wait  \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/name=postgresql \
  --timeout=120s



POSTGRES_PASSWORD=$(kubectl get secret dstk-psql-postgresql -o jsonpath="{.data.postgresql-password}" | base64 --decode)

kubectl port-forward svc/dstk-psql-postgresql 5432:5432 &
PID=$!
sleep 1
PGPASSWORD="$POSTGRES_PASSWORD" psql  --host 127.0.0.1 -U postgres -d postgres -p 5432 < ../pkg/sharding_engine/simple/schema.sql

kill $PID
wait $PID