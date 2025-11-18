<template>
  <div ref="containerRef" class="field-ignore-animation">
    <div class="scene-container">
      <!-- LynqForm -->
      <div
        class="box-wrapper"
        :class="{
          'focused': focusedBox === 'lynqform',
          'dimmed': focusedBox && focusedBox !== 'lynqform',
          'visible': stage >= 1
        }"
      >
        <div class="yaml-box">
          <div class="box-header">
            <span class="box-title">LynqForm</span>
            <span v-if="stage >= 5" class="status-badge updated">Updated</span>
          </div>
          <div class="yaml-content">
            <div class="yaml-line">
              <span class="key">deployments</span><span class="punctuation">:</span>
            </div>
            <div class="yaml-line indent-1">
              <span class="punctuation">-</span> <span class="key">id</span><span class="punctuation">:</span> <span class="value">app</span>
            </div>
            <div class="yaml-line indent-2 highlight-yellow">
              <span class="key">ignoreFields</span><span class="punctuation">:</span>
            </div>
            <div class="yaml-line indent-3 highlight-yellow">
              <span class="punctuation">-</span> <span class="string">"$.spec.replicas"</span>
              <span class="lock-icon">🔒</span>
            </div>
            <div class="yaml-line indent-2">
              <span class="key">spec</span><span class="punctuation">:</span>
            </div>
            <div class="yaml-line indent-3">
              <span class="key">apiVersion</span><span class="punctuation">:</span> <span class="value">apps/v1</span>
            </div>
            <div class="yaml-line indent-3">
              <span class="key">kind</span><span class="punctuation">:</span> <span class="value">Deployment</span>
            </div>
            <div class="yaml-line indent-3">
              <span class="key">spec</span><span class="punctuation">:</span>
            </div>
            <div class="yaml-line indent-4 highlight-gray">
              <span class="key">replicas</span><span class="punctuation">:</span> <span class="value">3</span>
              <span class="badge-label gray">ignored</span>
            </div>
            <div class="yaml-line indent-4">
              <span class="key">template</span><span class="punctuation">:</span>
            </div>
            <div class="yaml-line indent-5">
              <span class="key">spec</span><span class="punctuation">:</span>
            </div>
            <div class="yaml-line indent-6">
              <span class="key">containers</span><span class="punctuation">:</span>
            </div>
            <div
              class="yaml-line indent-7"
              :class="{
                'highlight-green': stage >= 1 && stage < 5,
                'highlight-orange pulse-highlight': stage === 5
              }"
            >
              <span class="punctuation">-</span> <span class="key">image</span><span class="punctuation">:</span> <span class="value-animated">{{ data.templateImage }}</span>
              <span v-if="stage < 5" class="badge-label green">managed</span>
              <span v-if="stage >= 5" class="badge-label orange">updated</span>
            </div>
          </div>
        </div>
      </div>

      <!-- LynqNode -->
      <div
        class="box-wrapper"
        :class="{
          'focused': focusedBox === 'lynqnode',
          'dimmed': focusedBox && focusedBox !== 'lynqnode',
          'visible': stage >= 2
        }"
      >
        <div class="yaml-box">
          <div class="box-header">
            <span class="box-title">LynqNode</span>
            <span v-if="stage >= 7" class="status-badge updated">Updated</span>
          </div>
          <div class="yaml-content">
            <div class="yaml-line">
              <span class="key">spec</span><span class="punctuation">:</span>
            </div>
            <div class="yaml-line indent-1">
              <span class="key">deployments</span><span class="punctuation">:</span>
            </div>
            <div class="yaml-line indent-2">
              <span class="punctuation">-</span> <span class="key">id</span><span class="punctuation">:</span> <span class="value">app</span>
            </div>
            <div
              class="yaml-line indent-3 highlight-yellow"
              :class="{ 'pulse-highlight': stage === 9 }"
            >
              <span class="key">ignoreFields</span><span class="punctuation">:</span>
            </div>
            <div
              class="yaml-line indent-4 highlight-yellow"
              :class="{ 'pulse-highlight': stage === 9 }"
            >
              <span class="punctuation">-</span> <span class="string">"$.spec.replicas"</span>
              <span class="lock-icon" :class="{ 'shake': stage === 9 }">🔒</span>
              <span v-if="stage === 9" class="skip-icon">⏭️</span>
            </div>
            <div class="yaml-line indent-3">
              <span class="key">spec</span><span class="punctuation">:</span>
            </div>
            <div class="yaml-line indent-4">
              <span class="key">replicas</span><span class="punctuation">:</span> <span class="value">3</span>
              <span class="badge-label gray">ignored</span>
            </div>
            <div class="yaml-line indent-4">
              <span class="key">template</span><span class="punctuation">:</span>
            </div>
            <div class="yaml-line indent-5">
              <span class="key">spec</span><span class="punctuation">:</span>
            </div>
            <div class="yaml-line indent-6">
              <span class="key">containers</span><span class="punctuation">:</span>
            </div>
            <div
              class="yaml-line indent-7"
              :class="{ 'highlight-orange pulse-highlight': stage === 7 }"
            >
              <span class="punctuation">-</span> <span class="key">image</span><span class="punctuation">:</span> <span class="value-animated">{{ data.lynqnodeImage }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Deployment -->
      <div
        class="box-wrapper"
        :class="{
          'focused': focusedBox === 'deployment',
          'dimmed': focusedBox && focusedBox !== 'deployment',
          'visible': stage >= 3
        }"
      >
        <div class="yaml-box">
          <div class="box-header">
            <span class="box-title">Deployment</span>
            <span v-if="stage >= 11" class="status-badge success">Synced</span>
          </div>
          <div class="yaml-content">
            <div class="yaml-line">
              <span class="key">spec</span><span class="punctuation">:</span>
            </div>
            <div
              class="yaml-line indent-1 highlight-gray"
              :class="{ 'pulse-highlight': stage === 4 || stage === 11 }"
            >
              <span class="key">replicas</span><span class="punctuation">:</span> <span class="value-animated">{{ data.deploymentReplicas }}</span>
              <span class="lock-icon" :class="{ 'shake': stage === 11 || stage === 12 }">🔒</span>
              <span class="badge-label gray">ignored</span>
              <span v-if="stage === 4" class="badge-special hpa">⚖️ HPA</span>
            </div>
            <div class="yaml-line indent-1">
              <span class="key">template</span><span class="punctuation">:</span>
            </div>
            <div class="yaml-line indent-2">
              <span class="key">spec</span><span class="punctuation">:</span>
            </div>
            <div class="yaml-line indent-3">
              <span class="key">containers</span><span class="punctuation">:</span>
            </div>
            <div
              class="yaml-line indent-4"
              :class="{ 'highlight-green pulse-highlight': stage === 11 }"
            >
              <span class="punctuation">-</span> <span class="key">image</span><span class="punctuation">:</span> <span class="value-animated">{{ data.deploymentImage }}</span>
              <span v-if="stage >= 11" class="badge-special synced">✓ Synced</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Status Message -->
    <div class="status-bar" :class="{ 'show': message }">
      {{ message }}
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue';

