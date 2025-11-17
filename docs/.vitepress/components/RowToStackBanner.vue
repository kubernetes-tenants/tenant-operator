<template>
  <div ref="bannerRef" class="feature-card row-stack-banner">
    <div class="feature__visualization">
      <!-- Database Table (Left) -->
      <div
        class="database-table"
        :class="{
          'show': animationState >= AnimationState.TABLE_APPEAR,
          'slide-left': animationState >= AnimationState.STACK_APPEAR
        }"
      >
        <div class="table-header">
          <div class="header-cell">uid</div>
          <div class="header-cell">domain</div>
          <div class="header-cell">plan</div>
        </div>
        <div class="table-body">
          <div
            v-for="(row, index) in rows"
            :key="row.uid"
            class="table-row"
            :class="{
              'show': animationState >= AnimationState.TABLE_APPEAR,
              'highlight': currentFocusRow === index && animationState >= AnimationState.NAMESPACE_FOCUS
            }"
            :style="{ animationDelay: `${0.3 + index * 0.2}s` }"
          >
            <div class="table-cell">{{ row.uid }}</div>
            <div class="table-cell">{{ row.domain }}</div>
            <div class="table-cell plan-cell">
              <span
                class="plan-badge"
                :class="row.plan"
                :key="`${row.uid}-${row.plan}`"
              >
                {{ row.plan }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- K8s Stack Browser (Right) -->
      <div
        class="k8s-browser"
        :class="{
          'show': animationState >= AnimationState.STACK_APPEAR
        }"
      >
        <div class="browser-header">
          <div class="browser-title">Kubernetes Cluster</div>
        </div>
        <div class="browser-content">
          <div
            v-for="(ns, index) in namespaces"
            :key="ns.uid"
            class="namespace-card"
            :class="{
              'show': animationState >= AnimationState.STACK_APPEAR,
              'focused': currentFocusRow === index && animationState >= AnimationState.NAMESPACE_FOCUS
            }"
            :style="{
              animationDelay: `${0.6 + index * 0.2}s`,
              '--card-offset': `${getCardOffset(index)}px`
            }"
          >
            <div class="namespace-header">
              <span class="namespace-icon">📦</span>
              <span class="namespace-name">{{ ns.uid }}</span>
            </div>
            <div class="namespace-resources">
              <div
                v-for="resource in ns.resources"
                :key="resource.name"
                class="resource-item"
                :class="{
                  'removing': resource.removing,
                  'updating': resource.updating
                }"
              >
                <span class="resource-icon">{{ resource.icon }}</span>
                <span class="resource-name">{{ resource.name }}</span>
                <span v-if="resource.detail" class="resource-detail">{{ resource.detail }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="feature__content">
      <h3 class="feature__title">Your Database, Your Infrastructure</h3>
      <p class="feature__description">
        Change data, Lynq changes infrastructure. Zero downtime, zero manual work.
      </p>
      <a href="/advanced-use-cases" class="use-cases-link">Explore Use Cases →</a>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue';

// Template ref
const bannerRef = ref(null);

// Animation states
const AnimationState = {
  INITIAL: 0,
  TABLE_APPEAR: 1,
  STACK_APPEAR: 2,
  NAMESPACE_FOCUS: 3,
  SCENARIO_PLAY: 4,
};

const animationState = ref(AnimationState.INITIAL);
const currentFocusRow = ref(null);
const hasStarted = ref(false);

// Initial data
const initialRows = [
  { uid: 'customer-1', domain: 'a.acme.com', plan: 'basic' },
  { uid: 'customer-2', domain: 'b.acme.com', plan: 'basic' },
  { uid: 'customer-3', domain: 'b.acme.com', plan: 'basic' },
];

const initialNamespaces = [
  {
    uid: 'customer-1',
    resources: [
      { name: 'Ingress', icon: '🌐', detail: 'a.acme.com', removing: false, updating: false },
      { name: 'Service', icon: '🔌', detail: '', removing: false, updating: false },
      { name: 'Deployment', icon: '📦', detail: 'replicas: 1', removing: false, updating: false },
    ]
  },
  {
    uid: 'customer-2',
    resources: [
      { name: 'Ingress', icon: '🌐', detail: 'b.acme.com', removing: false, updating: false },
      { name: 'Service', icon: '🔌', detail: '', removing: false, updating: false },
      { name: 'Deployment', icon: '📦', detail: 'replicas: 1', removing: false, updating: false },
    ]
  },
  {
    uid: 'customer-3',
    resources: [
      { name: 'Ingress', icon: '🌐', detail: 'b.acme.com', removing: false, updating: false },
      { name: 'Service', icon: '🔌', detail: '', removing: false, updating: false },
      { name: 'Deployment', icon: '📦', detail: 'replicas: 1', removing: false, updating: false },
    ]
  },
];

const rows = ref(JSON.parse(JSON.stringify(initialRows)));
const namespaces = ref(JSON.parse(JSON.stringify(initialNamespaces)));

// Scenario definitions for each customer
const scenarios = [
  // === FORWARD SCENARIOS (Change) ===
  // Scenario 0: customer-1 premium upgrade
  {
    customerIndex: 0,
    type: 'premium',
    execute: async () => {
      // Update plan to premium
      rows.value[0].plan = 'premium';

      // Wait a bit for visual effect
      await sleep(800);

      // Update deployment replicas
      const deployment = namespaces.value[0].resources.find(r => r.name === 'Deployment');
      if (deployment) {
        deployment.updating = true;
        await sleep(1000);
        deployment.detail = 'replicas: 3';
        deployment.updating = false;
      }
    }
  },
  // Scenario 1: customer-2 domain change (subdomain → custom domain)
  {
    customerIndex: 1,
    type: 'domain',
    execute: async () => {
      // Update domain to custom domain
      const customDomain = 'custom.example.com';
      rows.value[1].domain = customDomain;

      await sleep(800);

      // Update ingress
      const ingress = namespaces.value[1].resources.find(r => r.name === 'Ingress');
      if (ingress) {
        ingress.updating = true;
        await sleep(1000);
        ingress.detail = customDomain;
        ingress.updating = false;
      }
    }
  },
  // Scenario 2: customer-3 deletion
  {
    customerIndex: 2,
    type: 'delete',
    execute: async () => {
      // Update plan to deleted
      rows.value[2].plan = 'deleted';

      await sleep(800);

      // Remove resources one by one
      const ns = namespaces.value[2];
      for (let i = ns.resources.length - 1; i >= 0; i--) {
        ns.resources[i].removing = true;
        await sleep(600);
      }

      // Clear all resources
      await sleep(400);
      ns.resources = [];
    }
  },
  // === REVERT SCENARIOS (Restore to original) ===
  // Scenario 3: customer-1 revert to basic
  {
    customerIndex: 0,
    type: 'revert-premium',
    execute: async () => {
      // Revert plan to basic
      rows.value[0].plan = 'basic';

      await sleep(800);

      // Revert deployment replicas
      const deployment = namespaces.value[0].resources.find(r => r.name === 'Deployment');
      if (deployment) {
        deployment.updating = true;
        await sleep(1000);
        deployment.detail = 'replicas: 1';
        deployment.updating = false;
      }
    }
  },
  // Scenario 4: customer-2 revert domain (custom domain → subdomain)
  {
    customerIndex: 1,
    type: 'revert-domain',
    execute: async () => {
      // Revert domain to original subdomain
      const originalDomain = 'b.acme.com';
      rows.value[1].domain = originalDomain;

      await sleep(800);

      // Revert ingress
      const ingress = namespaces.value[1].resources.find(r => r.name === 'Ingress');
      if (ingress) {
        ingress.updating = true;
        await sleep(1000);
        ingress.detail = originalDomain;
        ingress.updating = false;
      }
    }
  },
  // Scenario 5: customer-3 restore resources
  {
    customerIndex: 2,
    type: 'restore',
    execute: async () => {
      // Revert plan to basic
      rows.value[2].plan = 'basic';

      await sleep(800);

      // Restore resources one by one in original order (Ingress → Service → Deployment)
      const ns = namespaces.value[2];
      const resourcesToRestore = [
        { name: 'Ingress', icon: '🌐', detail: 'b.acme.com', removing: false, updating: false },
        { name: 'Service', icon: '🔌', detail: '', removing: false, updating: false },
        { name: 'Deployment', icon: '📦', detail: 'replicas: 1', removing: false, updating: false },
      ];

      for (const resource of resourcesToRestore) {
        ns.resources.push({ ...resource, updating: true });
        await sleep(600);
        // Remove updating flag
        ns.resources[ns.resources.length - 1].updating = false;
      }
    }
  },
];

// Helper function for async sleep
const sleep = (ms) => new Promise(resolve => setTimeout(resolve, ms));

// Reset data to initial state
const resetData = () => {
  rows.value = JSON.parse(JSON.stringify(initialRows));
  namespaces.value = JSON.parse(JSON.stringify(initialNamespaces));
};

// Calculate card offset for carousel effect
const getCardOffset = (index) => {
  const totalCards = 3;
  // Default to customer-1 (index 0) as focused for initial state
  const focused = currentFocusRow.value === null ? 0 : currentFocusRow.value;

  // Current card is centered
  if (index === focused) return 0;

  // Calculate relative position
  // For focused=0: index 1 should be +220 (below), index 2 should be -220 (above)
  // For focused=1: index 2 should be +220 (below), index 0 should be -220 (above)
  // For focused=2: index 0 should be +220 (below), index 1 should be -220 (above)
  let relativePos = index - focused;

  // Normalize to circular range [-1, 0, 1]
  // If relativePos > 1, wrap around (e.g., 2 becomes -1)
  // If relativePos < -1, wrap around (e.g., -2 becomes 1)
  if (relativePos > 1) relativePos -= totalCards;
  if (relativePos < -1) relativePos += totalCards;

  // relativePos: -1 = above (previous), 0 = center (focused), 1 = below (next)
  return relativePos * 220;
};

let animationTimer = null;

const playAnimation = () => {
  let step = 0;
  let scenarioIndex = 0;

  const nextStep = async () => {
    switch (step) {
      case 0:
        // Step 1: Table appears
        animationState.value = AnimationState.TABLE_APPEAR;
        animationTimer = setTimeout(nextStep, 1500);
        break;
      case 1:
        // Step 2: Stack appears
        animationState.value = AnimationState.STACK_APPEAR;
        animationTimer = setTimeout(nextStep, 2000);
        break;
      case 2:
        // Step 3: Start cycling - Focus on customer based on scenario
        const scenario = scenarios[scenarioIndex % scenarios.length];
        currentFocusRow.value = scenario.customerIndex;
        animationState.value = AnimationState.NAMESPACE_FOCUS;
        animationTimer = setTimeout(nextStep, 2000);
        break;
      case 3:
        // Step 4: Play scenario for focused customer
        animationState.value = AnimationState.SCENARIO_PLAY;
        const currentScenario = scenarios[scenarioIndex % scenarios.length];
        await currentScenario.execute();

        // Wait a bit after scenario completes
        await sleep(1500);

        scenarioIndex++;

        // If we've completed all 6 scenarios (3 changes + 3 reverts), reset and start over
        if (scenarioIndex % scenarios.length === 0) {
          // Extra pause before restarting
          await sleep(1000);
          resetData();
          await sleep(500);
        }

        // Go back to step 2 (focus next customer)
        step = 1; // Will increment to 2
        animationTimer = setTimeout(nextStep, 500);
        break;
    }
    step++;
  };

  nextStep();
};

let observer = null;

onMounted(() => {
  // Set up Intersection Observer to start animation when component is near center of viewport
  observer = new IntersectionObserver(
    (entries) => {
      entries.forEach((entry) => {
        // Start animation when component intersects with viewport (50% threshold)
        if (entry.isIntersecting && !hasStarted.value) {
          hasStarted.value = true;
          playAnimation();
          // Disconnect observer after starting (only run once)
          if (observer) {
            observer.disconnect();
          }
        }
      });
    },
    {
      threshold: 0.5, // Trigger when 50% of component is visible
      rootMargin: '0px', // No margin adjustment
    }
  );

  // Start observing the banner element
  if (bannerRef.value) {
    observer.observe(bannerRef.value);
  }
});

onUnmounted(() => {
  if (animationTimer) {
    clearTimeout(animationTimer);
  }
  if (observer) {
    observer.disconnect();
  }
});
</script>

<style scoped>
.row-stack-banner {
  padding: 3rem 2rem;
  background: linear-gradient(135deg, var(--vp-c-bg) 0%, var(--vp-c-bg-soft) 100%);
  border-radius: 12px;
  border: 1px solid var(--vp-c-divider);
}

.feature__visualization {
  position: relative;
  width: 100%;
  min-height: 480px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 2rem;
  margin-bottom: 2rem;
}

/* Database Table */
.database-table {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  width: 320px;
  background: var(--vp-c-bg);
  border: 1px solid var(--vp-c-divider);
  border-radius: 8px;
  opacity: 0;
  transition: all 0.8s cubic-bezier(0.4, 0, 0.2, 1);
  z-index: 10;
}

.database-table.show {
  opacity: 1;
}

.database-table.slide-left {
  left: 10%;
  transform: translateX(0);
}

.table-header {
  display: grid;
  grid-template-columns: 1fr 1fr 0.8fr;
  gap: 0.5rem;
  padding: 0.75rem 0.5rem;
  background: var(--vp-c-bg-soft);
  border-bottom: 1px solid var(--vp-c-divider);
  border-radius: 8px 8px 0 0;
  align-items: center;
}

.header-cell {
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  color: var(--vp-c-text-2);
  letter-spacing: 0.05em;
}

.table-body {
  padding: 0.75rem 0.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.table-row {
  display: grid;
  grid-template-columns: 1fr 1fr 0.8fr;
  gap: 0.5rem;
  padding: 0.65rem 0.5rem;
  border-radius: 4px;
  transition: all 0.3s ease;
  opacity: 0;
  transform: translateY(10px);
  align-items: center;
}

.table-row.show {
  animation: rowInsert 0.5s ease-out forwards;
}

@keyframes rowInsert {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.table-row.highlight {
  background: linear-gradient(90deg, rgba(102, 126, 234, 0.15), transparent);
  box-shadow: 0 0 0 2px rgba(102, 126, 234, 0.3);
}

.table-cell {
  font-size: 0.8rem;
  color: var(--vp-c-text-1);
  font-family: 'SF Mono', Monaco, monospace;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.plan-cell {
  display: flex;
  align-items: center;
}

.plan-badge {
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  transition: all 0.3s ease;
}

.plan-badge.basic {
  background: rgba(100, 116, 139, 0.15);
  color: #64748b;
  border: 1px solid rgba(100, 116, 139, 0.3);
}

.plan-badge.premium {
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.2), rgba(5, 150, 105, 0.2));
  color: #10b981;
  border: 1px solid rgba(16, 185, 129, 0.4);
  box-shadow: 0 0 12px rgba(16, 185, 129, 0.3);
  animation: premiumGlow 1s ease-out;
}

@keyframes premiumGlow {
  0% {
    box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.7);
  }
  100% {
    box-shadow: 0 0 12px rgba(16, 185, 129, 0.3);
  }
}

.plan-badge.deleted {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

/* K8s Stack Browser */
.k8s-browser {
  position: absolute;
  right: 5%;
  width: 420px;
  background: linear-gradient(135deg, rgba(30, 30, 40, 0.95), rgba(20, 20, 30, 0.95));
  border: 1px solid rgba(100, 108, 255, 0.3);
  border-radius: 8px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
  opacity: 0;
  transform: translateX(50px);
  transition: all 0.8s cubic-bezier(0.4, 0, 0.2, 1);
  z-index: 8;
}

.k8s-browser.show {
  opacity: 1;
  transform: translateX(0);
}

.browser-header {
  padding: 0.75rem 1rem;
  background: rgba(100, 108, 255, 0.1);
  border-bottom: 1px solid rgba(100, 108, 255, 0.2);
  border-radius: 8px 8px 0 0;
}

.browser-title {
  font-size: 0.85rem;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.9);
}

.browser-content {
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  height: 400px;
  min-height: 400px;
  overflow: hidden;
  position: relative;
}

/* Namespace Cards */
.namespace-card {
  position: absolute;
  width: calc(100% - 2rem);
  background: rgba(40, 40, 50, 0.6);
  border: 1px solid rgba(100, 108, 255, 0.2);
  border-radius: 6px;
  padding: 0.65rem;
  opacity: 0;
  left: 1rem;
  top: 50%;
  /* Always use CSS variable for positioning */
  transform: translateY(calc(-50% + var(--card-offset, 0px)));
  transition: transform 0.8s cubic-bezier(0.4, 0, 0.2, 1),
              opacity 0.6s ease,
              border 0.4s ease,
              box-shadow 0.4s ease,
              background 0.4s ease,
              scale 0.4s ease;
}

/* Show animation - fade in with staggered delay */
.namespace-card.show {
  animation: fadeIn 0.6s ease-out forwards;
}

.namespace-card.show:nth-child(1) {
  animation-delay: 0.6s;
}

.namespace-card.show:nth-child(2) {
  animation-delay: 0.75s;
}

.namespace-card.show:nth-child(3) {
  animation-delay: 0.9s;
}

@keyframes fadeIn {
  to {
    opacity: 1;
  }
}

/* Focused state - center card with scale effect */
.namespace-card.focused {
  scale: 1.08;
  border: 2px solid rgba(100, 108, 255, 0.6);
  box-shadow: 0 12px 40px rgba(100, 108, 255, 0.4);
  background: rgba(40, 40, 50, 0.98);
  z-index: 10;
}

.namespace-card.focused .namespace-header {
  border-bottom-color: rgba(100, 108, 255, 0.3);
}

.namespace-card.focused .namespace-icon {
  font-size: 1.3rem;
}

.namespace-card.focused .namespace-name {
  font-size: 1rem;
}

.namespace-card.focused .resource-item {
  padding: 0.7rem 0.6rem;
  font-size: 0.85rem;
  background: rgba(60, 60, 70, 0.7);
}

.namespace-card.focused .resource-icon {
  font-size: 1.1rem;
}

.namespace-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid rgba(100, 108, 255, 0.15);
}

