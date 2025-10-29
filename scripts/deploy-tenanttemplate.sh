#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Deploy TenantTemplate ===${NC}"
echo ""

# Configuration
NAMESPACE="${TENANT_NAMESPACE:-default}"
REGISTRY_NAME="${REGISTRY_NAME:-example-registry}"
TEMPLATE_NAME="${TEMPLATE_NAME:-example-template}"

echo -e "${BLUE}Configuration:${NC}"
echo "  Namespace:         $NAMESPACE"
echo "  Registry Name:     $REGISTRY_NAME"
echo "  Template Name:     $TEMPLATE_NAME"
echo ""
echo -e "${YELLOW}Note: All tenant resources will be created in the same namespace as the Tenant CR ($NAMESPACE)${NC}"
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

# Check if TenantRegistry exists
echo ""
echo -e "${YELLOW}Checking TenantRegistry availability...${NC}"
if ! kubectl get tenantregistry "$REGISTRY_NAME" -n "$NAMESPACE" &> /dev/null; then
    echo -e "${RED}Error: TenantRegistry '$REGISTRY_NAME' not found in namespace '$NAMESPACE'${NC}"
    echo ""
    echo "Please deploy TenantRegistry first:"
    echo "  ./scripts/deploy-tenantregistry.sh"
    exit 1
fi
echo -e "${GREEN}✓ TenantRegistry exists${NC}"

# Create TenantTemplate
echo ""
echo -e "${YELLOW}Creating TenantTemplate...${NC}"
cat <<EOF | kubectl apply -f -
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantTemplate
metadata:
  name: $TEMPLATE_NAME
  namespace: $NAMESPACE
  labels:
    app.kubernetes.io/name: tenant-operator
    app.kubernetes.io/component: template
spec:
  registryId: $REGISTRY_NAME

  # ConfigMap for tenant configuration
  # Note: All resources are created in the same namespace as the Tenant CR
  configMaps:
  - id: tenant-config
    nameTemplate: "{{ .uid }}-config"
    spec:
      apiVersion: v1
      kind: ConfigMap
      metadata:
        labels:
          tenant: "{{ .uid }}"
      data:
        tenant.uid: "{{ .uid }}"
        tenant.host: "{{ .host }}"
        tenant.plan: "{{ default \"basic\" .planId }}"
        tenant.maxUsers: "{{ default \"100\" .maxUsers }}"
        tenant.storageGb: "{{ default \"10\" .storageGb }}"

  # Deployment for tenant application
  deployments:
  - id: tenant-deployment
    nameTemplate: "{{ .uid }}-app"
    dependIds:
    - tenant-config
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
              - name: TENANT_UID
                value: "{{ .uid }}"
              - name: TENANT_HOST
                value: "{{ .host }}"
              - name: TENANT_PLAN
                value: "{{ default \"basic\" .planId }}"
              - name: TENANT_MAX_USERS
                value: "{{ default \"100\" .maxUsers }}"
              resources:
                requests:
                  memory: "64Mi"
                  cpu: "100m"
                limits:
                  memory: "128Mi"
                  cpu: "200m"

  # Service for tenant application
  services:
  - id: tenant-service
    nameTemplate: "{{ .uid }}-svc"
    dependIds:
    - tenant-deployment
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
EOF
echo -e "${GREEN}✓ TenantTemplate created${NC}"

# Wait a moment for processing
echo ""
echo -e "${YELLOW}Waiting for TenantTemplate to be processed (5s)...${NC}"
sleep 5

# Show TenantTemplate status
echo ""
echo -e "${BLUE}TenantTemplate Status:${NC}"
kubectl get tenanttemplate "$TEMPLATE_NAME" -n "$NAMESPACE" 2>/dev/null || echo -e "${YELLOW}Not yet available${NC}"

# Show TenantRegistry status
echo ""
echo -e "${BLUE}TenantRegistry Status:${NC}"
kubectl get tenantregistry "$REGISTRY_NAME" -n "$NAMESPACE" 2>/dev/null || echo -e "${YELLOW}Not yet available${NC}"

# Wait for Tenants to be created
echo ""
echo -e "${YELLOW}Waiting for Tenants to be created (30s sync interval)...${NC}"
for i in {1..6}; do
    TENANT_COUNT=$(kubectl get tenants -n "$NAMESPACE" --no-headers 2>/dev/null | wc -l | tr -d ' ')
    if [ "$TENANT_COUNT" -gt 0 ]; then
        echo -e "${GREEN}✓ Found $TENANT_COUNT Tenant(s)${NC}"
        break
    fi
    if [ $i -eq 6 ]; then
        echo -e "${YELLOW}⚠ No Tenants created yet, but this may be normal${NC}"
    else
        echo "  Attempt $i/6: Waiting for Tenants... ($TENANT_COUNT found)"
        sleep 5
    fi
done

# Show created Tenants
echo ""
echo -e "${BLUE}Created Tenants:${NC}"
kubectl get tenants -n "$NAMESPACE" 2>/dev/null || echo -e "${YELLOW}No tenants created yet${NC}"

# Show deployment info
echo ""
echo -e "${GREEN}=== TenantTemplate Deployment Complete ===${NC}"
echo ""
echo -e "${BLUE}Resources Created:${NC}"
echo "  TenantTemplate:   $TEMPLATE_NAME"
echo ""
echo -e "${BLUE}Expected Active Tenants (from MySQL):${NC}"
echo "  - tenant-alpha (activate=true)"
echo "  - tenant-beta (activate=true)"
echo "  - tenant-gamma (activate=true)"
echo "  - tenant-epsilon (activate=true)"
echo ""
echo -e "${BLUE}Useful Commands:${NC}"
echo "  # Watch TenantRegistry status"
echo "  kubectl get tenantregistry $REGISTRY_NAME -n $NAMESPACE -w"
echo ""
echo "  # Watch Tenant creation"
echo "  watch kubectl get tenants -n $NAMESPACE"
echo ""
echo "  # Check specific Tenant"
echo "  kubectl describe tenant tenant-alpha -n $NAMESPACE"
echo ""
echo "  # List tenant resources in namespace"
echo "  kubectl get all -n $NAMESPACE -l managed-by=tenant-operator"
echo ""
echo "  # Watch tenant pods"
echo "  watch kubectl get pods -n $NAMESPACE -l managed-by=tenant-operator"
echo ""
echo -e "${BLUE}Verify Deployment:${NC}"
echo "  # Check if active tenants were created"
echo "  kubectl get tenants -n $NAMESPACE | grep -E 'alpha|beta|gamma|epsilon'"
echo ""
echo "  # Check tenant resources"
echo "  kubectl get deployments,services,configmaps -n $NAMESPACE -l tenant"
echo ""
echo -e "${BLUE}Operator Logs:${NC}"
echo "  kubectl logs -n tenant-operator-system -l control-plane=controller-manager -f --all-containers"
echo ""
