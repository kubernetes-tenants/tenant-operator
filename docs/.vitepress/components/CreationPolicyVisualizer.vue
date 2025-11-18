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
      <h3>CreationPolicy Flow Visualizer</h3>
      <p>
        Delete or recreate resources to see how changes cascade through the system.
        Watch how finalizers ensure proper cleanup order. Use "Make Drift" to see how WhenNeeded (Watch) and Once (Static) resources behave differently.
      </p>
    </div>

    <!-- Vertical Flow Diagram -->
    <div class="flow-diagram-vertical">
      <!-- Stage 1: Database (Always exists) -->
      <div class="flow-stage stage-db active">
        <div class="stage-header">
          <div class="stage-icon">🗄️</div>
          <div class="stage-info">
            <div class="stage-title">Database</div>
            <div class="stage-subtitle">External Data Source</div>
          </div>
          <div class="stage-status status-always">Always Active</div>
        </div>
        <div class="stage-content">
          <div class="db-table">
            <div class="db-row">
              <span class="db-col">acme-corp</span>
              <span class="db-col">acme.com</span>
              <span class="db-status active">✓ active</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Connection 1 -->
      <svg class="connection-vertical" :class="{ active: resources.hub.exists }" viewBox="0 0 60 80">
        <defs>
          <linearGradient id="grad-v1" x1="0%" y1="0%" x2="0%" y2="100%">
            <stop offset="0%" stop-color="#42b883" stop-opacity="0" />
            <stop offset="50%" stop-color="#42b883" stop-opacity="0.8" />
            <stop offset="100%" stop-color="#42b883" stop-opacity="0" />
          </linearGradient>
        </defs>
        <path d="M 30 0 L 30 80" stroke="#42b883" stroke-width="2" opacity="0.3" />
        <path d="M 30 0 L 30 80" stroke="url(#grad-v1)" stroke-width="2.5" class="animated-line" />
        <circle r="4" fill="#42b883" class="flow-dot">
          <animateMotion dur="2s" repeatCount="indefinite" path="M 30 0 L 30 80" />
        </circle>
      </svg>

      <!-- Stage 2: LynqHub -->
      <div
        class="flow-stage stage-hub"
        :class="{
          active: resources.hub.exists,
          finalizing: resources.hub.state === 'finalizing',
          deleting: resources.hub.state === 'deleting',
          creating: resources.hub.state === 'creating'
        }"
      >
        <div class="stage-header">
          <div class="stage-icon">📋</div>
          <div class="stage-info">
            <div class="stage-title">LynqHub</div>
            <div class="stage-subtitle">Syncs every 30s</div>
          </div>
          <div class="stage-actions">
            <button
              v-if="resources.hub.exists && resources.hub.state === 'active'"
              class="action-btn delete-btn"
              @click="deleteResource('hub')"
            >
              <span class="btn-icon">🗑️</span>
              <span class="btn-label">Delete</span>
            </button>
            <button
              v-if="!resources.hub.exists"
              class="action-btn create-btn"
              @click="createResource('hub')"
            >
              <span class="btn-icon">✨</span>
              <span class="btn-label">Create</span>
            </button>
          </div>
        </div>
        <div v-if="resources.hub.exists" class="stage-status" :class="`status-${resources.hub.state}`">
          {{ stateLabels[resources.hub.state] }}
        </div>
      </div>

      <!-- Connection 2 -->
      <svg class="connection-vertical" :class="{ active: resources.hub.exists && resources.form.exists }" viewBox="0 0 60 80">
        <defs>
          <linearGradient id="grad-v2" x1="0%" y1="0%" x2="0%" y2="100%">
            <stop offset="0%" stop-color="#667eea" stop-opacity="0" />
            <stop offset="50%" stop-color="#667eea" stop-opacity="0.8" />
            <stop offset="100%" stop-color="#667eea" stop-opacity="0" />
          </linearGradient>
        </defs>
        <path d="M 30 0 L 30 80" stroke="#667eea" stroke-width="2" opacity="0.3" />
        <path d="M 30 0 L 30 80" stroke="url(#grad-v2)" stroke-width="2.5" class="animated-line" />
        <circle r="4" fill="#667eea" class="flow-dot">
          <animateMotion dur="2s" repeatCount="indefinite" path="M 30 0 L 30 80" begin="0.3s" />
        </circle>
      </svg>

      <!-- Stage 3: LynqForm -->
      <div
        class="flow-stage stage-form"
        :class="{
          active: resources.form.exists,
          finalizing: resources.form.state === 'finalizing',
          deleting: resources.form.state === 'deleting',
          creating: resources.form.state === 'creating',
          drifting: resources.form.state === 'drifting'
        }"
      >
        <div class="stage-header">
          <div class="stage-icon">📄</div>
          <div class="stage-info">
            <div class="stage-title">LynqForm</div>
            <div class="stage-subtitle">Template (v{{ resources.form.version || 1 }})</div>
          </div>
          <div class="stage-actions">
            <button
              v-if="resources.form.exists && resources.form.state === 'active' && resources.node.exists && resources.node.state === 'active'"
              class="action-btn drift-btn"
              @click="makeDrift"
              title="Simulate template change and propagate to resources"
            >
              <span class="btn-icon">🔄</span>
              <span class="btn-label">Make Drift</span>
            </button>
            <button
              v-if="resources.form.exists && resources.form.state === 'active'"
              class="action-btn delete-btn"
              @click="deleteResource('form')"
            >
              <span class="btn-icon">🗑️</span>
              <span class="btn-label">Delete</span>
            </button>
            <button
              v-if="!resources.form.exists"
              class="action-btn create-btn"
              @click="createResource('form')"
            >
              <span class="btn-icon">✨</span>
              <span class="btn-label">Create</span>
            </button>
          </div>
        </div>
        <div v-if="resources.form.exists" class="stage-status" :class="`status-${resources.form.state}`">
          {{ stateLabels[resources.form.state] }}
        </div>
      </div>

      <!-- Connection 3 -->
      <svg class="connection-vertical" :class="{ active: resources.hub.exists && resources.form.exists && resources.node.exists }" viewBox="0 0 60 80">
        <defs>
          <linearGradient id="grad-v3" x1="0%" y1="0%" x2="0%" y2="100%">
            <stop offset="0%" stop-color="#41d1ff" stop-opacity="0" />
            <stop offset="50%" stop-color="#41d1ff" stop-opacity="0.8" />
            <stop offset="100%" stop-color="#41d1ff" stop-opacity="0" />
          </linearGradient>
        </defs>
        <path d="M 30 0 L 30 80" stroke="#41d1ff" stroke-width="2" opacity="0.3" />
        <path d="M 30 0 L 30 80" stroke="url(#grad-v3)" stroke-width="2.5" class="animated-line" />
        <circle r="4" fill="#41d1ff" class="flow-dot">
          <animateMotion dur="2s" repeatCount="indefinite" path="M 30 0 L 30 80" begin="0.6s" />
        </circle>
      </svg>

      <!-- Stage 4: LynqNode -->
      <div
        class="flow-stage stage-node"
        :class="{
          active: resources.node.exists,
          finalizing: resources.node.state === 'finalizing',
          deleting: resources.node.state === 'deleting',
          creating: resources.node.state === 'creating',
          syncing: resources.node.state === 'syncing',
          disabled: !resources.hub.exists || !resources.form.exists
        }"
      >
        <div class="stage-header">
          <div class="stage-icon">🏢</div>
          <div class="stage-info">
            <div class="stage-title">LynqNode</div>
            <div class="stage-subtitle">acme-corp (v{{ resources.node.version || 1 }})</div>
          </div>
          <div class="stage-actions">
            <button
              v-if="resources.node.exists && resources.node.state === 'active'"
              class="action-btn delete-btn"
              @click="deleteResource('node')"
            >
              <span class="btn-icon">🗑️</span>
              <span class="btn-label">Delete</span>
            </button>
          </div>
        </div>
        <div v-if="resources.node.exists" class="stage-status" :class="`status-${resources.node.state}`">
          {{ stateLabels[resources.node.state] }}
        </div>
        <div v-if="!resources.hub.exists || !resources.form.exists" class="stage-warning">
          ⚠️ Requires LynqHub & LynqForm
        </div>
      </div>

      <!-- Connection 4 -->
      <svg v-if="anyClusterResourceExists" class="connection-vertical" :class="{ active: anyClusterResourceExists }" viewBox="0 0 60 100">
        <defs>
          <linearGradient id="grad-v4" x1="0%" y1="0%" x2="0%" y2="100%">
            <stop offset="0%" stop-color="#f59e0b" stop-opacity="0" />
            <stop offset="50%" stop-color="#f59e0b" stop-opacity="0.8" />
            <stop offset="100%" stop-color="#f59e0b" stop-opacity="0" />
          </linearGradient>
        </defs>
        <!-- Multiple paths for multiple resources -->
        <path d="M 20 0 L 20 100" stroke="#f59e0b" stroke-width="1.5" opacity="0.2" />
        <path d="M 30 0 L 30 100" stroke="#f59e0b" stroke-width="1.5" opacity="0.2" />
        <path d="M 40 0 L 40 100" stroke="#f59e0b" stroke-width="1.5" opacity="0.2" />

        <path d="M 20 0 L 20 100" stroke="url(#grad-v4)" stroke-width="2" class="animated-line" style="animation-delay: 0s" />
        <path d="M 30 0 L 30 100" stroke="url(#grad-v4)" stroke-width="2" class="animated-line" style="animation-delay: 0.2s" />
        <path d="M 40 0 L 40 100" stroke="url(#grad-v4)" stroke-width="2" class="animated-line" style="animation-delay: 0.4s" />
      </svg>

      <!-- Stage 5: Cluster Resources -->
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
            <!-- Deployment -->
            <div v-if="resources.clusterResources.deployment.exists"
                 class="resource-item"
                 :class="{
                   creating: resources.clusterResources.deployment.state === 'creating',
                   deleting: resources.clusterResources.deployment.state === 'deleting',
                   watching: resources.clusterResources.deployment.creationPolicy === 'WhenNeeded' && resources.clusterResources.deployment.state === 'active'
                 }">
              <div class="resource-info">
                <span class="resource-kind">Deployment</span>
                <span class="resource-name">acme-api</span>
                <span class="resource-version">v{{ resources.clusterResources.deployment.version || 1 }}</span>
              </div>
              <div class="resource-actions">
                <span v-if="resources.clusterResources.deployment.creationPolicy === 'WhenNeeded' && resources.clusterResources.deployment.state === 'active'" class="watch-badge" title="Watching for LynqNode changes">
                  <span class="watch-icon">👁️</span>
                  <span class="watch-text">Watch</span>
                </span>
                <span v-else-if="resources.clusterResources.deployment.creationPolicy === 'Once'" class="static-badge" title="Will not sync with LynqNode changes">
                  <span class="static-icon">🔒</span>
                  <span class="static-text">Static</span>
                </span>
                <span class="resource-state">{{ stateLabels[resources.clusterResources.deployment.state] }}</span>
                <button
                  v-if="resources.clusterResources.deployment.state === 'active'"
                  class="resource-delete-btn"
                  @click="deleteClusterResource('deployment')"
                  title="Delete Deployment"
                >
                  🗑️
                </button>
              </div>
            </div>

            <!-- Service -->
            <div v-if="resources.clusterResources.service.exists"
                 class="resource-item"
                 :class="{
                   creating: resources.clusterResources.service.state === 'creating',
                   deleting: resources.clusterResources.service.state === 'deleting',
                   watching: resources.clusterResources.service.creationPolicy === 'WhenNeeded' && resources.clusterResources.service.state === 'active'
                 }">
              <div class="resource-info">
                <span class="resource-kind">Service</span>
                <span class="resource-name">acme-svc</span>
                <span class="resource-version">v{{ resources.clusterResources.service.version || 1 }}</span>
              </div>
              <div class="resource-actions">
                <span v-if="resources.clusterResources.service.creationPolicy === 'WhenNeeded' && resources.clusterResources.service.state === 'active'" class="watch-badge" title="Watching for LynqNode changes">
                  <span class="watch-icon">👁️</span>
                  <span class="watch-text">Watch</span>
                </span>
                <span v-else-if="resources.clusterResources.service.creationPolicy === 'Once'" class="static-badge" title="Will not sync with LynqNode changes">
                  <span class="static-icon">🔒</span>
                  <span class="static-text">Static</span>
                </span>
                <span class="resource-state">{{ stateLabels[resources.clusterResources.service.state] }}</span>
                <button
                  v-if="resources.clusterResources.service.state === 'active'"
                  class="resource-delete-btn"
                  @click="deleteClusterResource('service')"
                  title="Delete Service"
                >
                  🗑️
                </button>
              </div>
            </div>

            <!-- Ingress -->
            <div v-if="resources.clusterResources.ingress.exists"
                 class="resource-item"
                 :class="{
                   creating: resources.clusterResources.ingress.state === 'creating',
                   deleting: resources.clusterResources.ingress.state === 'deleting',
                   watching: resources.clusterResources.ingress.creationPolicy === 'WhenNeeded' && resources.clusterResources.ingress.state === 'active'
                 }">
              <div class="resource-info">
                <span class="resource-kind">Ingress</span>
                <span class="resource-name">acme-ing</span>
                <span class="resource-version">v{{ resources.clusterResources.ingress.version || 1 }}</span>
              </div>
              <div class="resource-actions">
                <span v-if="resources.clusterResources.ingress.creationPolicy === 'WhenNeeded' && resources.clusterResources.ingress.state === 'active'" class="watch-badge" title="Watching for LynqNode changes">
                  <span class="watch-icon">👁️</span>
                  <span class="watch-text">Watch</span>
                </span>
                <span v-else-if="resources.clusterResources.ingress.creationPolicy === 'Once'" class="static-badge" title="Will not sync with LynqNode changes">
                  <span class="static-icon">🔒</span>
                  <span class="static-text">Static</span>
                </span>
                <span class="resource-state">{{ stateLabels[resources.clusterResources.ingress.state] }}</span>
                <button
                  v-if="resources.clusterResources.ingress.state === 'active'"
                  class="resource-delete-btn"
                  @click="deleteClusterResource('ingress')"
                  title="Delete Ingress"
                >
                  🗑️
                </button>
              </div>
            </div>

            <!-- Job (CreationPolicy=Once, DeletionPolicy=Retain) -->
            <div v-if="resources.clusterResources.job.exists"
                 class="resource-item resource-item-special"
                 :class="{
                   creating: resources.clusterResources.job.state === 'creating',
                   deleting: resources.clusterResources.job.state === 'deleting'
                 }">
              <div class="resource-info">
                <span class="resource-kind resource-kind-job">Job</span>
                <span class="resource-name">acme-init-job</span>
                <span class="resource-version">v{{ resources.clusterResources.job.version || 1 }}</span>
              </div>
              <div class="resource-actions">
                <span class="static-badge" title="Will not sync with LynqNode changes">
                  <span class="static-icon">🔒</span>
                  <span class="static-text">Static</span>
                </span>
                <span class="retain-badge">Retain</span>
                <span class="resource-state">{{ stateLabels[resources.clusterResources.job.state] }}</span>
                <button
                  v-if="resources.clusterResources.job.state === 'active'"
                  class="resource-delete-btn"
                  @click="deleteClusterResource('job')"
                  title="Delete Job (will be retained after LynqNode deletion)"
                >
                  🗑️
                </button>
              </div>
            </div>
          </div>

          <!-- Policy explanation -->
          <div v-if="anyClusterResourceExists" class="policy-explanation">
            <div class="explanation-item">
              <span class="watch-badge-small">
                <span class="watch-icon">👁️</span>
                <span>Watch</span>
              </span>
              <span class="explanation-text">
                Continuously syncs with LynqNode changes (CreationPolicy=WhenNeeded)
              </span>
            </div>
            <div class="explanation-item">
              <span class="static-badge-small">
                <span class="static-icon">🔒</span>
                <span>Static</span>
              </span>
              <span class="explanation-text">
                No sync after creation (CreationPolicy=Once)
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
import { ref, watch, onMounted, onBeforeUnmount, computed } from 'vue';