const containerRef = ref(null);
const stage = ref(0);
const message = ref('');
const hasStarted = ref(false);

const data = reactive({
  templateImage: '1.20',
  lynqnodeImage: '1.20',
  deploymentImage: '1.20',
  deploymentReplicas: 3,
});

// Which box(es) should be focused
const focusedBox = computed(() => {
  if (stage.value === 1) return 'lynqform';
  if (stage.value === 2) return null; // Show both lynqform + lynqnode
  if (stage.value === 3) return null; // Show lynqnode + deployment
  if (stage.value === 4) return 'deployment'; // HPA changes
  if (stage.value === 5) return 'lynqform'; // Template update
  if (stage.value === 6 || stage.value === 7) return 'lynqnode'; // LynqNode update
  if (stage.value === 8 || stage.value === 9) return null; // Show lynqnode + deployment
  if (stage.value >= 10) return 'deployment'; // Final result
  return null;
});

// Animation sequence
const animate = async () => {
  const wait = (ms) => new Promise(r => setTimeout(r, ms));

  // Stage 1: LynqForm appears
  stage.value = 1;
  message.value = 'LynqForm defines template with ignoreFields';
  await wait(2500);

  // Stage 2: LynqNode appears
  stage.value = 2;
  message.value = 'Template renders to LynqNode';
  await wait(2500);

  // Stage 3: Deployment appears
  stage.value = 3;
  data.deploymentReplicas = 3;
  data.deploymentImage = '1.20';
  message.value = 'Initial Creation: All fields applied';
  await wait(2500);

  // Stage 4: HPA changes (focus on Deployment)
  stage.value = 4;
  await wait(800);
  data.deploymentReplicas = 5;
  message.value = 'HPA scales replicas: 3 → 5 (ignored field)';
  await wait(2500);

  // Stage 5: Template updates (focus on LynqForm)
  stage.value = 5;
  await wait(800);
  data.templateImage = '1.21';
  message.value = 'Template updated: nginx:1.20 → 1.21';
  await wait(2500);

  // Stage 6: Focus on LynqNode
  stage.value = 6;
  message.value = 'Update propagating to LynqNode...';
  await wait(1000);

  // Stage 7: LynqNode updates
  stage.value = 7;
  data.lynqnodeImage = '1.21';
  message.value = 'LynqNode spec updated';
  await wait(2500);

  // Stage 8: Show LynqNode + Deployment
  stage.value = 8;
  message.value = 'Reconciliation starts...';
  await wait(1500);

  // Stage 9: ignoreFields processing
  stage.value = 9;
  message.value = 'ignoreFields filters $.spec.replicas';
  await wait(2500);

  // Stage 10: Focus on Deployment
  stage.value = 10;
  message.value = 'Applying changes to Deployment...';
  await wait(1000);

  // Stage 11: Deployment syncs
  stage.value = 11;
  data.deploymentImage = '1.21';
  message.value = 'Result: replicas ignored, image synced';
  await wait(3000);

  // Stage 12: Final emphasis
  stage.value = 12;
  message.value = 'ignoreFields skips specified fields during sync';
  await wait(2500);

  // Reset
  message.value = '';
  await wait(500);

  stage.value = 0;
  data.templateImage = '1.20';
  data.lynqnodeImage = '1.20';
  data.deploymentImage = '1.20';
  data.deploymentReplicas = 3;

  await wait(1000);
  animate();
};

