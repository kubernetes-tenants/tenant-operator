<template>
  <div class="how-it-works-diagram">
    <!-- Replay Button -->
    <button
      v-if="isCompleted"
      class="replay-button"
      @click="restartAnimation"
      aria-label="Replay animation"
    >
      <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
        <path d="M17 10C17 13.866 13.866 17 10 17C6.134 17 3 13.866 3 10C3 6.134 6.134 3 10 3C11.848 3 13.545 3.711 14.828 4.879" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
        <path d="M14 1V5H10" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
      </svg>
      <span>Replay</span>
    </button>

    <div class="diagram-stage">
      <!-- Stage 1: Database -->
      <div class="stage-item database-stage" :class="{ active: currentStage >= 1 }">
        <div class="stage-title">Your Database</div>
        <div class="database-table">
          <div class="table-header">node_configs</div>
          <div class="table-body">
            <div
              class="table-row"
              :class="{
                highlight: currentStage >= 1,
                interactive: isCompleted,
                inactive: !nodes.acme.active,
                'highlight-prompt': isCompleted && !hasInteracted
              }"
              @click="isCompleted && toggleNode('acme')"
            >
              <span class="row-id">acme-corp</span>
              <span class="row-domain">acme.com</span>
              <span class="row-status" :class="{ active: nodes.acme.active }">
                {{ nodes.acme.active ? '✓ active' : '✗ inactive' }}
              </span>
            </div>
            <div
              class="table-row"
              :class="{
                highlight: currentStage >= 1,
                interactive: isCompleted,
                inactive: !nodes.beta.active,
                'highlight-prompt': isCompleted && !hasInteracted
              }"
              @click="isCompleted && toggleNode('beta')"
            >
              <span class="row-id">beta-inc</span>
              <span class="row-domain">beta.io</span>
              <span class="row-status" :class="{ active: nodes.beta.active }">
                {{ nodes.beta.active ? '✓ active' : '✗ inactive' }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- Connection Line 1: Database → Registry -->
      <svg class="connection-line line-1" :class="{ active: currentStage >= 2 }" viewBox="0 0 200 100">
        <defs>
          <linearGradient id="line-gradient-1" x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" stop-color="#42b883" stop-opacity="0" />
            <stop offset="50%" stop-color="#42b883" stop-opacity="0.8" />
            <stop offset="100%" stop-color="#42b883" stop-opacity="0" />
          </linearGradient>
        </defs>
        <path d="M 0 50 L 200 50" stroke="#42b883" stroke-width="2" fill="none" opacity="0.3" />
        <path d="M 0 50 L 200 50" stroke="url(#line-gradient-1)" stroke-width="2.5" fill="none" class="animated-path" />
        <circle r="4" fill="#42b883" class="flow-dot">
          <animateMotion dur="2s" repeatCount="indefinite" path="M 0 50 L 200 50" />
        </circle>
      </svg>

      <!-- Stage 2: LynqHub -->
      <div class="stage-item registry-stage" :class="{ active: currentStage >= 2 }">
        <div class="k8s-cluster-label">Kubernetes Cluster</div>
        <div class="registry-box clickable" @click="openModal('registry')">
          <div class="resource-icon">📋</div>
          <div class="resource-title">LynqHub</div>
          <div class="resource-subtitle">Syncs every 30 seconds</div>
          <div class="click-overlay">
            <span class="click-overlay-text">Click to view YAML</span>
          </div>
        </div>
      </div>

      <!-- LynqForm (separate position) -->
      <div class="stage-item template-stage" :class="{ active: currentStage >= 3 }">
        <div class="template-box clickable" @click="openModal('template')">
          <div class="resource-icon small">📄</div>
          <div class="resource-title small">LynqForm</div>
          <div class="click-overlay">
            <span class="click-overlay-text">Click to view YAML</span>
          </div>
        </div>
      </div>

      <!-- Connection Line 2b: Template → Nodes -->
      <svg class="connection-line line-2" :class="{ active: currentStage >= 3 }" viewBox="0 0 80 140">
        <defs>
          <linearGradient id="line-gradient-2b" x1="0%" y1="0%" x2="0%" y2="100%">
            <stop offset="0%" stop-color="#667eea" stop-opacity="0" />
            <stop offset="50%" stop-color="#667eea" stop-opacity="0.8" />
            <stop offset="100%" stop-color="#667eea" stop-opacity="0" />
          </linearGradient>
        </defs>
        <path d="M 40 0 L 40 140" stroke="#667eea" stroke-width="2" fill="none" opacity="0.3" />
        <path d="M 40 0 L 40 140" stroke="url(#line-gradient-2b)" stroke-width="2.5" fill="none" class="animated-path" />
        <circle r="4" fill="#667eea" class="flow-dot">
          <animateMotion dur="1.5s" repeatCount="indefinite" path="M 40 0 L 40 140" begin="0.3s" />
        </circle>
      </svg>

      <!-- Stage 3: LynqNode CRs -->
      <div class="stage-item nodes-stage" :class="{ active: currentStage >= 3 }">
        <div class="stage-subtitle">LynqNode CRs (Auto-created)</div>
        <div class="node-crs">
          <div
            v-if="nodes.acme.active"
            class="node-cr"
            :class="{ 'fade-out': isCompleted && !nodes.acme.active }"
            :style="{ animationDelay: '0s' }"
          >
            <div class="cr-icon">🏢</div>
            <div class="cr-name">acme-corp</div>
          </div>
          <div
            v-if="nodes.beta.active"
            class="node-cr"
            :class="{ 'fade-out': isCompleted && !nodes.beta.active }"
            :style="{ animationDelay: '0.15s' }"
          >
            <div class="cr-icon">🏢</div>
            <div class="cr-name">beta-inc</div>
          </div>
        </div>
      </div>

      <!-- Connection Lines 3: Nodes → Resources -->
      <svg class="connection-line line-3" :class="{ active: currentStage >= 4 }" viewBox="0 0 200 200">
        <defs>
          <linearGradient id="line-gradient-3" x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" stop-color="#41d1ff" stop-opacity="0" />
            <stop offset="50%" stop-color="#41d1ff" stop-opacity="0.8" />
            <stop offset="100%" stop-color="#41d1ff" stop-opacity="0" />
          </linearGradient>
        </defs>
        <!-- Multiple paths for multiple resources -->
        <path d="M 0 60 Q 100 40 200 30" stroke="#41d1ff" stroke-width="1.5" fill="none" opacity="0.2" />
        <path d="M 0 100 Q 100 90 200 100" stroke="#41d1ff" stroke-width="1.5" fill="none" opacity="0.2" />
        <path d="M 0 140 Q 100 150 200 170" stroke="#41d1ff" stroke-width="1.5" fill="none" opacity="0.2" />

        <path d="M 0 60 Q 100 40 200 30" stroke="url(#line-gradient-3)" stroke-width="2" fill="none" class="animated-path" style="animation-delay: 0s" />
        <path d="M 0 100 Q 100 90 200 100" stroke="url(#line-gradient-3)" stroke-width="2" fill="none" class="animated-path" style="animation-delay: 0.3s" />
        <path d="M 0 140 Q 100 150 200 170" stroke="url(#line-gradient-3)" stroke-width="2" fill="none" class="animated-path" style="animation-delay: 0.6s" />
      </svg>

      <!-- Stage 4: Kubernetes Resources -->
      <div class="stage-item resources-stage" :class="{ active: currentStage >= 4 }">
        <div class="stage-subtitle">Kubernetes Resources</div>
        <div class="resources-grid">
          <div v-if="nodes.acme.active" class="resource-item" :style="{ animationDelay: '0s' }">
            <span class="resource-kind">Deploy</span>
            <span class="resource-name">acme-corp-api</span>
          </div>
          <div v-if="nodes.acme.active" class="resource-item" :style="{ animationDelay: '0.1s' }">
            <span class="resource-kind">Svc</span>
            <span class="resource-name">acme-corp-svc</span>
          </div>
          <div v-if="nodes.acme.active" class="resource-item" :style="{ animationDelay: '0.2s' }">
            <span class="resource-kind">Ingress</span>
            <span class="resource-name">acme-corp-ing</span>
          </div>
          <div v-if="nodes.beta.active" class="resource-item" :style="{ animationDelay: '0.3s' }">
            <span class="resource-kind">Deploy</span>
            <span class="resource-name">beta-inc-api</span>
          </div>
          <div v-if="nodes.beta.active" class="resource-item" :style="{ animationDelay: '0.4s' }">
            <span class="resource-kind">Svc</span>
            <span class="resource-name">beta-inc-svc</span>
          </div>
          <div v-if="nodes.beta.active" class="resource-item" :style="{ animationDelay: '0.5s' }">
            <span class="resource-kind">Ingress</span>
            <span class="resource-name">beta-inc-ing</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Progress indicator -->
    <div class="progress-dots">
      <div v-for="stage in 4" :key="stage" class="dot" :class="{ active: currentStage >= stage }" @click="currentStage = stage"></div>
    </div>

    <!-- Modal -->
    <Transition name="modal">
      <div v-if="showModal" class="modal-overlay" @click="closeModal">
        <div class="modal-content" @click.stop>
          <div class="modal-header">
            <div>
              <h3 class="modal-title">{{ modalContent?.title }}</h3>
              <p class="modal-subtitle">{{ modalContent?.subtitle }}</p>
            </div>
            <button class="modal-close" @click="closeModal" aria-label="Close modal">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
                <path d="M18 6L6 18M6 6l12 12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
              </svg>
            </button>
          </div>
          <div class="modal-body">
            <div class="code-header">
              <span class="code-language">yaml</span>
              <button class="copy-button" @click="copyToClipboard">
                <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
                  <rect x="5" y="5" width="9" height="9" rx="1" stroke="currentColor" stroke-width="1.5"/>
                  <path d="M3 10.5V3C3 2.44772 3.44772 2 4 2H10.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
                </svg>
                Copy
              </button>
            </div>
            <pre class="code-block"><code>{{ modalContent?.code }}</code></pre>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue';

const currentStage = ref(0);
const isCompleted = ref(false);
const hasInteracted = ref(false);
let timeoutIds = [];

// Node active states
const nodes = ref({
  acme: { active: true },
  beta: { active: true }
});

// Modal state
const showModal = ref(false);
const modalContent = ref(null);

const registryYaml = `# Connect to your node database
apiVersion: operator.lynq.sh/v1
kind: LynqHub
metadata:
  name: production-nodes
spec:
  source:
    type: mysql
    syncInterval: 30s
    mysql:
      host: mysql.default.svc.cluster.local
      port: 3306
      database: nodes
      table: node_configs
      passwordRef:
        name: mysql-credentials
        key: password
  valueMappings:
    uid: node_id
    hostOrUrl: domain
    activate: is_active
    extraValueMappings:
      planId: plan_id
      deployImage: deploy_image`;

const templateYaml = `# Define what to create per node
apiVersion: operator.lynq.sh/v1
kind: LynqForm
metadata:
  name: saas-stack
spec:
  registryId: production-nodes
  deployments:
    - id: api-deployment
      nameTemplate: "{{ .uid }}-api"
      spec:
        replicas: 2
        selector:
          matchLabels:
            app: "{{ .uid }}-api"
        template:
          metadata:
            labels:
              app: "{{ .uid }}-api"
          spec:
            containers:
              - name: api
                image: "{{ .deployImage | default \\"myapp:latest\\" }}"
                env:
                  - name: NODE_ID
                    value: "{{ .uid }}"
                  - name: NODE_HOST
                    value: "{{ .host }}"
  services:
    - id: api-service
      nameTemplate: "{{ .uid }}-svc"
      dependIds: ["api-deployment"]
      spec:
        selector:
          app: "{{ .uid }}-api"
        ports:
          - port: 80
            targetPort: 8080
  ingresses:
    - id: api-ingress
      nameTemplate: "{{ .uid }}-ingress"
      dependIds: ["api-service"]
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
                        number: 80`;

const openModal = (type) => {
  modalContent.value = type === 'registry' ? {
    title: 'LynqHub',
    subtitle: 'Connect to your node database',
    code: registryYaml
  } : {
    title: 'LynqForm',
    subtitle: 'Define what to create per node',
    code: templateYaml
  };
  showModal.value = true;
  document.body.style.overflow = 'hidden';
};

const closeModal = () => {
  showModal.value = false;
  document.body.style.overflow = '';
};

const copyToClipboard = async () => {
  try {
    await navigator.clipboard.writeText(modalContent.value.code);
  } catch (err) {
    console.error('Failed to copy:', err);
  }
};

const stageTimings = [
  { stage: 1, delay: 500 },
  { stage: 2, delay: 2500 },
  { stage: 3, delay: 4500 },
  { stage: 4, delay: 6500 },
];

const playAnimation = () => {
  isCompleted.value = false;
  currentStage.value = 0;

  // Reset all nodes to active when replaying
  nodes.value.acme.active = true;
  nodes.value.beta.active = true;

  // Clear any existing timeouts
  timeoutIds.forEach(id => clearTimeout(id));
  timeoutIds = [];

  // Start animation sequence
  stageTimings.forEach(({ stage, delay }) => {
    const timeoutId = setTimeout(() => {
      currentStage.value = stage;

      // Mark as completed when reaching final stage
      if (stage === 4) {
        setTimeout(() => {
          isCompleted.value = true;
        }, 1000);
      }
    }, delay);
    timeoutIds.push(timeoutId);
  });
};

const restartAnimation = () => {
  playAnimation();
};

const toggleNode = (nodeKey) => {
  if (!isCompleted.value) return;

  // Mark as interacted on first click
  if (!hasInteracted.value) {
    hasInteracted.value = true;
  }

  nodes.value[nodeKey].active = !nodes.value[nodeKey].active;
};

onMounted(() => {
  // Auto-play animation on mount
  playAnimation();

  // Add escape key listener
  window.addEventListener('keydown', handleEscapeKey);
});

onUnmounted(() => {
  // Clear all timeouts on unmount
  timeoutIds.forEach(id => clearTimeout(id));
  timeoutIds = [];

  // Remove escape key listener
  window.removeEventListener('keydown', handleEscapeKey);
});

const handleEscapeKey = (e) => {
  if (e.key === 'Escape' && showModal.value) {
    closeModal();
  }
};
</script>

<style scoped>
.how-it-works-diagram {
  position: relative;
  width: 100%;
  min-height: 650px;
  padding: 2.5rem 2rem;
  background: var(--vp-c-bg-soft);
  border-radius: 16px;
  overflow-x: auto;
}

/* Replay Button */
.replay-button {
  position: absolute;
  top: 1.5rem;
  left: 1.5rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.65rem 1.25rem;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
  border: 1.5px solid rgba(102, 126, 234, 0.3);
  border-radius: 8px;
  color: var(--vp-c-text-1);
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  z-index: 20;
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.15);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
}

