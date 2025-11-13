<template>
  <div class="policy-visualizer patch-policy">
    <div class="visualizer-head">
      <div class="head-copy">
        <h3>PatchStrategy Visualizer</h3>
        <p>
          The mermaid path (<code>ConflictPolicy → PatchStrategy → Runtime</code>) determines how the spec reaches
          the cluster. Trigger external drift and apply different patch strategies to see how fields react.
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
        :class="{ active: selectedStrategy === option.value }"
        @click="selectPolicy(option.value)"
      >
        <div class="option-title">{{ option.label }}</div>
        <p>{{ option.summary }}</p>
      </button>
    </div>

    <div class="policy-description" v-if="currentOption">
      <div class="description-head">
        <h4>{{ currentOption.label }} strategy</h4>
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
      <section class="visualization-actions" aria-label="Simulate patch strategy scenarios">
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

        <div class="history-log" aria-label="Patch events">
          <div class="history-title">Patch events</div>
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

      <section class="field-tracks" aria-label="Field-level behavior">
        <article v-for="field in fields" :key="field.key" class="field-track">
          <div class="field-track-head">
            <div>
              <strong>{{ field.label }}</strong>
              <small>{{ field.description }}</small>
            </div>
            <span class="status-badge" :class="`state-${field.state}`">
              {{ field.statusText }}
            </span>
          </div>
          <div class="field-values">
            <div>
              <span>Template</span>
              <code>{{ field.template }}</code>
            </div>
            <div>
              <span>Live</span>
              <code>{{ field.live }}</code>
            </div>
          </div>
        </article>
      </section>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onBeforeUnmount } from 'vue';

const legendStates = [
  { key: 'active', label: 'Active' },
  { key: 'reconciling', label: 'Reconciling' },
  { key: 'conflict', label: 'Conflict' },
  { key: 'retained', label: 'Retained' },
  { key: 'deleted', label: 'Removed' }
];

const stateLabels = {
  active: 'Active in cluster',
  reconciling: 'Reconciling',
  conflict: 'Drift detected',
  retained: 'Preserved',
  deleted: 'Removed'
};

const policyOptions = [
  {
    value: 'apply',
    label: 'apply (SSA)',
    summary: 'Server-Side Apply with field ownership tracking handled by Kubernetes.',
    badgeLabel: 'Field ownership aware',
    badgeState: 'active',
    highlights: [
      'Detects conflicts before overwriting other controllers',
      'Assigns managedFields to lynq',
      'Ideal when multiple actors manage the same resource'
    ]
  },
  {
    value: 'merge',
    label: 'merge (strategic)',
    summary: 'Strategic merge patch merges JSON fields without owner tracking.',
    badgeLabel: 'Broad changes',
    badgeState: 'reconciling',
    highlights: [
      'Older clusters compatibility',
      'List merge rules can be surprising',
      'Fewer safety checks compared to SSA'
    ]
  },
  {
    value: 'replace',
    label: 'replace (PUT)',
    summary: 'Sends the entire object and replaces what is stored on the API server.',
    badgeLabel: 'Full replacement',
    badgeState: 'deleted',
    highlights: [
      'Removes any fields not present in the template',
      'Great when Lynq is the exclusive author',
      'High risk in shared resources'
    ]
  }
];

const actions = [
  {
    key: 'externalChange',
    label: 'External Controller Change',
    description: 'Simulate another controller mutating replicas/labels/sidecar.',
    icon: '♻️'
  },
  {
    key: 'applyTemplate',
    label: 'Apply Template Update',
    description: 'Run the selected patch strategy to correct drift.',
    icon: '🚀'
  }
];

const selectedStrategy = ref(policyOptions[0].value);
const resource = ref(createResource());
const fields = ref(createFields());
const actionLog = ref([]);
const activeAction = ref(null);
let historyCounter = 0;
const timers = [];

const currentOption = computed(() =>
  policyOptions.find((option) => option.value === selectedStrategy.value)
);

function createResource() {
  return {
    name: 'deployments/app',
    namespace: 'prod',
    state: 'active',
    message: 'Live spec matches template.',
    tags: [
      { label: `strategy: ${selectedStrategy.value}`, state: 'active' },
      { label: 'manager: lynq', state: 'active' }
    ]
  };
}