let observer = null;

onMounted(() => {
  observer = new IntersectionObserver(
    (entries) => {
      if (entries[0].isIntersecting && !hasStarted.value) {
        hasStarted.value = true;
        animate();
        observer.disconnect();
      }
    },
    { threshold: 0.3 }
  );

  if (containerRef.value) {
    observer.observe(containerRef.value);
  }
});

onUnmounted(() => {
  if (observer) observer.disconnect();
});
</script>

<style scoped>
.field-ignore-animation {
  padding: 2rem;
  background: linear-gradient(135deg, var(--vp-c-bg) 0%, var(--vp-c-bg-soft) 100%);
  border-radius: 12px;
  border: 1px solid var(--vp-c-divider);
}

.scene-container {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  max-width: 650px;
  margin: 0 auto;
}

.box-wrapper {
  opacity: 0;
  transform: scale(0.95) translateY(20px);
  transition: all 0.6s cubic-bezier(0.4, 0, 0.2, 1);
  filter: blur(0px);
}

.box-wrapper.visible {
  opacity: 1;
  transform: scale(1) translateY(0);
}

.box-wrapper.focused {
  transform: scale(1.02);
  z-index: 10;
  filter: blur(0px);
}

.box-wrapper.dimmed {
  opacity: 0.4;
  transform: scale(0.98);
  filter: blur(1px);
}

.yaml-box {
  background: var(--vp-c-bg);
  border: 1px solid var(--vp-c-divider);
  border-radius: 8px;
  padding: 1.25rem;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  transition: box-shadow 0.3s ease;
}

.box-wrapper.focused .yaml-box {
  box-shadow: 0 8px 24px rgba(139, 92, 246, 0.15);
  border-color: var(--vp-c-brand-1);
}

.box-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 1rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--vp-c-divider);
}

.box-title {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--vp-c-text-1);
  font-family: 'SF Mono', Monaco, monospace;
}

.status-badge {
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
}

.status-badge.updated {
  background: rgba(251, 146, 60, 0.2);
  color: #fb923c;
  border: 1px solid rgba(251, 146, 60, 0.4);
  animation: fadeIn 0.3s ease;
}

.status-badge.success {
  background: rgba(16, 185, 129, 0.2);
  color: #10b981;
  border: 1px solid rgba(16, 185, 129, 0.4);
  animation: fadeIn 0.3s ease;
}

.yaml-content {
  font-family: 'SF Mono', Monaco, monospace;
  font-size: 0.85rem;
  line-height: 1.8;
}

.yaml-line {
  padding: 0.25rem 0.6rem;
  border-radius: 3px;
  transition: all 0.4s ease;
}

.indent-1 { padding-left: 1.5rem; }
.indent-2 { padding-left: 3rem; }
.indent-3 { padding-left: 4.5rem; }
.indent-4 { padding-left: 6rem; }
.indent-5 { padding-left: 7.5rem; }
.indent-6 { padding-left: 9rem; }
.indent-7 { padding-left: 10.5rem; }