.replay-button:hover {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.2) 0%, rgba(118, 75, 162, 0.2) 100%);
  border-color: rgba(102, 126, 234, 0.5);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.25);
}

.replay-button:active {
  transform: translateY(0);
  box-shadow: 0 2px 6px rgba(102, 126, 234, 0.2);
}

.replay-button svg {
  flex-shrink: 0;
  color: var(--vp-c-brand);
  animation: rotateIn 0.3s ease-out;
}

.replay-button:hover svg {
  animation: rotateIcon 0.6s ease-in-out;
}

@keyframes rotateIn {
  from {
    opacity: 0;
    transform: rotate(-90deg) scale(0.8);
  }
  to {
    opacity: 1;
    transform: rotate(0) scale(1);
  }
}

@keyframes rotateIcon {
  0%, 100% {
    transform: rotate(0);
  }
  50% {
    transform: rotate(180deg);
  }
}

.diagram-stage {
  position: relative;
  width: min-content;
  min-width: 1000px;
  height: 550px;
  display: grid;
  grid-template-columns: 280px 120px 240px 120px 300px;
  grid-template-rows: 140px 140px 220px;
  gap: 0;
  align-items: center;
  justify-items: center;
  margin: 0 auto;
}

/* Stage Items */
.stage-item {
  opacity: 0;
  transform: translateY(20px) scale(0.95);
  transition: all 0.8s cubic-bezier(0.4, 0, 0.2, 1);
  z-index: 10;
  position: relative;
}

