<template>
  <div class="policy-visualizer">
    <!-- Replay Button -->
    <button
      class="replay-button"
      @click="resetAll"
      aria-label="Reset all resources"
    >
      <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
        <path d="M17 10C17 13.866 13.866 17 10 17C6.134 17 3 13.866 3 10C3 6.134 6.134 3 10 3C11.848 3 13.545 3.711 14.828 4.879" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
        <path d="M14 1V5H10" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
      </svg>
      <span>Reset All</span>
    </button>

    <div class="visualizer-head">
      <h3>ConflictPolicy Flow Visualizer</h3>
      <p>
        Trigger ownership conflicts to see how ConflictPolicy handles conflicts.
        Compare <strong>Stuck</strong> (safe halt) vs <strong>Force</strong> (SSA force takeover).
      </p>
    </div>

    <!-- Vertical Flow Diagram -->
    <div class="flow-diagram-vertical">
      <!-- Stage 1: LynqForm -->
      <div class="flow-stage stage-form active">
        <div class="stage-header">
          <div class="stage-icon">📄</div>
          <div class="stage-info">
            <div class="stage-title">LynqForm</div>
            <div class="stage-subtitle">Template Definition</div>
          </div>
        </div>
        <div class="stage-content">
          <div class="template-resources">
            <div class="template-resource">
              <div class="template-resource-info">
                <span class="template-resource-icon">🌐</span>
                <span class="template-resource-name">Service</span>
                <span class="stuck-badge-small">
                  <span class="stuck-icon">⚠️</span>
                  <span>Stuck</span>
                </span>
              </div>
            </div>
            <div class="template-resource">
              <div class="template-resource-info">
                <span class="template-resource-icon">📦</span>
                <span class="template-resource-name">Deployment</span>
                <span class="force-badge-small">
                  <span class="force-icon">⚡</span>
                  <span>Force</span>
                </span>
              </div>
            </div>
          </div>
        </div>
        <div class="stage-status status-active">
          Active
        </div>
      </div>

      <!-- Connection 1 -->
      <svg class="connection-vertical" :class="{ active: true }" viewBox="0 0 60 80">
        <defs>
          <linearGradient id="grad-form-node-conflict" x1="0%" y1="0%" x2="0%" y2="100%">
            <stop offset="0%" stop-color="#667eea" stop-opacity="0" />
            <stop offset="50%" stop-color="#667eea" stop-opacity="0.8" />
            <stop offset="100%" stop-color="#667eea" stop-opacity="0" />
          </linearGradient>
        </defs>
        <path d="M 30 0 L 30 80" stroke="#667eea" stroke-width="2" opacity="0.3" />
        <path d="M 30 0 L 30 80" stroke="url(#grad-form-node-conflict)" stroke-width="2.5" class="animated-line" />
        <circle r="4" fill="#667eea" class="flow-dot">
          <animateMotion dur="2s" repeatCount="indefinite" path="M 30 0 L 30 80" />
        </circle>
      </svg>

      <!-- Stage 2: LynqNode -->
      <div
        class="flow-stage stage-node"
        :class="{
          active: resources.node.exists,
          degraded: resources.node.state === 'degraded'
        }"
      >
        <div class="stage-header">
          <div class="stage-icon">🏢</div>
          <div class="stage-info">
            <div class="stage-title">LynqNode</div>
            <div class="stage-subtitle">acme-corp</div>
          </div>
        </div>
        <div v-if="resources.node.exists" class="stage-status" :class="`status-${resources.node.state}`">
          {{ stateLabels[resources.node.state] }}
        </div>
        <div v-if="resources.node.state === 'degraded'" class="stage-warning">
          ⚠️ Conflict detected - manual intervention required
        </div>
      </div>

      <!-- Connection 2 -->
      <svg class="connection-vertical" :class="{ active: anyClusterResourceExists }" viewBox="0 0 60 100">
        <defs>
          <linearGradient id="grad-node-resources-conflict" x1="0%" y1="0%" x2="0%" y2="100%">
            <stop offset="0%" stop-color="#f59e0b" stop-opacity="0" />
            <stop offset="50%" stop-color="#f59e0b" stop-opacity="0.8" />
            <stop offset="100%" stop-color="#f59e0b" stop-opacity="0" />
          </linearGradient>
        </defs>
        <!-- Multiple paths for multiple resources -->
        <path d="M 20 0 L 20 100" stroke="#f59e0b" stroke-width="1.5" opacity="0.2" />
        <path d="M 40 0 L 40 100" stroke="#f59e0b" stroke-width="1.5" opacity="0.2" />

        <path d="M 20 0 L 20 100" stroke="url(#grad-node-resources-conflict)" stroke-width="2" class="animated-line" style="animation-delay: 0s" />
        <path d="M 40 0 L 40 100" stroke="url(#grad-node-resources-conflict)" stroke-width="2" class="animated-line" style="animation-delay: 0.3s" />
      </svg>

      <!-- Stage 3: Cluster Resources -->
      <div
        v-if="anyClusterResourceExists"
        class="flow-stage stage-resources"
        :class="{
          active: anyClusterResourceExists
        }"
      >
        <div class="stage-header">
          <div class="stage-icon">⚙️</div>
          <div class="stage-info">
            <div class="stage-title">Cluster Resources</div>
            <div class="stage-subtitle">Kubernetes Objects</div>
          </div>
        </div>
        <div class="stage-content">
          <div class="resources-list">
            <!-- Service (ConflictPolicy=Stuck) -->
            <div v-if="resources.service.exists"
                 class="resource-item"
                 :class="{
                   creating: resources.service.state === 'creating',
                   conflict: resources.service.state === 'conflict'
                 }">
              <div class="resource-info">
                <span class="resource-kind">Service</span>
                <span class="resource-name">acme-svc</span>
              </div>
              <div class="resource-actions">
                <span class="stuck-badge" title="Stops on conflict">
                  <span class="stuck-icon">⚠️</span>
                  <span class="stuck-text">Stuck</span>
                </span>
                <span v-if="resources.service.conflictOwner" class="conflict-owner" title="Conflicting field manager">
                  {{ resources.service.conflictOwner }}
                </span>
                <span class="resource-state">{{ stateLabels[resources.service.state] }}</span>
                <button
                  v-if="resources.service.state === 'active'"
                  class="resource-conflict-btn"
                  @click="triggerConflict('service')"
                  title="Simulate ownership conflict"
                >
                  ⚡
                </button>
                <button
                  v-if="resources.service.state === 'conflict'"
                  class="resource-resolve-btn"
                  @click="resolveConflict('service')"
                  title="Manually resolve conflict"
                >
                  🧹
                </button>
              </div>
            </div>

            <!-- Deployment (ConflictPolicy=Force) -->
            <div v-if="resources.deployment.exists"
                 class="resource-item"
                 :class="{
                   creating: resources.deployment.state === 'creating',
                   forcing: resources.deployment.state === 'forcing'
                 }">
              <div class="resource-info">
                <span class="resource-kind">Deployment</span>
                <span class="resource-name">acme-api</span>
              </div>
              <div class="resource-actions">
                <span class="force-badge" title="Force takes ownership">
                  <span class="force-icon">⚡</span>
                  <span class="force-text">Force</span>
                </span>
                <span v-if="resources.deployment.conflictOwner && resources.deployment.state === 'forcing'" class="conflict-owner conflict-owner-forcing" title="Evicting field manager">
                  Evicting {{ resources.deployment.conflictOwner }}
                </span>
                <span class="resource-state">{{ stateLabels[resources.deployment.state] }}</span>
                <button
                  v-if="resources.deployment.state === 'active'"
                  class="resource-conflict-btn"
                  @click="triggerConflict('deployment')"
                  title="Simulate ownership conflict"
                >
                  ⚡
                </button>
              </div>
            </div>
          </div>

          <!-- Policy explanation -->
          <div v-if="anyClusterResourceExists" class="policy-explanation">
            <div class="explanation-item">
              <span class="stuck-badge-small">
                <span class="stuck-icon">⚠️</span>
                <span>Stuck</span>
              </span>
              <span class="explanation-text">
                Halts reconciliation on conflict, marks LynqNode as Degraded (safe but requires manual fix)
              </span>
            </div>
            <div class="explanation-item">
              <span class="force-badge-small">
                <span class="force-icon">⚡</span>
                <span>Force</span>
              </span>
              <span class="explanation-text">
                Uses SSA force=true to take ownership (aggressive, keeps reconciliation healthy)
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Event Log -->
    <div class="event-log" :class="{ visible: eventLog.length > 0 }">
      <div class="event-log-title">
        <span>Event Log</span>
        <span class="event-count">{{ eventLog.length }}</span>
      </div>
      <div class="event-list">
        <div
          v-for="event in eventLog"
          :key="event.id"
          class="event-item"
          :class="`event-${event.type}`"
        >
          <div class="event-indicator"></div>
          <div class="event-content">
            <div class="event-message">{{ event.message }}</div>
            <div class="event-time">{{ event.time }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue';

const stateLabels = {
  creating: 'Creating...',
  active: 'Active',
  conflict: 'Conflict',
  forcing: 'Force Applying...',
  healthy: 'Healthy',
  degraded: 'Degraded'
};

const eventLog = ref([]);
let eventCounter = 0;
let timeoutIds = [];

const conflictingControllers = ['argo-rollouts', 'keda', 'cert-manager', 'istio'];

const resources = ref({
  node: {
    exists: true,
    state: 'healthy'
  },
  service: {
    exists: true,
    state: 'active',
    conflictPolicy: 'Stuck',
    conflictOwner: null
  },
  deployment: {
    exists: true,
    state: 'active',
    conflictPolicy: 'Force',
    conflictOwner: null
  }
});

// Helper to check if any cluster resource exists
const anyClusterResourceExists = computed(() => {
  return resources.value.service.exists || resources.value.deployment.exists;
});

const logEvent = (message, type = 'info') => {
  eventCounter += 1;
  eventLog.value = [
    {
      id: eventCounter,
      message,
      type,
      time: new Date().toLocaleTimeString()
    },
    ...eventLog.value
  ].slice(0, 10);
};

// Trigger conflict for a resource
const triggerConflict = (resourceType) => {
  const resource = resources.value[resourceType];
  if (resource.state !== 'active') return;

  // Pick random conflicting controller
  const conflictOwner = conflictingControllers[Math.floor(Math.random() * conflictingControllers.length)];
  resource.conflictOwner = conflictOwner;

  logEvent(`⚡ Ownership conflict detected: ${resourceType} field manager conflict with ${conflictOwner}`, 'conflict');

  if (resource.conflictPolicy === 'Stuck') {
    // ConflictPolicy=Stuck: Stop reconciliation, mark as conflict
    const timeoutId = setTimeout(() => {
      resource.state = 'conflict';
      resources.value.node.state = 'degraded';
      logEvent(`❌ ${resourceType} reconciliation halted (ConflictPolicy=Stuck)`, 'conflict');
      logEvent(`⚠️ LynqNode marked as Degraded, ResourceConflict event emitted`, 'warning');
    }, 600);
    timeoutIds.push(timeoutId);
  } else {
    // ConflictPolicy=Force: Use SSA force=true to take ownership
    const timeoutId = setTimeout(() => {
      resource.state = 'forcing';
      logEvent(`⚡ ${resourceType} using SSA force=true to take ownership from ${conflictOwner}...`, 'reconciling');

      const forceTimeout = setTimeout(() => {
        resource.state = 'active';
        resource.conflictOwner = null;
        logEvent(`✓ ${resourceType} ownership transferred to Lynq (${conflictOwner} evicted from managedFields)`, 'success');
      }, 1200);
      timeoutIds.push(forceTimeout);
    }, 600);
    timeoutIds.push(timeoutId);
  }
};

// Manually resolve conflict (for Stuck policy)
const resolveConflict = (resourceType) => {
  const resource = resources.value[resourceType];
  if (resource.state !== 'conflict') return;

  logEvent(`🧹 Manually resolving conflict for ${resourceType}...`, 'info');

  const timeoutId = setTimeout(() => {
    resource.state = 'active';
    resource.conflictOwner = null;

    // Check if all resources are healthy to restore node state
    const hasConflicts = resources.value.service.state === 'conflict' ||
                         resources.value.deployment.state === 'conflict';

    if (!hasConflicts) {
      resources.value.node.state = 'healthy';
      logEvent(`✓ LynqNode restored to Healthy state`, 'success');
    }

    logEvent(`✓ ${resourceType} conflict resolved, SSA completed`, 'success');
  }, 1000);
  timeoutIds.push(timeoutId);
};

// Reset all resources to initial state
const resetAll = () => {
  // Clear all timeouts
  timeoutIds.forEach(id => clearTimeout(id));
  timeoutIds = [];

  // Reset all resources
  resources.value = {
    node: {
      exists: true,
      state: 'healthy'
    },
    service: {
      exists: true,
      state: 'active',
      conflictPolicy: 'Stuck',
      conflictOwner: null
    },
    deployment: {
      exists: true,
      state: 'active',
      conflictPolicy: 'Force',
      conflictOwner: null
    }
  };

  eventLog.value = [];
  logEvent('🔄 All resources reset to initial state', 'info');
};

onMounted(() => {
  logEvent('✓ Initial state: All resources active, no conflicts', 'success');
});

onBeforeUnmount(() => {
  timeoutIds.forEach(id => clearTimeout(id));
  timeoutIds = [];
});
</script>

<style scoped>
.policy-visualizer {
  position: relative;
  width: 100%;
  min-height: 600px;
  padding: 2.5rem 2rem;
  background: var(--vp-c-bg-soft);
  border-radius: 16px;
}

/* Replay Button */
.replay-button {
  position: absolute;
  top: 1.5rem;
  right: 1.5rem;
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
  z-index: 30;
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.15);
}

