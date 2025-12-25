#!/bin/bash
set -e

CLUSTER_NAME="mindhit"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=== MindHit Local K8s (kind) Setup ==="

# Delete existing cluster if exists
kind delete cluster --name $CLUSTER_NAME 2>/dev/null || true

# Create cluster
echo "Creating kind cluster..."
kind create cluster --name $CLUSTER_NAME --config "$SCRIPT_DIR/kind-config.yaml"

# Set kubectl context
kubectl cluster-info --context kind-$CLUSTER_NAME

# Install NGINX Ingress Controller
echo "Installing NGINX Ingress Controller..."
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml

# Wait for Ingress to be ready
echo "Waiting for Ingress Controller to be ready..."
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s

echo ""
echo "=== Kind cluster ready! ==="
echo ""
echo "Add to /etc/hosts:"
echo "127.0.0.1 api.mindhit.local"
echo ""
echo "Run 'moonx infra:kind-deploy' to deploy"
