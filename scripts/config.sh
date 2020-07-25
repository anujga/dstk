#!/bin/bash

kubectl create configmap dstk-config \
  --from-file "$1" -o yaml --dry-run=client \
  | kubectl apply -f -