const stateLabels = {
  creating: 'Creating...',
  active: 'Active',
  finalizing: 'Finalizing...',
  deleting: 'Deleting...',
  deleted: 'Deleted',
  drifting: 'Drifting...',
  syncing: 'Syncing...'
};

const eventLog = ref([]);
let eventCounter = 0;
let timeoutIds = [];

const resources = ref({
  hub: { exists: true, state: 'active' },
  form: { exists: true, state: 'active', version: 1 },
  node: {
    exists: true,
    state: 'active',
    version: 1,
    creationPolicy: 'WhenNeeded' // Copied from LynqForm template
  },
  clusterResources: {
    deployment: {
      exists: true,
      state: 'active',
      createdOnce: false,
      version: 1,
      creationPolicy: 'WhenNeeded',
      deletionPolicy: 'Delete'
    },
    service: {
      exists: true,
      state: 'active',
      createdOnce: false,
      version: 1,
      creationPolicy: 'WhenNeeded',
      deletionPolicy: 'Delete'
    },
    ingress: {
      exists: true,
      state: 'active',
      createdOnce: false,
      version: 1,
      creationPolicy: 'WhenNeeded',
      deletionPolicy: 'Delete'
    },
    job: {
      exists: true,
      state: 'active',
      createdOnce: false,
      version: 1,
      creationPolicy: 'Once',
      deletionPolicy: 'Retain'
    }
  }
});