.replay-button:hover {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.2) 0%, rgba(118, 75, 162, 0.2) 100%);
  border-color: rgba(102, 126, 234, 0.5);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.25);
}

.replay-button svg {
  color: var(--vp-c-brand);
}

.replay-button:hover svg {
  animation: rotateIcon 0.6s ease-in-out;
}

@keyframes rotateIcon {
  0%, 100% { transform: rotate(0); }
  50% { transform: rotate(180deg); }
}

/* Header */
.visualizer-head {
  margin-bottom: 2rem;
}

.visualizer-head h3 {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--vp-c-text-1);
  margin: 0 0 0.5rem;
}

.visualizer-head p {
  font-size: 1rem;
  line-height: 1.6;
  color: var(--vp-c-text-2);
  margin: 0;
}

/* Vertical Flow Diagram */
.flow-diagram-vertical {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0;
  max-width: 700px;
  margin: 0 auto;
}

/* Flow Stages */
.flow-stage {
  position: relative;
  width: 100%;
  padding: 1.5rem;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.05) 0%, rgba(118, 75, 162, 0.05) 100%);
  border: 2px solid rgba(102, 126, 234, 0.2);
  border-radius: 12px;
  transition: all 0.4s ease;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}

.flow-stage.active {
  opacity: 1;
}

