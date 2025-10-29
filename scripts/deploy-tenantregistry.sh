#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Deploy TenantRegistry ===${NC}"
echo ""

# Configuration
NAMESPACE="${TENANT_NAMESPACE:-default}"
MYSQL_NAMESPACE="${MYSQL_NAMESPACE:-tenant-operator-test}"
MYSQL_HOST="${MYSQL_HOST:-mysql.$MYSQL_NAMESPACE.svc.cluster.local}"
MYSQL_PORT="${MYSQL_PORT:-3306}"
MYSQL_DATABASE="${MYSQL_DATABASE:-tenant_registry}"
MYSQL_USER="${MYSQL_USER:-tenant_user}"
MYSQL_PASSWORD="${MYSQL_PASSWORD:-tenant_password}"
REGISTRY_NAME="${REGISTRY_NAME:-example-registry}"

echo -e "${BLUE}Configuration:${NC}"
echo "  Namespace:         $NAMESPACE"
echo "  Registry Name:     $REGISTRY_NAME"
echo "  MySQL Host:        $MYSQL_HOST"
echo "  MySQL Database:    $MYSQL_DATABASE"
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

# Check if MySQL is running
echo ""
echo -e "${YELLOW}Checking MySQL availability...${NC}"
if ! kubectl get service mysql -n "$MYSQL_NAMESPACE" &> /dev/null; then
    echo -e "${RED}Error: MySQL service not found in namespace '$MYSQL_NAMESPACE'${NC}"
    echo ""
    echo "Please deploy MySQL first:"
    echo "  ./scripts/deploy-mysql.sh"
    exit 1
fi

if ! kubectl get pods -n "$MYSQL_NAMESPACE" -l app=mysql -o jsonpath='{.items[0].status.phase}' 2>/dev/null | grep -q "Running"; then
    echo -e "${RED}Error: MySQL pod is not running${NC}"
    echo ""
    echo "Please check MySQL deployment:"
    echo "  kubectl get pods -n $MYSQL_NAMESPACE"
    exit 1
fi
echo -e "${GREEN}✓ MySQL is available${NC}"

# Create MySQL password secret
echo ""
echo -e "${YELLOW}Creating MySQL password secret...${NC}"
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: mysql-credentials
  namespace: $NAMESPACE
type: Opaque
stringData:
  password: "$MYSQL_PASSWORD"
EOF
echo -e "${GREEN}✓ Secret created${NC}"

# Create TenantRegistry
echo ""
echo -e "${YELLOW}Creating TenantRegistry...${NC}"
cat <<EOF | kubectl apply -f -
apiVersion: operator.kubernetes-tenants.org/v1
kind: TenantRegistry
metadata:
  name: $REGISTRY_NAME
  namespace: $NAMESPACE
  labels:
    app.kubernetes.io/name: tenant-operator
    app.kubernetes.io/component: registry
spec:
  source:
    type: mysql
    syncInterval: 30s
    mysql:
      host: "$MYSQL_HOST"
      port: $MYSQL_PORT
      database: "$MYSQL_DATABASE"
      username: "$MYSQL_USER"
      passwordRef:
        name: mysql-credentials
        key: password
      table: tenants

  # Required mappings
  valueMappings:
    uid: uid
    hostOrUrl: host_or_url
    activate: activate

  # Extra mappings for template variables
  extraValueMappings:
    deployImage: deploy_image
    planId: plan_id
    maxUsers: max_users
    storageGb: storage_gb
    customConfig: custom_config
EOF
echo -e "${GREEN}✓ TenantRegistry created${NC}"

# Wait a moment for registry to process
echo ""
echo -e "${YELLOW}Waiting for TenantRegistry to initialize (5s)...${NC}"
sleep 5

# Check TenantRegistry status
echo -e "${YELLOW}Checking TenantRegistry status...${NC}"
for i in {1..10}; do
    STATUS=$(kubectl get tenantregistry "$REGISTRY_NAME" -n "$NAMESPACE" -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}' 2>/dev/null || echo "")
    if [ "$STATUS" == "True" ]; then
        echo -e "${GREEN}✓ TenantRegistry is Ready (database connected)${NC}"
        break
    fi
    if [ $i -eq 10 ]; then
        echo -e "${YELLOW}⚠ TenantRegistry not ready yet (database connection may be pending)${NC}"
    else
        echo "  Attempt $i/10: Waiting for database connection..."
        sleep 3
    fi
done

# Show TenantRegistry status
echo ""
echo -e "${BLUE}TenantRegistry Status:${NC}"
kubectl get tenantregistry "$REGISTRY_NAME" -n "$NAMESPACE" 2>/dev/null || echo -e "${YELLOW}Not yet available${NC}"

# Show detailed status
echo ""
echo -e "${BLUE}TenantRegistry Details:${NC}"
kubectl describe tenantregistry "$REGISTRY_NAME" -n "$NAMESPACE" | tail -20

# Show deployment info
echo ""
echo -e "${GREEN}=== TenantRegistry Deployment Complete ===${NC}"
echo ""
echo -e "${BLUE}Resources Created:${NC}"
echo "  Secret:           mysql-credentials"
echo "  TenantRegistry:   $REGISTRY_NAME"
echo ""
echo -e "${BLUE}Next Steps:${NC}"
echo "  1. Deploy TenantTemplate:"
echo "     ./scripts/deploy-tenanttemplate.sh"
echo ""
echo "  2. Check TenantRegistry status:"
echo "     kubectl get tenantregistry $REGISTRY_NAME -n $NAMESPACE -w"
echo ""
echo "  3. View detailed status:"
echo "     kubectl describe tenantregistry $REGISTRY_NAME -n $NAMESPACE"
echo ""
echo -e "${BLUE}Note:${NC}"
echo "  - TenantRegistry is 'Ready' when database connection succeeds"
echo "  - Tenants will NOT be created until TenantTemplate is deployed"
echo "  - TenantRegistry will show 'desired' count from database"
echo "  - No Tenant CRs will exist until template is available"
echo ""
