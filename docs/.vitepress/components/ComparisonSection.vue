<template>
  <div class="comparison-section" ref="sectionRef">
    <p class="section-intro" :class="{ visible: isVisible }">
      Managing hundreds or thousands of tenants in Kubernetes shouldn't require custom scripts, manual updates, or rebuilding Helm charts every time your data changes.
    </p>

    <div class="comparison-container">
      <div
        v-for="(pair, index) in comparisonPairs"
        :key="`pair-${index}`"
        :ref="el => pairRefs[index] = el"
        class="comparison-pair"
        :class="{ visible: isVisible, highlighted: (hoveredIndex !== null ? hoveredIndex : activeIndex) === index }"
        :style="{ animationDelay: `${0.2 + index * 0.2}s` }"
        @mouseenter="hoveredIndex = index"
        @mouseleave="hoveredIndex = null"
      >
        <!-- Problem Card -->
        <div class="comparison-card problem-card">
          <div class="card-label problem-label">❌ Problem</div>
          <strong class="card-title">{{ pair.problem.title }}</strong>
          <p class="card-description">{{ pair.problem.description }}</p>
        </div>

        <!-- Arrow Connector -->
        <div class="arrow-connector" :class="{ visible: isVisible }" :style="{ animationDelay: `${0.4 + index * 0.2}s` }">
          <svg width="60" height="40" viewBox="0 0 60 40" fill="none">
            <path d="M5 20 L45 20" stroke="currentColor" stroke-width="2" stroke-dasharray="4 4" class="arrow-line"/>
            <path d="M40 15 L50 20 L40 25" stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round" class="arrow-head"/>
          </svg>
          <span class="arrow-label">becomes</span>
        </div>

        <!-- Solution Card -->
        <div class="comparison-card solution-card">
          <div class="card-label solution-label">✅ Solution</div>
          <strong class="card-title">{{ pair.solution.title }}</strong>
          <p class="card-description">{{ pair.solution.description }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue';

const sectionRef = ref(null);
const isVisible = ref(false);
const hoveredIndex = ref(null);
const activeIndex = ref(null);
const pairRefs = ref([]);

const comparisonPairs = [
  {
    problem: {
      title: 'Helm Charts',
      description: 'Static values files. Adding 10 new tenants? Manually create 10 new releases and track them separately.'
    },
    solution: {
      title: 'Database-Driven Automation',
      description: 'Your existing database is the source of truth. No YAML commits, no manual kubectl commands.'
    }
  },
  {
    problem: {
      title: 'GitOps Only',
      description: 'New customer signs up? Commit YAML, wait for CI/CD, manually sync. Not dynamic enough for real-time provisioning.'
    },
    solution: {
      title: 'Real-Time Synchronization',
      description: 'Add a tenant → resources created in 30 seconds. Deactivate a tenant → everything cleaned up automatically.'
    }
  },
  {
    problem: {
      title: 'Custom Scripts',
      description: 'kubectl apply in bash loops. Works until you need drift detection, conflict handling, or dependency ordering.'
    },
    solution: {
      title: 'Production-Grade Control',
      description: 'Built-in policies, drift detection, conflict resolution, dependency management, and comprehensive observability.'
    }
  }
];

let observer = null;

const updateActiveIndex = () => {
  // Only update if not hovering
  if (hoveredIndex.value !== null) return;

  const viewportCenter = window.innerHeight / 2;
  let closestIndex = null;
  let closestDistance = Infinity;

  pairRefs.value.forEach((el, index) => {
    if (!el) return;

    const rect = el.getBoundingClientRect();
    const elementCenter = rect.top + rect.height / 2;
    const distance = Math.abs(elementCenter - viewportCenter);

    // Only consider elements that are at least partially visible
    if (rect.top < window.innerHeight && rect.bottom > 0) {
      if (distance < closestDistance) {
        closestDistance = distance;
        closestIndex = index;
      }
    }
  });

  activeIndex.value = closestDistance < window.innerHeight * 0.6 ? closestIndex : null;
};

const handleScroll = () => {
  requestAnimationFrame(updateActiveIndex);
};

onMounted(() => {
  observer = new IntersectionObserver(
    (entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting && !isVisible.value) {
          isVisible.value = true;
          // Initial active index calculation
          setTimeout(updateActiveIndex, 100);
        }
      });
    },
    {
      threshold: 0.2,
      rootMargin: '0px 0px -100px 0px'
    }
  );

  if (sectionRef.value) {
    observer.observe(sectionRef.value);
  }

  // Add scroll listener
  window.addEventListener('scroll', handleScroll, { passive: true });
});

onUnmounted(() => {
  if (observer && sectionRef.value) {
    observer.unobserve(sectionRef.value);
  }

  // Remove scroll listener
  window.removeEventListener('scroll', handleScroll);
});
</script>

<style scoped>
.comparison-section {
  margin: 2.5rem 0;
}

