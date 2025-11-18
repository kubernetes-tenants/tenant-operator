<template>
  <div class="dependency-animation-container">
    <div class="animation-title">Cycle Detection</div>
    <svg ref="svgRef" viewBox="0 0 400 280" class="dependency-svg">
      <!-- Background -->
      <defs>
        <marker id="arrowhead-cycle-ab" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto">
          <polygon points="0 0, 10 3, 0 6" class="arrow-marker arrow-marker-ab"/>
        </marker>
        <marker id="arrowhead-cycle-ba" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto">
          <polygon points="0 0, 10 3, 0 6" class="arrow-marker arrow-marker-ba"/>
        </marker>
      </defs>

      <!-- Edges (arrows) - drawn first so nodes appear on top -->
      <g class="edges">
        <!-- a → b edge (upper arc) -->
        <path
          class="edge edge-a-b"
          d="M 148 138 Q 200 100 252 138"
          stroke="var(--vp-c-text-3)"
          stroke-width="2"
          fill="none"
          marker-end="url(#arrowhead-cycle-ab)"
        />
        <!-- b → a edge (lower arc, creating cycle) -->
        <path
          class="edge edge-b-a"
          d="M 252 162 Q 200 200 148 162"
          stroke="var(--vp-c-text-3)"
          stroke-width="2"
          fill="none"
          marker-end="url(#arrowhead-cycle-ba)"
        />
      </g>

      <!-- Nodes -->
      <g class="nodes">
        <!-- Node a -->
        <g transform="translate(120, 150)">
          <g class="node node-a">
            <circle r="30" class="node-circle"/>
            <text x="0" y="0" text-anchor="middle" dominant-baseline="middle" class="node-label">
              a
            </text>
          </g>
        </g>

        <!-- Node b -->
        <g transform="translate(280, 150)">
          <g class="node node-b">
            <circle r="30" class="node-circle"/>
            <text x="0" y="0" text-anchor="middle" dominant-baseline="middle" class="node-label">
              b
            </text>
          </g>
        </g>
      </g>

      <!-- Error icon (centered between nodes) -->
      <g transform="translate(200, 150)">
        <g class="error-icon">
          <circle r="18" fill="#ef4444" opacity="0.15"/>
          <path
            d="M 0 -8 L 0 2 M 0 6 L 0 8"
            stroke="#ef4444"
            stroke-width="2.5"
            stroke-linecap="round"
          />
        </g>
      </g>

      <!-- Cycle indicator -->
      <g class="cycle-indicator" transform="translate(200, 235)">
        <text x="0" y="0" text-anchor="middle" class="cycle-text">
          ❌ Cycle detected: a → b → a
        </text>
      </g>
    </svg>
    <div class="animation-description">
      Circular dependencies are detected and rejected during reconciliation
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

/* Node styles */
.node-circle {
  fill: var(--vp-c-bg);
  stroke: var(--vp-c-divider);
  stroke-width: 2;
  transition: all 0.3s ease;
}

.node-label {
  font-size: 16px;
  font-weight: 600;
  fill: var(--vp-c-text-1);
}

/* Edge styles */
.edge {
  stroke-dasharray: 200;
  stroke-dashoffset: 200;
  opacity: 0.5;
}

.edge-a-b {
  animation: drawEdgeCycle 10s ease-in-out infinite;
}

.edge-b-a {
  animation: drawEdgeCycle 10s ease-in-out infinite;
  animation-delay: 1s;
}

@keyframes drawEdgeCycle {
  0%, 10% {
    stroke-dashoffset: 200;
    opacity: 0.3;
    stroke: var(--vp-c-text-3);
  }
  20% {
    stroke-dashoffset: 0;
    opacity: 0.8;
    stroke: var(--vp-c-text-3);
  }
  30%, 40% {
    stroke-dashoffset: 0;
    opacity: 1;
    stroke: #ef4444;
  }
  100% {
    stroke-dashoffset: 0;
    opacity: 1;
    stroke: #ef4444;
  }
}

/* Node animations */
.node-a {
  animation: nodeShake 10s ease-in-out infinite;
}

.node-b {
  animation: nodeShake 10s ease-in-out infinite;
  animation-delay: 1s;
}

.node-a .node-circle,
.node-b .node-circle {
  animation: nodeCycleError 10s ease-in-out infinite;
}

.node-b .node-circle {
  animation-delay: 1s;
}

@keyframes nodeCycleError {
  0%, 20% {
    stroke: var(--vp-c-divider);
    fill: var(--vp-c-bg);
  }
  30%, 40% {
    stroke: #ef4444;
    fill: rgba(239, 68, 68, 0.1);
  }
  100% {
    stroke: #ef4444;
    fill: rgba(239, 68, 68, 0.1);
  }
}

@keyframes nodeShake {
  0%, 40% {
    transform: translate(0, 0);
  }
  41% {
    transform: translate(-3px, 0);
  }
  42% {
    transform: translate(3px, 0);
  }
  43% {
    transform: translate(-3px, 0);
  }
  44% {
    transform: translate(3px, 0);
  }
  45%, 100% {
    transform: translate(0, 0);
  }
}

/* Error icon animation */
.error-icon {
  opacity: 0;
  animation: errorAppear 10s ease-in-out infinite;
}

@keyframes errorAppear {
  0%, 30% {
    opacity: 0;
    transform: scale(0);
  }
  35% {
    opacity: 1;
    transform: scale(1.2);
  }
  40% {
    opacity: 1;
    transform: scale(1);
  }
  100% {
    opacity: 1;
    transform: scale(1);
  }
}

/* Cycle indicator animation */
.cycle-text {
  font-size: 12px;
  font-weight: 600;
  fill: #ef4444;
  opacity: 0;
  animation: cycleTextAppear 10s ease-in-out infinite;
}

@keyframes cycleTextAppear {
  0%, 35% {
    opacity: 0;
    transform: translateY(10px);
  }
  40% {
    opacity: 1;
    transform: translateY(0);
  }
  100% {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Arrow marker animation */
.arrow-marker {
  opacity: 0;
}

.arrow-marker-ab {
  animation: arrowCycleAppear 10s ease-in-out infinite;
}

.arrow-marker-ba {
  animation: arrowCycleAppear 10s ease-in-out infinite;
  animation-delay: 1s;
}

@keyframes arrowCycleAppear {
  0%, 10% {
    opacity: 0;
    fill: var(--vp-c-text-3);
  }
  20% {
    opacity: 1;
    fill: var(--vp-c-text-3);
  }
  30%, 100% {
    opacity: 1;
    fill: #ef4444;
  }
}
</style>
