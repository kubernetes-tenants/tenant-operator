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
      <h3>DeletionPolicy Flow Visualizer</h3>
      <p>
        Remove resources from LynqForm template or delete LynqNode to see how DeletionPolicy controls cleanup behavior.
        Compare <strong>Delete</strong> (automatic removal) vs <strong>Retain</strong> (keeps in cluster with orphan markers).
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
            <div class="template-resource" :class="{ removed: resources.deployment.removedFromTemplate }">
              <div class="template-resource-info">
                <span class="template-resource-icon">📦</span>
                <span class="template-resource-name">Deployment</span>
                <span class="delete-badge-small">
                  <span class="delete-icon">🗑️</span>
                  <span>Delete</span>
                </span>
              </div>
              <button
                v-if="!resources.deployment.removedFromTemplate"
                class="template-remove-btn"
                @click="removeFromTemplate('deployment')"
                title="Remove from template"
              >
                ✕
              </button>
              <span v-else class="template-removed-label">Removed</span>
            </div>
            <div class="template-resource" :class="{ removed: resources.pvc.removedFromTemplate }">
              <div class="template-resource-info">
                <span class="template-resource-icon">💾</span>
                <span class="template-resource-name">PVC</span>
                <span class="retain-badge-small">
                  <span class="retain-icon">🔒</span>
                  <span>Retain</span>
                </span>
              </div>
              <button
                v-if="!resources.pvc.removedFromTemplate"
                class="template-remove-btn"
                @click="removeFromTemplate('pvc')"
                title="Remove from template"
              >
                ✕
              </button>
              <span v-else class="template-removed-label">Removed</span>
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
          <linearGradient id="grad-form-node" x1="0%" y1="0%" x2="0%" y2="100%">
            <stop offset="0%" stop-color="#667eea" stop-opacity="0" />
            <stop offset="50%" stop-color="#667eea" stop-opacity="0.8" />
            <stop offset="100%" stop-color="#667eea" stop-opacity="0" />
          </linearGradient>
        </defs>
        <path d="M 30 0 L 30 80" stroke="#667eea" stroke-width="2" opacity="0.3" />
        <path d="M 30 0 L 30 80" stroke="url(#grad-form-node)" stroke-width="2.5" class="animated-line" />
        <circle r="4" fill="#667eea" class="flow-dot">
          <animateMotion dur="2s" repeatCount="indefinite" path="M 30 0 L 30 80" />
        </circle>
      </svg>

      <!-- Stage 2: LynqNode -->
      <div
        class="flow-stage stage-node"
        :class="{
          active: resources.node.exists,
          finalizing: resources.node.state === 'finalizing',
          deleting: resources.node.state === 'deleting'
        }"
      >
        <div class="stage-header">
          <div class="stage-icon">🏢</div>
          <div class="stage-info">
            <div class="stage-title">LynqNode</div>
            <div class="stage-subtitle">acme-corp</div>
          </div>
          <div class="stage-actions">
            <button
              v-if="resources.node.exists && resources.node.state === 'active'"
              class="action-btn delete-btn"
              @click="deleteNode"
              title="Delete LynqNode (triggers finalizer)"
            >
              <span class="btn-icon">🗑️</span>
              <span class="btn-label">Delete LynqNode</span>
            </button>
          </div>
        </div>
        <div v-if="resources.node.exists" class="stage-status" :class="`status-${resources.node.state}`">
          {{ stateLabels[resources.node.state] }}
        </div>
      </div>

      <!-- Connection 2 -->
      <svg v-if="anyClusterResourceExists" class="connection-vertical" :class="{ active: anyClusterResourceExists }" viewBox="0 0 60 100">
        <defs>
          <linearGradient id="grad-node-resources" x1="0%" y1="0%" x2="0%" y2="100%">
            <stop offset="0%" stop-color="#f59e0b" stop-opacity="0" />
            <stop offset="50%" stop-color="#f59e0b" stop-opacity="0.8" />
            <stop offset="100%" stop-color="#f59e0b" stop-opacity="0" />
          </linearGradient>
        </defs>
        <!-- Multiple paths for multiple resources -->
        <path d="M 20 0 L 20 100" stroke="#f59e0b" stroke-width="1.5" opacity="0.2" />
        <path d="M 40 0 L 40 100" stroke="#f59e0b" stroke-width="1.5" opacity="0.2" />

        <path d="M 20 0 L 20 100" stroke="url(#grad-node-resources)" stroke-width="2" class="animated-line" style="animation-delay: 0s" />
        <path d="M 40 0 L 40 100" stroke="url(#grad-node-resources)" stroke-width="2" class="animated-line" style="animation-delay: 0.3s" />
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
            <!-- Deployment (DeletionPolicy=Delete) -->
            <div v-if="resources.deployment.exists"
                 class="resource-item"
                 :class="{
                   creating: resources.deployment.state === 'creating',
                   deleting: resources.deployment.state === 'deleting'
                 }">
              <div class="resource-info">
                <span class="resource-kind">Deployment</span>
                <span class="resource-name">acme-api</span>
              </div>
              <div class="resource-actions">
                <span class="delete-badge" title="Will be deleted with LynqNode">
                  <span class="delete-icon">🗑️</span>
                  <span class="delete-text">Delete</span>
                </span>
                <span class="tracking-badge" title="Uses ownerReference for automatic cleanup">
                  <span class="tracking-text">ownerRef</span>
                </span>
                <span class="resource-state">{{ stateLabels[resources.deployment.state] }}</span>
                <button
                  v-if="resources.deployment.state === 'active'"
                  class="resource-delete-btn"
                  @click="deleteResource('deployment')"
                  title="Delete Deployment"
                >
                  🗑️
                </button>
              </div>
            </div>

            <!-- PVC (DeletionPolicy=Retain) -->
            <div v-if="resources.pvc.exists"
                 class="resource-item resource-item-retained"
                 :class="{
                   creating: resources.pvc.state === 'creating',
                   retained: resources.pvc.state === 'retained'
                 }">
              <div class="resource-info">
                <span class="resource-kind resource-kind-pvc">PVC</span>
                <span class="resource-name">acme-data</span>
              </div>
              <div class="resource-actions">
                <span class="retain-badge" title="Will be retained when LynqNode is deleted">
                  <span class="retain-icon">🔒</span>
                  <span class="retain-text">Retain</span>
                </span>
                <span class="tracking-badge" title="Uses label-based tracking (no ownerReference)">
                  <span class="tracking-text">labels</span>
                </span>
                <span class="resource-state">{{ stateLabels[resources.pvc.state] }}</span>
                <button
                  v-if="resources.pvc.state === 'active' || resources.pvc.state === 'retained'"
                  class="resource-delete-btn"
                  @click="deleteResource('pvc')"
                  title="Delete PVC manually"
                >
                  🗑️
                </button>
              </div>
            </div>

            <!-- Deployment Orphan Info (when retained) -->
            <div v-if="resources.deployment.state === 'retained'" class="orphan-info">
              <div class="orphan-title">🏷️ Orphan Markers Added</div>
              <div class="orphan-markers">
                <code>lynq.sh/orphaned: "true"</code>
                <code>lynq.sh/orphaned-at: "{{ new Date().toISOString() }}"</code>
                <code>lynq.sh/orphaned-reason: "{{ getOrphanReason('deployment') }}"</code>
              </div>
              <p class="orphan-note">
                Resource stays in cluster. Manual cleanup required or will be re-adopted if template is re-applied.
              </p>
            </div>

            <!-- PVC Orphan Info (when retained) -->
            <div v-if="resources.pvc.state === 'retained'" class="orphan-info">
              <div class="orphan-title">🏷️ Orphan Markers Added</div>
              <div class="orphan-markers">
                <code>lynq.sh/orphaned: "true"</code>
                <code>lynq.sh/orphaned-at: "{{ new Date().toISOString() }}"</code>
                <code>lynq.sh/orphaned-reason: "{{ getOrphanReason('pvc') }}"</code>
              </div>
              <p class="orphan-note">
                Resource stays in cluster. Manual cleanup required or will be re-adopted if template is re-applied.
              </p>
            </div>
          </div>

          <!-- Policy explanation -->
          <div v-if="anyClusterResourceExists" class="policy-explanation">
            <div class="explanation-item">
              <span class="delete-badge-small">
                <span class="delete-icon">🗑️</span>
                <span>Delete</span>
              </span>
              <span class="explanation-text">
                Automatic cleanup via ownerReference (Kubernetes garbage collector)
              </span>
            </div>
            <div class="explanation-item">
              <span class="retain-badge-small">
                <span class="retain-icon">🔒</span>
                <span>Retain</span>
              </span>
              <span class="explanation-text">
                Stays in cluster with orphan markers (label-based tracking, no ownerReference)
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
  finalizing: 'Finalizing...',
  deleting: 'Deleting...',
  deleted: 'Deleted',
  retained: 'Retained'
};