.flow-stage.degraded {
  border-color: rgba(237, 137, 54, 0.5);
  background: linear-gradient(135deg, rgba(237, 137, 54, 0.08) 0%, rgba(237, 137, 54, 0.03) 100%);
  animation: degradedPulse 2s ease-in-out infinite;
}

@keyframes degradedPulse {
  0%, 100% {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05), 0 0 0 0 rgba(237, 137, 54, 0.4);
  }
  50% {
    box-shadow: 0 4px 20px rgba(237, 137, 54, 0.15), 0 0 0 8px rgba(237, 137, 54, 0);
  }
}

/* Stage Header */
.stage-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1rem;
}

.stage-icon {
  font-size: 2.5rem;
  line-height: 1;
  flex-shrink: 0;
}

.stage-info {
  flex: 1;
}

.stage-title {
  font-size: 1.2rem;
  font-weight: 700;
  color: var(--vp-c-text-1);
  margin-bottom: 0.25rem;
}

.stage-subtitle {
  font-size: 0.9rem;
  color: var(--vp-c-text-3);
}

/* Stage Status */
.stage-status {
  display: inline-block;
  padding: 0.4rem 0.9rem;
  border-radius: 6px;
  font-size: 0.85rem;
  font-weight: 600;
  margin-top: 0.5rem;
}

