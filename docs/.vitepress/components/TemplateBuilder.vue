<template>
  <div class="template-builder">
    <div class="builder-container">
      <!-- Left Panel: Form UI -->
      <div class="form-panel">
        <div class="panel-header">
          <h3>Form Builder</h3>
          <button @click="resetBuilder" class="btn-secondary">🔄 Reset</button>
        </div>

        <!-- Hub Settings -->
        <section class="form-section">
          <h4>Hub Settings</h4>
          <div class="form-group">
            <label>Hub ID *</label>
            <input
              v-model="hubId"
              type="text"
              placeholder="my-hub"
              class="form-input"
            />
            <span class="hint">Reference to your LynqHub</span>
          </div>
        </section>

        <!-- Template Variables -->
        <section class="form-section collapsible">
          <h4 @click="toggleSection('variables')" class="section-toggle">
            <span class="toggle-icon">{{ sectionsExpanded.variables ? '▼' : '▶' }}</span>
            Available Variables
          </h4>
          <div v-show="sectionsExpanded.variables" class="variables-info">
            <div class="variable-item">
              <code>.uid</code>
              <span>Node unique identifier</span>
            </div>
            <div class="variable-item">
              <code>.host</code>
              <span>Extracted host from URL</span>
            </div>
            <div class="variable-item">
              <code>.hostOrUrl</code>
              <span>Original URL/host value</span>
            </div>
            <div class="variable-item">
              <code>.activate</code>
              <span>Activation status</span>
            </div>
            <div class="info-box">
              <strong>Usage:</strong> Use variables in templates like <code v-pre>{{ .uid }}-app</code>
            </div>
          </div>
        </section>

        <!-- Resources -->
        <section class="form-section">
          <div class="section-header">
            <h4>Resources</h4>
            <button @click="showAddResource = true" class="btn-primary">+ Add Resource</button>
          </div>

          <!-- Resource List -->
          <div v-if="resources.length === 0" class="empty-state">
            <p>No resources added yet</p>
            <p class="hint">Click "Add Resource" to start building your template</p>
          </div>

          <div v-else class="resource-list">
            <div
              v-for="(resource, index) in resources"
              :key="index"
              class="resource-card"
              @click="editResource(index)"
            >
              <div class="resource-header">
                <span class="resource-type">{{ resource.type }}</span>
                <button @click.stop="removeResource(index)" class="btn-icon">🗑️</button>
              </div>
              <div class="resource-id">{{ resource.id }}</div>
              <div v-if="resource.dependIds?.length" class="resource-deps">
                Depends on: {{ resource.dependIds.join(', ') }}
              </div>
            </div>
          </div>
        </section>
      </div>

      <!-- Right Panel: YAML Editor/Preview -->
      <div class="preview-panel">
        <div class="panel-header">
          <div class="tab-buttons">
            <button
              @click="activeTab = 'preview'"
              :class="['tab-btn', { active: activeTab === 'preview' }]"
            >
              Preview
            </button>
            <button
              @click="activeTab = 'editor'"
              :class="['tab-btn', { active: activeTab === 'editor' }]"
            >
              Editor
            </button>
          </div>
          <div class="header-actions">
            <button v-if="activeTab === 'preview'" @click="copyYaml" class="btn-secondary">📋 Copy</button>
            <button v-if="activeTab === 'preview'" @click="downloadYaml" class="btn-secondary">💾 Download</button>
            <button v-if="activeTab === 'editor' && yamlEditorContent !== generatedYaml" @click="importFromYaml" class="btn-primary">⬅️ Import to UI</button>
          </div>
        </div>

        <!-- Preview Tab -->
        <div v-show="activeTab === 'preview'" class="yaml-preview">
          <pre><code>{{ generatedYaml }}</code></pre>
        </div>

        <!-- Editor Tab -->
        <div v-show="activeTab === 'editor'" class="yaml-editor-container">
          <textarea
            v-model="yamlEditorContent"
            class="yaml-editor"
            placeholder="Paste your LynqForm YAML here..."
            spellcheck="false"
          ></textarea>
          <div v-if="parseError" class="parse-error">
            ⚠️ {{ parseError }}
          </div>
          <div v-if="activeTab === 'editor'" class="editor-hint">
            💡 Paste your YAML and click "Import to UI" to load it into the form
          </div>
        </div>

        <div v-if="copySuccess" class="copy-toast">✓ Copied to clipboard!</div>
      </div>
    </div>

    <!-- Add/Edit Resource Modal -->
    <div v-if="showAddResource || editingIndex !== null" class="modal-overlay" @click="closeModal">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>{{ editingIndex !== null ? 'Edit Resource' : 'Add Resource' }}</h3>
          <button @click="closeModal" class="btn-close">✕</button>
        </div>

        <div class="modal-body">
          <!-- Resource Type -->
          <div class="form-group">
            <label>Resource Type *</label>
            <select v-model="currentResource.type" class="form-input">
              <option value="">Select type...</option>
              <option value="namespaces">Namespace</option>
              <option value="deployments">Deployment</option>
              <option value="statefulSets">StatefulSet</option>
              <option value="daemonSets">DaemonSet</option>
              <option value="services">Service</option>
              <option value="ingresses">Ingress</option>
              <option value="configMaps">ConfigMap</option>
              <option value="secrets">Secret</option>
              <option value="persistentVolumeClaims">PersistentVolumeClaim</option>
              <option value="jobs">Job</option>
              <option value="cronJobs">CronJob</option>
              <option value="horizontalPodAutoscalers">HorizontalPodAutoscaler</option>
              <option value="manifests">Manifest (Raw YAML)</option>
            </select>
          </div>

          <!-- Resource ID -->
          <div class="form-group">
            <label>Resource ID *</label>
            <input
              v-model="currentResource.id"
              type="text"
              placeholder="my-resource"
              class="form-input"
            />
            <span class="hint">Unique identifier for this resource</span>
          </div>

          <!-- Name Template -->
          <div class="form-group">
            <label>Name Template *</label>
            <input
              v-model="currentResource.nameTemplate"
              type="text"
              placeholder="{{ .uid }}-app"
              class="form-input"
            />
            <span class="hint">Go template for resource name</span>
          </div>

          <!-- Dependencies -->
          <div class="form-group">
            <label>Dependencies</label>
            <div class="dep-selector">
              <div v-if="availableDeps.length === 0" class="hint">
                No other resources available for dependencies
              </div>
              <div v-else>
                <label v-for="dep in availableDeps" :key="dep.id" class="checkbox-label">
                  <input
                    type="checkbox"
                    :value="dep.id"
                    v-model="currentResource.dependIds"
                  />
                  {{ dep.id }} ({{ dep.type }})
                </label>
              </div>
            </div>
          </div>

          <!-- Policies -->
          <div class="form-section-divider">
            <h5>Policies</h5>
          </div>

          <!-- Creation Policy -->
          <div class="form-group">
            <label>Creation Policy</label>
            <select v-model="currentResource.creationPolicy" class="form-input">
              <option value="WhenNeeded">WhenNeeded (default) - Reapply on changes</option>
              <option value="Once">Once - Create only once, never reapply</option>
            </select>
            <span class="hint">Controls when resource is created/updated</span>
          </div>

          <!-- Deletion Policy -->
          <div class="form-group">
            <label>Deletion Policy</label>
            <select v-model="currentResource.deletionPolicy" class="form-input">
              <option value="Delete">Delete (default) - Remove on node deletion</option>
              <option value="Retain">Retain - Keep resource after deletion</option>
            </select>
            <span class="hint">What happens when node is deleted</span>
          </div>

          <!-- Conflict Policy -->
          <div class="form-group">
            <label>Conflict Policy</label>
            <select v-model="currentResource.conflictPolicy" class="form-input">
              <option value="Stuck">Stuck (default) - Fail on conflicts</option>
              <option value="Force">Force - Take ownership forcefully</option>
            </select>
            <span class="hint">How to handle resource conflicts</span>
          </div>

          <!-- Patch Strategy -->
          <div class="form-group">
            <label>Patch Strategy</label>
            <select v-model="currentResource.patchStrategy" class="form-input">
              <option value="apply">apply (default) - Server-Side Apply</option>
              <option value="merge">merge - Strategic merge patch</option>
              <option value="replace">replace - Full replacement</option>
            </select>
            <span class="hint">How updates are applied</span>
          </div>

          <!-- Wait for Ready -->
          <div class="form-group">
            <label class="checkbox-label">
              <input type="checkbox" v-model="currentResource.waitForReady" />
              Wait for resource to be ready before continuing
            </label>
          </div>

          <!-- Timeout -->
          <div class="form-group">
            <label>Timeout (seconds)</label>
            <input
              v-model.number="currentResource.timeoutSeconds"
              type="number"
              min="1"
              max="3600"
              placeholder="300"
              class="form-input"
            />
            <span class="hint">Max wait time for readiness (default: 300, max: 3600)</span>
          </div>
        </div>

        <div class="modal-footer">
          <button @click="closeModal" class="btn-secondary">Cancel</button>
          <button @click="saveResource" class="btn-primary" :disabled="!isResourceValid">
            {{ editingIndex !== null ? 'Update' : 'Add' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue';
import yaml from 'js-yaml';

// State
const hubId = ref('my-hub');
const resources = ref([]);
const showAddResource = ref(false);
const editingIndex = ref(null);
const copySuccess = ref(false);
const activeTab = ref('preview');
const yamlEditorContent = ref('');
const parseError = ref('');

const sectionsExpanded = ref({
  variables: false
});

const currentResource = ref({
  type: '',
  id: '',
  nameTemplate: '',
  dependIds: [],
  creationPolicy: 'WhenNeeded',
  deletionPolicy: 'Delete',
  conflictPolicy: 'Stuck',
  patchStrategy: 'apply',
  waitForReady: true,
  timeoutSeconds: 300
});

// Watch for tab changes and sync editor content
watch(activeTab, (newTab) => {
  if (newTab === 'editor' && !yamlEditorContent.value) {
    yamlEditorContent.value = generatedYaml.value;
  }
  parseError.value = '';
});

// Computed
const availableDeps = computed(() => {
  if (editingIndex.value !== null) {
    return resources.value.filter((_, index) => index !== editingIndex.value);
  }
  return resources.value;
});

const isResourceValid = computed(() => {
  return currentResource.value.type &&
         currentResource.value.id &&
         currentResource.value.nameTemplate;
});

const generatedYaml = computed(() => {
  if (!hubId.value) {
    return '# Set Hub ID to start building your form';
  }

  const template = {
    apiVersion: 'operator.lynq.sh/v1',
    kind: 'LynqForm',
    metadata: {
      name: 'my-form'
    },
    spec: {
      hubId: hubId.value
    }
  };

  // Group resources by type
  resources.value.forEach(resource => {
    if (!template.spec[resource.type]) {
      template.spec[resource.type] = [];
    }

    const resourceSpec = {
      id: resource.id,
      nameTemplate: resource.nameTemplate
    };

    if (resource.dependIds && resource.dependIds.length > 0) {
      resourceSpec.dependIds = resource.dependIds;
    }

    // Add policies only if non-default
    if (resource.creationPolicy && resource.creationPolicy !== 'WhenNeeded') {
      resourceSpec.creationPolicy = resource.creationPolicy;
    }

    if (resource.deletionPolicy && resource.deletionPolicy !== 'Delete') {
      resourceSpec.deletionPolicy = resource.deletionPolicy;
    }

    if (resource.conflictPolicy && resource.conflictPolicy !== 'Stuck') {
      resourceSpec.conflictPolicy = resource.conflictPolicy;
    }

    if (resource.patchStrategy && resource.patchStrategy !== 'apply') {
      resourceSpec.patchStrategy = resource.patchStrategy;
    }

    if (resource.waitForReady === false) {
      resourceSpec.waitForReady = false;
    }

    if (resource.timeoutSeconds && resource.timeoutSeconds !== 300) {
      resourceSpec.timeoutSeconds = resource.timeoutSeconds;
    }

    // Minimal spec placeholder
    resourceSpec.spec = {
      '# Add your resource spec here': null
    };

    template.spec[resource.type].push(resourceSpec);
  });

  return yaml.dump(template, {
    indent: 2,
    lineWidth: -1,
    noRefs: true
  });
});

// Methods
const toggleSection = (section) => {
  sectionsExpanded.value[section] = !sectionsExpanded.value[section];
};

const resetBuilder = () => {
  if (confirm('Are you sure you want to reset? All resources will be removed.')) {
    hubId.value = 'my-hub';
    resources.value = [];
  }
};

const editResource = (index) => {
  editingIndex.value = index;
  currentResource.value = { ...resources.value[index] };
  if (!currentResource.value.dependIds) {
    currentResource.value.dependIds = [];
  }
};

const removeResource = (index) => {
  if (confirm('Remove this resource?')) {
    resources.value.splice(index, 1);
  }
};

const saveResource = () => {
  const resource = {
    type: currentResource.value.type,
    id: currentResource.value.id,
    nameTemplate: currentResource.value.nameTemplate,
    dependIds: currentResource.value.dependIds || [],
    creationPolicy: currentResource.value.creationPolicy,
    deletionPolicy: currentResource.value.deletionPolicy,
    conflictPolicy: currentResource.value.conflictPolicy,
    patchStrategy: currentResource.value.patchStrategy,
    waitForReady: currentResource.value.waitForReady,
    timeoutSeconds: currentResource.value.timeoutSeconds
  };

  if (editingIndex.value !== null) {
    resources.value[editingIndex.value] = resource;
  } else {
    resources.value.push(resource);
  }

  closeModal();
};

const closeModal = () => {
  showAddResource.value = false;
  editingIndex.value = null;
  currentResource.value = {
    type: '',
    id: '',
    nameTemplate: '',
    dependIds: [],
    creationPolicy: 'WhenNeeded',
    deletionPolicy: 'Delete',
    conflictPolicy: 'Stuck',
    patchStrategy: 'apply',
    waitForReady: true,
    timeoutSeconds: 300
  };
};

const copyYaml = async () => {
  try {
    await navigator.clipboard.writeText(generatedYaml.value);
    copySuccess.value = true;
    setTimeout(() => {
      copySuccess.value = false;
    }, 2000);
  } catch (err) {
    console.error('Failed to copy:', err);
  }
};

const downloadYaml = () => {
  const blob = new Blob([generatedYaml.value], { type: 'text/yaml' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = 'lynqform-template.yaml';
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
};

const importFromYaml = () => {
  parseError.value = '';

  try {
    const parsed = yaml.load(yamlEditorContent.value);

    if (!parsed || !parsed.spec) {
      parseError.value = 'Invalid LynqForm: missing spec field';
      return;
    }

    // Extract hub ID
    if (parsed.spec.hubId) {
      hubId.value = parsed.spec.hubId;
    }

    // Resource type mapping (API field name to internal type)
    const resourceTypes = [
      'serviceAccounts', 'deployments', 'statefulSets', 'daemonSets',
      'services', 'configMaps', 'secrets', 'persistentVolumeClaims',
      'ingresses', 'jobs', 'cronJobs', 'horizontalPodAutoscalers',
      'namespaces', 'manifests'
    ];

    // Extract resources
    const extractedResources = [];
    resourceTypes.forEach(type => {
      if (parsed.spec[type] && Array.isArray(parsed.spec[type])) {
        parsed.spec[type].forEach(resource => {
          if (resource.id) {
            extractedResources.push({
              type: type,
              id: resource.id,
              nameTemplate: resource.nameTemplate || '',
              dependIds: resource.dependIds || [],
              creationPolicy: resource.creationPolicy || 'WhenNeeded',
              deletionPolicy: resource.deletionPolicy || 'Delete',
              conflictPolicy: resource.conflictPolicy || 'Stuck',
              patchStrategy: resource.patchStrategy || 'apply',
              waitForReady: resource.waitForReady !== false,
              timeoutSeconds: resource.timeoutSeconds || 300
            });
          }
        });
      }
    });

    if (extractedResources.length === 0) {
      parseError.value = 'No resources with "id" field found in template';
      return;
    }

    // Update resources
    resources.value = extractedResources;

    // Switch to preview tab to show result
    activeTab.value = 'preview';

    // Show success message
    copySuccess.value = true;
    setTimeout(() => {
      copySuccess.value = false;
    }, 2000);

  } catch (error) {
    parseError.value = `YAML parsing error: ${error.message}`;
  }
};
</script>

<style scoped>
.template-builder {
  margin: 2rem 0;
}

.builder-container {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.5rem;
  min-height: 600px;
}

.form-panel,
.preview-panel {
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

.panel-header h3 {
  margin: 0;
  font-size: 1.1rem;
  color: var(--vp-c-text-1);
}

.header-actions {
  display: flex;
  gap: 0.5rem;
}

.form-panel {
  overflow-y: auto;
  max-height: 800px;
}

.form-section {
  padding: 1.5rem;
  border-bottom: 1px solid var(--vp-c-divider);
}

.form-section:last-child {
  border-bottom: none;
}

.form-section h4 {
  margin: 0 0 1rem 0;
  font-size: 0.95rem;
  color: var(--vp-c-text-1);
}

.section-toggle {
  cursor: pointer;
  user-select: none;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.toggle-icon {
  font-size: 0.8rem;
  color: var(--vp-c-text-2);
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.section-header h4 {
  margin: 0;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group:last-child {
  margin-bottom: 0;
}

.form-section-divider {
  margin: 1.5rem 0 1rem 0;
  padding-top: 1rem;
  border-top: 2px solid var(--vp-c-divider);
}

.form-section-divider h5 {
  margin: 0 0 0.5rem 0;
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--vp-c-brand);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-size: 0.9rem;
  font-weight: 500;
  color: var(--vp-c-text-1);
}

.form-input {
  width: 100%;
  padding: 0.5rem 0.75rem;
  background: var(--vp-c-bg);
  border: 1px solid var(--vp-c-divider);
  border-radius: 4px;
  color: var(--vp-c-text-1);
  font-size: 0.9rem;
  font-family: inherit;
}

.form-input:focus {
  outline: none;
  border-color: var(--vp-c-brand);
}

.hint {
  display: block;
  margin-top: 0.25rem;
  font-size: 0.85rem;
  color: var(--vp-c-text-2);
  font-style: italic;
}

.variables-info {
  margin-top: 1rem;
}

.variable-item {
  display: flex;
  justify-content: space-between;
  padding: 0.5rem;
  margin-bottom: 0.5rem;
  background: var(--vp-c-bg-soft);
  border-radius: 4px;
}

.variable-item code {
  color: var(--vp-c-brand);
  font-weight: 500;
}

.variable-item span {
  color: var(--vp-c-text-2);
  font-size: 0.85rem;
}

.info-box {
  margin-top: 1rem;
  padding: 0.75rem;
  background: var(--vp-c-brand-soft);
  border-left: 3px solid var(--vp-c-brand);
  border-radius: 4px;
  font-size: 0.85rem;
}

.info-box code {
  background: var(--vp-c-bg);
  padding: 0.2rem 0.4rem;
  border-radius: 3px;
}

.empty-state {
  padding: 2rem;
  text-align: center;
  color: var(--vp-c-text-2);
}

.empty-state p {
  margin: 0.5rem 0;
}

.resource-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.resource-card {
  padding: 1rem;
  background: var(--vp-c-bg-soft);
  border: 1px solid var(--vp-c-divider);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.resource-card:hover {
  border-color: var(--vp-c-brand);
  background: var(--vp-c-bg);
}

.resource-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.resource-type {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  background: var(--vp-c-brand-soft);
  color: var(--vp-c-brand);
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
}

.resource-id {
  font-weight: 600;
  color: var(--vp-c-text-1);
  margin-bottom: 0.25rem;
}

.resource-deps {
  font-size: 0.85rem;
  color: var(--vp-c-text-2);
}

.tab-buttons {
  display: flex;
  gap: 0.5rem;
}

.tab-btn {
  padding: 0.5rem 1rem;
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--vp-c-text-2);
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.tab-btn:hover {
  color: var(--vp-c-text-1);
}

.tab-btn.active {
  color: var(--vp-c-brand);
  border-bottom-color: var(--vp-c-brand);
}

.yaml-preview {
  flex: 1;
  overflow: auto;
  padding: 1rem;
  background: var(--vp-c-bg);
}

.yaml-preview pre {
  margin: 0;
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 0.85rem;
  line-height: 1.6;
  color: var(--vp-c-text-1);
}

.yaml-editor-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
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
  resize: none;
  outline: none;
}

.parse-error {
  padding: 0.75rem 1rem;
  background: var(--vp-c-danger-soft);
  color: #ef5350;
  font-size: 0.85rem;
  border-top: 1px solid var(--vp-c-divider);
}

.editor-hint {
  padding: 0.75rem 1rem;
  background: var(--vp-c-brand-soft);
  color: var(--vp-c-brand);
  font-size: 0.85rem;
  border-top: 1px solid var(--vp-c-divider);
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.copy-toast {
  position: fixed;
  bottom: 2rem;
  right: 2rem;
  padding: 0.75rem 1.5rem;
  background: var(--vp-c-brand);
  color: white;
  border-radius: 6px;
  font-weight: 500;
  animation: slideIn 0.3s ease;
  z-index: 1000;
}

@keyframes slideIn {
  from {
    transform: translateY(100%);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

/* Buttons */
.btn-primary,
.btn-secondary,
.btn-icon {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 6px;
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-primary {
  background: var(--vp-c-brand);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: var(--vp-c-brand-dark);
  transform: translateY(-1px);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  background: var(--vp-c-bg-soft);
  color: var(--vp-c-text-1);
  border: 1px solid var(--vp-c-divider);
}

.btn-secondary:hover {
  background: var(--vp-c-bg);
  border-color: var(--vp-c-brand);
}

.btn-icon {
  padding: 0.25rem 0.5rem;
  background: transparent;
  font-size: 1rem;
}

.btn-icon:hover {
  background: var(--vp-c-danger-soft);
}

/* Modal */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  padding: 2rem;
}

.modal-content {
  background: var(--vp-c-bg);
  border-radius: 8px;
  max-width: 600px;
  width: 100%;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.3);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid var(--vp-c-divider);
}

.modal-header h3 {
  margin: 0;
  color: var(--vp-c-text-1);
}

.btn-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: var(--vp-c-text-2);
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
}

.btn-close:hover {
  background: var(--vp-c-bg-soft);
  color: var(--vp-c-text-1);
}

.modal-body {
  flex: 1;
  overflow-y: auto;
  padding: 1.5rem;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1.5rem;
  border-top: 1px solid var(--vp-c-divider);
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem;
  cursor: pointer;
  border-radius: 4px;
  font-size: 0.9rem;
}

.checkbox-label:hover {
  background: var(--vp-c-bg-soft);
}

.checkbox-label input[type="checkbox"] {
  cursor: pointer;
}

.dep-selector {
  padding: 0.75rem;
  background: var(--vp-c-bg-soft);
  border-radius: 4px;
}

/* Responsive */
@media (max-width: 1024px) {
  .builder-container {
    grid-template-columns: 1fr;
  }

  .preview-panel {
    min-height: 400px;
  }
}
</style>
