#!/usr/bin/env bash
echo "starting"
set -euo pipefail

echo "🗑️ Tearing down existing cluster..."
kind delete cluster || true

REG_NAME='kind-registry'
export REG_PORT=5001

# Create container registry, if it doesn't exist (When ready for github actions)
if [ "$(docker inspect -f '{{.State.Running}}' "${REG_NAME}" 2>/dev/null || true)" != 'true' ]; then
   docker run \
     -d --restart=always -p "127.0.0.1:${REG_PORT}:5000" --network bridge --name "${REG_NAME}" \
     registry:2
fi

echo "🏗️ Creating new Kind cluster with custom network routing..."
kind create cluster --config kind-config.yaml

# Connect the registry to the cluster network (When ready for github actions)
# (This allows KIND nodes to access the registry container via its container name)
if [ "$(docker inspect -f='{{json .NetworkSettings.Networks.kind}}' "${REG_NAME}")" = 'null' ]; then
  docker network connect "kind" "${REG_NAME}"
fi

# 4. Inject registry configuration to containerd on running nodes
REGISTRY_DIR="/etc/containerd/certs.d/localhost:${REG_PORT}"
REGISTRY_DIR_CONTAINER="/etc/containerd/certs.d/${REG_NAME}:5000"
for node in $(kind get nodes); do
  # Map localhost:5001 to the insecure registry
  docker exec "${node}" mkdir -p "${REGISTRY_DIR}"
  cat <<EOF | docker exec -i "${node}" cp /dev/stdin "${REGISTRY_DIR}/hosts.toml"
server="http://${REG_NAME}:5000"

[host."http://${REG_NAME}:5000"]
  capabilities = ["pull", "resolve"]
EOF

  # Map kind-registry:5000 to the insecure registry (to satisfy direct container name pulls)
  docker exec "${node}" mkdir -p "${REGISTRY_DIR_CONTAINER}"
  cat <<EOF | docker exec -i "${node}" cp /dev/stdin "${REGISTRY_DIR_CONTAINER}/hosts.toml"
server = "http://${REG_NAME}:5000"

[host."http://${REG_NAME}:5000"]
  capabilities = ["pull", "resolve"]
EOF
done


# Document the registry in a ConfigMap
# This tells local tools (and Kubernetes) where to find the registry
envsubst '$REG_PORT' <  configmap_registry.tmpl.yaml | kubectl apply -f -

echo "🛡️ Installing Calico Network Engine..."
kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/v3.26.1/manifests/calico.yaml

echo "🔧 Configuring Calico interface detection for Kind..."
kubectl set env daemonset/calico-node -n kube-system IP_AUTODETECTION_METHOD=interface=eth0

echo "⏳ Waiting for Calico network infrastructure to stabilize..."
kubectl rollout status daemonset/calico-node -n kube-system --timeout=120s

echo "🤖 Installing ArgoCD Engine..."
kubectl create namespace argocd || true
kubectl apply --server-side -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

echo "🔌 Patching ArgoCD Service to use a Cloud LoadBalancer..."
kubectl patch svc argocd-server -n argocd -p '{"spec": {"type": "LoadBalancer"}}'

echo "⏳ Waiting for ArgoCD API Server to be fully functional..."
kubectl rollout status deployment/argocd-server -n argocd --timeout=180s

echo "⚙️ Applying Default ArgoCD AppProject rules..."
kubectl apply -f default-project.yaml

# need to enable github access and load in the demo kafka applications
kubectl apply -f ~/projects/microservices-demo/localfiles/new/repo-credentials.yaml

# echo "🌿 Pre-seeding Root App-of-Apps..."
# # A brief pause ensures the webhook configurations are completely settled
# sleep 5 
# kubectl apply -f cluster-gitops/bootstrap/root-application.yaml
# sleep 30
# kind load image-archive localfiles/service-a.tar
# kind load image-archive localfiles/service-b.tar
# kubectl apply -f localfiles/secrets/default-secrets.yaml
# kubectl apply -f localfiles/secrets/infra-messaging-secrets.yaml

# Install the Ingress to mimic a cloud load balancer
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
helm install ingress-nginx ingress-nginx/ingress-nginx --namespace ingress-nginx --create-namespace -f ingress-values.yaml


kubectl apply --server-side -f https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.30/releases/cnpg-1.30.0.yaml
kubectl apply -f terraform/onprem/postgres_cluster.yaml

echo "🚀 Cluster bootstrap 100% complete!"
echo "--------------------------------------------------"
echo "Next steps: Run your local port-forward command for port 8080"
echo "and apply your root App-of-Apps manifest to kick off the sync!"

# Install cloud-provider-kind to get load balancer access

#KIND_EXPERIMENTAL_PROVIDER=podman DOCKER_HOST=unix://$XDG_RUNTIME_DIR/podman/podman.sock cloud-provider-kind