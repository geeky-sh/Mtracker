#!/usr/bin/env bash
# Deploy all Kubernetes manifests. Run this after building and pushing your Docker image.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

REGISTRY="${1:-ghcr.io/aash/mtracker-api}"
TAG="${2:-latest}"

echo "==> Building API Docker image: $REGISTRY:$TAG"
docker build -t "$REGISTRY:$TAG" apps/api

echo "==> Pushing image…"
docker push "$REGISTRY:$TAG"

echo "==> Applying Kubernetes manifests…"
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/postgres/pvc.yaml
kubectl apply -f k8s/postgres/secret.yaml       # must exist (copy from secret.example.yaml)
kubectl apply -f k8s/postgres/deployment.yaml
kubectl apply -f k8s/postgres/service.yaml
kubectl apply -f k8s/api/configmap.yaml
kubectl apply -f k8s/api/secret.yaml            # must exist (copy from secret.example.yaml)
kubectl apply -f k8s/api/deployment.yaml
kubectl apply -f k8s/api/service.yaml

echo "==> Done. Check status with: kubectl get all -n mtracker"
