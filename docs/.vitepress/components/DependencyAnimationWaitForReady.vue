<template>
  <div class="dependency-animation-container">
    <div class="animation-title">Wait For Ready</div>
    <svg ref="svgRef" viewBox="0 0 400 300" class="dependency-svg">
      <!-- Background -->
      <defs>
        <marker id="arrowhead-ready-db-app" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto">
          <polygon points="0 0, 10 3, 0 6" fill="var(--vp-c-text-3)" class="arrow-marker arrow-marker-db-app"/>
        </marker>
      </defs>

      <!-- Edges (arrows) - drawn first so nodes appear on top -->
      <g class="edges">
        <!-- db → app -->
        <path
          class="edge edge-db-app"
          d="M 200 98 L 200 192"
          stroke="var(--vp-c-text-3)"
          stroke-width="2"
          fill="none"
          marker-end="url(#arrowhead-ready-db-app)"
        />
      </g>

      <!-- Nodes -->
      <g class="nodes">
        <!-- DB node -->
        <g class="node node-db" transform="translate(200, 70)">
          <rect x="-60" y="-25" width="120" height="50" rx="8" class="node-bg"/>
          <text x="0" y="-5" text-anchor="middle" dominant-baseline="middle" class="node-label">
            db
          </text>
          <text x="0" y="8" text-anchor="middle" dominant-baseline="middle" class="node-type">
            Deployment
          </text>

          <!-- Progress bar background -->
          <rect x="-40" y="18" width="80" height="4" rx="2" class="progress-bg"/>
          <!-- Progress bar fill -->
          <rect x="-40" y="18" width="80" height="4" rx="2" class="progress-fill" style="transform-origin: left center;"/>

          <!-- Waiting text -->
          <text x="0" y="32" text-anchor="middle" class="waiting-text">
            Waiting for Ready...
          </text>
        </g>

        <!-- App node -->
        <g class="node node-app" transform="translate(200, 220)">
          <rect x="-60" y="-25" width="120" height="50" rx="8" class="node-bg"/>
          <text x="0" y="-2" text-anchor="middle" dominant-baseline="middle" class="node-label">
            app
          </text>
          <text x="0" y="10" text-anchor="middle" dominant-baseline="middle" class="node-type">
            Deployment
          </text>
        </g>
      </g>

      <!-- Timeout badge -->
      <g class="timeout-badge" transform="translate(275, 70)">
        <rect x="0" y="-10" width="80" height="20" rx="10" class="badge-bg"/>
        <text x="40" y="2" text-anchor="middle" dominant-baseline="middle" class="badge-text">
          timeout: 300s
        </text>
      </g>
    </svg>
    <div class="animation-description">
      <code>waitForReady: true</code> ensures resource is ready before dependent workloads start
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const svgRef = ref(null)

onMounted(() => {
  // Animation starts automatically
})
</script>

<style scoped>
.dependency-animation-container {
  margin: 2rem 0;
  padding: 1.5rem;
  background: var(--vp-c-bg-soft);
  border-radius: 12px;
  border: 1px solid var(--vp-c-divider);
}

.animation-title {
  text-align: center;
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--vp-c-text-2);
  margin-bottom: 1rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.dependency-svg {
  width: 100%;
  height: auto;
  max-width: 400px;
  margin: 0 auto;
  display: block;
}

.animation-description {
  text-align: center;
  font-size: 0.85rem;
  color: var(--vp-c-text-2);
  margin-top: 1rem;
}

.animation-description code {
  background: var(--vp-c-bg);
  padding: 0.2em 0.4em;
  border-radius: 4px;
  font-size: 0.9em;
}

/* Node styles */
.node-bg {
  fill: var(--vp-c-bg);
  stroke: var(--vp-c-divider);
  stroke-width: 2;
  transition: all 0.3s ease;
}

.node-label {
  font-size: 14px;
  font-weight: 600;
  fill: var(--vp-c-text-1);
}

.node-type {
  font-size: 10px;
  fill: var(--vp-c-text-3);
}

/* Progress bar */
.progress-bg {
  fill: var(--vp-c-divider);
  opacity: 0;
}

