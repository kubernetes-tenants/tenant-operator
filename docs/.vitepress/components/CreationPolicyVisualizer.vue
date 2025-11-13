<template>
  <div class="policy-visualizer creation-policy">
    <div class="visualizer-head">
      <div class="head-copy">
        <h3>CreationPolicy Visualizer</h3>
        <p>
          Follow the mermaid flow (<code>LynqForm → CreationPolicy → Runtime</code>) by selecting an option
          and driving real reconciliation events. Watch how Lynq reacts to drift, deletions, and template
          revisions.
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
      <section class="visualization-actions" aria-label="Simulate creation policy scenarios">
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
              <dt>Template revision</dt>
              <dd>rev {{ templateRevision }}</dd>
            </div>
            <div>
              <dt>Live revision</dt>
              <dd>rev {{ liveRevision }}</dd>
            </div>
            <div>
              <dt>Tracking mode</dt>
              <dd>{{ resource.tracking }}</dd>
            </div>
          </dl>

          <div class="resource-tags">
            <span
              v-for="tag in resource.tags"
              :key="tag.label"
              class="status-badge"
              :class="`state-${tag.state}`"
            >
              {{ tag.label }}
            </span>
          </div>
        </div>

        <div class="history-log" aria-label="Recent reconciliation events">
          <div class="history-title">Recent events</div>
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
  { key: 'reconciling', label: 'Reconciling' },
  { key: 'deleted', label: 'Deleted' },
  { key: 'conflict', label: 'Conflict' },
  { key: 'retained', label: 'Retained/Frozen' }
];

const policyOptions = [
  {
    value: 'WhenNeeded',
    label: 'WhenNeeded',
    summary: 'Continuously enforces template spec. Re-applies when it drifts or disappears.',
    badgeLabel: 'Continuous sync',
    badgeState: 'active',
    highlights: [
      'SSA apply keeps live objects aligned with the database',
      'Any drift triggers reconciling → active animation',
      'Manual deletes are back-filled automatically'
    ]
  },
  {
    value: 'Once',
    label: 'Once',
    summary: 'Creates the resource a single time. Live objects remain untouched unless deleted.',
    badgeLabel: 'Create + hold',
    badgeState: 'retained',
    highlights: [
      'Template revisions do not touch the existing object',
      'Great for one-time jobs and bootstrap scripts',
      'Deleting the object lets Lynq recreate it using the newest spec'
    ]
  }
];

const actions = [
  {
    key: 'drift',
    label: 'Spec Drift / Template Update',
    description: 'Simulate a spec change or drift detection coming from the database.',
    icon: '🌀'
  },
  {
    key: 'manualDelete',
    label: 'Manual Delete in Cluster',
    description: 'Pretend an engineer removed the object outside of Lynq.',
    icon: '🗑️'
  }
];

const stateLabels = {
  active: 'Active in cluster',
  reconciling: 'Reconciling',
  deleted: 'Deleted',
  conflict: 'Conflict',
  retained: 'Frozen'
};

const selectedPolicy = ref(policyOptions[0].value);
const templateRevision = ref(1);
const liveRevision = ref(1);
const resource = ref(createResource());
const actionLog = ref([]);
const activeAction = ref(null);
let historyCounter = 0;
const timers = [];

const currentOption = computed(() =>
  policyOptions.find((option) => option.value === selectedPolicy.value)
);

function createResource() {
  return {
    name: 'deployments/app',
    namespace: 'prod-cluster',
    state: 'active',
    tracking: selectedPolicy.value === 'WhenNeeded' ? 'SSA drift correction' : 'Single apply',
    message: 'Revision 1 is live inside the cluster.',
    tags: [
      { label: 'owner: lynq', state: 'active' },
      { label: selectedPolicy.value, state: selectedPolicy.value === 'WhenNeeded' ? 'active' : 'retained' }
    ]
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
  templateRevision.value = 1;
  liveRevision.value = 1;
  resource.value = createResource();
  resource.value.message = 'Revision 1 is live inside the cluster.';
  actionLog.value = [
    {
      id: ++historyCounter,
      text: 'Initial apply ensured revision 1 matches the template.',
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
  if (key === 'drift') {
    handleDrift();
  } else if (key === 'manualDelete') {
    handleManualDelete();
  }
};

const handleDrift = () => {
  clearTimers();
  templateRevision.value += 1;
  resource.value = {
    ...resource.value,
    templateRevision: templateRevision.value
  };

  if (selectedPolicy.value === 'WhenNeeded') {
    resource.value = {
      ...resource.value,
      state: 'reconciling',
      message: `Detected drift. Applying revision ${templateRevision.value} with SSA.`
    };
    logEvent(`Revision ${templateRevision.value} scheduled for SSA apply.`, 'reconciling');
    schedule(() => {
      liveRevision.value = templateRevision.value;
      resource.value = {
        ...resource.value,
        state: 'active',
        message: `Live object now matches revision ${liveRevision.value}.`
      };
      logEvent(`Revision ${liveRevision.value} is active.`, 'active');
    }, 1100);
  } else {
    resource.value = {
      ...resource.value,
      state: 'retained',
      message: `Template advanced to revision ${templateRevision.value}. Live object stays at revision ${liveRevision.value}.`
    };
    logEvent(
      `Revision ${templateRevision.value} ignored. CreationPolicy=Once keeps the first apply.`,
      'retained'
    );
  }
};

const handleManualDelete = () => {
  clearTimers();
  resource.value = {
    ...resource.value,
    state: 'deleted',
    message: 'Resource missing from cluster. Click recreate to watch Lynq re-apply it.'
  };
  logEvent('Cluster object disappeared. Awaiting recreation.', 'deleted');
};

const recreateResource = () => {
  if (resource.value.state !== 'deleted') {
    return;
  }
  resource.value = {
    ...resource.value,
    state: 'reconciling',
    message: 'Recreating object from template...'
  };
  logEvent('Recreating resource from template.', 'reconciling');

  schedule(() => {
    if (selectedPolicy.value === 'WhenNeeded') {
      liveRevision.value = templateRevision.value;
    } else {
      // "Once" only updates when a fresh object is created
      liveRevision.value = templateRevision.value;
    }
    resource.value = {
      ...resource.value,
      state: 'active',
      message: `Resource restored. Live revision is ${liveRevision.value}.`
    };
    logEvent('Resource restored to the cluster.', 'active');
  }, 900);
};

onBeforeUnmount(() => {
  clearTimers();
});

resetVisualizer();
</script>