.namespace-icon {
  font-size: 1rem;
}

.namespace-name {
  font-size: 0.85rem;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.9);
  font-family: 'SF Mono', Monaco, monospace;
}

.namespace-resources {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  min-height: 151px; /* Maintain height even when resources are removed */
}

.resource-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.4rem;
  background: rgba(60, 60, 70, 0.4);
  border-radius: 4px;
  font-size: 0.72rem;
  color: rgba(255, 255, 255, 0.7);
  transition: all 0.3s ease;
}

.resource-item.removing {
  animation: fadeOutRemove 0.8s ease-out forwards;
}

@keyframes fadeOutRemove {
  0% {
    opacity: 1;
    transform: translateX(0);
  }
  100% {
    opacity: 0;
    transform: translateX(20px);
  }
}

.resource-icon {
  font-size: 0.9rem;
}

.resource-name {
  flex: 1;
  font-weight: 500;
}

.resource-detail {
  font-size: 0.7rem;
  color: rgba(100, 200, 255, 0.9);
  font-family: 'SF Mono', Monaco, monospace;
}

/* Resource updating animation */
.resource-item.updating {
  animation: resourceUpdate 1s ease-in-out;
  border-color: rgba(16, 185, 129, 0.5);
  box-shadow: 0 0 15px rgba(16, 185, 129, 0.3);
}

