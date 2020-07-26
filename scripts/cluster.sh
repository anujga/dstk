#!/bin/sh
set -o errexit

. postgres.sh

# todo: different ingress for prod
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/static/provider/kind/deploy.yaml

kubectl apply -f charts/prom.yaml

kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s

kubectl wait  \
  --for=condition=ready pod/prometheus-dstk-0 \
  --timeout=120s

