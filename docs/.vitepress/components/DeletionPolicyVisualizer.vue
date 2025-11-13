<template>
  <div class="policy-visualizer deletion-policy">
    <div class="visualizer-head">
      <div class="head-copy">
        <h3>DeletionPolicy Visualizer</h3>
        <p>
          Explore the mermaid branch (<code>LynqForm → DeletionPolicy → Runtime</code>) by simulating LynqNode
          deletion and template removals. Compare how <code>Delete</code> and <code>Retain</code> react.
        </p>
      </div>
      <div class="legend" aria-label="Legend">
        <div v-for="state in legendStates" :key="state.key" class="legend-item">
          <span class="legend-dot" :class="`state-${state.key}`"></span>
          <span>{{ state.label }}</span>
        </div>
      </div>
    </div>

    <div class="policy-options" role="list">
      <button
        v-for="option in policyOptions"
        :key="option.value"
        class="policy-option"
        :class="{ active: selectedPolicy === option.value }"
        @click="selectPolicy(option.value)"
      >
        <div class="option-title">{{ option.label }}</div>
        <p>{{ option.summary }}</p>
      </button>
    </div>

    <div class="policy-description" v-if="currentOption">
      <div class="description-head">
        <h4>{{ currentOption.label }} policy</h4>
        <span class="status-badge" :class="`state-${currentOption.badgeState}`">
          {{ currentOption.badgeLabel }}
        </span>
      </div>
      <p>{{ currentOption.summary }}</p>
      <ul>
        <li v-for="point in currentOption.highlights" :key="point">{{ point }}</li>
      </ul>
    </div>

    <div class="visualization-area">
      <section class="visualization-actions" aria-label="Simulate deletion policy scenarios">
        <article
          v-for="action in actions"
          :key="action.key"
          class="action-card"
          :class="{ active: activeAction === action.key }"
          @click="runAction(action.key)"
        >
          <div class="action-icon">{{ action.icon }}</div>
          <div>
            <div class="action-title">{{ action.label }}</div>
            <p>{{ action.description }}</p>
          </div>
          <span class="action-hint">Click to simulate</span>
        </article>
      </section>

      <section class="resource-stage">
        <div class="resource-box" :class="`state-${resource.state}`">
          <div class="resource-box-header">
            <div>
              <strong>{{ resource.name }}</strong>
              <small>{{ resource.namespace }}</small>
            </div>
            <button
              v-if="resource.state === 'deleted'"
              type="button"
              class="recreate-btn"
              @click.stop="recreateResource"
            >
              ↺ Recreate
            </button>
          </div>

          <div class="resource-message">
            <div class="state-label">{{ stateLabels[resource.state] }}</div>
            <p>{{ resource.message }}</p>
          </div>

          <dl class="resource-meta">
            <div>
              <dt>Tracking</dt>
              <dd>{{ trackingLabel }}</dd>
            </div>
            <div>
              <dt>Owner reference</dt>
              <dd>{{ ownerReference }}</dd>
            </div>
            <div>
              <dt>Labels</dt>
              <dd>{{ resource.labels }}</dd>
            </div>
          </dl>

          <div v-if="resource.state === 'retained'" class="orphan-markers">
            <div class="orphan-title">Orphan markers</div>
            <div class="marker-chips">
              <span v-for="marker in orphanMarkers" :key="marker" class="status-badge state-retained">
                {{ marker }}
              </span>
            </div>
          </div>
        </div>

        <div class="history-log" aria-label="Recent retention events">
          <div class="history-title">Lifecycle log</div>
          <ul>
            <li v-for="entry in actionLog" :key="entry.id">
              <span class="history-state" :class="`state-${entry.state}`"></span>
              <div>
                <p>{{ entry.text }}</p>
                <small>{{ entry.meta }}</small>
              </div>
            </li>
          </ul>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onBeforeUnmount } from 'vue';

const legendStates = [
  { key: 'active', label: 'Active' },
  { key: 'retained', label: 'Retained' },
  { key: 'deleted', label: 'Deleted' },
  { key: 'reconciling', label: 'Reconciling' }
];

const stateLabels = {
  active: 'Active in cluster',
  retained: 'Retained (No GC)',
  deleted: 'Deleted',
  reconciling: 'Recreating'
};

const policyOptions = [
  {
    value: 'Delete',
    label: 'Delete (default)',
    summary: 'Resources are garbage collected via ownerReference when the LynqNode disappears.',
    badgeLabel: 'Garbage collected',
    badgeState: 'deleted',
    highlights: [
      'ownerReference ties the resource lifetime to the LynqNode',
      'Great for stateless resources that must vanish with the node',
      'Simplifies cleanup in dev/test clusters'
    ]
  },
  {
    value: 'Retain',
    label: 'Retain',
    summary: 'Resources stay in the cluster. The operator switches to label-based tracking and marks orphans.',
    badgeLabel: 'Safe retention',
    badgeState: 'retained',
    highlights: [
      'No ownerReference → Kubernetes GC will not touch the resource',
      'Finalizer writes orphan labels + annotations for discovery',
      'Ideal for PVCs, backups, or expensive resources'
    ]
  }
];