.stage-status.status-active {
  background: rgba(66, 184, 131, 0.15);
  color: #42b883;
  border: 1px solid rgba(66, 184, 131, 0.3);
}

.stage-status.status-healthy {
  background: rgba(66, 184, 131, 0.15);
  color: #42b883;
  border: 1px solid rgba(66, 184, 131, 0.3);
}

.stage-status.status-degraded {
  background: rgba(237, 137, 54, 0.15);
  color: #ed8936;
  border: 1px solid rgba(237, 137, 54, 0.3);
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}

/* Stage Warning */
.stage-warning {
  padding: 0.75rem 1rem;
  background: rgba(237, 137, 54, 0.1);
  border: 1px solid rgba(237, 137, 54, 0.3);
  border-radius: 8px;
  font-size: 0.85rem;
  color: #ed8936;
  margin-top: 1rem;
  font-weight: 500;
}

/* Stage Content */
.stage-content {
  margin-top: 1rem;
}

/* Template Resources */
.template-resources {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.template-resource {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1rem;
  background: var(--vp-c-bg);
  border: 1px solid var(--vp-c-divider);
  border-radius: 6px;
  transition: all 0.3s ease;
}

.template-resource-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.template-resource-icon {
  font-size: 1.2rem;
  line-height: 1;
}

.template-resource-name {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--vp-c-text-1);
  min-width: 100px;
}

