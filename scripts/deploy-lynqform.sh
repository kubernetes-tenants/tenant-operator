#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Deploy LynqForm ===${NC}"
echo ""

# Configuration
NAMESPACE="${NODE_NAMESPACE:-default}"
REGISTRY_NAME="${REGISTRY_NAME:-example-registry}"
TEMPLATE_NAME="${TEMPLATE_NAME:-example-template}"

echo -e "${BLUE}Configuration:${NC}"
echo "  Namespace:         $NAMESPACE"
echo "  Registry Name:     $REGISTRY_NAME"
echo "  Template Name:     $TEMPLATE_NAME"
echo ""
echo -e "${YELLOW}Note: Each node will have its own dynamically created namespace (tenant-<uid>)${NC}"
echo -e "${YELLOW}      All tenant resources will be deployed into their respective namespaces${NC}"
echo ""

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}Error: kubectl is not installed${NC}"
    exit 1
fi

# Check if namespace exists
echo -e "${YELLOW}Ensuring namespace exists...${NC}"
if [ "$NAMESPACE" != "default" ]; then
    kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
fi
echo -e "${GREEN}✓ Namespace ready${NC}"

# Check if LynqHub exists
echo ""
echo -e "${YELLOW}Checking LynqHub availability...${NC}"
if ! kubectl get lynqhub "$REGISTRY_NAME" -n "$NAMESPACE" &> /dev/null; then
    echo -e "${RED}Error: LynqHub '$REGISTRY_NAME' not found in namespace '$NAMESPACE'${NC}"
    echo ""
    echo "Please deploy LynqHub first:"
    echo "  ./scripts/deploy-lynqhub.sh"
    exit 1
fi
echo -e "${GREEN}✓ LynqHub exists${NC}"

# Create LynqForm
echo ""
echo -e "${YELLOW}Creating LynqForm...${NC}"
cat <<EOF | kubectl apply -f -
apiVersion: operator.lynq.sh/v1
kind: LynqForm
metadata:
  name: $TEMPLATE_NAME
  namespace: $NAMESPACE
  labels:
    app.kubernetes.io/name: lynq
    app.kubernetes.io/component: template
spec:
  hubId: $REGISTRY_NAME

  # Namespace per node (dynamically created)
  # Each node gets its own isolated namespace
  namespaces:
  - id: node-namespace
    nameTemplate: "node-{{ .uid }}"
    spec:
      apiVersion: v1
      kind: Namespace
      metadata:
        labels:
          node-id: "{{ .uid }}"
          node-host: "{{ .host }}"
          node-plan: "{{ default \"basic\" .planId }}"
          managed-by: lynq

  # ConfigMap for node configuration
  # Deployed into the tenant's own namespace
  configMaps:
  - id: node-config
    nameTemplate: "{{ .uid }}-config"
    targetNamespace: "node-{{ .uid }}"
    dependIds:
    - node-namespace
    spec:
      apiVersion: v1
      kind: ConfigMap
      metadata:
        labels:
          tenant: "{{ .uid }}"
      data:
        node.uid: "{{ .uid }}"
        node.host: "{{ .host }}"
        node.plan: "{{ default \"basic\" .planId }}"
        node.maxUsers: "{{ default \"100\" .maxUsers }}"
        node.storageGb: "{{ default \"10\" .storageGb }}"

  # Deployment for node application
  deployments:
  - id: node-deployment
    nameTemplate: "{{ .uid }}-app"
    targetNamespace: "node-{{ .uid }}"
    dependIds:
    - node-namespace
    - node-config
    ignoreFields:
    - "$.spec.replicas"
    spec:
      apiVersion: apps/v1
      kind: Deployment
      metadata:
        labels:
          app: "{{ .uid }}"
          tenant: "{{ .uid }}"
      spec:
        replicas: 1
        selector:
          matchLabels:
            app: "{{ .uid }}"
        template:
          metadata:
            labels:
              app: "{{ .uid }}"
              tenant: "{{ .uid }}"
          spec:
            containers:
            - name: app
              image: "{{ default \"nginx:1.21\" .deployImage }}"
              ports:
              - containerPort: 80
                name: http
              env:
              - name: NODE_UID
                value: "{{ .uid }}"
              - name: NODE_HOST
                value: "{{ .host }}"
              - name: NODE_PLAN
                value: "{{ default \"basic\" .planId }}"
              - name: NODE_MAX_USERS
                value: "{{ default \"100\" .maxUsers }}"
              resources:
                requests:
                  memory: "64Mi"
                  cpu: "100m"
                limits:
                  memory: "128Mi"
                  cpu: "200m"

  # Service for node application
  services:
  - id: node-service
    nameTemplate: "{{ .uid }}-svc"
    targetNamespace: "node-{{ .uid }}"
    dependIds:
    - node-namespace
    - node-deployment
    spec:
      apiVersion: v1
      kind: Service
      metadata:
        labels:
          app: "{{ .uid }}"
          tenant: "{{ .uid }}"
      spec:
        type: ClusterIP
        selector:
          app: "{{ .uid }}"
        ports:
        - port: 80
          targetPort: 80
          protocol: TCP
          name: http

  # HorizontalPodAutoscaler for node application
  horizontalPodAutoscalers:
  - id: node-hpa
    nameTemplate: "{{ .uid }}-hpa"
    targetNamespace: "node-{{ .uid }}"
    dependIds:
    - node-namespace
    - node-deployment
    spec:
      apiVersion: autoscaling/v2
      kind: HorizontalPodAutoscaler
      metadata:
        labels:
          app: "{{ .uid }}"
          tenant: "{{ .uid }}"
      spec:
        scaleTargetRef:
          apiVersion: apps/v1
          kind: Deployment
          name: "{{ .uid }}-app"
        minReplicas: 1
        maxReplicas: 5
        metrics:
        - type: Resource
          resource:
            name: cpu
            target:
              type: Utilization
              averageUtilization: 70
        - type: Resource
          resource:
            name: memory
            target:
              type: Utilization
              averageUtilization: 80
