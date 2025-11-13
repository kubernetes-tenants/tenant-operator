#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Deploy Lynq to Minikube ===${NC}"
echo ""

# Configuration
PROFILE="${MINIKUBE_PROFILE:-lynq}"
NAMESPACE="${OPERATOR_NAMESPACE:-lynq-system}"

# Generate timestamp-based image tag for development
# This ensures each build is treated as a new image by Kubernetes
if [ -z "$IMG" ]; then
    TIMESTAMP=$(date +%Y%m%d-%H%M%S)
    IMG="lynq:dev-$TIMESTAMP"
    echo -e "${BLUE}Generated development image tag: $IMG${NC}"
else
    echo -e "${BLUE}Using provided image tag: $IMG${NC}"
fi

# Check if minikube is installed
if ! command -v minikube &> /dev/null; then
    echo -e "${RED}Error: minikube is not installed${NC}"
    exit 1
fi

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}Error: kubectl is not installed${NC}"
    exit 1
fi

# Check if minikube is running
if ! minikube status -p "$PROFILE" &> /dev/null; then
    echo -e "${RED}Error: Minikube cluster '$PROFILE' is not running${NC}"
    echo ""
    echo "Start the cluster first:"
    echo "  ./scripts/setup-minikube.sh"
    exit 1
fi

# Force switch to minikube context
MINIKUBE_CONTEXT="$PROFILE"
CURRENT_CONTEXT=$(kubectl config current-context 2>/dev/null || echo "none")

if [ "$CURRENT_CONTEXT" != "$MINIKUBE_CONTEXT" ]; then
    echo -e "${YELLOW}Current context: $CURRENT_CONTEXT${NC}"
    echo -e "${YELLOW}Switching to minikube context: $MINIKUBE_CONTEXT${NC}"
    kubectl config use-context "$MINIKUBE_CONTEXT"
    echo -e "${GREEN}✓ Context switched${NC}"
else
    echo -e "${GREEN}✓ Already using minikube context: $MINIKUBE_CONTEXT${NC}"
fi

# Verify context is minikube
CURRENT_CONTEXT=$(kubectl config current-context)
if [ "$CURRENT_CONTEXT" != "$MINIKUBE_CONTEXT" ]; then
    echo -e "${RED}Error: Failed to switch to minikube context${NC}"
    exit 1
fi

echo ""
echo -e "${BLUE}Configuration:${NC}"
echo "  Profile:    $PROFILE"
echo "  Context:    $CURRENT_CONTEXT"
echo "  Image:      $IMG"
echo "  Namespace:  $NAMESPACE"
echo ""

# Clean up old development images in minikube (keep last 3)
echo ""
echo -e "${YELLOW}Cleaning up old development images in minikube...${NC}"
OLD_IMAGES=$(minikube -p "$PROFILE" image ls 2>/dev/null | grep "lynq:dev-" | sort -r | tail -n +4 || echo "")
if [ -n "$OLD_IMAGES" ]; then
    echo "$OLD_IMAGES" | while read -r old_img; do
        echo "  Removing: $old_img"
        minikube -p "$PROFILE" image rm "$old_img" 2>/dev/null || true
    done
    echo -e "${GREEN}✓ Old images cleaned up${NC}"
else
    echo -e "${GREEN}✓ No old images to clean up${NC}"
fi

# Build the operator image locally
echo ""
echo -e "${YELLOW}Building operator image locally...${NC}"
make docker-build IMG="$IMG"
echo -e "${GREEN}✓ Image built: $IMG${NC}"

# Load image into minikube
echo ""
echo -e "${YELLOW}Loading image into minikube...${NC}"
minikube -p "$PROFILE" image load "$IMG"
echo -e "${GREEN}✓ Image loaded into minikube${NC}"

# Verify image exists in minikube
echo ""
echo -e "${YELLOW}Verifying image in minikube...${NC}"
if minikube -p "$PROFILE" image ls | grep -q "$IMG"; then
    echo -e "${GREEN}✓ Image verified in minikube${NC}"
else
    echo -e "${RED}Error: Image not found in minikube${NC}"
    echo ""
    echo "Available images in minikube:"
    minikube -p "$PROFILE" image ls | head -10
    exit 1
fi

# Install or update CRDs
echo ""
echo -e "${YELLOW}Installing/Updating CRDs...${NC}"
make install
echo -e "${GREEN}✓ CRDs installed${NC}"

# Ensure cert-manager is ready
echo ""
echo -e "${YELLOW}Checking cert-manager status...${NC}"
if ! kubectl get namespace cert-manager &> /dev/null; then
    echo -e "${RED}Error: cert-manager is not installed${NC}"
    echo ""
    echo "Please install cert-manager first:"
    echo "  kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.2/cert-manager.yaml"
    echo ""
    echo "Or recreate the cluster with cert-manager:"
    echo "  ./scripts/cleanup-minikube.sh"
    echo "  ./scripts/setup-minikube.sh"
    exit 1
fi