.stage-item.active {
  opacity: 1;
  transform: translateY(0) scale(1);
}

.stage-title {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--vp-c-text-1);
  margin-bottom: 0.75rem;
  text-align: center;
}

.stage-subtitle {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--vp-c-text-2);
  margin-bottom: 0.75rem;
  text-align: center;
}

/* Database Stage */
.database-stage {
  grid-column: 1;
  grid-row: 1 / 4;
  align-self: center;
  justify-self: start;
  padding-left: 1rem;
}

.database-table {
  background: linear-gradient(135deg, rgba(66, 184, 131, 0.1) 0%, rgba(66, 184, 131, 0.05) 100%);
  border: 2px solid rgba(66, 184, 131, 0.3);
  border-radius: 12px;
  overflow: hidden;
  width: 260px;
  box-shadow: 0 4px 16px rgba(66, 184, 131, 0.15);
}

.table-header {
  background: rgba(66, 184, 131, 0.2);
  padding: 0.75rem 1rem;
  font-weight: 600;
  font-size: 0.9rem;
  color: var(--vp-c-text-1);
  border-bottom: 1px solid rgba(66, 184, 131, 0.3);
}

.table-body {
  padding: 0.5rem;
}

.table-row {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  background: var(--vp-c-bg);
  border-radius: 6px;
  margin-bottom: 0.5rem;
  opacity: 0.5;
  transform: scale(0.98);
  transition: all 0.3s ease;
}

