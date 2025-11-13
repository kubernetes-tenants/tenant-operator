<template>
  <section
    class="qs-step"
    :class="{ 'qs-step--done': completed }"
    :id="`qs-step-${step}`"
    ref="rootEl"
  >
    <header class="step-head">
      <div class="step-meta">
        <span class="chip">Step {{ step }}</span>
        <span v-if="duration" class="chip ghost">{{ duration }}</span>
      </div>
      <h4>{{ title }}</h4>
      <p class="focus">{{ focus }}</p>
    </header>

    <div class="cmd-block">
      <div class="cmd-label">
        Command
        <button class="copy-btn" @click="copyCommand" :disabled="copyStatus === 'copied'">
          <span v-if="copyStatus === 'idle'">Copy</span>
          <span v-else-if="copyStatus === 'copied'">Copied ✓</span>
          <span v-else>Retry</span>
        </button>
      </div>
      <code>{{ command }}</code>
    </div>

    <div v-if="prerequisites.length" class="pre-reqs">
      <p class="label">Prerequisites</p>
      <ul>
        <li v-for="item in prerequisites" :key="item">{{ item }}</li>
      </ul>
    </div>

    <div class="creates">
      <p class="label">This step creates</p>
      <ul>
        <li v-for="item in creates" :key="item">{{ item }}</li>
      </ul>
    </div>

    <div v-if="checklistState.length" class="checklist">
      <p class="label">Post-run checks (click to toggle)</p>
      <div class="checks">
        <button
          v-for="(item, idx) in checklistState"
          :key="item.label"
          class="check"
          :class="{ checked: item.done }"
          @click="toggleChecklist(idx)"
        >
          <span class="box">{{ item.done ? '✓' : '' }}</span>
          <span>{{ item.label }}</span>
        </button>
      </div>
      <div v-if="progress.total" class="progress">
        <div class="bar">
          <span class="fill" :style="{ width: `${progress.percent}%` }"></span>
        </div>
        <span class="percent">{{ progress.percent }}% done</span>
      </div>
    </div>

    <footer class="next-hint">
      <div class="status">
        <button class="complete-btn" @click="completed = !completed">
          {{ completed ? 'Unmark step' : 'Mark step done' }}
        </button>
        <span v-if="completed" class="status-pill">Completed</span>
      </div>
      <p v-if="nextHint">
        <strong>Move to next step:</strong> {{ nextHint }}
      </p>
    </footer>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';

const props = withDefaults(
  defineProps<{
    step: number;
    title: string;
    duration?: string;
    command: string;
    focus: string;
    creates?: string[];
    prerequisites?: string[];
    checklist?: string[];
    nextHint?: string;
    nextTargetId?: string;
  }>(),
  {
    duration: '',
    creates: () => [],
    prerequisites: () => [],
    checklist: () => [],
    nextHint: '',
    nextTargetId: '',
  }
);

const completed = ref(false);
const copyStatus = ref<'idle' | 'copied' | 'error'>('idle');
const checklistState = ref(props.checklist.map((label) => ({ label, done: false })));
const rootEl = ref<HTMLElement | null>(null);

watch(
  () => props.checklist,
  (newList) => {
    checklistState.value = newList.map((label) => ({ label, done: false }));
  }
);

const progress = computed(() => {
  const total = checklistState.value.length;
  const done = checklistState.value.filter((item) => item.done).length;
  return {
    total,
    percent: total ? Math.round((done / total) * 100) : 0,
  };
});

const scrollToNext = () => {
  if (!props.nextTargetId || typeof window === 'undefined') return;
  window.requestAnimationFrame(() => {
    const targetEl = document.getElementById(props.nextTargetId!);
    if (targetEl) {
      const computed = window.getComputedStyle(document.documentElement);
      const navOffset = computed.getPropertyValue('--vp-nav-height');
      const layoutOffset = computed.getPropertyValue('--vp-layout-top-height');
      const nav = navOffset ? parseInt(navOffset, 10) || 0 : 0;
      const layout = layoutOffset ? parseInt(layoutOffset, 10) || 0 : 0;
      const offset = nav + layout;
      const top = targetEl.getBoundingClientRect().top + window.scrollY - offset;
      window.scrollTo({ top, behavior: 'smooth' });
    }
  });
};

watch(
  () => completed.value,
  (isDone, previous) => {
    if (isDone && !previous) {
      scrollToNext();
    }
  }
);