echo -e "${YELLOW}Waiting for cert-manager to be ready...${NC}"
kubectl wait --namespace cert-manager \
    --for=condition=ready pod \
    --selector=app.kubernetes.io/instance=cert-manager \
    --timeout=300s
echo -e "${GREEN}✓ cert-manager is ready${NC}"

# Check if operator is already deployed
echo ""
EXISTING_DEPLOYMENT=$(kubectl get deployment -n "$NAMESPACE" -l control-plane=controller-manager -o name 2>/dev/null || echo "")

if [ -n "$EXISTING_DEPLOYMENT" ]; then
    echo -e "${YELLOW}Existing deployment found, performing rolling update...${NC}"

    # Delete the deployment to force recreation with new image
    echo -e "${YELLOW}Deleting existing deployment...${NC}"
    kubectl delete deployment -n "$NAMESPACE" -l control-plane=controller-manager --ignore-not-found=true

    # Wait for pods to terminate
    echo -e "${YELLOW}Waiting for old pods to terminate...${NC}"
    kubectl wait --for=delete pod -n "$NAMESPACE" -l control-plane=controller-manager --timeout=60s 2>/dev/null || true

    echo -e "${GREEN}✓ Old deployment removed${NC}"
fi

# Deploy the operator
echo ""
echo -e "${YELLOW}Deploying operator...${NC}"
make deploy IMG="$IMG"
echo -e "${GREEN}✓ Operator deployed${NC}"

# Patch imagePullPolicy to IfNotPresent (use local image if available)
echo ""
echo -e "${YELLOW}Patching imagePullPolicy to IfNotPresent (prefer local image)...${NC}"
DEPLOYMENT_NAME=$(kubectl get deployment -n "$NAMESPACE" -l control-plane=controller-manager -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
if [ -n "$DEPLOYMENT_NAME" ]; then
    kubectl patch deployment "$DEPLOYMENT_NAME" -n "$NAMESPACE" \
        --type='json' \
        -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/imagePullPolicy", "value": "IfNotPresent"}]'
    echo -e "${GREEN}✓ ImagePullPolicy set to IfNotPresent${NC}"
else
    echo -e "${RED}Error: Could not find deployment${NC}"
    exit 1
fi

# Wait for webhook certificate to be ready
echo ""
echo -e "${YELLOW}Waiting for webhook certificate to be ready...${NC}"
kubectl wait --namespace "$NAMESPACE" \
    --for=condition=Ready certificate \
    --selector=app.kubernetes.io/component=webhook \
    --timeout=120s 2>/dev/null || echo -e "${YELLOW}Certificate not found or not ready yet${NC}"

# Wait for operator to be ready
echo ""
echo -e "${YELLOW}Waiting for operator to be ready...${NC}"
kubectl wait --for=condition=Available deployment "$DEPLOYMENT_NAME" -n "$NAMESPACE" --timeout=300s

# Check pod status
echo ""
echo -e "${GREEN}✓ Operator is ready!${NC}"
echo ""
echo -e "${BLUE}Deployment Status:${NC}"
kubectl get deployment -n "$NAMESPACE" -l control-plane=controller-manager
echo ""
echo -e "${BLUE}Pod Status:${NC}"
kubectl get pods -n "$NAMESPACE" -l control-plane=controller-manager

# Show recent logs
echo ""
echo -e "${BLUE}Recent Logs (last 20 lines):${NC}"
POD_NAME=$(kubectl get pods -n "$NAMESPACE" -l control-plane=controller-manager -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
if [ -n "$POD_NAME" ]; then
    kubectl logs -n "$NAMESPACE" "$POD_NAME" --tail=20 --all-containers || true
else
    echo -e "${YELLOW}No pods found${NC}"
fi

# Display useful commands
echo ""
echo -e "${GREEN}=== Deployment Complete ===${NC}"
echo ""
echo -e "${BLUE}Deployed Image:${NC}"
echo "  $IMG"
echo ""
echo -e "${BLUE}Useful Commands:${NC}"
echo "  Watch pods:              kubectl get pods -n $NAMESPACE -w"
echo "  View logs:               kubectl logs -n $NAMESPACE -l control-plane=controller-manager -f --all-containers"
echo "  Check CRDs:              kubectl get crd | grep lynq"
echo "  List registries:         kubectl get lynqhubs -A"
echo "  List templates:          kubectl get lynqforms -A"
echo "  List nodes:              kubectl get lynqnodes -A"
echo "  Describe operator:       kubectl describe deployment -n $NAMESPACE -l control-plane=controller-manager"
echo "  List images:             minikube -p $PROFILE image ls | grep lynq"
echo ""
echo -e "${BLUE}Apply samples:${NC}"
echo "  kubectl apply -f config/samples/"
echo ""
echo -e "${BLUE}Redeploy after code changes:${NC}"
echo "  ./scripts/deploy-to-minikube.sh  # Builds and deploys with fresh timestamp tag"
echo ""
echo -e "${BLUE}Note:${NC}"
echo "  - Each deployment uses a unique timestamp-based image tag"
echo "  - This ensures Kubernetes always uses the latest code changes"
echo "  - Old dev images are automatically cleaned up (keeps last 3)"
echo ""