EOF
echo -e "${GREEN}✓ LynqForm created${NC}"

# Wait a moment for processing
echo ""
echo -e "${YELLOW}Waiting for LynqForm to be processed (5s)...${NC}"
sleep 5

# Show LynqForm status
echo ""
echo -e "${BLUE}LynqForm Status:${NC}"
kubectl get lynqform "$TEMPLATE_NAME" -n "$NAMESPACE" 2>/dev/null || echo -e "${YELLOW}Not yet available${NC}"

# Show LynqHub status
echo ""
echo -e "${BLUE}LynqHub Status:${NC}"
kubectl get lynqhub "$REGISTRY_NAME" -n "$NAMESPACE" 2>/dev/null || echo -e "${YELLOW}Not yet available${NC}"

# Wait for Tenants to be created
echo ""
echo -e "${YELLOW}Waiting for LynqNodes to be created (30s sync interval)...${NC}"
for i in {1..6}; do
    NODE_COUNT=$(kubectl get lynqnodes -n "$NAMESPACE" --no-headers 2>/dev/null | wc -l | tr -d ' ')
    if [ "$NODE_COUNT" -gt 0 ]; then
        echo -e "${GREEN}✓ Found $NODE_COUNT LynqNode(s)${NC}"
        break
    fi
    if [ $i -eq 6 ]; then
        echo -e "${YELLOW}⚠ No LynqNodes created yet, but this may be normal${NC}"
    else
        echo "  Attempt $i/6: Waiting for LynqNodes... ($NODE_COUNT found)"
        sleep 5
    fi
done

# Show created Tenants
echo ""
echo -e "${BLUE}Created LynqNodes:${NC}"
kubectl get lynqnodes -n "$NAMESPACE" 2>/dev/null || echo -e "${YELLOW}No nodes created yet${NC}"

# Show deployment info
echo ""
echo -e "${GREEN}=== LynqForm Deployment Complete ===${NC}"
echo ""
echo -e "${BLUE}Resources Created:${NC}"
echo "  LynqForm:   $TEMPLATE_NAME"
echo ""
echo -e "${BLUE}Template includes:${NC}"
echo "  - Namespace (dynamically created per node: tenant-<uid>)"
echo "  - ConfigMap (node configuration)"
echo "  - Deployment (application, replicas ignored for HPA)"
echo "  - Service (ClusterIP)"
echo "  - HorizontalPodAutoscaler (min: 1, max: 5, CPU: 70%, Memory: 80%)"
echo ""
echo -e "${BLUE}Namespace Isolation:${NC}"
echo "  Each node gets its own namespace for complete resource isolation"
echo "  Expected namespaces:"
echo "    - node-alpha"
echo "    - node-beta"
echo "    - node-gamma"
echo "    - node-epsilon"
echo ""
echo -e "${BLUE}Expected Active LynqNodes (from MySQL):${NC}"
echo "  - node-alpha (activate=true)"
echo "  - node-beta (activate=true)"
echo "  - node-gamma (activate=true)"
echo "  - node-epsilon (activate=true)"
echo ""
echo -e "${BLUE}Useful Commands:${NC}"
echo "  # Watch LynqHub status"
echo "  kubectl get lynqhub $REGISTRY_NAME -n $NAMESPACE -w"
echo ""
echo "  # Watch LynqNode creation"
echo "  watch kubectl get lynqnodes -n $NAMESPACE"
echo ""
echo "  # List all node namespaces"
echo "  kubectl get namespaces -l managed-by=lynq"
echo ""
echo "  # Check specific LynqNode"
echo "  kubectl describe lynqnode node-alpha -n $NAMESPACE"
echo ""
echo "  # List resources in a specific node namespace"
echo "  kubectl get all -n node-alpha"
echo ""
echo "  # Watch all node pods across all namespaces"
echo "  watch kubectl get pods -A -l managed-by=lynq"
echo ""
echo "  # Get all resources for a specific node (across namespaces)"
echo "  kubectl get all -A -l tenant-id=alpha"
echo ""
echo -e "${BLUE}Verify Deployment:${NC}"
echo "  # Check if active nodes were created"
echo "  kubectl get lynqnodes -n $NAMESPACE | grep -E 'alpha|beta|gamma|epsilon'"
echo ""
echo "  # Check if node namespaces were created"
echo "  kubectl get namespaces | grep node-"
echo ""
echo "  # Check resources in a specific node namespace"
echo "  kubectl get deployments,services,configmaps,hpa -n node-alpha"
echo ""
echo "  # Check all node resources across all namespaces"
echo "  kubectl get deployments,services,configmaps,hpa -A -l managed-by=lynq"
echo ""
echo "  # Check HPA status for all nodes"
echo "  kubectl get hpa -A -l managed-by=lynq"
echo ""
echo -e "${BLUE}Operator Logs:${NC}"
echo "  kubectl logs -n lynq-system -l control-plane=controller-manager -f --all-containers"
echo ""