const eventLog = ref([]);
let eventCounter = 0;
let timeoutIds = [];

const resources = ref({
  node: {
    exists: true,
    state: 'active'
  },
  deployment: {
    exists: true,
    state: 'active',
    deletionPolicy: 'Delete',
    hasOwnerReference: true,
    removedFromTemplate: false
  },
  pvc: {
    exists: true,
    state: 'active',
    deletionPolicy: 'Retain',
    hasOwnerReference: false,
    removedFromTemplate: false
  }
});

// Helper to check if any cluster resource exists
const anyClusterResourceExists = computed(() => {
  return resources.value.deployment.exists || resources.value.pvc.exists;
});

// Get orphan reason for a resource
const getOrphanReason = (resourceType) => {
  const resource = resources.value[resourceType];
  if (resource.removedFromTemplate) {
    return 'RemovedFromTemplate';
  }
  return 'LynqNodeDeleted';
};

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

// Remove resource from template
const removeFromTemplate = (resourceType) => {
  const resource = resources.value[resourceType];
  if (resource.removedFromTemplate) return;

  // Mark as removed from template
  resource.removedFromTemplate = true;
  logEvent(`📝 ${resourceType} removed from LynqForm template`, 'info');

  const timeoutId = setTimeout(() => {
    if (resource.deletionPolicy === 'Delete') {
      // DeletionPolicy=Delete: Remove resource from cluster
      logEvent(`🗑️ ${resourceType} has DeletionPolicy=Delete, removing from cluster...`, 'deleting');
      resource.state = 'deleting';

      const deleteTimeout = setTimeout(() => {
        resource.exists = false;
        resource.state = 'deleted';
        logEvent(`✓ ${resourceType} deleted (removed from template)`, 'deleted');
      }, 1000);
      timeoutIds.push(deleteTimeout);
    } else {
      // DeletionPolicy=Retain: Keep in cluster with orphan markers
      logEvent(`🔒 ${resourceType} has DeletionPolicy=Retain, adding orphan markers (reason: RemovedFromTemplate)...`, 'info');

      const retainTimeout = setTimeout(() => {
        resource.state = 'retained';
        logEvent(`✓ ${resourceType} retained with orphan markers (lynq.sh/orphaned-reason=RemovedFromTemplate)`, 'success');
      }, 800);
      timeoutIds.push(retainTimeout);
    }
  }, 600);
  timeoutIds.push(timeoutId);
};

