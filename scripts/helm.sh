#!/bin/sh
set -o errexit


# CRD needs to be installed prior to operator
# https://helm.sh/docs/chart_best_practices/custom_resource_definitions/
#kubectl apply -f https://raw.githubusercontent.com/fluxcd/helm-operator/{{ version }}/deploy/crds.yaml


helm repo add fluxcd https://charts.fluxcd.io

kubectl create namespace flux

helm upgrade -i helm-operator fluxcd/helm-operator \
    --namespace flux \
    --set helm.versions=v3


kubectl wait \
  --namespace flux \
  --for=condition=ready pod \
  --timeout=120s --all