// Helper to check if any cluster resource exists
const anyClusterResourceExists = computed(() => {
  return Object.values(resources.value.clusterResources).some(r => r.exists);
});

// Helper to check if all cluster resources exist
const allClusterResourcesExist = computed(() => {
  return Object.values(resources.value.clusterResources).every(r => r.exists);
});

// Watch for Hub and Form both being active to auto-create Node
watch(
  () => [resources.value.hub.exists, resources.value.hub.state, resources.value.form.exists, resources.value.form.state],
  ([hubExists, hubState, formExists, formState]) => {
    // Auto-create LynqNode when both Hub and Form are active
    if (hubExists && hubState === 'active' && formExists && formState === 'active') {
      if (!resources.value.node.exists) {
        const timeoutId = setTimeout(() => {
          logEvent('🔄 LynqHub and LynqForm both active, auto-creating LynqNode...', 'reconciling');
          const createTimeout = setTimeout(() => {
            createResource('node');
          }, 600);
          timeoutIds.push(createTimeout);
        }, 800);
        timeoutIds.push(timeoutId);
      }
    }
  },
  { deep: true }
);

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

// Delete resource with finalizer simulation
const deleteResource = (resourceKey) => {
  const resource = resources.value[resourceKey];
  if (!resource.exists || resource.state !== 'active') return;

  logEvent(`🗑️ Delete request for ${getResourceLabel(resourceKey)}`, 'info');

  // Step 1: Enter finalizing state
  resource.state = 'finalizing';
  logEvent(`⏳ Finalizer triggered for ${getResourceLabel(resourceKey)}`, 'finalizing');

  const timeoutId = setTimeout(() => {
    // Step 2: Delete child resources first
    cleanupChildResources(resourceKey);
  }, 600);
  timeoutIds.push(timeoutId);
};

