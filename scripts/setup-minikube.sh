#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Minikube Setup Script for Lynq ===${NC}"
echo ""

# Default values
DRIVER="${MINIKUBE_DRIVER:-docker}"
CPUS="${MINIKUBE_CPUS:-2}"
MEMORY="${MINIKUBE_MEMORY:-2048}"
DISK_SIZE="${MINIKUBE_DISK_SIZE:-5g}"
K8S_VERSION="${MINIKUBE_K8S_VERSION:-v1.28.3}"
PROFILE="${MINIKUBE_PROFILE:-lynq}"

# Check if minikube is installed
if ! command -v minikube &> /dev/null; then
    echo -e "${RED}Error: minikube is not installed${NC}"
    echo ""
    echo "Please install minikube first:"
    echo "  macOS: brew install minikube"
    echo "  Linux: curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64"
    exit 1
fi

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}Error: kubectl is not installed${NC}"
    echo ""
    echo "Please install kubectl first:"
    echo "  macOS: brew install kubectl"
    echo "  Linux: curl -LO https://dl.k8s.io/release/\$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
    exit 1
fi

# Display configuration
echo -e "${YELLOW}Configuration:${NC}"
echo "  Profile:        $PROFILE"
echo "  Driver:         $DRIVER"
echo "  CPUs:           $CPUS"
echo "  Memory:         $MEMORY MB"
echo "  Disk Size:      $DISK_SIZE"
echo "  K8s Version:    $K8S_VERSION"
echo ""

read -p "Continue with these settings? (Y/n): " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Nn]$ ]]; then
    echo -e "${YELLOW}Setup cancelled${NC}"
    echo ""
    echo "You can customize settings with environment variables:"
    echo "  MINIKUBE_DRIVER=docker|hyperkit|virtualbox"
echo "  MINIKUBE_CPUS=2"
echo "  MINIKUBE_MEMORY=2048"
    echo "  MINIKUBE_DISK_SIZE=5g"
    echo "  MINIKUBE_K8S_VERSION=v1.28.3"
    echo "  MINIKUBE_PROFILE=lynq"
    exit 0
fi

# Check if profile already exists
if minikube profile list 2>/dev/null | grep -q "^| $PROFILE "; then
    echo -e "${YELLOW}Profile '$PROFILE' already exists${NC}"
    read -p "Delete existing profile and create new one? (y/N): " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${YELLOW}Deleting existing profile...${NC}"
        minikube delete -p "$PROFILE"
    else
        echo -e "${YELLOW}Using existing profile${NC}"
        minikube start -p "$PROFILE"
        kubectl config use-context "$PROFILE"
        echo -e "${GREEN}✓ Switched to existing cluster${NC}"
        exit 0
    fi
fi

# Start minikube
echo -e "${YELLOW}Starting minikube cluster...${NC}"
minikube start \
    -p "$PROFILE" \
    --driver="$DRIVER" \
    --cpus="$CPUS" \
    --memory="$MEMORY" \
    --disk-size="$DISK_SIZE" \
    --kubernetes-version="$K8S_VERSION" \
    --addons=ingress,metrics-server,dashboard

echo -e "${GREEN}✓ Minikube cluster started${NC}"

# Set kubectl context
kubectl config use-context "$PROFILE"
echo -e "${GREEN}✓ Kubectl context set to '$PROFILE'${NC}"

# Wait for cluster to be ready
echo -e "${YELLOW}Waiting for cluster to be ready...${NC}"
kubectl wait --for=condition=Ready nodes --all --timeout=300s
echo -e "${GREEN}✓ Cluster is ready${NC}"

# Enable additional addons
echo -e "${YELLOW}Configuring addons...${NC}"

# Wait for ingress controller to be ready
echo -e "${YELLOW}Waiting for ingress controller...${NC}"
kubectl wait --namespace ingress-nginx \
    --for=condition=ready pod \
    --selector=app.kubernetes.io/component=controller \
    --timeout=300s 2>/dev/null || echo -e "${YELLOW}Ingress controller not ready yet${NC}"

# Install cert-manager for webhook certificates
echo ""
echo -e "${YELLOW}Installing cert-manager...${NC}"
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.2/cert-manager.yaml
echo -e "${GREEN}✓ cert-manager manifests applied${NC}"

# Wait for cert-manager to be ready
echo -e "${YELLOW}Waiting for cert-manager to be ready...${NC}"
kubectl wait --namespace cert-manager \
    --for=condition=ready pod \
    --selector=app.kubernetes.io/instance=cert-manager \
    --timeout=300s
echo -e "${GREEN}✓ cert-manager is ready${NC}"

# Install CRDs
echo ""
echo -e "${YELLOW}Installing Lynq CRDs...${NC}"
if [ -f "config/crd/bases/operator.lynq.sh_lynqhubs.yaml" ]; then
    kubectl apply -f config/crd/bases/
    echo -e "${GREEN}✓ CRDs installed${NC}"
else
    echo -e "${YELLOW}⚠ CRD files not found, skipping...${NC}"
fi

# Create operator namespace
echo ""
echo -e "${YELLOW}Creating operator namespace...${NC}"
kubectl create namespace lynq-system --dry-run=client -o yaml | kubectl apply -f -
echo -e "${GREEN}✓ Namespace 'lynq-system' created${NC}"

# Create test namespace
echo -e "${YELLOW}Creating test namespace...${NC}"
kubectl create namespace lynq-test --dry-run=client -o yaml | kubectl apply -f -
echo -e "${GREEN}✓ Namespace 'lynq-test' created${NC}"

# Display cluster info
echo ""
echo -e "${GREEN}=== Setup Complete ===${NC}"
echo ""
echo -e "${BLUE}Cluster Information:${NC}"
echo "  Profile:        $PROFILE"
echo "  Context:        $(kubectl config current-context)"
echo "  Kubernetes:     $(kubectl version --short 2>/dev/null | grep Server || kubectl version --output=json | jq -r .serverVersion.gitVersion)"
echo "  Nodes:          $(kubectl get nodes --no-headers | wc -l | tr -d ' ')"
echo ""
echo -e "${BLUE}Useful Commands:${NC}"
echo "  Check cluster status:    minikube status -p $PROFILE"
echo "  Open dashboard:          minikube dashboard -p $PROFILE"
echo "  Stop cluster:            minikube stop -p $PROFILE"
echo "  Delete cluster:          ./scripts/cleanup-minikube.sh"
echo "  Run operator locally:    make run"
echo "  Build & deploy operator: make docker-build docker-push deploy IMG=<your-registry>/lynq:tag"
echo ""
echo -e "${BLUE}Get cluster IP for testing:${NC}"
echo "  minikube ip -p $PROFILE"
echo ""
echo -e "${BLUE}Access services via LoadBalancer:${NC}"
echo "  minikube tunnel -p $PROFILE  # Run in separate terminal"
echo ""