.table-row.highlight {
  opacity: 1;
  transform: scale(1);
  box-shadow: 0 2px 8px rgba(66, 184, 131, 0.2);
}

.table-row.interactive {
  cursor: pointer;
  user-select: none;
}

.table-row.interactive:hover {
  background: var(--vp-c-bg-soft);
  box-shadow: 0 3px 12px rgba(66, 184, 131, 0.3);
  transform: scale(1.02);
}

.table-row.interactive:active {
  transform: scale(0.98);
}

.table-row.inactive {
  opacity: 0.4;
  background: var(--vp-c-bg-mute);
}

.table-row.inactive .row-status {
  color: var(--vp-c-text-3);
}

.table-row.highlight-prompt {
  animation: promptPulse 2s ease-in-out infinite;
  position: relative;
}

.table-row.highlight-prompt:first-of-type::after {
  content: '👇 Click to toggle';
  position: absolute;
  top: -28px;
  left: 50%;
  transform: translateX(-50%);
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--vp-c-brand);
  background: var(--vp-c-bg);
  padding: 0.25rem 0.75rem;
  border-radius: 4px;
  border: 1px solid var(--vp-c-brand);
  white-space: nowrap;
  animation: promptBounce 1.5s ease-in-out infinite;
  pointer-events: none;
  z-index: 10;
}