// Delete LynqNode with finalizer simulation
const deleteNode = () => {
  if (!resources.value.node.exists || resources.value.node.state !== 'active') return;

  logEvent('🗑️ Delete request for LynqNode', 'info');

  // Step 1: Enter finalizing state
  resources.value.node.state = 'finalizing';
  logEvent('⏳ Finalizer triggered for LynqNode', 'finalizing');

  const timeoutId = setTimeout(() => {
    // Step 2: Handle DeletionPolicy for each resource
    handleResourceDeletion();
  }, 800);
  timeoutIds.push(timeoutId);
};

// Handle resource deletion based on DeletionPolicy
const handleResourceDeletion = () => {
  let deletedCount = 0;
  let totalToCheck = 0;

  // Deployment (DeletionPolicy=Delete)
  if (resources.value.deployment.exists) {
    totalToCheck++;
    logEvent('🗑️ Deployment has DeletionPolicy=Delete, removing via ownerReference...', 'deleting');

    const deploymentTimeout = setTimeout(() => {
      resources.value.deployment.state = 'deleting';

      const deleteTimeout = setTimeout(() => {
        resources.value.deployment.exists = false;
        resources.value.deployment.state = 'deleted';
        deletedCount++;
        logEvent('✓ Deployment deleted (ownerReference removed by Kubernetes GC)', 'deleted');

        checkNodeDeletionComplete(deletedCount, totalToCheck);
      }, 700);
      timeoutIds.push(deleteTimeout);
    }, 400);
    timeoutIds.push(deploymentTimeout);
  }

  // PVC (DeletionPolicy=Retain)
  if (resources.value.pvc.exists && resources.value.pvc.state === 'active') {
    totalToCheck++;
    logEvent('🔒 PVC has DeletionPolicy=Retain, removing tracking labels and adding orphan markers...', 'info');

    const pvcTimeout = setTimeout(() => {
      resources.value.pvc.state = 'retained';
      deletedCount++;
      logEvent('✓ PVC retained with orphan markers (lynq.sh/orphaned=true)', 'success');

      checkNodeDeletionComplete(deletedCount, totalToCheck);
    }, 400);
    timeoutIds.push(pvcTimeout);
  }

  // If no resources to handle
  if (totalToCheck === 0) {
    completeNodeDeletion();
  }
};