// Poll until child resource is completely deleted
const waitForChildDeletion = (childKey, callback) => {
  const checkInterval = 200; // Check every 200ms
  const maxWaitTime = 10000; // Max 10 seconds
  let elapsedTime = 0;

  const checkDeletion = () => {
    elapsedTime += checkInterval;

    if (!resources.value[childKey].exists) {
      // Child is completely deleted
      logEvent(`✓ ${getResourceLabel(childKey)} completely removed`, 'deleted');
      callback();
    } else if (elapsedTime >= maxWaitTime) {
      // Timeout - force proceed
      logEvent(`⚠️ Timeout waiting for ${getResourceLabel(childKey)} deletion`, 'warning');
      callback();
    } else {
      // Keep checking
      const timeoutId = setTimeout(checkDeletion, checkInterval);
      timeoutIds.push(timeoutId);
    }
  };

  const timeoutId = setTimeout(checkDeletion, checkInterval);
  timeoutIds.push(timeoutId);
};

// Delete individual cluster resource
const deleteClusterResource = (resourceType) => {
  const resource = resources.value.clusterResources[resourceType];
  if (!resource.exists || resource.state !== 'active') return;

  logEvent(`🗑️ Deleting ${resourceType}...`, 'deleting');
  resource.state = 'deleting';

  const timeoutId = setTimeout(() => {
    resource.exists = false;
    resource.state = 'deleted';
    logEvent(`✓ ${resourceType} deleted`, 'deleted');

    // Handle reconciliation based on policy
    handleClusterResourceReconciliation(resourceType);
  }, 1000);
  timeoutIds.push(timeoutId);
};

// Handle reconciliation for individual cluster resources
const handleClusterResourceReconciliation = (resourceType) => {
  // Only reconcile if LynqNode is active
  if (!resources.value.node.exists || resources.value.node.state !== 'active') return;

  const resource = resources.value.clusterResources[resourceType];
  const resourceCreationPolicy = resource.creationPolicy;

  // CreationPolicy=Once: 리소스가 존재하지 않으면 생성 (삭제되었으므로 다시 생성)
  // CreationPolicy=WhenNeeded: 항상 재생성 (drift 감지)

  if (resourceCreationPolicy === 'Once') {
    // Once: 삭제되었으니 다시 생성해야 함
    const timeoutId = setTimeout(() => {
      logEvent(`🔄 ${resourceType} CreationPolicy=Once: Recreating (deleted, needs to exist)`, 'reconciling');
      const createTimeout = setTimeout(() => {
        createClusterResource(resourceType);
      }, 600);
      timeoutIds.push(createTimeout);
    }, 800);
    timeoutIds.push(timeoutId);
  } else {
    // WhenNeeded: 항상 재생성
    const timeoutId = setTimeout(() => {
      logEvent(`🔄 ${resourceType} CreationPolicy=WhenNeeded: Recreating (drift detected)`, 'reconciling');
      const createTimeout = setTimeout(() => {
        createClusterResource(resourceType);
      }, 600);
      timeoutIds.push(createTimeout);
    }, 800);
    timeoutIds.push(timeoutId);
  }
};