@keyframes promptPulse {
  0%, 100% {
    box-shadow: 0 2px 8px rgba(66, 184, 131, 0.2), 0 0 0 0 rgba(102, 126, 234, 0.4);
  }
  50% {
    box-shadow: 0 4px 16px rgba(66, 184, 131, 0.4), 0 0 0 8px rgba(102, 126, 234, 0);
  }
}

@keyframes promptBounce {
  0%, 100% {
    transform: translateX(-50%) translateY(0);
  }
  50% {
    transform: translateX(-50%) translateY(-4px);
  }
}

.table-row:last-child {
  margin-bottom: 0;
}

.row-id {
  font-weight: 600;
  font-size: 0.85rem;
  color: var(--vp-c-text-1);
}

.row-domain {
  font-size: 0.85rem;
  color: var(--vp-c-text-2);
}

.row-status {
  font-size: 0.8rem;
  font-weight: 500;
}

.row-status.active {
  color: #42b883;
}

/* Connection Lines */
.connection-line {
  position: relative;
  opacity: 0;
  transition: opacity 0.5s ease;
  z-index: 1;
}

.connection-line.active {
  opacity: 1;
}

.line-1 {
  grid-column: 2;
  grid-row: 1 / 4;
  width: 120px;
  height: 80px;
  align-self: center;
}

