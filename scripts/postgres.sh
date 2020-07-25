#!/bin/sh
set -o errexit

helm repo add bitnami https://charts.bitnami.com/bitnami
helm upgrade -i pq bitnami/postgresql

# if you internet is slow, increase timeout but most likely it would be some error
# if it did not succeed in 2 min
kubectl wait  \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/name=postgresql \
  --timeout=120s

export POSTGRES_PASSWORD=$(kubectl get secret --namespace default pq-postgresql -o jsonpath="{.data.postgresql-password}" | base64 --decode)

kubectl port-forward --namespace default svc/pq-postgresql 5432:5432

PGPASSWORD="$POSTGRES_PASSWORD" psql  --host 127.0.0.1 -U postgres -d postgres -p 5432 < ../pkg/sharding_engine/simple/schema.sql