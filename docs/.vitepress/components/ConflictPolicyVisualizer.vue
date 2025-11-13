<template>
  <div class="policy-visualizer conflict-policy">
    <div class="visualizer-head">
      <div class="head-copy">
        <h3>ConflictPolicy Visualizer</h3>
        <p>
          Walk the mermaid route (<code>CreationPolicy → ConflictPolicy → Patch → Runtime</code>) to see what happens
          when another controller owns the same fields. Trigger conflicts and compare <code>Stuck</code> vs
          <code>Force</code>.
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
      <section class="visualization-actions" aria-label="Simulate conflict policy scenarios">
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
          </div>

          <div class="resource-message">
            <div class="state-label">{{ stateLabels[resource.state] }}</div>
            <p>{{ resource.message }}</p>
          </div>

          <div class="resource-tags owners">
            <span
              v-for="owner in resource.owners"
              :key="owner.label"
              class="status-badge"
              :class="`state-${owner.state}`"
            >
              {{ owner.label }}
            </span>
          </div>

          <dl class="resource-meta">
            <div>
              <dt>Conflict owner</dt>
              <dd>{{ resource.conflictOwner ?? 'None' }}</dd>
            </div>
            <div>
              <dt>Node condition</dt>
              <dd>{{ nodeCondition }}</dd>
            </div>
            <div>
              <dt>Events</dt>
              <dd>{{ eventHint }}</dd>
            </div>
          </dl>
        </div>

        <div class="history-log" aria-label="Conflict log">
          <div class="history-title">Conflict log</div>
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
  { key: 'conflict', label: 'Conflict' },
  { key: 'reconciling', label: 'Force applying' }
];

const stateLabels = {
  active: 'Active in cluster',
  conflict: 'Conflict detected',
  reconciling: 'Force applying'
};

const policyOptions = [
  {
    value: 'Stuck',
    label: 'Stuck (default)',
    summary: 'Stops reconciliation when another field manager owns the resource.',
    badgeLabel: 'Safe but halted',
    badgeState: 'conflict',
    highlights: [
      'Fires ResourceConflict events and marks the node degraded',
      'Requires manual intervention or renaming',
      'Protects other controllers from being overwritten'
    ]
  },
  {
    value: 'Force',
    label: 'Force',
    summary: 'Uses Server-Side Apply with force=true to take ownership.',
    badgeLabel: 'Aggressive takeover',
    badgeState: 'reconciling',
    highlights: [
      'SSA force apply evicts other managers from overlapping fields',
      'Keeps reconciliation healthy but may overwrite live changes',
      'Ideal when Lynq is the sole source of truth'
    ]
  }
];

const actions = [
  {
    key: 'conflict',
    label: 'Simulate Ownership Conflict',
    description: 'Another controller writes to the same field set.',
    icon: '⚡️'
  },
  {
    key: 'resolve',
    label: 'Resolve / Clear Conflict',
    description: 'Delete conflicting object or allow Lynq to take over again.',
    icon: '🧹'
  }
];

const conflictingControllers = ['argo-rollouts', 'keda', 'cert-manager', 'traffic-operator'];

const selectedPolicy = ref(policyOptions[0].value);
const resource = ref(createResource());
const actionLog = ref([]);
const activeAction = ref(null);
let historyCounter = 0;
const timers = [];

const currentOption = computed(() =>
  policyOptions.find((option) => option.value === selectedPolicy.value)
);

const nodeCondition = computed(() =>
  resource.value.state === 'conflict' ? 'Degraded (ResourceConflict)' : 'Healthy'
);

const eventHint = computed(() =>
  resource.value.state === 'conflict'
    ? 'ResourceConflict events emitted'
    : selectedPolicy.value === 'Force'
      ? 'SSA force apply events'
      : 'Normal SSA events'
);

function createResource() {
  return {
    name: 'services/app-svc',
    namespace: 'prod-core',
    state: 'active',
    message: 'Service managed by Lynq SSA.',
    owners: [{ label: 'lynq (manager=lynq)', state: 'active' }],
    conflictOwner: null
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
  actionLog.value = [
    {
      id: ++historyCounter,
      text: 'Awaiting conflicts. Lynq is the current field manager.',
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
  if (key === 'conflict') {
    simulateConflict();
  } else if (key === 'resolve') {
    resolveConflict();
  }
};

const simulateConflict = () => {
  clearTimers();
  const external = conflictingControllers[Math.floor(Math.random() * conflictingControllers.length)];
  resource.value = {
    ...resource.value,
    conflictOwner: external
  };

  if (selectedPolicy.value === 'Stuck') {
    resource.value = {
      ...resource.value,
      state: 'conflict',
      message: `Conflict with ${external}. Lynq halts and emits ResourceConflict events.`,
      owners: [
        { label: 'lynq (paused)', state: 'conflict' },
        { label: `${external} (active)`, state: 'conflict' }
      ]
    };
    logEvent(`Conflict detected with ${external}.`, 'conflict');
  } else {
    resource.value = {
      ...resource.value,
      state: 'reconciling',
      message: `Force applying against ${external}…`,
      owners: [
        { label: 'lynq (forcing)', state: 'reconciling' },
        { label: `${external} (will be removed)`, state: 'conflict' }
      ]
    };
    logEvent(`Force apply issued to beat ${external}.`, 'reconciling');
    schedule(() => {
      resource.value = {
        ...resource.value,
        state: 'active',
        message: `SSA force succeeded. ${external} removed from managedFields.`,
        owners: [{ label: 'lynq (owner)', state: 'active' }],
        conflictOwner: null
      };
      logEvent(`Ownership transferred back to Lynq.`, 'active');
    }, 1100);
  }
};

const resolveConflict = () => {
  clearTimers();
  if (resource.value.state === 'conflict') {
    resource.value = {
      ...resource.value,
      state: 'reconciling',
      message: 'Conflict cleared manually. Re-applying template…',
      owners: resource.value.owners.map((owner) =>
        owner.label.startsWith('lynq') ? { ...owner, state: 'reconciling' } : owner
      )
    };
    logEvent('Operator manually resolved the conflict.', 'reconciling');
    schedule(() => {
      resource.value = {
        ...resource.value,
        state: 'active',
        message: 'SSA completed. Node condition is healthy again.',
        owners: [{ label: 'lynq (owner)', state: 'active' }],
        conflictOwner: null
      };
      logEvent('Service reconciled successfully.', 'active');
    }, 900);
  } else {
    resource.value = {
      ...resource.value,
      message: 'No conflict present. Everything already healthy.'
    };
    logEvent('No conflict to resolve.', 'active');
  }
};

onBeforeUnmount(() => {
  clearTimers();
});

resetVisualizer();
</script>