/* Resources List */
.resources-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.resource-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1rem;
  background: var(--vp-c-bg);
  border: 1px solid var(--vp-c-divider);
  border-radius: 6px;
  transition: all 0.3s ease;
}

.resource-item.creating {
  border-color: rgba(66, 184, 131, 0.4);
  background: linear-gradient(135deg, rgba(66, 184, 131, 0.05) 0%, rgba(66, 184, 131, 0.02) 100%);
  animation: fadeIn 1s ease-out;
}

.resource-item.conflict {
  border-color: rgba(237, 137, 54, 0.4);
  background: linear-gradient(135deg, rgba(237, 137, 54, 0.08) 0%, rgba(237, 137, 54, 0.03) 100%);
  animation: conflictPulse 2s ease-in-out infinite;
}

.resource-item.forcing {
  border-color: rgba(102, 126, 234, 0.4);
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.08) 0%, rgba(102, 126, 234, 0.03) 100%);
  animation: forcingPulse 1.5s ease-in-out infinite;
}

@keyframes fadeIn {
  from { opacity: 0.3; transform: scale(0.98); }
  to { opacity: 1; transform: scale(1); }
}

@keyframes conflictPulse {
  0%, 100% {
    box-shadow: 0 0 0 0 rgba(237, 137, 54, 0.4);
  }
  50% {
    box-shadow: 0 0 0 4px rgba(237, 137, 54, 0);
  }
}

@keyframes forcingPulse {
  0%, 100% {
    box-shadow: 0 0 0 0 rgba(102, 126, 234, 0.4);
  }
  50% {
    box-shadow: 0 0 0 4px rgba(102, 126, 234, 0);
  }
}

.resource-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.resource-kind {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--vp-c-brand);
  background: rgba(var(--vp-c-brand-rgb), 0.1);
  padding: 0.3rem 0.6rem;
  border-radius: 4px;
  min-width: 75px;
  text-align: center;
}

.resource-name {
  font-size: 0.85rem;
  color: var(--vp-c-text-2);
  font-family: monospace;
}

.resource-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.stuck-badge {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.25rem 0.6rem;
  background: rgba(237, 137, 54, 0.15);
  color: #ed8936;
  border: 1px solid rgba(237, 137, 54, 0.3);
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.stuck-icon {
  font-size: 0.8rem;
  line-height: 1;
}

.stuck-text {
  line-height: 1;
}

.force-badge {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.25rem 0.6rem;
  background: rgba(102, 126, 234, 0.15);
  color: #667eea;
  border: 1px solid rgba(102, 126, 234, 0.3);
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.force-icon {
  font-size: 0.8rem;
  line-height: 1;
}

.force-text {
  line-height: 1;
}

.conflict-owner {
  font-size: 0.7rem;
  font-weight: 600;
  color: #e53e3e;
  padding: 0.25rem 0.5rem;
  background: rgba(229, 62, 62, 0.1);
  border: 1px solid rgba(229, 62, 62, 0.3);
  border-radius: 4px;
  font-family: monospace;
}

.conflict-owner-forcing {
  color: #667eea;
  background: rgba(102, 126, 234, 0.1);
  border-color: rgba(102, 126, 234, 0.3);
}

.resource-state {
  font-size: 0.75rem;
  color: var(--vp-c-text-3);
  min-width: 60px;
  text-align: center;
}

.resource-conflict-btn {
  padding: 0.3rem 0.5rem;
  background: linear-gradient(135deg, rgba(237, 137, 54, 0.05) 0%, rgba(237, 137, 54, 0.02) 100%);
  border: 1px solid rgba(237, 137, 54, 0.3);
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.3s ease;
  font-size: 0.9rem;
  line-height: 1;
}

.resource-conflict-btn:hover {
  background: linear-gradient(135deg, rgba(237, 137, 54, 0.1) 0%, rgba(237, 137, 54, 0.05) 100%);
  border-color: rgba(237, 137, 54, 0.5);
  transform: scale(1.1);
}

.resource-resolve-btn {
  padding: 0.3rem 0.5rem;
  background: linear-gradient(135deg, rgba(66, 184, 131, 0.05) 0%, rgba(66, 184, 131, 0.02) 100%);
  border: 1px solid rgba(66, 184, 131, 0.3);
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.3s ease;
  font-size: 0.9rem;
  line-height: 1;
}

.resource-resolve-btn:hover {
  background: linear-gradient(135deg, rgba(66, 184, 131, 0.1) 0%, rgba(66, 184, 131, 0.05) 100%);
  border-color: rgba(66, 184, 131, 0.5);
  transform: scale(1.1);
}

/* Policy Explanation */
.policy-explanation {
  margin-top: 1rem;
  padding: 1rem;
  background: var(--vp-c-bg-soft);
  border: 1px solid var(--vp-c-divider);
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.explanation-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.stuck-badge-small {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.2rem 0.5rem;
  background: rgba(237, 137, 54, 0.15);
  color: #ed8936;
  border: 1px solid rgba(237, 137, 54, 0.3);
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  white-space: nowrap;
}

.force-badge-small {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.2rem 0.5rem;
  background: rgba(102, 126, 234, 0.15);
  color: #667eea;
  border: 1px solid rgba(102, 126, 234, 0.3);
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  white-space: nowrap;
}

.explanation-text {
  font-size: 0.85rem;
  color: var(--vp-c-text-2);
  line-height: 1.4;
}

/* Connections */
.connection-vertical {
  width: 60px;
  height: 100px;
  opacity: 0;
  transition: opacity 0.5s ease;
  margin: -2px 0;
}

.connection-vertical.active {
  opacity: 1;
}

.animated-line {
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
  filter: drop-shadow(0 0 4px currentColor);
}

/* Event Log */
.event-log {
  margin-top: 2.5rem;
  padding: 1.5rem;
  background: var(--vp-c-bg);
  border: 1px solid var(--vp-c-divider);
  border-radius: 12px;
  max-height: 0;
  overflow: hidden;
  opacity: 0;
  transition: all 0.5s ease;
}

.event-log.visible {
  max-height: 500px;
  opacity: 1;
}

.event-log-title {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--vp-c-divider);
  font-size: 1rem;
  font-weight: 700;
  color: var(--vp-c-text-1);
}

.event-count {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--vp-c-text-3);
  background: var(--vp-c-bg-soft);
  padding: 0.25rem 0.6rem;
  border-radius: 12px;
}

.event-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  max-height: 400px;
  overflow-y: auto;
}