@keyframes resourceUpdate {
  0%, 100% {
    background: rgba(60, 60, 70, 0.4);
  }
  50% {
    background: rgba(16, 185, 129, 0.25);
  }
}

/* Feature Content */
.feature__content {
  text-align: center;
  max-width: 600px;
  margin: 50px auto 0;
}

.feature__title {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--vp-c-text-1);
  margin-bottom: 0.75rem;
}

.feature__description {
  font-size: 1rem;
  line-height: 1.6;
  color: var(--vp-c-text-2);
  margin-bottom: 1rem;
}

.use-cases-link {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--vp-c-brand-1);
  text-decoration: none;
  transition: all 0.2s ease;
}

.use-cases-link:hover {
  color: var(--vp-c-brand-2);
  gap: 0.5rem;
}

/* Responsive */
@media (max-width: 968px) {
  .feature__visualization {
    min-height: 350px;
  }

  .database-table {
    width: 280px;
  }

  .k8s-browser {
    width: 340px;
  }

  .browser-content {
    height: 280px;
    min-height: 280px;
  }
}

@media (max-width: 768px) {
  .row-stack-banner {
    padding: 2rem 1rem;
  }

  .feature__visualization {
    min-height: 300px;
  }

  .database-table {
    width: 240px;
    font-size: 0.85rem;
  }

  .k8s-browser {
    width: 300px;
  }

  .browser-content {
    height: 260px;
    min-height: 260px;
  }
}
</style>