function createFields() {
  return [
    {
      key: 'replicas',
      label: 'spec.replicas',
      description: 'Desired replica count',
      template: '5',
      live: '5',
      state: 'active',
      statusText: 'Managed by Lynq'
    },
    {
      key: 'labels',
      label: 'metadata.labels.release',
      description: 'Traffic routing label',
      template: 'stable',
      live: 'stable',
      state: 'active',
      statusText: 'Managed by Lynq'
    },
    {
      key: 'sidecar',
      label: 'spec.template.spec.containers[sidecar]',
      description: 'External injected container',
      template: 'lynq-sidecar:v1',
      live: 'lynq-sidecar:v1',
      state: 'retained',
      statusText: 'Managed by other controller'
    },
    {
      key: 'status',
      label: 'status.availableReplicas',
      description: 'Runtime status (read-only)',
      template: 'runtime',
      live: '5 Ready',
      state: 'retained',
      statusText: 'Ignored by template'
    }
  ];
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
  fields.value = createFields();
  actionLog.value = [
    {
      id: ++historyCounter,
      text: 'Ready to apply patches. Cluster and template are aligned.',
      state: 'active',
      meta: 'start'
    }
  ];
  activeAction.value = null;
};

const selectPolicy = (value) => {
  if (selectedStrategy.value !== value) {
    selectedStrategy.value = value;
  }
};

watch(selectedStrategy, () => {
  resetVisualizer();
});

const runAction = (key) => {
  activeAction.value = key;
  if (key === 'externalChange') {
    simulateExternalChange();
  } else if (key === 'applyTemplate') {
    applyTemplate();
  }
};

const simulateExternalChange = () => {
  clearTimers();
  resource.value = {
    ...resource.value,
    state: 'conflict',
    message: 'Another controller modified replicas, labels, and the sidecar.'
  };
  updateField('replicas', {
    live: '4 (HPA)',
    state: 'conflict',
    statusText: 'External HPA scaled'
  });
  updateField('labels', {
    live: 'hotfix',
    state: 'conflict',
    statusText: 'ArgoCD override'
  });
  updateField('sidecar', {
    live: 'traffic-agent:v2',
    state: 'conflict',
    statusText: 'Injected by traffic-operator'
  });
  logEvent('External changes introduced drift.', 'conflict');
};

const applyTemplate = () => {
  clearTimers();
  resource.value = {
    ...resource.value,
    state: 'reconciling',
    message: `Applying template with ${selectedStrategy.value} strategy...`
  };
  logEvent(`Applying template via ${selectedStrategy.value}.`, 'reconciling');
  runStrategyEffects();
  schedule(() => {
    resource.value = {
      ...resource.value,
      state: 'active',
      message: 'Cluster back in sync with template.',
      tags: [
        { label: `strategy: ${selectedStrategy.value}`, state: 'active' },
        { label: 'manager: lynq', state: 'active' }
      ]
    };
    logEvent('Patch completed successfully.', 'active');
  }, 1100);
};

const updateField = (key, patch) => {
  fields.value = fields.value.map((field) =>
    field.key === key ? { ...field, ...patch } : field
  );
};

const runStrategyEffects = () => {
  if (selectedStrategy.value === 'apply') {
    updateField('replicas', {
      live: '5 (SSA)',
      state: 'active',
      statusText: 'SSA owns this field'
    });
    updateField('labels', {
      live: 'stable',
      state: 'active',
      statusText: 'SSA updated label'
    });
    updateField('sidecar', {
      live: 'traffic-agent:v2',
      state: 'retained',
      statusText: 'Preserved (other owner)'
    });
    updateField('status', {
      live: '4 Ready',
      state: 'retained',
      statusText: 'Status untouched'
    });
  } else if (selectedStrategy.value === 'merge') {
    updateField('replicas', {
      live: '5 (merged)',
      state: 'active',
      statusText: 'Merged replica value'
    });
    updateField('labels', {
      live: 'stable',
      state: 'active',
      statusText: 'Merged label'
    });
    updateField('sidecar', {
      live: 'lynq-sidecar:v1 (overrode traffic-agent)',
      state: 'conflict',
      statusText: 'List merge may clobber other controllers'
    });
    updateField('status', {
      live: 'reset by merge',
      state: 'conflict',
      statusText: 'Risky: status may be overwritten'
    });
    logEvent('Strategic merge overwrote sidecar + status fields.', 'conflict');
  } else {
    updateField('replicas', {
      live: '5 (replace)',
      state: 'active',
      statusText: 'Exact match enforced'
    });
    updateField('labels', {
      live: 'stable',
      state: 'active',
      statusText: 'Exact labels only'
    });
    updateField('sidecar', {
      live: 'removed (not in template)',
      state: 'deleted',
      statusText: 'Field removed'
    });
    updateField('status', {
      live: 'reset (needs repopulation)',
      state: 'deleted',
      statusText: 'All unspecified fields wiped'
    });
    logEvent('Replace wiped fields not present in the template.', 'deleted');
  }
};

onBeforeUnmount(() => {
  clearTimers();
});

resetVisualizer();
</script>