// Check if all resources are handled
const checkNodeDeletionComplete = (deletedCount, totalToCheck) => {
  if (deletedCount === totalToCheck) {
    const timeoutId = setTimeout(() => {
      completeNodeDeletion();
    }, 400);
    timeoutIds.push(timeoutId);
  }
};

// Complete node deletion
const completeNodeDeletion = () => {
  logEvent('🗑️ Removing finalizer from LynqNode', 'deleting');
  resources.value.node.state = 'deleting';

  const timeoutId = setTimeout(() => {
    resources.value.node.exists = false;
    resources.value.node.state = 'deleted';
    logEvent('✓ LynqNode deleted', 'deleted');
  }, 700);
  timeoutIds.push(timeoutId);
};

// Delete individual resource
const deleteResource = (resourceType) => {
  const resource = resources.value[resourceType];
  if (!resource.exists) return;

  if (resourceType === 'pvc' && resource.state === 'retained') {
    // Manual cleanup of retained resource
    logEvent(`🗑️ Manually deleting retained ${resourceType}...`, 'deleting');
    resource.state = 'deleting';

    const timeoutId = setTimeout(() => {
      resource.exists = false;
      resource.state = 'deleted';
      logEvent(`✓ ${resourceType} manually deleted`, 'deleted');
    }, 1000);
    timeoutIds.push(timeoutId);
    return;
  }

  if (resource.state !== 'active') return;

  logEvent(`🗑️ Deleting ${resourceType}...`, 'deleting');
  resource.state = 'deleting';

  const timeoutId = setTimeout(() => {
    resource.exists = false;
    resource.state = 'deleted';
    logEvent(`✓ ${resourceType} deleted`, 'deleted');

    // Recreate after deletion if LynqNode is active
    if (resources.value.node.exists && resources.value.node.state === 'active') {
      const recreateTimeout = setTimeout(() => {
        logEvent(`🔄 LynqNode controller detected missing ${resourceType}, recreating...`, 'reconciling');

        const createTimeout = setTimeout(() => {
          recreateResource(resourceType);
        }, 600);
        timeoutIds.push(createTimeout);
      }, 800);
      timeoutIds.push(recreateTimeout);
    }
  }, 1000);
  timeoutIds.push(timeoutId);
};

// Recreate resource
const recreateResource = (resourceType) => {
  const resource = resources.value[resourceType];

  logEvent(`✨ Creating ${resourceType} (DeletionPolicy=${resource.deletionPolicy})...`, 'creating');
  resource.state = 'creating';
  resource.exists = true;

  const timeoutId = setTimeout(() => {
    resource.state = 'active';

    if (resource.deletionPolicy === 'Delete') {
      logEvent(`✓ ${resourceType} created with ownerReference`, 'success');
    } else {
      logEvent(`✓ ${resourceType} created with label-based tracking (no ownerReference)`, 'success');
    }
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
      state: 'active'
    },
    deployment: {
      exists: true,
      state: 'active',
      deletionPolicy: 'Delete',
      hasOwnerReference: true,
      removedFromTemplate: false
    },
    pvc: {
      exists: true,
      state: 'active',
      deletionPolicy: 'Retain',
      hasOwnerReference: false,
      removedFromTemplate: false
    }
  };

  eventLog.value = [];
  logEvent('🔄 All resources reset to initial state', 'info');
};