// Create individual cluster resource
const createClusterResource = (resourceType) => {
  const resource = resources.value.clusterResources[resourceType];
  if (resource.exists) {
    // 리소스가 이미 존재함
    if (resource.creationPolicy === 'Once') {
      logEvent(`⏭️ ${resourceType} already exists, skipping update (CreationPolicy=Once)`, 'info');
    } else {
      logEvent(`🔄 ${resourceType} already exists, would update (CreationPolicy=WhenNeeded)`, 'info');
    }
    return;
  }

  const resourceCreationPolicy = resource.creationPolicy;
  const deletionPolicy = resource.deletionPolicy;

  // Use current LynqNode version when creating
  const currentNodeVersion = resources.value.node.version || 1;

  logEvent(`✨ Creating ${resourceType} (CreationPolicy=${resourceCreationPolicy}, DeletionPolicy=${deletionPolicy}) with version v${currentNodeVersion}...`, 'creating');
  resource.state = 'creating';
  resource.exists = true;

  const timeoutId = setTimeout(() => {
    resource.state = 'active';
    // Set resource version to current LynqNode version
    resource.version = currentNodeVersion;

    // Mark as created once (for tracking purposes)
    resource.createdOnce = true;

    if (resourceCreationPolicy === 'Once') {
      logEvent(`✓ ${resourceType} created with v${currentNodeVersion} (CreationPolicy=Once, will skip updates while existing)`, 'success');
    } else {
      logEvent(`✓ ${resourceType} created with v${currentNodeVersion} (CreationPolicy=WhenNeeded, will sync on changes)`, 'success');
    }
  }, 1000);
  timeoutIds.push(timeoutId);
};

// Create all cluster resources (apply LynqNode's CreationPolicy except for resources with their own policy)
const createAllClusterResources = () => {
  const nodePolicy = resources.value.node.creationPolicy;

  Object.keys(resources.value.clusterResources).forEach((resourceType, index) => {
    const resource = resources.value.clusterResources[resourceType];

    // Apply LynqNode's policy to resources, but respect individual resource policies
    // Job always has CreationPolicy=Once regardless of LynqNode
    if (resourceType !== 'job' && resource.creationPolicy !== nodePolicy) {
      resource.creationPolicy = nodePolicy;
      logEvent(`📝 ${resourceType} CreationPolicy updated to ${nodePolicy} from LynqNode`, 'info');
    }

    const timeoutId = setTimeout(() => {
      createClusterResource(resourceType);
    }, 300 * index);
    timeoutIds.push(timeoutId);
  });
};

// Delete all cluster resources (respecting DeletionPolicy)
const deleteAllClusterResources = (callback) => {
  const resourceTypes = Object.keys(resources.value.clusterResources);

  // Filter resources that exist and should be deleted (DeletionPolicy=Delete)
  const resourcesToDelete = resourceTypes.filter(rt => {
    const res = resources.value.clusterResources[rt];
    return res.exists && res.deletionPolicy === 'Delete';
  });

  // Resources with DeletionPolicy=Retain
  const resourcesToRetain = resourceTypes.filter(rt => {
    const res = resources.value.clusterResources[rt];
    return res.exists && res.deletionPolicy === 'Retain';
  });

  // Log retained resources
  resourcesToRetain.forEach(rt => {
    logEvent(`🔒 ${rt} has DeletionPolicy=Retain, will be kept in cluster`, 'info');
  });

  // If no resources to delete, call callback immediately
  if (resourcesToDelete.length === 0) {
    logEvent('✓ No cluster resources to delete (all retained or already deleted)', 'info');
    if (callback) callback();
    return;
  }

  let deletedCount = 0;
  const totalToDelete = resourcesToDelete.length;

  resourcesToDelete.forEach((resourceType, index) => {
    const timeoutId = setTimeout(() => {
      const resource = resources.value.clusterResources[resourceType];
      resource.state = 'deleting';
      logEvent(`🗑️ Deleting ${resourceType} (DeletionPolicy=Delete)...`, 'deleting');

      const deleteTimeout = setTimeout(() => {
        resource.exists = false;
        resource.state = 'deleted';
        deletedCount++;

        logEvent(`✓ ${resourceType} deleted (${deletedCount}/${totalToDelete})`, 'deleted');

        // Check if all resources have been deleted
        if (deletedCount === totalToDelete) {
          logEvent(`✓ All ${totalToDelete} cluster resources deleted (${resourcesToRetain.length} retained)`, 'success');
          if (callback) callback();
        }
      }, 700);
      timeoutIds.push(deleteTimeout);
    }, 300 * index);
    timeoutIds.push(timeoutId);
  });
};

