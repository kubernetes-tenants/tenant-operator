<template>
  <div class="dependency-graph-visualizer">
    <div class="controls-section">
      <div class="preset-buttons">
        <span class="label">Load Example:</span>
        <button
          v-for="preset in presetList"
          :key="preset.key"
          @click="loadPreset(preset.key)"
          class="preset-btn"
          :class="{ active: selectedPreset === preset.key }"
        >
          {{ preset.label }}
        </button>
      </div>
    </div>

    <div class="editor-graph-container">
      <!-- Left: YAML Editor -->
      <div class="editor-panel">
        <div class="panel-header">
          <h4>LynqForm YAML</h4>
          <button @click="analyzeYaml" class="analyze-btn">
            🔍 Analyze Dependencies
          </button>
        </div>
        <textarea
          v-model="yamlInput"
          class="yaml-editor"
          placeholder="Paste your LynqForm YAML here..."
          spellcheck="false"
        ></textarea>
        <div v-if="parseError" class="parse-error">
          ⚠️ {{ parseError }}
        </div>
      </div>

      <!-- Right: Graph Visualization -->
      <div class="graph-panel">
        <div class="panel-header">
          <h4>Dependency Graph</h4>
          <div v-if="nodes.length > 0" class="resource-count">
            {{ nodes.length }} resources found
          </div>
        </div>

        <div v-if="nodes.length === 0" class="empty-state">
          <div class="empty-icon">📊</div>
          <p>Paste a LynqForm YAML and click "Analyze Dependencies"</p>
          <p class="hint">or load an example from the buttons above</p>
        </div>

        <div v-else class="graph-wrapper">
          <!-- Zoom Controls -->
          <div class="zoom-controls">
            <button @click="zoomIn" class="zoom-btn" title="Zoom In">➕</button>
            <button @click="zoomOut" class="zoom-btn" title="Zoom Out">➖</button>
            <button @click="resetZoom" class="zoom-btn" title="Reset">🔄</button>
            <span class="zoom-level">{{ Math.round(scale * 100) }}%</span>
          </div>

          <svg
            ref="svgRef"
            class="dependency-graph"
            :viewBox="viewBox"
            preserveAspectRatio="xMidYMid meet"
            @wheel.prevent="handleWheel"
            @mousedown="startPan"
            @mousemove="handlePan"
            @mouseup="endPan"
            @mouseleave="endPan"
          >
          <defs>
            <marker
              id="arrowhead"
              markerWidth="10"
              markerHeight="7"
              refX="9"
              refY="3.5"
              orient="auto"
            >
              <polygon points="0 0, 10 3.5, 0 7" :fill="arrowColor" />
            </marker>
            <marker
              id="arrowhead-error"
              markerWidth="10"
              markerHeight="7"
              refX="9"
              refY="3.5"
              orient="auto"
            >
              <polygon points="0 0, 10 3.5, 0 7" fill="#ef5350" />
            </marker>
          </defs>

          <g :transform="transformString">

          <!-- Edges (dependencies) -->
          <g class="edges">
            <line
              v-for="edge in edges"
              :key="`${edge.from}-${edge.to}`"
              :x1="getEdgePoints(edge.from, edge.to).x1"
              :y1="getEdgePoints(edge.from, edge.to).y1"
              :x2="getEdgePoints(edge.from, edge.to).x2"
              :y2="getEdgePoints(edge.from, edge.to).y2"
              :stroke="edge.isInCycle ? '#ef5350' : edgeColor"
              :stroke-width="edge.isInCycle ? 3 : 2"
              :marker-end="edge.isInCycle ? 'url(#arrowhead-error)' : 'url(#arrowhead)'"
              class="edge-line"
              :class="{ 'edge-error': edge.isInCycle }"
            />
          </g>

          <!-- Nodes (resources) -->
          <g class="nodes">
            <g
              v-for="node in nodes"
              :key="node.id"
              :transform="`translate(${node.x}, ${node.y})`"
              class="node"
              :class="{
                'node-error': node.isInCycle,
                'node-highlighted': highlightedNode === node.id
              }"
              @mouseenter="highlightedNode = node.id"
              @mouseleave="highlightedNode = null"
            >
              <circle
                :r="nodeRadius"
                :fill="node.isInCycle ? 'var(--vp-c-danger-soft)' : nodeColor"
                :stroke="node.isInCycle ? '#ef5350' : nodeBorderColor"
                stroke-width="2"
              />

              <!-- Execution order badge -->
              <circle
                v-if="node.order !== null && !hasCycle"
                :cx="nodeRadius - 8"
                :cy="-nodeRadius + 8"
                r="12"
                fill="var(--vp-c-brand)"
                stroke="var(--vp-c-bg)"
                stroke-width="2"
                class="order-badge"
              />
              <text
                v-if="node.order !== null && !hasCycle"
                :x="nodeRadius - 8"
                :y="-nodeRadius + 8"
                text-anchor="middle"
                dominant-baseline="central"
                class="order-text"
              >
                {{ node.order }}
              </text>

              <!-- Node label -->
              <text
                y="5"
                text-anchor="middle"
                class="node-label"
                :fill="node.isInCycle ? '#ef5350' : 'var(--vp-c-text-1)'"
              >
                {{ node.id }}
              </text>

              <!-- Node type -->
              <text
                y="20"
                text-anchor="middle"
                class="node-type"
              >
                {{ node.type }}
              </text>
            </g>
          </g>
          </g>

          </svg>
        </div>

        <!-- Analysis Results -->
        <div v-if="nodes.length > 0" class="analysis-results">
          <div v-if="hasCycle" class="error-message">
            <strong>⚠️ Dependency Cycle Detected!</strong>
            <p>{{ cycleMessage }}</p>
            <p class="hint">Remove one or more dependencies to break the cycle.</p>
          </div>

          <div v-else class="success-message">
            <strong>✅ Valid Dependency Graph</strong>
            <p><strong>Execution order:</strong> {{ executionOrderText }}</p>
            <p class="hint">Resources are applied in topological order. Resources at the same level can execute in parallel.</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Resource Details Table -->
    <div v-if="nodes.length > 0" class="resource-details">
      <h4>Resource Dependencies</h4>
      <div class="resource-table">
        <div class="table-header">
          <div class="col-order">Order</div>
          <div class="col-id">Resource ID</div>
          <div class="col-type">Type</div>
          <div class="col-deps">Dependencies</div>
        </div>
        <div
          v-for="node in sortedNodes"
          :key="node.id"
          class="table-row"
          :class="{ 'row-error': node.isInCycle }"
          @mouseenter="highlightedNode = node.id"
          @mouseleave="highlightedNode = null"
        >
          <div class="col-order">
            <span v-if="!hasCycle" class="order-badge-small">{{ node.order || '-' }}</span>
            <span v-else class="error-badge">⚠️</span>
          </div>
          <div class="col-id">{{ node.id }}</div>
          <div class="col-type">
            <span class="type-badge">{{ node.type }}</span>
          </div>
          <div class="col-deps">
            <span v-if="node.dependIds.length === 0" class="no-deps">None</span>
            <span v-else class="dep-list">{{ node.dependIds.join(', ') }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';
import yaml from 'js-yaml';

const yamlInput = ref('');
const parseError = ref('');
const selectedPreset = ref('');
const highlightedNode = ref(null);

// Graph properties
const width = 900;
const height = 500;
const nodeRadius = 45;
const viewBox = `0 0 ${width} ${height}`;

// Zoom and pan state
const scale = ref(1);
const translateX = ref(0);
const translateY = ref(0);
const isPanning = ref(false);
const panStartX = ref(0);
const panStartY = ref(0);

// Theme colors
const nodeColor = computed(() => 'var(--vp-c-bg-soft)');
const nodeBorderColor = computed(() => 'var(--vp-c-brand-light)');
const edgeColor = computed(() => 'var(--vp-c-text-3)');
const arrowColor = computed(() => 'var(--vp-c-text-3)');

// Node data
const nodes = ref([]);
const edges = ref([]);
const hasCycle = ref(false);
const cycleMessage = ref('');

// Preset examples
const presetList = [
  { key: 'simple', label: 'Simple App' },
  { key: 'multi-tier', label: 'Multi-Tier Stack' },
  { key: 'parallel', label: 'Parallel Resources' },
  { key: 'cycle', label: 'Cycle Example' }
];

const presets = {
  simple: `apiVersion: operator.lynq.sh/v1
kind: LynqForm
metadata:
  name: simple-app
spec:
  hubId: my-hub

  # 1. Secret (no dependencies)
  secrets:
    - id: app-secret
      nameTemplate: "{{ .uid }}-secret"
      spec:
        stringData:
          password: "{{ randAlphaNum 32 }}"

  # 2. Deployment (depends on secret)
  deployments:
    - id: app-deployment
      dependIds: ["app-secret"]
      nameTemplate: "{{ .uid }}-app"
      spec:
        replicas: 2
        template:
          spec:
            containers:
            - name: app
              image: nginx:latest

  # 3. Service (depends on deployment)
  services:
    - id: app-service
      dependIds: ["app-deployment"]
      nameTemplate: "{{ .uid }}-svc"
      spec:
        ports:
        - port: 80`,

  'multi-tier': `apiVersion: operator.lynq.sh/v1
kind: LynqForm
metadata:
  name: multi-tier-app
spec:
  hubId: my-hub

  # Level 1: Namespace
  manifests:
    - id: node-namespace
      spec:
        apiVersion: v1
        kind: Namespace
        metadata:
          name: "node-{{ .uid }}"

  # Level 2: Secrets and PVC (parallel)
  secrets:
    - id: db-credentials
      dependIds: ["node-namespace"]
      nameTemplate: "{{ .uid }}-db-secret"
      spec:
        stringData:
          password: "{{ randAlphaNum 32 }}"

  persistentVolumeClaims:
    - id: data-pvc
      dependIds: ["node-namespace"]
      nameTemplate: "{{ .uid }}-data"
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 10Gi

  # Level 3: Database
  statefulSets:
    - id: postgres
      dependIds: ["db-credentials", "data-pvc"]
      nameTemplate: "{{ .uid }}-db"
      spec:
        serviceName: "{{ .uid }}-db"
        replicas: 1
        template:
          spec:
            containers:
            - name: postgres
              image: postgres:15

  # Level 4: Application
  deployments:
    - id: app
      dependIds: ["postgres"]
      nameTemplate: "{{ .uid }}-app"
      spec:
        replicas: 3
        template:
          spec:
            containers:
            - name: app
              image: myapp:latest

  # Level 5: Service
  services:
    - id: app-service
      dependIds: ["app"]
      nameTemplate: "{{ .uid }}-svc"
      spec:
        ports:
        - port: 8080

  # Level 6: Ingress
  ingresses:
    - id: app-ingress
      dependIds: ["app-service"]
      nameTemplate: "{{ .uid }}-ingress"
      spec:
        rules:
        - host: "{{ .host }}"
          http:
            paths:
            - path: /
              pathType: Prefix
              backend:
                service:
                  name: "{{ .uid }}-svc"
                  port:
                    number: 8080`,

  parallel: `apiVersion: operator.lynq.sh/v1
kind: LynqForm
metadata:
  name: parallel-app
spec:
  hubId: my-hub

  # Level 1: Multiple independent resources (can run in parallel)
  secrets:
    - id: api-secret
      nameTemplate: "{{ .uid }}-api-secret"
      spec:
        stringData:
          key: "{{ .apiKey }}"

  configMaps:
    - id: app-config
      nameTemplate: "{{ .uid }}-config"
      spec:
        data:
          app.conf: "config data"

  persistentVolumeClaims:
    - id: cache-pvc
      nameTemplate: "{{ .uid }}-cache"
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 5Gi

  # Level 2: Two apps depending on the same resources (parallel)
  deployments:
    - id: web-app
      dependIds: ["api-secret", "app-config"]
      nameTemplate: "{{ .uid }}-web"
      spec:
        replicas: 2
        template:
          spec:
            containers:
            - name: web
              image: web:latest

    - id: worker-app
      dependIds: ["api-secret", "app-config"]
      nameTemplate: "{{ .uid }}-worker"
      spec:
        replicas: 1
        template:
          spec:
            containers:
            - name: worker
              image: worker:latest

  # Level 3: Services (parallel)
  services:
    - id: web-service
      dependIds: ["web-app"]
      nameTemplate: "{{ .uid }}-web-svc"
      spec:
        ports:
        - port: 80

    - id: worker-service
      dependIds: ["worker-app"]
      nameTemplate: "{{ .uid }}-worker-svc"
      spec:
        ports:
        - port: 8080`,

  cycle: `apiVersion: operator.lynq.sh/v1
kind: LynqForm
metadata:
  name: invalid-cycle
spec:
  hubId: my-hub

  # This form has a dependency cycle - INVALID!
  deployments:
    - id: service-a
      dependIds: ["service-b"]  # A depends on B
      nameTemplate: "{{ .uid }}-a"
      spec:
        replicas: 1

  services:
    - id: service-b
      dependIds: ["config-c"]  # B depends on C
      nameTemplate: "{{ .uid }}-b"
      spec:
        ports:
        - port: 80

  configMaps:
    - id: config-c
      dependIds: ["service-a"]  # C depends on A - CYCLE!
      nameTemplate: "{{ .uid }}-c"
      spec:
        data:
          config: "data"`
};

// Resource type mapping
const resourceTypeMap = {
  serviceAccounts: 'ServiceAccount',
  deployments: 'Deployment',
  statefulSets: 'StatefulSet',
  daemonSets: 'DaemonSet',
  services: 'Service',
  configMaps: 'ConfigMap',
  secrets: 'Secret',
  persistentVolumeClaims: 'PVC',
  jobs: 'Job',
  cronJobs: 'CronJob',
  ingresses: 'Ingress',
  podDisruptionBudgets: 'PodDisruptionBudget',
  networkPolicies: 'NetworkPolicy',
  horizontalPodAutoscalers: 'HPA',
  manifests: 'Custom',
  namespaces: 'Namespace'
};

// Load preset
const loadPreset = (key) => {
  selectedPreset.value = key;
  yamlInput.value = presets[key];
  analyzeYaml();
};

// Parse YAML and extract dependencies
const analyzeYaml = () => {
  parseError.value = '';
  nodes.value = [];
  edges.value = [];

  try {
    const parsed = yaml.load(yamlInput.value);

    if (!parsed || !parsed.spec) {
      parseError.value = 'Invalid LynqForm: missing spec field';
      return;
    }

    const extractedNodes = [];

    // Extract resources from all resource types
    Object.keys(resourceTypeMap).forEach(resourceType => {
      if (parsed.spec[resourceType]) {
        parsed.spec[resourceType].forEach(resource => {
          if (resource.id) {
            extractedNodes.push({
              id: resource.id,
              type: resourceTypeMap[resourceType],
              dependIds: resource.dependIds || [],
              x: 0,
              y: 0,
              order: null,
              isInCycle: false
            });
          }
        });
      }
    });

    if (extractedNodes.length === 0) {
      parseError.value = 'No resources with "id" field found in form';
      return;
    }

    nodes.value = extractedNodes;
    calculateLayout(); // This will detect cycles and apply appropriate layout
    calculateExecutionOrder();

  } catch (error) {
    parseError.value = `YAML parsing error: ${error.message}`;
  }
};

// Calculate graph layout
const calculateLayout = () => {
  // Build edges first (needed for cycle detection)
  edges.value = [];
  nodes.value.forEach(node => {
    node.dependIds.forEach(depId => {
      edges.value.push({
        from: node.id,
        to: depId,
        isInCycle: false
      });
    });
  });

  // Detect cycles to identify cycle nodes and mark edges
  detectCycles();

  // Apply appropriate layout based on cycle detection
  if (hasCycle.value) {
    // Use circular layout for cycle detection visualization
    layoutForCycle();
  } else {
    // Use hierarchical layout for normal DAG
    layoutHierarchical();
  }
};

// Hierarchical layout for DAG (no cycles)
const layoutHierarchical = () => {
  const levels = buildLevels();

  levels.forEach((level, levelIndex) => {
    const y = (height / (levels.length + 1)) * (levelIndex + 1);
    const spacing = width / (level.length + 1);

    level.forEach((nodeId, nodeIndex) => {
      const node = nodes.value.find(n => n.id === nodeId);
      if (node) {
        node.x = spacing * (nodeIndex + 1);
        node.y = y;
      }
    });
  });
};

// Circular layout for cycle visualization
const layoutForCycle = () => {
  const cycleNodes = nodes.value.filter(n => n.isInCycle);
  const nonCycleNodes = nodes.value.filter(n => !n.isInCycle);

  // Position cycle nodes in a circle
  if (cycleNodes.length > 0) {
    const centerX = width / 2;
    const centerY = height / 2;
    const radius = Math.min(width, height) / 3;

    cycleNodes.forEach((node, index) => {
      const angle = (2 * Math.PI * index) / cycleNodes.length - Math.PI / 2;
      node.x = centerX + radius * Math.cos(angle);
      node.y = centerY + radius * Math.sin(angle);
    });
  }

  // Position non-cycle nodes around the circle
  if (nonCycleNodes.length > 0) {
    // Try to place them based on their relationships
    const placed = new Set();

    // First, place nodes that have dependencies on cycle nodes
    nonCycleNodes.forEach(node => {
      const hasCycleDep = node.dependIds.some(depId =>
        cycleNodes.find(cn => cn.id === depId)
      );

      if (hasCycleDep) {
        // Find the cycle node it depends on
        const depCycleNode = cycleNodes.find(cn => node.dependIds.includes(cn.id));
        if (depCycleNode) {
          // Place it outside the cycle, near the dependency
          const dx = depCycleNode.x - width / 2;
          const dy = depCycleNode.y - height / 2;
          const distance = Math.sqrt(dx * dx + dy * dy);
          const scale = 1.8; // Place further out

          node.x = width / 2 + dx * scale;
          node.y = height / 2 + dy * scale;
          placed.add(node.id);
        }
      }
    });

    // Place nodes that depend on the non-cycle nodes we just placed
    nonCycleNodes.forEach(node => {
      if (!placed.has(node.id)) {
        const hasDep = node.dependIds.some(depId => placed.has(depId));

        if (hasDep) {
          const depNode = nodes.value.find(n => placed.has(n.id) && node.dependIds.includes(n.id));
          if (depNode) {
            const dx = depNode.x - width / 2;
            const dy = depNode.y - height / 2;
            const distance = Math.sqrt(dx * dx + dy * dy);
            const scale = 1.3;

            node.x = width / 2 + dx * scale;
            node.y = height / 2 + dy * scale;
            placed.add(node.id);
          }
        }
      }
    });

    // Place any remaining non-cycle nodes in a separate area
    const remaining = nonCycleNodes.filter(n => !placed.has(n.id));
    if (remaining.length > 0) {
      const startY = 50;
      const spacing = Math.min(width / (remaining.length + 1), 150);

      remaining.forEach((node, index) => {
        node.x = spacing * (index + 1);
        node.y = startY;
        placed.add(node.id);
      });
    }
  }
};

// Build levels for layout
const buildLevels = () => {
  const levels = [];
  const visited = new Set();
  const inDegree = {};

  nodes.value.forEach(node => {
    inDegree[node.id] = node.dependIds.length;
  });

  let currentLevel = nodes.value.filter(n => inDegree[n.id] === 0).map(n => n.id);

  while (currentLevel.length > 0) {
    levels.push([...currentLevel]);
    currentLevel.forEach(nodeId => visited.add(nodeId));

    const nextLevel = [];
    nodes.value.forEach(node => {
      if (!visited.has(node.id)) {
        const allDepsVisited = node.dependIds.every(depId => visited.has(depId));
        if (allDepsVisited) {
          nextLevel.push(node.id);
        }
      }
    });

    currentLevel = nextLevel;
  }

  return levels;
};

// Detect cycles using standard DFS algorithm
const detectCycles = () => {
  const WHITE = 0; // Not visited
  const GRAY = 1;  // Currently visiting (in recursion stack)
  const BLACK = 2; // Completely visited

  const color = {};
  const parent = {};
  const cycleNodes = new Set();
  let cycleStart = null;
  let cycleEnd = null;

  // Initialize all nodes as WHITE
  nodes.value.forEach(node => {
    color[node.id] = WHITE;
    parent[node.id] = null;
  });

  const dfs = (nodeId) => {
    color[nodeId] = GRAY;

    const node = nodes.value.find(n => n.id === nodeId);
    if (node) {
      for (const depId of node.dependIds) {
        // Check if dependency exists in nodes
        if (!nodes.value.find(n => n.id === depId)) {
          continue; // Skip non-existent dependencies
        }

        if (color[depId] === GRAY) {
          // Found a back edge (cycle detected)
          cycleStart = depId;
          cycleEnd = nodeId;
          return true;
        }

        if (color[depId] === WHITE) {
          parent[depId] = nodeId;
          if (dfs(depId)) {
            return true;
          }
        }
      }
    }

    color[nodeId] = BLACK;
    return false;
  };

  // Try to find cycle from each unvisited node
  for (const node of nodes.value) {
    if (color[node.id] === WHITE) {
      if (dfs(node.id)) {
        // Reconstruct cycle path
        const cycle = [];
        let current = cycleEnd;
        cycle.push(current);
        cycleNodes.add(current);

        while (current !== cycleStart) {
          current = parent[current];
          if (!current) break; // Safety check
          cycle.push(current);
          cycleNodes.add(current);
        }

        cycle.reverse();
        cycle.push(cycleStart); // Complete the cycle

        hasCycle.value = true;
        cycleMessage.value = `Dependency cycle: ${cycle.join(' → ')}`;

        // Mark nodes and edges in cycle
        nodes.value.forEach(n => {
          n.isInCycle = cycleNodes.has(n.id);
        });

        edges.value.forEach(edge => {
          edge.isInCycle = cycleNodes.has(edge.from) && cycleNodes.has(edge.to);
        });

        return;
      }
    }
  }

  // No cycle found
  hasCycle.value = false;
  nodes.value.forEach(n => { n.isInCycle = false; });
  edges.value.forEach(e => { e.isInCycle = false; });
};

// Calculate execution order
const calculateExecutionOrder = () => {
  if (hasCycle.value) {
    nodes.value.forEach(n => { n.order = null; });
    return;
  }

  const inDegree = {};
  const queue = [];
  let order = 1;

  nodes.value.forEach(node => {
    inDegree[node.id] = node.dependIds.length;
    if (inDegree[node.id] === 0) {
      queue.push(node.id);
    }
  });

  while (queue.length > 0) {
    const levelSize = queue.length;

    for (let i = 0; i < levelSize; i++) {
      const nodeId = queue.shift();
      const node = nodes.value.find(n => n.id === nodeId);
      if (node) {
        node.order = order;
      }

      nodes.value.forEach(n => {
        if (n.dependIds.includes(nodeId)) {
          inDegree[n.id]--;
          if (inDegree[n.id] === 0) {
            queue.push(n.id);
          }
        }
      });
    }

    order++;
  }
};

// Get node position
const getNodePosition = (nodeId) => {
  const node = nodes.value.find(n => n.id === nodeId);
  return node ? { x: node.x, y: node.y } : { x: 0, y: 0 };
};

// Get edge points (from node border to node border)
const getEdgePoints = (fromId, toId) => {
  const from = getNodePosition(fromId);
  const to = getNodePosition(toId);

  // Calculate angle between nodes
  const dx = to.x - from.x;
  const dy = to.y - from.y;
  const angle = Math.atan2(dy, dx);

  // Calculate border points
  // From node: move FROM center TOWARDS the target (add radius offset)
  const x1 = from.x + nodeRadius * Math.cos(angle);
  const y1 = from.y + nodeRadius * Math.sin(angle);

  // To node: move FROM target center AWAY from source (subtract radius offset)
  const x2 = to.x - nodeRadius * Math.cos(angle);
  const y2 = to.y - nodeRadius * Math.sin(angle);

  return { x1, y1, x2, y2 };
};

// Sorted nodes for table
const sortedNodes = computed(() => {
  return [...nodes.value].sort((a, b) => {
    if (a.order === null && b.order === null) return 0;
    if (a.order === null) return 1;
    if (b.order === null) return -1;
    return a.order - b.order;
  });
});

// Execution order text
const executionOrderText = computed(() => {
  if (hasCycle.value) return 'Cannot determine (cycle detected)';

  const orderGroups = {};
  nodes.value.forEach(node => {
    if (node.order !== null) {
      if (!orderGroups[node.order]) {
        orderGroups[node.order] = [];
      }
      orderGroups[node.order].push(node.id);
    }
  });

  const orderedGroups = Object.keys(orderGroups)
    .sort((a, b) => parseInt(a) - parseInt(b))
    .map(order => {
      const group = orderGroups[order];
      return group.length > 1 ? `[${group.join(', ')}]` : group[0];
    });

  return orderedGroups.join(' → ');
});

// Transform string for SVG
const transformString = computed(() => {
  return `translate(${translateX.value}, ${translateY.value}) scale(${scale.value})`;
});

// Zoom functions
const zoomIn = () => {
  scale.value = Math.min(scale.value * 1.2, 3);
};

const zoomOut = () => {
  scale.value = Math.max(scale.value / 1.2, 0.3);
};

const resetZoom = () => {
  scale.value = 1;
  translateX.value = 0;
  translateY.value = 0;
};

// Mouse wheel zoom
const handleWheel = (event) => {
  const delta = event.deltaY;
  const zoomFactor = delta > 0 ? 0.9 : 1.1;

  const newScale = Math.min(Math.max(scale.value * zoomFactor, 0.3), 3);

  if (newScale !== scale.value) {
    scale.value = newScale;
  }
};

// Pan functions
const startPan = (event) => {
  isPanning.value = true;
  panStartX.value = event.clientX - translateX.value;
  panStartY.value = event.clientY - translateY.value;
  event.currentTarget.style.cursor = 'grabbing';
};

const handlePan = (event) => {
  if (!isPanning.value) return;

  translateX.value = event.clientX - panStartX.value;
  translateY.value = event.clientY - panStartY.value;
};

const endPan = (event) => {
  if (isPanning.value) {
    isPanning.value = false;
    event.currentTarget.style.cursor = 'grab';
  }
};

// Load default preset on mount
loadPreset('simple');
</script>

<style scoped>
.dependency-graph-visualizer {
  margin: 2rem 0;
}

.controls-section {
  margin-bottom: 1.5rem;
  padding: 1rem;
  background: var(--vp-c-bg-soft);
  border-radius: 8px;
  border: 1px solid var(--vp-c-divider);
}

.preset-buttons {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.preset-buttons .label {
  font-weight: 600;
  color: var(--vp-c-text-1);
  font-size: 0.9rem;
}

.preset-btn {
  padding: 0.5rem 1rem;
  background: var(--vp-c-bg);
  border: 1px solid var(--vp-c-divider);
  border-radius: 6px;
  color: var(--vp-c-text-1);
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.preset-btn:hover {
  border-color: var(--vp-c-brand);
  background: var(--vp-c-brand-soft);
  color: var(--vp-c-brand);
}

.preset-btn.active {
  background: var(--vp-c-brand);
  color: white;
  border-color: var(--vp-c-brand);
}

.editor-graph-container {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.5rem;
  margin-bottom: 1.5rem;
}

.editor-panel,
.graph-panel {
  display: flex;
  flex-direction: column;
  background: var(--vp-c-bg);
  border: 1px solid var(--vp-c-divider);
  border-radius: 8px;
  overflow: hidden;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  background: var(--vp-c-bg-soft);
  border-bottom: 1px solid var(--vp-c-divider);
}

.panel-header h4 {
  margin: 0;
  font-size: 0.95rem;
  color: var(--vp-c-text-1);
}

.resource-count {
  font-size: 0.85rem;
  color: var(--vp-c-text-2);
  padding: 0.25rem 0.75rem;
  background: var(--vp-c-brand-soft);
  border-radius: 12px;
}

.analyze-btn {
  padding: 0.4rem 0.75rem;
  background: var(--vp-c-brand);
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.analyze-btn:hover {
  background: var(--vp-c-brand-dark);
  transform: translateY(-1px);
}

.yaml-editor {
  flex: 1;
  padding: 1rem;
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 0.85rem;
  line-height: 1.6;
  background: var(--vp-c-bg);
  color: var(--vp-c-text-1);
  border: none;
  resize: vertical;
  min-height: 400px;
  outline: none;
}

.parse-error {
  padding: 0.75rem 1rem;
  background: var(--vp-c-danger-soft);
  color: #ef5350;
  font-size: 0.85rem;
  border-top: 1px solid var(--vp-c-divider);
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem 2rem;
  color: var(--vp-c-text-2);
  text-align: center;
}

.empty-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
  opacity: 0.5;
}

.empty-state p {
  margin: 0.25rem 0;
  font-size: 0.9rem;
}

.empty-state .hint {
  font-size: 0.85rem;
  font-style: italic;
  color: var(--vp-c-text-3);
}

.graph-wrapper {
  position: relative;
  flex: 1;
  overflow: hidden;
}

.zoom-controls {
  position: absolute;
  top: 1rem;
  right: 1rem;
  display: flex;
  gap: 0.5rem;
  align-items: center;
  background: var(--vp-c-bg-soft);
  padding: 0.5rem;
  border-radius: 8px;
  border: 1px solid var(--vp-c-divider);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  z-index: 10;
}

.zoom-btn {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--vp-c-bg);
  border: 1px solid var(--vp-c-divider);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 1rem;
}

.zoom-btn:hover {
  background: var(--vp-c-brand-soft);
  border-color: var(--vp-c-brand);
  transform: scale(1.05);
}

.zoom-btn:active {
  transform: scale(0.95);
}

.zoom-level {
  font-size: 0.85rem;
  color: var(--vp-c-text-2);
  font-weight: 600;
  min-width: 50px;
  text-align: center;
  padding: 0 0.25rem;
}

.dependency-graph {
  width: 100%;
  height: 500px;
  padding: 1rem;
  cursor: grab;
  user-select: none;
}

.dependency-graph:active {
  cursor: grabbing;
}

.edge-line {
  transition: all 0.3s ease;
}

.edge-error {
  stroke-dasharray: 4 4;
  animation: dashFlow 1s linear infinite;
}

@keyframes dashFlow {
  from { stroke-dashoffset: 0; }
  to { stroke-dashoffset: 8; }
}

.node {
  cursor: pointer;
  transition: all 0.3s ease;
}

.node:hover circle {
  filter: brightness(1.1);
  stroke-width: 3;
}

.node-highlighted circle {
  filter: drop-shadow(0 0 8px var(--vp-c-brand));
}

.node-error circle {
  animation: errorPulse 2s ease-in-out infinite;
}

@keyframes errorPulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}

.order-badge {
  animation: badgePop 0.5s ease-out;
}

@keyframes badgePop {
  0% { transform: scale(0); }
  50% { transform: scale(1.2); }
  100% { transform: scale(1); }
}

.order-text {
  font-weight: 700;
  font-size: 12px;
  fill: white;
  pointer-events: none;
}

.node-label {
  font-weight: 600;
  font-size: 14px;
  pointer-events: none;
}

.node-type {
  font-size: 11px;
  fill: var(--vp-c-text-2);
  pointer-events: none;
}

.analysis-results {
  padding: 1rem;
  border-top: 1px solid var(--vp-c-divider);
}

.error-message,
.success-message {
  padding: 1rem;
  border-radius: 6px;
  border-left: 4px solid;
}

.error-message {
  background: var(--vp-c-danger-soft);
  border-color: #ef5350;
}

.error-message strong {
  color: #ef5350;
  display: block;
  margin-bottom: 0.5rem;
}

.success-message {
  background: var(--vp-c-success-soft);
  border-color: var(--vp-c-brand);
}

.success-message strong {
  color: var(--vp-c-brand);
  display: block;
  margin-bottom: 0.5rem;
}

.error-message p,
.success-message p {
  margin: 0.25rem 0;
  color: var(--vp-c-text-1);
  font-size: 0.9rem;
}

.hint {
  font-size: 0.85rem !important;
  color: var(--vp-c-text-2) !important;
  font-style: italic;
}

.resource-details {
  padding: 1.5rem;
  background: var(--vp-c-bg-soft);
  border-radius: 8px;
  border: 1px solid var(--vp-c-divider);
}

.resource-details h4 {
  margin: 0 0 1rem 0;
  color: var(--vp-c-text-1);
  font-size: 0.95rem;
}

.resource-table {
  background: var(--vp-c-bg);
  border-radius: 6px;
  overflow: hidden;
  border: 1px solid var(--vp-c-divider);
}

.table-header,
.table-row {
  display: grid;
  grid-template-columns: 80px 1fr 150px 2fr;
  gap: 1rem;
  padding: 0.75rem 1rem;
  align-items: center;
}

.table-header {
  background: var(--vp-c-bg-soft);
  font-weight: 600;
  font-size: 0.85rem;
  color: var(--vp-c-text-2);
  border-bottom: 1px solid var(--vp-c-divider);
}

.table-row {
  border-bottom: 1px solid var(--vp-c-divider);
  transition: all 0.2s ease;
}

.table-row:last-child {
  border-bottom: none;
}

.table-row:hover {
  background: var(--vp-c-bg-soft);
}

.table-row.row-error {
  background: var(--vp-c-danger-soft);
}

.order-badge-small {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  background: var(--vp-c-brand);
  color: white;
  border-radius: 50%;
  font-weight: 700;
  font-size: 0.85rem;
}

.error-badge {
  font-size: 1.2rem;
}

.type-badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  background: var(--vp-c-brand-soft);
  color: var(--vp-c-brand);
  border-radius: 4px;
  font-size: 0.8rem;
  font-weight: 500;
}

.no-deps {
  color: var(--vp-c-text-3);
  font-style: italic;
  font-size: 0.85rem;
}

.dep-list {
  color: var(--vp-c-text-1);
  font-size: 0.875rem;
}

/* Responsive */
@media (max-width: 1200px) {
  .editor-graph-container {
    grid-template-columns: 1fr;
  }

  .yaml-editor {
    min-height: 300px;
  }
}

@media (max-width: 768px) {
  .table-header,
  .table-row {
    grid-template-columns: 60px 1fr;
    gap: 0.5rem;
  }

  .col-type,
  .col-deps {
    grid-column: 2;
  }

  .col-id {
    font-weight: 600;
  }

  .preset-buttons {
    flex-direction: column;
    align-items: stretch;
  }

  .preset-btn {
    width: 100%;
  }
}
</style>
