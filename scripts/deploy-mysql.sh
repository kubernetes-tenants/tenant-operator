#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Deploy MySQL Test Database ===${NC}"
echo ""

# Configuration
NAMESPACE="${MYSQL_NAMESPACE:-lynq-test}"
MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD:-rootpassword}"
MYSQL_DATABASE="${MYSQL_DATABASE:-tenant_registry}"
MYSQL_USER="${MYSQL_USER:-tenant_user}"
MYSQL_PASSWORD="${MYSQL_PASSWORD:-tenant_password}"

echo -e "${BLUE}Configuration:${NC}"
echo "  Namespace:      $NAMESPACE"
echo "  Database:       $MYSQL_DATABASE"
echo "  User:           $MYSQL_USER"
echo ""

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}Error: kubectl is not installed${NC}"
    exit 1
fi

# Create namespace if not exists
echo -e "${YELLOW}Ensuring namespace exists...${NC}"
kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
echo -e "${GREEN}✓ Namespace ready${NC}"

# Create MySQL Secret
echo ""
echo -e "${YELLOW}Creating MySQL secret...${NC}"
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: mysql-secret
  namespace: $NAMESPACE
type: Opaque
stringData:
  root-password: "$MYSQL_ROOT_PASSWORD"
  password: "$MYSQL_PASSWORD"
EOF
echo -e "${GREEN}✓ Secret created${NC}"

# Create Init SQL ConfigMap
echo ""
echo -e "${YELLOW}Creating init SQL ConfigMap...${NC}"
cat <<'EOF' | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: mysql-init-sql
  namespace: lynq-test
data:
  init.sql: |
    -- Create tenants table
    CREATE TABLE IF NOT EXISTS tenants (
      id INT AUTO_INCREMENT PRIMARY KEY,
      uid VARCHAR(255) NOT NULL UNIQUE,
      host_or_url VARCHAR(255) NOT NULL,
      activate BOOLEAN NOT NULL DEFAULT TRUE,
      deploy_image VARCHAR(255),
      plan_id VARCHAR(50),
      max_users INT,
      storage_gb INT,
      custom_config JSON,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
      INDEX idx_uid (uid),
      INDEX idx_activate (activate)
    );

    -- Insert test data
    INSERT INTO tenants (uid, host_or_url, activate, deploy_image, plan_id, max_users, storage_gb, custom_config)
    VALUES
      ('node-alpha', 'https://alpha.example.com', TRUE, 'nginx:1.21', 'enterprise', 1000, 100, '{"features": ["advanced-analytics", "custom-domain"]}'),
      ('node-beta', 'https://beta.example.com', TRUE, 'nginx:1.21', 'professional', 500, 50, '{"features": ["analytics", "api-access"]}'),
      ('node-gamma', 'https://gamma.example.com', TRUE, 'nginx:1.21', 'basic', 100, 10, '{"features": ["basic-support"]}'),
      ('node-delta', 'https://delta.example.com', FALSE, 'nginx:1.21', 'professional', 500, 50, '{"features": ["analytics"]}'),
      ('node-epsilon', 'https://epsilon.example.com', TRUE, 'nginx:1.22', 'enterprise', 2000, 200, '{"features": ["advanced-analytics", "custom-domain", "sso"]}'),
      ('node-zeta', 'https://zeta.example.com', FALSE, 'nginx:1.20', 'basic', 50, 5, '{"features": []}')
    ON DUPLICATE KEY UPDATE
      host_or_url=VALUES(host_or_url),
      activate=VALUES(activate),
      deploy_image=VALUES(deploy_image),
      plan_id=VALUES(plan_id),
      max_users=VALUES(max_users),
      storage_gb=VALUES(storage_gb),
      custom_config=VALUES(custom_config),
      updated_at=CURRENT_TIMESTAMP;

    -- Show inserted data
    SELECT
      id,
      uid,
      host_or_url,
      activate,
      plan_id,
      deploy_image
    FROM tenants
    ORDER BY id;
EOF
echo -e "${GREEN}✓ Init SQL ConfigMap created${NC}"

# Create PersistentVolumeClaim
echo ""
echo -e "${YELLOW}Creating PersistentVolumeClaim...${NC}"
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: mysql-pvc
  namespace: $NAMESPACE
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
EOF
echo -e "${GREEN}✓ PVC created${NC}"

