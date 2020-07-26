#!/bin/sh
set -o errexit


tmp_dir=$(mktemp -d -t ci-XXXXXXXXXX)
echo $tmp_dir
cd $tmp_dir
git clone https://github.com/coreos/kube-prometheus.git

cd kube-prometheus
kubectl create -f manifests/setup
until kubectl get servicemonitors --all-namespaces
do
  date
  sleep 1
  echo ""
done

kubectl create -f manifests/

kubectl wait \
  --namespace monitoring \
  --for=condition=ready pod \
  --timeout=120s --all