// Cleanup child resources (called by finalizer)
const cleanupChildResources = (resourceKey) => {
  const childMap = {
    hub: ['node'],
    form: ['node'],
    node: ['allClusterResources'],
    clusterResources: []
  };

  const children = childMap[resourceKey] || [];

  if (children.length === 0) {
    // No children, proceed to delete self
    completeDeletion(resourceKey);
    return;
  }

  // Special handling for cluster resources
  if (children.includes('allClusterResources')) {
    logEvent(`🧹 Cleaning up all cluster resources of ${getResourceLabel(resourceKey)}...`, 'cleaning');

    deleteAllClusterResources(() => {
      logEvent(`✓ All cluster resources cleaned up`, 'success');
      const timeoutId = setTimeout(() => {
        completeDeletion(resourceKey);
      }, 400);
      timeoutIds.push(timeoutId);
    });
    return;
  }

  logEvent(`🧹 Cleaning up ${children.length} child resource(s) of ${getResourceLabel(resourceKey)}...`, 'cleaning');

  // Delete children sequentially and wait for each to complete
  const deleteChildrenSequentially = (index) => {
    if (index >= children.length) {
      // All children processed, now complete parent deletion
      logEvent(`✓ All child resources of ${getResourceLabel(resourceKey)} cleaned up`, 'success');
      const timeoutId = setTimeout(() => {
        completeDeletion(resourceKey);
      }, 400);
      timeoutIds.push(timeoutId);
      return;
    }

    const childKey = children[index];
    if (resources.value[childKey].exists) {
      // Trigger child deletion
      const timeoutId = setTimeout(() => {
        deleteResource(childKey);

        // Wait for this child to be completely deleted before moving to next
        waitForChildDeletion(childKey, () => {
          deleteChildrenSequentially(index + 1);
        });
      }, 600);
      timeoutIds.push(timeoutId);
    } else {
      // Child already deleted, move to next
      deleteChildrenSequentially(index + 1);
    }
  };

  deleteChildrenSequentially(0);
};

// Complete the deletion after finalizer cleanup
const completeDeletion = (resourceKey) => {
  const resource = resources.value[resourceKey];

  logEvent(`🗑️ Removing finalizer from ${getResourceLabel(resourceKey)}`, 'deleting');
  resource.state = 'deleting';

  const timeoutId = setTimeout(() => {
    resource.exists = false;
    resource.state = 'deleted';
    logEvent(`✓ ${getResourceLabel(resourceKey)} deleted`, 'deleted');

    // Check for reconciliation after deletion
    handleReconciliation(resourceKey);
  }, 700);
  timeoutIds.push(timeoutId);
};

// Create resource with proper sequencing
const createResource = (resourceKey) => {
  const resource = resources.value[resourceKey];
  if (resource.exists) return;

  // Check dependencies
  if (!canCreate(resourceKey)) {
    logEvent(`⚠️ Cannot create ${getResourceLabel(resourceKey)}: missing dependencies`, 'warning');
    return;
  }

  // Set creating state
  resource.state = 'creating';
  resource.exists = true;

  // LynqNode: Copy CreationPolicy from LynqForm template (default: WhenNeeded)
  if (resourceKey === 'node') {
    resource.creationPolicy = 'WhenNeeded';
    logEvent(`✨ Creating ${getResourceLabel(resourceKey)} (CreationPolicy=WhenNeeded from LynqForm)...`, 'creating');
  } else {
    logEvent(`✨ Creating ${getResourceLabel(resourceKey)}...`, 'creating');
  }

  const timeoutId = setTimeout(() => {
    resource.state = 'active';
    logEvent(`✓ ${getResourceLabel(resourceKey)} created and active`, 'success');

    // Auto-create children based on LynqNode's policy
    if (resourceKey === 'node') {
      autoCreateChildren(resourceKey);
    }
  }, 1000);
  timeoutIds.push(timeoutId);
};

// Auto-create child resources
const autoCreateChildren = (resourceKey) => {
  // Only LynqNode auto-creates cluster resources
  if (resourceKey === 'node') {
    const nodePolicy = resources.value.node.creationPolicy;
    const timeoutId = setTimeout(() => {
      logEvent(`🔄 Auto-creating cluster resources (LynqNode CreationPolicy=${nodePolicy})`, 'info');
      const createTimeout = setTimeout(() => {
        createAllClusterResources();
      }, 600);
      timeoutIds.push(createTimeout);
    }, 1200);
    timeoutIds.push(timeoutId);
  }
};

// Handle reconciliation after deletion
const handleReconciliation = (resourceKey) => {
  // LynqNode: Auto-recreated by LynqHub controller if Hub and Form are both active
  if (resourceKey === 'node') {
    if (resources.value.hub.exists && resources.value.hub.state === 'active' &&
        resources.value.form.exists && resources.value.form.state === 'active') {
      const timeoutId = setTimeout(() => {
        logEvent('🔄 LynqHub controller detected missing LynqNode, reconciling...', 'reconciling');

        const reconcileTimeout = setTimeout(() => {
          createResource('node');
        }, 800);
        timeoutIds.push(reconcileTimeout);
      }, 1000);
      timeoutIds.push(timeoutId);
    }
  }
};

// Check if resource can be created (dependencies met)
const canCreate = (resourceKey) => {
  const dependencies = {
    hub: [],
    form: [],
    node: ['hub', 'form'],
    clusterResources: ['node']
  };

  const required = dependencies[resourceKey] || [];
  return required.every(dep => resources.value[dep].exists && resources.value[dep].state === 'active');
};

// Get human-readable resource label
const getResourceLabel = (resourceKey) => {
  const labels = {
    hub: 'LynqHub',
    form: 'LynqForm',
    node: 'LynqNode',
    clusterResources: 'Cluster Resources'
  };
  return labels[resourceKey] || resourceKey;
};

