#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Minikube Cleanup Script ===${NC}"
echo ""

# Check if minikube is installed
if ! command -v minikube &> /dev/null; then
    echo -e "${RED}Error: minikube is not installed${NC}"
    exit 1
fi

# Check if minikube cluster exists
if ! minikube status &> /dev/null; then
    echo -e "${YELLOW}No active minikube cluster found${NC}"
    echo ""
    read -p "Do you want to delete all minikube profiles? (y/N): " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${YELLOW}Deleting all minikube profiles...${NC}"
        minikube delete --all --purge
        echo -e "${GREEN}✓ All minikube profiles deleted${NC}"
    fi
else
    echo -e "${YELLOW}Active minikube cluster detected${NC}"
    echo ""

    # Show current cluster info
    echo "Current cluster status:"
    minikube status || true
    echo ""

    # Clean up MySQL resources first
    if command -v kubectl &> /dev/null && kubectl config current-context &> /dev/null; then
        echo ""
        read -p "Do you want to clean up MySQL test database resources? (y/N): " -n 1 -r
        echo ""
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            MYSQL_NAMESPACE="${MYSQL_NAMESPACE:-tenant-operator-test}"

            if kubectl get namespace "$MYSQL_NAMESPACE" &> /dev/null; then
                echo -e "${YELLOW}Cleaning up MySQL resources in namespace $MYSQL_NAMESPACE...${NC}"

                # Delete MySQL resources
                kubectl delete deployment mysql -n "$MYSQL_NAMESPACE" --ignore-not-found=true 2>/dev/null || true
                kubectl delete service mysql -n "$MYSQL_NAMESPACE" --ignore-not-found=true 2>/dev/null || true
                kubectl delete pvc mysql-pvc -n "$MYSQL_NAMESPACE" --ignore-not-found=true 2>/dev/null || true
                kubectl delete configmap mysql-init-sql -n "$MYSQL_NAMESPACE" --ignore-not-found=true 2>/dev/null || true
                kubectl delete secret mysql-secret -n "$MYSQL_NAMESPACE" --ignore-not-found=true 2>/dev/null || true

                echo -e "${GREEN}✓ MySQL resources cleaned up${NC}"

                # Option to delete the entire namespace
                echo ""
                read -p "Do you want to delete the entire namespace '$MYSQL_NAMESPACE'? (y/N): " -n 1 -r
                echo ""
                if [[ $REPLY =~ ^[Yy]$ ]]; then
                    kubectl delete namespace "$MYSQL_NAMESPACE" --ignore-not-found=true 2>/dev/null || true
                    echo -e "${GREEN}✓ Namespace '$MYSQL_NAMESPACE' deleted${NC}"
                fi
            else
                echo -e "${YELLOW}MySQL namespace '$MYSQL_NAMESPACE' not found, skipping...${NC}"
            fi
        fi

        # Clean up operator resources
        echo ""
        read -p "Do you want to clean up Tenant Operator resources? (y/N): " -n 1 -r
        echo ""
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            OPERATOR_NAMESPACE="${OPERATOR_NAMESPACE:-tenant-operator-system}"

            if kubectl get namespace "$OPERATOR_NAMESPACE" &> /dev/null; then
                echo -e "${YELLOW}Cleaning up Tenant Operator resources...${NC}"

                # Delete Tenant CRs first
                kubectl delete tenants --all --all-namespaces --ignore-not-found=true 2>/dev/null || true
                kubectl delete tenanttemplates --all --all-namespaces --ignore-not-found=true 2>/dev/null || true
                kubectl delete tenantregistries --all --all-namespaces --ignore-not-found=true 2>/dev/null || true

                # Delete operator deployment
                kubectl delete deployment -n "$OPERATOR_NAMESPACE" -l control-plane=controller-manager --ignore-not-found=true 2>/dev/null || true

                echo -e "${GREEN}✓ Operator resources cleaned up${NC}"

                # Option to delete CRDs
                echo ""
                read -p "Do you want to delete Tenant Operator CRDs? (y/N): " -n 1 -r
                echo ""
                if [[ $REPLY =~ ^[Yy]$ ]]; then
                    kubectl delete crd tenants.operator.kubernetes-tenants.org --ignore-not-found=true 2>/dev/null || true
                    kubectl delete crd tenanttemplates.operator.kubernetes-tenants.org --ignore-not-found=true 2>/dev/null || true
                    kubectl delete crd tenantregistries.operator.kubernetes-tenants.org --ignore-not-found=true 2>/dev/null || true
                    echo -e "${GREEN}✓ CRDs deleted${NC}"
                fi
            else
                echo -e "${YELLOW}Operator namespace '$OPERATOR_NAMESPACE' not found, skipping...${NC}"
            fi
        fi
    fi

    echo ""
    read -p "Do you want to delete the current minikube cluster? (y/N): " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${YELLOW}Stopping minikube cluster...${NC}"
        minikube stop || true

        echo -e "${YELLOW}Deleting minikube cluster...${NC}"
        minikube delete --purge

        echo -e "${GREEN}✓ Minikube cluster deleted${NC}"
    else
        echo -e "${YELLOW}Cleanup cancelled${NC}"
        exit 0
    fi
fi

# Optional: Clean up kubectl context
echo ""
read -p "Do you want to remove minikube kubectl context? (y/N): " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    kubectl config delete-context minikube 2>/dev/null && echo -e "${GREEN}✓ Kubectl context removed${NC}" || echo -e "${YELLOW}Context not found or already removed${NC}"
    kubectl config delete-cluster minikube 2>/dev/null && echo -e "${GREEN}✓ Kubectl cluster removed${NC}" || echo -e "${YELLOW}Cluster not found or already removed${NC}"
fi

# Optional: Clean up cached images
echo ""
read -p "Do you want to clean up Docker/Podman images cached by minikube? (y/N): " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}Cleaning up cached images...${NC}"
    rm -rf ~/.minikube/cache/images/* 2>/dev/null && echo -e "${GREEN}✓ Cached images cleaned${NC}" || echo -e "${YELLOW}No cached images found${NC}"
fi

echo ""
echo -e "${GREEN}=== Cleanup Complete ===${NC}"
echo ""
echo "To start a fresh minikube cluster, run:"
echo "  minikube start"