const copyCommand = async () => {
  if (copyStatus.value === 'copied') return;
  try {
    if (typeof navigator !== 'undefined' && navigator.clipboard) {
      await navigator.clipboard.writeText(props.command);
      copyStatus.value = 'copied';
      setTimeout(() => (copyStatus.value = 'idle'), 1600);
    } else {
      throw new Error('Clipboard not available');
    }
  } catch (err) {
    console.warn('[QuickstartStep] copy failed', err);
    copyStatus.value = 'error';
    setTimeout(() => (copyStatus.value = 'idle'), 2000);
  }
};

const toggleChecklist = (idx: number) => {
  checklistState.value[idx].done = !checklistState.value[idx].done;
};
</script>

<style scoped>
.qs-step {
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 18px;
  padding: 1.75rem;
  margin: 1.5rem 0;
  background: var(--vp-c-bg-alt);
  transition: border-color 0.3s ease, box-shadow 0.3s ease;
}

.qs-step--done {
  border-color: var(--vp-c-brand);
  box-shadow: 0 12px 30px rgba(0, 0, 0, 0.18);
}

.step-head h4 {
  margin: 0.4rem 0;
  font-size: 1.35rem;
}

.focus {
  margin: 0;
  color: var(--vp-c-text-2);
}

.chip {
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 999px;
  padding: 0.2rem 0.75rem;
  font-size: 0.75rem;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.chip.ghost {
  border-color: transparent;
  background: rgba(255, 255, 255, 0.05);
}

.step-meta {
  display: flex;
  gap: 0.6rem;
  align-items: center;
  flex-wrap: wrap;
}

.cmd-block {
  margin: 1.25rem 0;
  padding: 1rem;
  border-radius: 12px;
  background: rgba(0, 0, 0, 0.25);
}

.cmd-label {
  display: flex;
  justify-content: space-between;
  font-size: 0.85rem;
  color: var(--vp-c-text-2);
  margin-bottom: 0.5rem;
}

.copy-btn {
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 8px;
  padding: 0.25rem 0.65rem;
  background: transparent;
  color: inherit;
  cursor: pointer;
  transition: all 0.2s ease;
}

.copy-btn:disabled {
  opacity: 0.6;
  cursor: default;
}

code {
  font-family: var(--vp-font-family-mono);
  font-size: 0.95rem;
  white-space: nowrap;
  overflow-x: auto;
  display: block;
}

.label {
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.2em;
  color: var(--vp-c-brand);
  margin-bottom: 0.4rem;
}

ul {
  margin: 0;
  padding-left: 1.1rem;
  color: var(--vp-c-text-2);
}

.checklist {
  margin-top: 1rem;
}

.checks {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.check {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.6rem 0.75rem;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.05);
  cursor: pointer;
  text-align: left;
  transition: all 0.2s ease;
}

.check.checked {
  border-color: var(--vp-c-brand);
  background: rgba(65, 209, 255, 0.08);
}

.box {
  width: 20px;
  height: 20px;
  border-radius: 6px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 0.9rem;
}

.check.checked .box {
  border-color: var(--vp-c-brand);
  background: var(--vp-c-brand);
  color: #0f1117;
  font-weight: 700;
}

.progress {
  margin-top: 0.6rem;
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.bar {
  flex: 1;
  height: 6px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.08);
  overflow: hidden;
}

.fill {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, var(--vp-c-brand), #41d1ff);
  transition: width 0.2s ease;
}

.percent {
  font-size: 0.85rem;
  color: var(--vp-c-text-2);
}

.next-hint {
  margin-top: 1.2rem;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  padding-top: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.status {
  display: flex;
  gap: 0.75rem;
  align-items: center;
  flex-wrap: wrap;
}

.complete-btn {
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 10px;
  padding: 0.5rem 0.9rem;
  background: transparent;
  cursor: pointer;
  transition: all 0.2s ease;
}

.complete-btn:hover {
  border-color: var(--vp-c-brand);
  color: var(--vp-c-brand);
}

.status-pill {
  padding: 0.35rem 0.8rem;
  border-radius: 999px;
  background: rgba(65, 209, 255, 0.15);
  color: var(--vp-c-brand);
  font-weight: 600;
}

@media (max-width: 600px) {
  .qs-step {
    padding: 1.25rem;
  }

  .cmd-block code {
    white-space: normal;
  }
}
</style>