.progress-fill {
  fill: #3b82f6;
  opacity: 0;
  transform: scaleX(0);
  animation: progressFill 10s ease-in-out infinite;
}

@keyframes progressFill {
  0%, 20% {
    transform: scaleX(0);
  }
  30% {
    transform: scaleX(0.3);
  }
  40%, 50% {
    transform: scaleX(0.8);
  }
  100% {
    transform: scaleX(0.8);
  }
}

/* Waiting text */
.waiting-text {
  font-size: 9px;
  fill: var(--vp-c-text-3);
  opacity: 0;
  animation: waitingTextBlink 10s ease-in-out infinite;
}

@keyframes waitingTextBlink {
  0%, 20% {
    opacity: 0;
  }
  25% {
    opacity: 1;
  }
  30% {
    opacity: 0.5;
  }
  35% {
    opacity: 1;
  }
  40% {
    opacity: 0.5;
  }
  45% {
    opacity: 1;
  }
  50%, 100% {
    opacity: 0;
  }
}

/* Timeout badge */
.timeout-badge {
  opacity: 0;
  animation: badgeFade 10s ease-in-out infinite;
}

.badge-bg {
  fill: rgba(59, 130, 246, 0.15);
  stroke: #3b82f6;
  stroke-width: 1;
}

.badge-text {
  font-size: 8px;
  fill: #3b82f6;
  font-weight: 600;
}

@keyframes badgeFade {
  0%, 20% {
    opacity: 0;
  }
  25% {
    opacity: 1;
  }
  50% {
    opacity: 1;
  }
  55% {
    opacity: 0;
  }
  100% {
    opacity: 0;
  }
}

/* Edge styles */
.edge {
  stroke-dasharray: 200;
  stroke-dashoffset: 200;
  opacity: 0.3;
}

.edge-db-app {
  animation: drawEdgeReady 10s ease-in-out infinite;
}

@keyframes drawEdgeReady {
  0%, 50% {
    stroke-dashoffset: 200;
    opacity: 0.3;
  }
  60% {
    stroke-dashoffset: 0;
    opacity: 1;
  }
  100% {
    stroke-dashoffset: 0;
    opacity: 1;
  }
}

/* Node background animations */
.node-db .node-bg {
  animation: nodeDbActivate 10s ease-in-out infinite;
}

.node-db .progress-bg {
  animation: progressBgShow 10s ease-in-out infinite;
}

.node-db .progress-fill {
  animation: progressFillShow 10s ease-in-out infinite;
}

.node-app .node-bg {
  animation: nodeAppActivate 10s ease-in-out infinite;
}

@keyframes nodeDbActivate {
  0%, 10% {
    stroke: var(--vp-c-divider);
    fill: var(--vp-c-bg);
  }
  20%, 45% {
    stroke: #3b82f6;
    fill: rgba(59, 130, 246, 0.1);
  }
  50%, 100% {
    stroke: #10b981;
    fill: rgba(16, 185, 129, 0.1);
  }
}

@keyframes nodeAppActivate {
  0%, 50% {
    stroke: var(--vp-c-divider);
    fill: var(--vp-c-bg);
  }
  60%, 65% {
    stroke: #3b82f6;
    fill: rgba(59, 130, 246, 0.1);
  }
  70%, 100% {
    stroke: #10b981;
    fill: rgba(16, 185, 129, 0.1);
  }
}

@keyframes progressBgShow {
  0%, 20% {
    opacity: 0;
  }
  25%, 50% {
    opacity: 1;
  }
  55%, 100% {
    opacity: 0;
  }
}

@keyframes progressFillShow {
  0%, 20% {
    opacity: 0;
  }
  25%, 50% {
    opacity: 1;
  }
  55%, 100% {
    opacity: 0;
  }
}

/* Arrow marker animation */
.arrow-marker {
  opacity: 0;
}

.arrow-marker-db-app {
  animation: arrowReadyAppear 10s ease-in-out infinite;
}

@keyframes arrowReadyAppear {
  0%, 50% {
    opacity: 0;
  }
  60% {
    opacity: 1;
  }
  100% {
    opacity: 1;
  }
}
</style>
