<template>
  <div class="dependency-animation-container">
    <div class="animation-title">Parallel Execution</div>
    <svg ref="svgRef" viewBox="0 0 400 280" class="dependency-svg">
      <!-- Background -->
      <defs>
        <marker id="arrowhead-parallel-1" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto">
          <polygon points="0 0, 10 3, 0 6" fill="var(--vp-c-text-3)" class="arrow-marker arrow-marker-1"/>
        </marker>
        <marker id="arrowhead-parallel-2" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto">
          <polygon points="0 0, 10 3, 0 6" fill="var(--vp-c-text-3)" class="arrow-marker arrow-marker-2"/>
        </marker>
      </defs>

      <!-- Edges (arrows) - drawn first so nodes appear on top -->
      <g class="edges">
        <!-- secret → app-a -->
        <path
          class="edge edge-1"
          d="M 185 78 L 115 162"
          stroke="var(--vp-c-text-3)"
          stroke-width="2"
          fill="none"
          marker-end="url(#arrowhead-parallel-1)"
        />
        <!-- secret → app-b -->
        <path
          class="edge edge-2"
          d="M 215 78 L 285 162"
          stroke="var(--vp-c-text-3)"
          stroke-width="2"
          fill="none"
          marker-end="url(#arrowhead-parallel-2)"
        />
      </g>

      <!-- Nodes -->
      <g class="nodes">
        <!-- Secret node -->
        <g class="node node-secret" transform="translate(200, 60)">
          <rect x="-50" y="-20" width="100" height="40" rx="8" class="node-bg"/>
          <text x="0" y="-2" text-anchor="middle" dominant-baseline="middle" class="node-label">
            secret
          </text>
          <text x="0" y="10" text-anchor="middle" dominant-baseline="middle" class="node-type">
            Secret
          </text>
        </g>

        <!-- App-a node -->
        <g class="node node-app-a" transform="translate(100, 180)">
          <rect x="-50" y="-20" width="100" height="40" rx="8" class="node-bg"/>
          <text x="0" y="-2" text-anchor="middle" dominant-baseline="middle" class="node-label">
            app-a
          </text>
          <text x="0" y="10" text-anchor="middle" dominant-baseline="middle" class="node-type">
            Deployment
          </text>
        </g>

        <!-- App-b node -->
        <g class="node node-app-b" transform="translate(300, 180)">
          <rect x="-50" y="-20" width="100" height="40" rx="8" class="node-bg"/>
          <text x="0" y="-2" text-anchor="middle" dominant-baseline="middle" class="node-label">
            app-b
          </text>
          <text x="0" y="10" text-anchor="middle" dominant-baseline="middle" class="node-type">
            Deployment
          </text>
        </g>
      </g>
    </svg>
    <div class="animation-description">
      Both <code>app-a</code> and <code>app-b</code> execute in parallel after <code>secret</code> is ready
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
  font-size: 13px;
  font-weight: 600;
  fill: var(--vp-c-text-1);
  transform: translateY(-4px);
}

.node-type {
  font-size: 10px;
  fill: var(--vp-c-text-3);
  transform: translateY(-4px);
}

/* Edge styles */
.edge {
  stroke-dasharray: 200;
  stroke-dashoffset: 200;
  opacity: 0.3;
}

.edge-1,
.edge-2 {
  animation: drawEdge 8s ease-in-out infinite;
}

@keyframes drawEdge {
  0%, 25% {
    stroke-dashoffset: 200;
    opacity: 0.3;
  }
  37.5% {
    stroke-dashoffset: 0;
    opacity: 1;
  }
  100% {
    stroke-dashoffset: 0;
    opacity: 1;
  }
}

/* Node background animations */
.node-secret .node-bg {
  animation: nodeActivateSecret 8s ease-in-out infinite;
}

.node-app-a .node-bg,
.node-app-b .node-bg {
  animation: nodeActivateApps 8s ease-in-out infinite;
}

/* Secret node: 0-25% */
@keyframes nodeActivateSecret {
  0% {
    stroke: var(--vp-c-divider);
    fill: var(--vp-c-bg);
  }
  12.5% {
    stroke: #3b82f6;
    fill: rgba(59, 130, 246, 0.1);
  }
  25%, 100% {
    stroke: #10b981;
    fill: rgba(16, 185, 129, 0.1);
  }
}

/* App nodes: 37.5-50% (parallel execution) */
@keyframes nodeActivateApps {
  0%, 37.5% {
    stroke: var(--vp-c-divider);
    fill: var(--vp-c-bg);
  }
  43.75% {
    stroke: #3b82f6;
    fill: rgba(59, 130, 246, 0.1);
  }
  50%, 100% {
    stroke: #10b981;
    fill: rgba(16, 185, 129, 0.1);
  }
}

/* Arrow marker animation */
.arrow-marker {
  opacity: 0;
}

.arrow-marker-1,
.arrow-marker-2 {
  animation: arrowAppear 8s ease-in-out infinite;
}

@keyframes arrowAppear {
  0%, 25% {
    opacity: 0;
  }
  37.5%, 100% {
    opacity: 1;
  }
}
</style>