// Make drift: Simulate template change and propagate to resources
const makeDrift = () => {
  if (!resources.value.form.exists || resources.value.form.state !== 'active') return;
  if (!resources.value.node.exists || resources.value.node.state !== 'active') return;

  // Step 1: LynqForm enters drifting state
  resources.value.form.state = 'drifting';
  resources.value.form.version = (resources.value.form.version || 1) + 1;
  logEvent(`📝 LynqForm updated to v${resources.value.form.version} (drift detected)`, 'info');

  // Step 2: After 800ms, LynqNode starts syncing
  const step2Timeout = setTimeout(() => {
    resources.value.node.state = 'syncing';
    logEvent(`🔄 LynqNode syncing from LynqForm v${resources.value.form.version}...`, 'reconciling');

    // Step 3: After 1000ms, LynqNode completes sync
    const step3Timeout = setTimeout(() => {
      resources.value.node.version = resources.value.form.version;
      resources.value.node.state = 'active';
      resources.value.form.state = 'active';
      logEvent(`✓ LynqNode synced to v${resources.value.node.version}`, 'success');

      // Step 4: After 600ms, propagate to ClusterResources (only WhenNeeded)
      const step4Timeout = setTimeout(() => {
        propagateDriftToResources();
      }, 600);
      timeoutIds.push(step4Timeout);
    }, 1000);
    timeoutIds.push(step3Timeout);
  }, 800);
  timeoutIds.push(step2Timeout);
};

// Propagate drift to cluster resources (only WhenNeeded resources)
const propagateDriftToResources = () => {
  Object.keys(resources.value.clusterResources).forEach((resourceType, index) => {
    const resource = resources.value.clusterResources[resourceType];

    if (resource.exists && resource.creationPolicy === 'WhenNeeded') {
      const timeout = setTimeout(() => {
        logEvent(`🔄 Updating ${resourceType} (Watch: detected drift from LynqNode v${resources.value.node.version})`, 'reconciling');

        const updateTimeout = setTimeout(() => {
          resource.version = resources.value.node.version;
          logEvent(`✓ ${resourceType} updated to v${resources.value.node.version}`, 'success');
        }, 800);
        timeoutIds.push(updateTimeout);
      }, 400 * index);
      timeoutIds.push(timeout);
    } else if (resource.exists && resource.creationPolicy === 'Once') {
      const timeout = setTimeout(() => {
        logEvent(`⏭️ ${resourceType} skipping update (Static: CreationPolicy=Once)`, 'info');
      }, 400 * index);
      timeoutIds.push(timeout);
    }
  });
};

// Reset all resources to initial state
const resetAll = () => {
  // Clear all timeouts
  timeoutIds.forEach(id => clearTimeout(id));
  timeoutIds = [];

  // Reset all resources
  resources.value = {
    hub: { exists: true, state: 'active' },
    form: { exists: true, state: 'active', version: 1 },
    node: {
      exists: true,
      state: 'active',
      version: 1,
      creationPolicy: 'WhenNeeded'
    },
    clusterResources: {
      deployment: {
        exists: true,
        state: 'active',
        createdOnce: false,
        version: 1,
        creationPolicy: 'WhenNeeded',
        deletionPolicy: 'Delete'
      },
      service: {
        exists: true,
        state: 'active',
        createdOnce: false,
        version: 1,
        creationPolicy: 'WhenNeeded',
        deletionPolicy: 'Delete'
      },
      ingress: {
        exists: true,
        state: 'active',
        createdOnce: false,
        version: 1,
        creationPolicy: 'WhenNeeded',
        deletionPolicy: 'Delete'
      },
      job: {
        exists: true,
        state: 'active',
        createdOnce: false,
        version: 1,
        creationPolicy: 'Once',
        deletionPolicy: 'Retain'
      }
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

.flow-stage.disabled {
  opacity: 0.5;
  filter: grayscale(50%);
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

.flow-stage.creating {
  border-color: rgba(66, 184, 131, 0.4);
  background: linear-gradient(135deg, rgba(66, 184, 131, 0.05) 0%, rgba(66, 184, 131, 0.02) 100%);
  animation: fadeIn 1s ease-out;
}

.flow-stage.drifting {
  border-color: rgba(102, 126, 234, 0.5);
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.08) 0%, rgba(102, 126, 234, 0.03) 100%);
  animation: driftingPulse 1.5s ease-in-out infinite;
}

.flow-stage.syncing {
  border-color: rgba(65, 209, 255, 0.5);
  background: linear-gradient(135deg, rgba(65, 209, 255, 0.08) 0%, rgba(65, 209, 255, 0.03) 100%);
  animation: syncingPulse 1.5s ease-in-out infinite;
}

@keyframes driftingPulse {
  0%, 100% {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05), 0 0 0 0 rgba(102, 126, 234, 0.4);
  }
  50% {
    box-shadow: 0 4px 20px rgba(102, 126, 234, 0.15), 0 0 0 8px rgba(102, 126, 234, 0);
  }
}

@keyframes syncingPulse {
  0%, 100% {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05), 0 0 0 0 rgba(65, 209, 255, 0.4);
  }
  50% {
    box-shadow: 0 4px 20px rgba(65, 209, 255, 0.15), 0 0 0 8px rgba(65, 209, 255, 0);
  }
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

.create-btn {
  background: linear-gradient(135deg, rgba(66, 184, 131, 0.05) 0%, rgba(66, 184, 131, 0.02) 100%);
  border-color: rgba(66, 184, 131, 0.3);
  color: #42b883;
}

.create-btn:hover {
  background: linear-gradient(135deg, rgba(66, 184, 131, 0.1) 0%, rgba(66, 184, 131, 0.05) 100%);
  border-color: rgba(66, 184, 131, 0.5);
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(66, 184, 131, 0.2);
}

.drift-btn {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.05) 0%, rgba(102, 126, 234, 0.02) 100%);
  border-color: rgba(102, 126, 234, 0.3);
  color: #667eea;
}