const actions = [
  {
    key: 'nodeDelete',
    label: 'Delete LynqNode',
    description: 'Simulate hub/form deletion cascading down to this LynqNode.',
    icon: '🧨'
  },
  {
    key: 'templateRemoval',
    label: 'Removed from Template',
    description: 'Resource removed from the LynqForm spec during reconciliation.',
    icon: '🧩'
  }
];

const selectedPolicy = ref(policyOptions[0].value);
const resource = ref(createResource());
const actionLog = ref([]);
const activeAction = ref(null);
const orphanMarkers = ref([]);
let historyCounter = 0;
const timers = [];

const currentOption = computed(() =>
  policyOptions.find((option) => option.value === selectedPolicy.value)
);

const trackingLabel = computed(() =>
  selectedPolicy.value === 'Delete' ? 'ownerReference → Kubernetes GC' : 'Label tracking (lynq.sh/node*)'
);

const ownerReference = computed(() =>
  selectedPolicy.value === 'Delete' ? 'lynqnode/prod-template' : 'not set'
);

function createResource() {
  return {
    name: 'persistentVolumeClaims/data',
    namespace: 'prod',
    state: 'active',
    message: 'Data volume is attached to its LynqNode.',
    labels: 'lynq.sh/node=acme-prod'
  };
}

const schedule = (callback, delay) => {
  const timer = setTimeout(() => {
    callback();
    const index = timers.indexOf(timer);
    if (index >= 0) {
      timers.splice(index, 1);
    }
  }, delay);
  timers.push(timer);
};

const clearTimers = () => {
  timers.forEach((timer) => clearTimeout(timer));
  timers.length = 0;
};

const logEvent = (text, state) => {
  historyCounter += 1;
  actionLog.value = [
    {
      id: historyCounter,
      text,
      state,
      meta: new Date().toLocaleTimeString()
    },
    ...actionLog.value
  ].slice(0, 4);
};

const resetVisualizer = () => {
  clearTimers();
  resource.value = createResource();
  orphanMarkers.value = [];
  actionLog.value = [
    {
      id: ++historyCounter,
      text: 'Resource synced with template and waiting for lifecycle events.',
      state: 'active',
      meta: 'start'
    }
  ];
  activeAction.value = null;
};

const selectPolicy = (value) => {
  if (selectedPolicy.value !== value) {
    selectedPolicy.value = value;
  }
};

watch(selectedPolicy, () => {
  resetVisualizer();
});

const runAction = (key) => {
  activeAction.value = key;
  if (key === 'nodeDelete') {
    simulateNodeDeletion();
  } else if (key === 'templateRemoval') {
    simulateTemplateRemoval();
  }
};

const simulateNodeDeletion = () => {
  clearTimers();
  if (selectedPolicy.value === 'Delete') {
    resource.value = {
      ...resource.value,
      state: 'reconciling',
      message: 'Finalizer removes the resource via ownerReference.'
    };
    logEvent('Finalizer issued delete for PVC.', 'reconciling');
    schedule(() => {
      resource.value = {
        ...resource.value,
        state: 'deleted',
        message: 'PVC removed with the LynqNode. Use recreate if needed.'
      };
      logEvent('PVC deleted through garbage collector.', 'deleted');
    }, 1000);
  } else {
    resource.value = {
      ...resource.value,
      state: 'retained',
      message: 'LynqNode gone, but PVC stays (labels only).',
      labels: 'lynq.sh/orphaned=true'
    };
    orphanMarkers.value = [
      'lynq.sh/orphaned=true',
      'lynq.sh/orphaned-reason=LynqNodeDeleted'
    ];
    logEvent('PVC marked orphaned after LynqNode deletion.', 'retained');
  }
};

const simulateTemplateRemoval = () => {
  clearTimers();
  if (selectedPolicy.value === 'Delete') {
    resource.value = {
      ...resource.value,
      state: 'reconciling',
      message: 'Resource pruned after being removed from the template.'
    };
    logEvent('Resource scheduled for deletion (removed from templates).', 'reconciling');
    schedule(() => {
      resource.value = {
        ...resource.value,
        state: 'deleted',
        message: 'Resource deleted because it no longer exists in the form.'
      };
      logEvent('Resource deleted after template removal.', 'deleted');
    }, 900);
  } else {
    resource.value = {
      ...resource.value,
      state: 'retained',
      message: 'Template no longer references it. Resource stays with orphan markers.',
      labels: 'lynq.sh/orphaned=true'
    };
    orphanMarkers.value = [
      'lynq.sh/orphaned=true',
      'lynq.sh/orphaned-reason=RemovedFromTemplate'
    ];
    logEvent('Resource retained and marked orphaned (template removal).', 'retained');
  }
};

const recreateResource = () => {
  if (resource.value.state !== 'deleted') {
    return;
  }
  resource.value = {
    ...resource.value,
    state: 'reconciling',
    message: 'Recreating resource with policy defaults...'
  };
  logEvent('Manual recreate requested.', 'reconciling');
  schedule(() => {
    resource.value = {
      ...resource.value,
      state: 'active',
      message: 'Resource recreated and attached to LynqNode.'
    };
    logEvent('Resource recreated successfully.', 'active');
  }, 900);
};

onBeforeUnmount(() => {
  clearTimers();
});

resetVisualizer();
</script>