.key { color: #8b5cf6; font-weight: 500; }
.value { color: #10b981; }
.value-animated {
  color: #10b981;
  display: inline-block;
  transition: transform 0.3s ease;
}
.string { color: #f59e0b; }
.punctuation { color: var(--vp-c-text-2); }

.highlight-yellow {
  background: rgba(250, 204, 21, 0.15);
  border-left: 3px solid #facc15;
}

.highlight-gray {
  background: rgba(148, 163, 184, 0.1);
  border-left: 3px solid #94a3b8;
}

.highlight-green {
  background: rgba(16, 185, 129, 0.15);
  border-left: 3px solid #10b981;
}

.highlight-orange {
  background: rgba(251, 146, 60, 0.15);
  border-left: 3px solid #fb923c;
}

.pulse-highlight {
  animation: pulseHighlight 0.8s ease-out;
}

@keyframes pulseHighlight {
  0% {
    transform: scale(1);
    background: rgba(251, 146, 60, 0.3);
  }
  100% {
    transform: scale(1);
  }
}

.badge-label {
  display: inline-block;
  margin-left: 0.5rem;
  padding: 0.15rem 0.4rem;
  border-radius: 3px;
  font-size: 0.65rem;
  font-weight: 600;
  text-transform: uppercase;
}

.badge-label.gray {
  background: rgba(148, 163, 184, 0.2);
  color: #94a3b8;
}

.badge-label.green {
  background: rgba(16, 185, 129, 0.2);
  color: #10b981;
}

.badge-label.orange {
  background: rgba(251, 146, 60, 0.2);
  color: #fb923c;
  animation: fadeIn 0.3s ease;
}

.lock-icon {
  margin-left: 0.5rem;
  font-size: 0.9rem;
  display: inline-block;
  transition: transform 0.2s ease;
}

.lock-icon.shake {
  animation: shake 0.5s ease-in-out infinite;
}

.skip-icon {
  margin-left: 0.5rem;
  font-size: 0.9rem;
  animation: fadeIn 0.3s ease;
}

@keyframes shake {
  0%, 100% { transform: rotate(0deg); }
  25% { transform: rotate(-12deg); }
  75% { transform: rotate(12deg); }
}

.badge-special {
  display: inline-block;
  margin-left: 0.75rem;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 600;
  animation: fadeIn 0.3s ease;
}

.badge-special.hpa {
  background: rgba(251, 146, 60, 0.2);
  color: #fb923c;
  border: 1px solid rgba(251, 146, 60, 0.4);
}

.badge-special.synced {
  background: rgba(16, 185, 129, 0.2);
  color: #10b981;
  border: 1px solid rgba(16, 185, 129, 0.4);
}

@keyframes fadeIn {
  from { opacity: 0; transform: scale(0.9); }
  to { opacity: 1; transform: scale(1); }
}

.status-bar {
  margin-top: 2rem;
  padding: 0.75rem 1.5rem;
  background: linear-gradient(135deg, rgba(139, 92, 246, 0.1), rgba(6, 182, 212, 0.1));
  border: 1px solid var(--vp-c-brand-1);
  border-radius: 8px;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--vp-c-text-1);
  text-align: center;
  opacity: 0;
  transform: translateY(10px);
  transition: all 0.3s ease;
}

.status-bar.show {
  opacity: 1;
  transform: translateY(0);
}

@media (max-width: 768px) {
  .yaml-content {
    font-size: 0.75rem;
  }

  .indent-1 { padding-left: 1rem; }
  .indent-2 { padding-left: 2rem; }
  .indent-3 { padding-left: 3rem; }
  .indent-4 { padding-left: 4rem; }
  .indent-5 { padding-left: 5rem; }
  .indent-6 { padding-left: 6rem; }
  .indent-7 { padding-left: 7rem; }

  .status-bar {
    font-size: 0.75rem;
  }
}

@media (max-width: 480px) {
  .scene-container {
    gap: 1rem;
  }

  .yaml-content {
    font-size: 0.7rem;
  }

  .indent-1 { padding-left: 0.75rem; }
  .indent-2 { padding-left: 1.5rem; }
  .indent-3 { padding-left: 2.25rem; }
  .indent-4 { padding-left: 3rem; }
  .indent-5 { padding-left: 3.75rem; }
  .indent-6 { padding-left: 4.5rem; }
  .indent-7 { padding-left: 5.25rem; }
}
</style>