# Create MySQL Deployment
echo ""
echo -e "${YELLOW}Creating MySQL Deployment...${NC}"
cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mysql
  namespace: $NAMESPACE
  labels:
    app: mysql
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mysql
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: mysql
    spec:
      containers:
      - name: mysql
        image: mysql:8.0
        env:
        - name: MYSQL_ROOT_PASSWORD
          valueFrom:
            secretKeyRef:
              name: mysql-secret
              key: root-password
        - name: MYSQL_DATABASE
          value: "$MYSQL_DATABASE"
        - name: MYSQL_USER
          value: "$MYSQL_USER"
        - name: MYSQL_PASSWORD
          valueFrom:
            secretKeyRef:
              name: mysql-secret
              key: password
        ports:
        - containerPort: 3306
          name: mysql
        volumeMounts:
        - name: mysql-storage
          mountPath: /var/lib/mysql
        - name: init-sql
          mountPath: /docker-entrypoint-initdb.d
        startupProbe:
          tcpSocket:
            port: 3306
          initialDelaySeconds: 10
          periodSeconds: 5
          timeoutSeconds: 2
          failureThreshold: 30
        livenessProbe:
          tcpSocket:
            port: 3306
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          tcpSocket:
            port: 3306
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 2
          failureThreshold: 3
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
      volumes:
      - name: mysql-storage
        persistentVolumeClaim:
          claimName: mysql-pvc
      - name: init-sql
        configMap:
          name: mysql-init-sql
EOF
echo -e "${GREEN}✓ Deployment created${NC}"

# Create MySQL Service
echo ""
echo -e "${YELLOW}Creating MySQL Service...${NC}"
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: mysql
  namespace: $NAMESPACE
  labels:
    app: mysql
spec:
  type: ClusterIP
  ports:
  - port: 3306
    targetPort: 3306
    protocol: TCP
    name: mysql
  selector:
    app: mysql
EOF
echo -e "${GREEN}✓ Service created${NC}"

# Wait for MySQL to be ready
echo ""
echo -e "${YELLOW}Waiting for MySQL to be ready...${NC}"
kubectl wait --for=condition=Available deployment/mysql -n "$NAMESPACE" --timeout=300s
echo -e "${GREEN}✓ MySQL is ready${NC}"

# Wait for pod to be fully ready
echo ""
echo -e "${YELLOW}Waiting for MySQL pod to be ready...${NC}"
kubectl wait --for=condition=Ready pod -l app=mysql -n "$NAMESPACE" --timeout=300s
echo -e "${GREEN}✓ MySQL pod is ready${NC}"

# Show deployment info
echo ""
echo -e "${GREEN}=== MySQL Deployment Complete ===${NC}"
echo ""
echo -e "${BLUE}Connection Information:${NC}"
echo "  Host:           mysql.$NAMESPACE.svc.cluster.local"
echo "  Port:           3306"
echo "  Database:       $MYSQL_DATABASE"
echo "  User:           $MYSQL_USER"
echo "  Password:       $MYSQL_PASSWORD"
echo "  Root Password:  $MYSQL_ROOT_PASSWORD"
echo ""
echo -e "${BLUE}Kubernetes Resources:${NC}"
echo "  Namespace:      $NAMESPACE"
echo "  Deployment:     mysql"
echo "  Service:        mysql"
echo "  Secret:         mysql-secret"
echo "  ConfigMap:      mysql-init-sql"
echo "  PVC:            mysql-pvc"
echo ""
echo -e "${BLUE}Verify Installation:${NC}"
echo "  kubectl get all -n $NAMESPACE"
echo "  kubectl logs -n $NAMESPACE deployment/mysql"
echo ""
echo -e "${BLUE}Access MySQL:${NC}"
echo "  kubectl run -it --rm mysql-client --image=mysql:8.0 --restart=Never -n $NAMESPACE -- \\"
echo "    mysql -h mysql.$NAMESPACE.svc.cluster.local -u $MYSQL_USER -p$MYSQL_PASSWORD $MYSQL_DATABASE"
echo ""
echo -e "${BLUE}Test Query:${NC}"
echo "  kubectl exec -it deployment/mysql -n $NAMESPACE -- \\"
echo "    mysql -u $MYSQL_USER -p$MYSQL_PASSWORD $MYSQL_DATABASE -e 'SELECT * FROM tenants;'"
echo ""
echo -e "${BLUE}Create LynqHub:${NC}"
echo "  Apply the sample: config/samples/lynqnodes_v1_lynqhub.yaml"
echo ""