.line-2 {
  grid-column: 3;
  grid-row: 2 / 3;
  width: 80px;
  height: 140px;
  align-self: end;
  top: 120px;
  /* margin-bottom: -40px; */
}

.line-3 {
  grid-column: 4;
  grid-row: 1 / 4;
  width: 120px;
  height: 200px;
  align-self: center;
}

.animated-path {
  stroke-dasharray: 0 1000;
  animation: flowLine 3s ease-in-out infinite;
}

@keyframes flowLine {
  0% {
    stroke-dasharray: 0 1000;
    opacity: 0;
  }
  40% {
    opacity: 1;
  }
  100% {
    stroke-dasharray: 1000 0;
    opacity: 0;
  }
}

.flow-dot {
  filter: drop-shadow(0 0 6px currentColor);
  opacity: 0.9;
}

/* Registry Stage */
.registry-stage {
  grid-column: 3;
  grid-row: 1;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  align-items: center;
  align-self: center;
}

.k8s-cluster-label {
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--vp-c-brand);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 0.5rem;
}

.registry-box {
  position: relative;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
  border: 2px solid rgba(102, 126, 234, 0.3);
  border-radius: 12px;
  padding: 1.25rem 1rem;
  width: 200px;
  text-align: center;
  box-shadow: 0 4px 16px rgba(102, 126, 234, 0.15);
}

.registry-box.clickable,
.template-box.clickable {
  cursor: pointer;
  transition: all 0.3s ease;
}

.registry-box.clickable:hover,
.template-box.clickable:hover {
  transform: translateY(-4px);
  border-color: rgba(102, 126, 234, 0.5);
  box-shadow: 0 6px 24px rgba(102, 126, 234, 0.25);
}

.registry-box.clickable:active,
.template-box.clickable:active {
  transform: translateY(-2px);
}

.click-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  border-radius: inherit;
  opacity: 0;
  transition: opacity 0.3s ease;
  pointer-events: none;
}

.clickable:hover .click-overlay {
  opacity: 1;
}

.click-overlay-text {
  font-size: 0.85rem;
  font-weight: 600;
  color: white;
  text-align: center;
  padding: 0.5rem 1rem;
  border-radius: 6px;
  background: rgba(102, 126, 234, 0.9);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

.template-box {
  position: relative;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.08) 0%, rgba(118, 75, 162, 0.08) 100%);
  border: 1.5px solid rgba(102, 126, 234, 0.25);
  border-radius: 8px;
  padding: 0.75rem 1rem;
  width: 180px;
  text-align: center;
  opacity: 0;
  transform: scale(0.9);
  transition: all 0.5s ease;
}

.template-stage.active .template-box {
  opacity: 1;
  transform: scale(1);
}

.resource-icon {
  font-size: 1.75rem;
  margin-bottom: 0.4rem;
}

.resource-icon.small {
  font-size: 1.35rem;
  margin-bottom: 0.3rem;
}

.resource-title {
  font-weight: 600;
  font-size: 1rem;
  color: var(--vp-c-text-1);
  margin-bottom: 0.25rem;
}

.resource-title.small {
  font-size: 0.875rem;
}