onMounted(() => {
  logEvent('✓ Initial state: All resources active', 'success');
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

.flow-stage.finalizing {
  border-color: rgba(245, 158, 11, 0.5);
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.08) 0%, rgba(245, 158, 11, 0.03) 100%);
  animation: finalizingPulse 2s ease-in-out infinite;
}

.flow-stage.deleting {
  border-color: rgba(229, 62, 62, 0.4);
  background: linear-gradient(135deg, rgba(229, 62, 62, 0.05) 0%, rgba(229, 62, 62, 0.02) 100%);
  animation: fadeOut 0.7s ease-out;
}

@keyframes finalizingPulse {
  0%, 100% {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05), 0 0 0 0 rgba(245, 158, 11, 0.4);
  }
  50% {
    box-shadow: 0 4px 20px rgba(245, 158, 11, 0.15), 0 0 0 8px rgba(245, 158, 11, 0);
  }
}

@keyframes fadeOut {
  from { opacity: 1; transform: scale(1); }
  to { opacity: 0.3; transform: scale(0.98); }
}

@keyframes fadeIn {
  from { opacity: 0.3; transform: scale(0.98); }
  to { opacity: 1; transform: scale(1); }
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

/* Stage Actions */
.stage-actions {
  display: flex;
  gap: 0.5rem;
  flex-shrink: 0;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.6rem 1rem;
  border: 1.5px solid;
  border-radius: 8px;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
}

.delete-btn {
  background: linear-gradient(135deg, rgba(229, 62, 62, 0.05) 0%, rgba(229, 62, 62, 0.02) 100%);
  border-color: rgba(229, 62, 62, 0.3);
  color: #e53e3e;
}

.delete-btn:hover {
  background: linear-gradient(135deg, rgba(229, 62, 62, 0.1) 0%, rgba(229, 62, 62, 0.05) 100%);
  border-color: rgba(229, 62, 62, 0.5);
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(229, 62, 62, 0.2);
}

.btn-icon {
  font-size: 1rem;
  line-height: 1;
}

.btn-label {
  line-height: 1;
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

.stage-status.status-finalizing {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
  border: 1px solid rgba(245, 158, 11, 0.3);
  animation: pulse 2s ease-in-out infinite;
}

.stage-status.status-deleting {
  background: rgba(229, 62, 62, 0.15);
  color: #e53e3e;
  border: 1px solid rgba(229, 62, 62, 0.3);
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
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

.template-resource.removed {
  opacity: 0.5;
  border-color: rgba(229, 62, 62, 0.3);
  background: linear-gradient(135deg, rgba(229, 62, 62, 0.03) 0%, rgba(229, 62, 62, 0.01) 100%);
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

.template-remove-btn {
  padding: 0.3rem 0.6rem;
  background: linear-gradient(135deg, rgba(229, 62, 62, 0.05) 0%, rgba(229, 62, 62, 0.02) 100%);
  border: 1px solid rgba(229, 62, 62, 0.3);
  border-radius: 4px;
  color: #e53e3e;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  line-height: 1;
}

.template-remove-btn:hover {
  background: linear-gradient(135deg, rgba(229, 62, 62, 0.1) 0%, rgba(229, 62, 62, 0.05) 100%);
  border-color: rgba(229, 62, 62, 0.5);
  transform: scale(1.1);
}

.template-removed-label {
  font-size: 0.75rem;
  font-weight: 600;
  color: #e53e3e;
  padding: 0.3rem 0.6rem;
  background: rgba(229, 62, 62, 0.1);
  border-radius: 4px;
  text-transform: uppercase;
  letter-spacing: 0.03em;
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

.resource-item.deleting {
  border-color: rgba(229, 62, 62, 0.4);
  background: linear-gradient(135deg, rgba(229, 62, 62, 0.05) 0%, rgba(229, 62, 62, 0.02) 100%);
  animation: fadeOut 0.7s ease-out;
}

.resource-item-retained {
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.03) 0%, rgba(245, 158, 11, 0.01) 100%);
  border-color: rgba(245, 158, 11, 0.2);
}

.resource-item.retained {
  border-color: rgba(245, 158, 11, 0.4);
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.08) 0%, rgba(245, 158, 11, 0.03) 100%);
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

.resource-kind-pvc {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
  border: 1px solid rgba(245, 158, 11, 0.3);
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

.delete-badge {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.25rem 0.6rem;
  background: rgba(229, 62, 62, 0.15);
  color: #e53e3e;
  border: 1px solid rgba(229, 62, 62, 0.3);
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.delete-icon {
  font-size: 0.8rem;
  line-height: 1;
}

.delete-text {
  line-height: 1;
}

.retain-badge {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.25rem 0.6rem;
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
  border: 1px solid rgba(245, 158, 11, 0.3);
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.retain-icon {
  font-size: 0.8rem;
  line-height: 1;
}

.retain-text {
  line-height: 1;
}

.tracking-badge {
  padding: 0.25rem 0.5rem;
  background: rgba(102, 126, 234, 0.15);
  color: #667eea;
  border: 1px solid rgba(102, 126, 234, 0.3);
  border-radius: 4px;
  font-size: 0.65rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.tracking-text {
  line-height: 1;
}

.resource-state {
  font-size: 0.75rem;
  color: var(--vp-c-text-3);
  min-width: 60px;
  text-align: center;
}

.resource-delete-btn {
  padding: 0.3rem 0.5rem;
  background: linear-gradient(135deg, rgba(229, 62, 62, 0.05) 0%, rgba(229, 62, 62, 0.02) 100%);
  border: 1px solid rgba(229, 62, 62, 0.3);
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.3s ease;
  font-size: 0.9rem;
  line-height: 1;
}

.resource-delete-btn:hover {
  background: linear-gradient(135deg, rgba(229, 62, 62, 0.1) 0%, rgba(229, 62, 62, 0.05) 100%);
  border-color: rgba(229, 62, 62, 0.5);
  transform: scale(1.1);
}

/* Orphan Info */
.orphan-info {
  margin-top: 1rem;
  padding: 1rem;
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.05) 0%, rgba(245, 158, 11, 0.02) 100%);
  border: 1px solid rgba(245, 158, 11, 0.3);
  border-radius: 8px;
}

.orphan-title {
  font-size: 0.9rem;
  font-weight: 600;
  color: #f59e0b;
  margin-bottom: 0.75rem;
}

.orphan-markers {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}

.orphan-markers code {
  font-size: 0.75rem;
  background: var(--vp-c-bg);
  padding: 0.4rem 0.6rem;
  border-radius: 4px;
  border: 1px solid var(--vp-c-divider);
  color: var(--vp-c-text-2);
}

.orphan-note {
  font-size: 0.85rem;
  color: var(--vp-c-text-2);
  margin: 0;
  line-height: 1.5;
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

.delete-badge-small {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.2rem 0.5rem;
  background: rgba(229, 62, 62, 0.15);
  color: #e53e3e;
  border: 1px solid rgba(229, 62, 62, 0.3);
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  white-space: nowrap;
}

.retain-badge-small {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.2rem 0.5rem;
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
  border: 1px solid rgba(245, 158, 11, 0.3);
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

.event-item.event-finalizing {
  border-left-color: #f59e0b;
}

.event-item.event-deleting {
  border-left-color: #f59e0b;
}

.event-item.event-deleted {
  border-left-color: #e53e3e;
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

.event-item.event-finalizing .event-indicator {
  background: #f59e0b;
  box-shadow: 0 0 6px rgba(245, 158, 11, 0.4);
  animation: pulse 2s ease-in-out infinite;
}

.event-item.event-deleting .event-indicator {
  background: #f59e0b;
  box-shadow: 0 0 6px rgba(245, 158, 11, 0.4);
}

.event-item.event-deleted .event-indicator {
  background: #e53e3e;
  box-shadow: 0 0 6px rgba(229, 62, 62, 0.4);
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

  .stage-actions {
    width: 100%;
    justify-content: flex-end;
    margin-top: 0.5rem;
  }
}
</style>