.event-item {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 0.75rem;
  background: var(--vp-c-bg-soft);
  border-radius: 8px;
  border-left: 3px solid transparent;
  animation: slideInLeft 0.3s ease-out;
}

@keyframes slideInLeft {
  from {
    opacity: 0;
    transform: translateX(-10px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

.event-item.event-info {
  border-left-color: #667eea;
}

.event-item.event-success {
  border-left-color: #42b883;
}

.event-item.event-conflict {
  border-left-color: #ed8936;
}

.event-item.event-warning {
  border-left-color: #ed8936;
}

.event-item.event-reconciling {
  border-left-color: #41d1ff;
}

.event-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-top: 0.3rem;
  flex-shrink: 0;
}

.event-item.event-info .event-indicator {
  background: #667eea;
  box-shadow: 0 0 6px rgba(102, 126, 234, 0.4);
}

.event-item.event-success .event-indicator {
  background: #42b883;
  box-shadow: 0 0 6px rgba(66, 184, 131, 0.4);
}

.event-item.event-conflict .event-indicator {
  background: #ed8936;
  box-shadow: 0 0 6px rgba(237, 137, 54, 0.4);
  animation: pulse 2s ease-in-out infinite;
}

.event-item.event-warning .event-indicator {
  background: #ed8936;
  box-shadow: 0 0 6px rgba(237, 137, 54, 0.4);
}

.event-item.event-reconciling .event-indicator {
  background: #41d1ff;
  box-shadow: 0 0 6px rgba(65, 209, 255, 0.4);
  animation: pulse 2s ease-in-out infinite;
}

.event-content {
  flex: 1;
}

.event-message {
  font-size: 0.9rem;
  line-height: 1.4;
  color: var(--vp-c-text-1);
  margin-bottom: 0.25rem;
}

.event-time {
  font-size: 0.75rem;
  color: var(--vp-c-text-3);
}

/* Responsive */
@media (max-width: 768px) {
  .policy-visualizer {
    padding: 2rem 1.5rem;
  }

  .replay-button {
    top: 1rem;
    right: 1rem;
    padding: 0.5rem 1rem;
    font-size: 0.85rem;
  }

  .visualizer-head h3 {
    font-size: 1.5rem;
  }

  .stage-header {
    flex-wrap: wrap;
  }
}
</style>