.resource-subtitle {
  font-size: 0.75rem;
  color: var(--vp-c-text-2);
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

/* Template Stage */
.template-stage {
  grid-column: 3;
  grid-row: 2;
  align-self: center;
  z-index: 15;
  position: relative;
}

/* Nodes Stage */
.nodes-stage {
  grid-column: 3;
  grid-row: 3;
  align-self: start;
  padding-top: 120px;
}

.node-crs {
  display: flex;
  gap: 1rem;
  justify-content: center;
}

.node-cr {
  background: linear-gradient(135deg, rgba(65, 209, 255, 0.1) 0%, rgba(65, 209, 255, 0.05) 100%);
  border: 2px solid rgba(65, 209, 255, 0.3);
  border-radius: 10px;
  padding: 0.85rem 1.25rem;
  text-align: center;
  width: 100px;
  opacity: 0;
  transform: translateY(20px);
  animation: slideInUp 0.6s ease-out forwards;
  box-shadow: 0 4px 12px rgba(65, 209, 255, 0.15);
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.node-cr.fade-out {
  opacity: 0;
  transform: translateY(20px);
}

@keyframes slideInUp {
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.cr-icon {
  font-size: 1.35rem;
  margin-bottom: 0.4rem;
}

.cr-name {
  font-weight: 600;
  font-size: 0.8rem;
  color: var(--vp-c-text-1);
}

/* Resources Stage */
.resources-stage {
  grid-column: 5;
  grid-row: 1 / 4;
  align-self: center;
  justify-self: end;
  padding-right: 1rem;
}

.resources-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 0.5rem;
  width: 280px;
}

.resource-item {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 0.75rem;
  align-items: center;
  background: var(--vp-c-bg);
  border: 1px solid var(--vp-c-divider);
  border-radius: 8px;
  padding: 0.65rem 0.85rem;
  opacity: 0;
  transform: translateX(-20px);
  animation: slideInRight 0.5s ease-out forwards;
  transition: opacity 0.3s ease, transform 0.3s ease, border-color 0.3s ease, box-shadow 0.3s ease;
}

.resource-item:hover {
  border-color: var(--vp-c-brand);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transform: translateX(4px);
}

@keyframes slideInRight {
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

.resource-kind {
  font-size: 0.7rem;
  font-weight: 600;
  color: var(--vp-c-brand);
  background: rgba(var(--vp-c-brand-rgb), 0.1);
  padding: 0.2rem 0.45rem;
  border-radius: 4px;
  white-space: nowrap;
}

.resource-name {
  font-size: 0.8rem;
  color: var(--vp-c-text-2);
  font-family: monospace;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* Progress Dots */
.progress-dots {
  display: flex;
  justify-content: center;
  gap: 0.75rem;
  margin-top: 2rem;
}

.dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: var(--vp-c-divider);
  cursor: pointer;
  transition: all 0.3s ease;
}

.dot:hover {
  background: var(--vp-c-text-2);
  transform: scale(1.2);
}

.dot.active {
  background: var(--vp-c-brand);
  transform: scale(1.3);
  box-shadow: 0 0 8px rgba(var(--vp-c-brand-rgb), 0.5);
}

/* Responsive */
@media (max-width: 1200px) {
  .how-it-works-diagram {
    padding: 2rem 1.5rem;
    min-height: auto;
  }

  .diagram-stage {
    grid-template-columns: 1fr;
    grid-template-rows: auto auto auto auto auto auto auto auto auto;
    gap: 1.5rem;
    height: auto;
    min-width: 0;
    width: 100%;
  }

  .database-stage {
    grid-column: 1;
    grid-row: 1;
    justify-self: center;
    padding-left: 0;
  }

  .line-1 {
    grid-column: 1;
    grid-row: 2;
    width: 80px;
    height: 60px;
    transform: rotate(90deg);
  }

  .registry-stage {
    grid-column: 1;
    grid-row: 3;
  }

  .template-stage {
    grid-column: 1;
    grid-row: 5;
  }

  .line-2 {
    grid-column: 1;
    grid-row: 6;
    width: 60px;
    height: 60px;
    /* transform: rotate(90deg); */
    margin-bottom: 0;
    align-self: center;
    top: 0;
  }

  .nodes-stage {
    grid-column: 1;
    grid-row: 7;
    padding-top: 0;
  }

  .line-3 {
    grid-column: 1;
    grid-row: 8;
    width: 80px;
    height: 80px;
    transform: rotate(90deg);
  }

  .resources-stage {
    grid-column: 1;
    grid-row: 9;
    justify-self: center;
    padding-right: 0;
  }

  .database-table {
    width: 100%;
    max-width: 350px;
  }

  .resources-grid {
    width: 100%;
    max-width: 350px;
  }
}

@media (max-width: 640px) {
  .how-it-works-diagram {
    padding: 1.5rem 1rem;
  }

  .replay-button {
    top: 1rem;
    left: 1rem;
    padding: 0.5rem 1rem;
    font-size: 0.85rem;
    gap: 0.4rem;
  }

  .replay-button svg {
    width: 16px;
    height: 16px;
  }

  .stage-title {
    font-size: 1rem;
  }

  .stage-subtitle {
    font-size: 0.85rem;
  }

  .node-crs {
    flex-direction: column;
    gap: 0.75rem;
  }

  .node-cr {
    width: 100%;
  }

  .database-table,
  .resources-grid {
    max-width: 280px;
  }
}

/* Modal Styles */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal-content {
  background: var(--vp-c-bg);
  border: 1px solid var(--vp-c-divider);
  border-radius: 12px;
  width: 100%;
  max-width: 800px;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  overflow: hidden;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 1.5rem;
  border-bottom: 1px solid var(--vp-c-divider);
  background: var(--vp-c-bg-soft);
}

.modal-title {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--vp-c-text-1);
}

.modal-subtitle {
  margin: 0.25rem 0 0;
  font-size: 0.9rem;
  color: var(--vp-c-text-2);
}

.modal-close {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  color: var(--vp-c-text-2);
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.2s ease;
}

.modal-close:hover {
  background: var(--vp-c-bg-mute);
  color: var(--vp-c-text-1);
}

.modal-body {
  padding: 1.5rem;
  overflow-y: auto;
  flex: 1;
}

.code-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
  padding: 0.5rem 0.75rem;
  background: var(--vp-c-bg-soft);
  border-radius: 6px 6px 0 0;
}

.code-language {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--vp-c-text-2);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.copy-button {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.4rem 0.75rem;
  background: transparent;
  border: 1px solid var(--vp-c-divider);
  border-radius: 6px;
  color: var(--vp-c-text-2);
  font-size: 0.8rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.copy-button:hover {
  background: var(--vp-c-bg);
  border-color: var(--vp-c-brand);
  color: var(--vp-c-brand);
}

.code-block {
  margin: 0;
  padding: 1.25rem;
  background: var(--vp-code-block-bg);
  border: 1px solid var(--vp-c-divider);
  border-radius: 0 0 6px 6px;
  overflow-x: auto;
  font-family: 'Menlo', 'Monaco', 'Courier New', monospace;
  font-size: 0.875rem;
  line-height: 1.7;
  color: var(--vp-c-text-1);
}

.code-block code {
  display: block;
  white-space: pre;
}

/* Modal Transitions */
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.3s ease;
}

.modal-enter-active .modal-content,
.modal-leave-active .modal-content {
  transition: transform 0.3s ease, opacity 0.3s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .modal-content,
.modal-leave-to .modal-content {
  transform: scale(0.95);
  opacity: 0;
}

@media (max-width: 768px) {
  .modal-content {
    max-width: 100%;
    max-height: 90vh;
  }

  .modal-header {
    padding: 1.25rem;
  }

  .modal-body {
    padding: 1.25rem;
  }

  .modal-title {
    font-size: 1.25rem;
  }

  .code-block {
    font-size: 0.8rem;
    padding: 1rem;
  }
}
</style>