.section-intro {
  font-size: 1.05rem;
  color: var(--vp-c-text-2);
  text-align: center;
  max-width: 800px;
  margin: 0 auto 3rem;
  line-height: 1.6;
  opacity: 0;
  transform: translateY(20px);
  transition: all 0.8s cubic-bezier(0.4, 0, 0.2, 1);
}

.section-intro.visible {
  opacity: 1;
  transform: translateY(0);
}

.comparison-container {
  display: flex;
  flex-direction: column;
  gap: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.comparison-pair {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  gap: 1.5rem;
  align-items: center;
  opacity: 0;
  transform: translateY(30px);
  transition: all 0.8s cubic-bezier(0.4, 0, 0.2, 1);
  padding: 1rem;
  border-radius: 12px;
  position: relative;
}

.comparison-pair.visible {
  opacity: 1;
  transform: translateY(0);
}

.comparison-pair.highlighted {
  background: var(--vp-c-bg-soft);
  box-shadow: 0 4px 20px rgba(102, 126, 234, 0.15);
}

.comparison-pair.highlighted::before {
  content: '';
  position: absolute;
  inset: -2px;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.3), rgba(118, 75, 162, 0.3));
  border-radius: 12px;
  z-index: -1;
  opacity: 0;
  animation: glowPulse 2s ease-in-out infinite;
}

.comparison-pair.highlighted::before {
  opacity: 1;
}

@keyframes glowPulse {
  0%, 100% {
    opacity: 0.3;
  }
  50% {
    opacity: 0.6;
  }
}

.comparison-card {
  padding: 1.5rem;
  border-radius: 10px;
  border: 1px solid var(--vp-c-divider);
  background: var(--vp-c-bg);
  transition: all 0.3s ease;
  position: relative;
}

.card-label {
  position: absolute;
  top: -10px;
  left: 1rem;
  padding: 0.25rem 0.75rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  background: var(--vp-c-bg);
  border: 1px solid;
}

.problem-label {
  border-color: var(--vp-c-divider);
  color: var(--vp-c-text-2);
}

.solution-label {
  border-color: var(--vp-c-brand-light);
  color: var(--vp-c-brand);
}

.problem-card {
  border-color: var(--vp-c-divider);
}

.solution-card {
  border-color: var(--vp-c-brand-light);
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.1);
}

.comparison-pair.highlighted .problem-card {
  border-color: var(--vp-c-text-3);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.comparison-pair.highlighted .solution-card {
  border-color: var(--vp-c-brand);
  box-shadow: 0 4px 16px rgba(102, 126, 234, 0.25);
}

.card-title {
  display: block;
  margin-bottom: 0.75rem;
  margin-top: 0.5rem;
  font-size: 1.05rem;
  font-weight: 600;
  color: var(--vp-c-text-1);
}

.card-description {
  margin: 0;
  font-size: 0.9rem;
  color: var(--vp-c-text-2);
  line-height: 1.6;
}

/* Arrow Connector */
.arrow-connector {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  opacity: 0;
  transition: all 0.6s cubic-bezier(0.4, 0, 0.2, 1);
}

.arrow-connector.visible {
  opacity: 1;
}

.arrow-connector svg {
  color: var(--vp-c-text-3);
  transition: all 0.3s ease;
}

.comparison-pair.highlighted .arrow-connector svg {
  color: var(--vp-c-brand);
  transform: scale(1.1);
}

.arrow-line {
  animation: arrowFlow 2s ease-in-out infinite;
}

@keyframes arrowFlow {
  0% {
    stroke-dashoffset: 0;
  }
  100% {
    stroke-dashoffset: -8;
  }
}

.arrow-head {
  transition: transform 0.3s ease;
}

.comparison-pair.highlighted .arrow-head {
  animation: arrowBounce 1s ease-in-out infinite;
}

@keyframes arrowBounce {
  0%, 100% {
    transform: translateX(0);
  }
  50% {
    transform: translateX(5px);
  }
}

.arrow-label {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--vp-c-text-3);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  transition: color 0.3s ease;
}

.comparison-pair.highlighted .arrow-label {
  color: var(--vp-c-brand);
}

/* Responsive */
@media (max-width: 968px) {
  .comparison-pair {
    grid-template-columns: 1fr;
    grid-template-rows: auto auto auto;
    gap: 1rem;
  }

  .arrow-connector {
    transform: rotate(90deg);
    margin: 0.5rem 0;
  }

  .arrow-label {
    transform: rotate(-90deg);
    white-space: nowrap;
  }
}

@media (max-width: 768px) {
  .section-intro {
    font-size: 1rem;
    margin-bottom: 2rem;
  }

  .comparison-pair {
    padding: 0.75rem;
  }

  .comparison-card {
    padding: 1.25rem;
  }

  .card-title {
    font-size: 1rem;
  }

  .card-description {
    font-size: 0.875rem;
  }
}
</style>