.drift-btn:hover {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(102, 126, 234, 0.05) 100%);
  border-color: rgba(102, 126, 234, 0.5);
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.2);
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

.stage-status.status-always {
  background: rgba(66, 184, 131, 0.15);
  color: #42b883;
  border: 1px solid rgba(66, 184, 131, 0.3);
}

.stage-status.status-creating {
  background: rgba(102, 126, 234, 0.15);
  color: #667eea;
  border: 1px solid rgba(102, 126, 234, 0.3);
  animation: pulse 2s ease-in-out infinite;
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

.stage-status.status-deleted {
  background: rgba(113, 128, 150, 0.15);
  color: #718096;
  border: 1px solid rgba(113, 128, 150, 0.3);
}

.stage-status.status-drifting {
  background: rgba(102, 126, 234, 0.15);
  color: #667eea;
  border: 1px solid rgba(102, 126, 234, 0.3);
  animation: pulse 2s ease-in-out infinite;
}

.stage-status.status-syncing {
  background: rgba(65, 209, 255, 0.15);
  color: #41d1ff;
  border: 1px solid rgba(65, 209, 255, 0.3);
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

/* Database Table */
.db-table {
  width: 100%;
}

.db-row {
  display: grid;
  grid-template-columns: 1fr 1fr auto;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  background: var(--vp-c-bg);
  border-radius: 6px;
  border: 1px solid var(--vp-c-divider);
}

.db-col {
  color: var(--vp-c-text-2);
  font-family: monospace;
  font-size: 0.85rem;
}

.db-status {
  font-weight: 600;
  font-size: 0.85rem;
}

.db-status.active {
  color: #42b883;
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

.resource-item.watching {
  border-color: rgba(66, 184, 131, 0.4);
  background: linear-gradient(135deg, rgba(66, 184, 131, 0.03) 0%, rgba(66, 184, 131, 0.01) 100%);
  animation: watchingPulse 3s ease-in-out infinite;
}

@keyframes watchingPulse {
  0%, 100% {
    box-shadow: 0 0 0 0 rgba(66, 184, 131, 0.4);
  }
  50% {
    box-shadow: 0 0 0 4px rgba(66, 184, 131, 0);
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

.resource-version {
  font-size: 0.7rem;
  font-weight: 600;
  color: var(--vp-c-text-3);
  background: var(--vp-c-bg-soft);
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  border: 1px solid var(--vp-c-divider);
  font-family: monospace;
  transition: all 0.3s ease;
}

.resource-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.once-badge {
  padding: 0.25rem 0.5rem;
  background: rgba(113, 128, 150, 0.15);
  color: #718096;
  border: 1px solid rgba(113, 128, 150, 0.3);
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.retain-badge {
  padding: 0.25rem 0.5rem;
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
  border: 1px solid rgba(245, 158, 11, 0.3);
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.watch-badge {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.25rem 0.6rem;
  background: rgba(66, 184, 131, 0.15);
  color: #42b883;
  border: 1px solid rgba(66, 184, 131, 0.3);
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  animation: watchBadgePulse 2s ease-in-out infinite;
}

.watch-icon {
  font-size: 0.85rem;
  line-height: 1;
  animation: eyeBlink 3s ease-in-out infinite;
}

@keyframes watchBadgePulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.7;
  }
}

@keyframes eyeBlink {
  0%, 90%, 100% {
    opacity: 1;
  }
  95% {
    opacity: 0.3;
  }
}

.watch-text {
  line-height: 1;
}

.static-badge {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.25rem 0.6rem;
  background: rgba(113, 128, 150, 0.15);
  color: #718096;
  border: 1px solid rgba(113, 128, 150, 0.3);
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.static-icon {
  font-size: 0.8rem;
  line-height: 1;
}

.static-text {
  line-height: 1;
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

.watch-badge-small {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.2rem 0.5rem;
  background: rgba(66, 184, 131, 0.15);
  color: #42b883;
  border: 1px solid rgba(66, 184, 131, 0.3);
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  white-space: nowrap;
}

.static-badge-small {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.2rem 0.5rem;
  background: rgba(113, 128, 150, 0.15);
  color: #718096;
  border: 1px solid rgba(113, 128, 150, 0.3);
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

.resource-item-special {
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.03) 0%, rgba(245, 158, 11, 0.01) 100%);
  border-color: rgba(245, 158, 11, 0.2);
}

.resource-kind-job {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
  border: 1px solid rgba(245, 158, 11, 0.3);
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

/* Connections */
.connection-vertical {
  width: 60px;
  height: 80px;
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

.event-item.event-creating {
  border-left-color: #41d1ff;
}

.event-item.event-finalizing {
  border-left-color: #f59e0b;
}

.event-item.event-cleaning {
  border-left-color: #ed8936;
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

.event-item.event-warning {
  border-left-color: #ed8936;
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

.event-item.event-creating .event-indicator {
  background: #41d1ff;
  box-shadow: 0 0 6px rgba(65, 209, 255, 0.4);
}

.event-item.event-finalizing .event-indicator {
  background: #f59e0b;
  box-shadow: 0 0 6px rgba(245, 158, 11, 0.4);
  animation: pulse 2s ease-in-out infinite;
}

.event-item.event-cleaning .event-indicator {
  background: #ed8936;
  box-shadow: 0 0 6px rgba(237, 137, 54, 0.4);
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

.event-item.event-warning .event-indicator {
  background: #ed8936;
  box-shadow: 0 0 6px rgba(237, 137, 54, 0.4);
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
