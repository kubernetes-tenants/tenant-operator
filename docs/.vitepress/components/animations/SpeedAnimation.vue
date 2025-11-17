<template>
  <div class="animation-container">
    <svg viewBox="0 0 280 200" class="animation-svg" :class="{ 'animate': inView }">
      <g opacity="0.35">
        <text x="10" y="22" fill="var(--vp-c-text-3)" font-size="11" font-weight="600">
          Traditional GitOps
        </text>
        <path class="slow-path" d="M 30 60 Q 70 35, 110 60 T 190 60 Q 220 35, 250 60" stroke="#6b7280" stroke-width="3" fill="none" stroke-linecap="round"/>
        <circle cx="30" cy="60" r="5" fill="#6b7280" opacity="0.5"/>
        <circle cx="110" cy="60" r="5" fill="#6b7280" opacity="0.5"/>
        <circle cx="190" cy="60" r="5" fill="#6b7280" opacity="0.5"/>
        <circle cx="250" cy="60" r="5" fill="#6b7280" opacity="0.5"/>
        <text x="30" y="78" fill="var(--vp-c-text-3)" font-size="9" text-anchor="middle">Git</text>
        <text x="110" y="78" fill="var(--vp-c-text-3)" font-size="9" text-anchor="middle">CI</text>
        <text x="190" y="78" fill="var(--vp-c-text-3)" font-size="9" text-anchor="middle">CD</text>
        <text x="250" y="78" fill="var(--vp-c-text-3)" font-size="9" text-anchor="middle">K8s</text>
      </g>

      <line x1="20" y1="110" x2="260" y2="110" stroke="var(--vp-c-divider)" stroke-width="1.5" stroke-dasharray="6 4"/>

      <g>
        <text x="10" y="132" fill="#06b6d4" font-size="11" font-weight="600">
          Lynq Direct
        </text>
        <path class="fast-path" d="M 30 165 L 250 165" stroke="#06b6d4" stroke-width="4" fill="none" stroke-linecap="round"/>
        <circle class="travel-dot" cx="30" cy="165" r="5" fill="#06b6d4"/>
        <circle cx="30" cy="165" r="5" fill="#06b6d4"/>
        <circle cx="250" cy="165" r="5" fill="#06b6d4"/>
        <text x="30" y="183" fill="#06b6d4" font-size="9" text-anchor="middle" font-weight="600">DB</text>
        <text x="250" y="183" fill="#06b6d4" font-size="9" text-anchor="middle" font-weight="600">K8s</text>
      </g>
    </svg>
  </div>
</template>

<script setup>
defineProps({
  inView: {
    type: Boolean,
    default: false
  }
})
</script>

<style scoped>
.animation-container {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem 0;
}

.animation-svg {
  width: 100%;
  height: 200px;
  max-width: 100%;
}

.slow-path {
  stroke-dasharray: 500;
  stroke-dashoffset: 500;
  animation: pathDrawSlow 6s ease-out infinite;
  animation-play-state: paused;
}

.animation-svg.animate .slow-path {
  animation-play-state: running;
}

@keyframes pathDrawSlow {
  0% {
    stroke-dashoffset: 500;
  }
  50%, 100% {
    stroke-dashoffset: 0;
  }
}

.fast-path {
  stroke-dasharray: 300;
  stroke-dashoffset: 300;
  opacity: 0;
  animation: pathDrawFast 6s ease-out infinite;
  animation-play-state: paused;
}

.animation-svg.animate .fast-path {
  animation-play-state: running;
}

@keyframes pathDrawFast {
  0%, 16% {
    stroke-dashoffset: 300;
    opacity: 0;
  }
  20% {
    opacity: 1;
  }
  40%, 100% {
    stroke-dashoffset: 0;
    opacity: 1;
  }
}

.travel-dot {
  opacity: 0;
  animation: dotTravel 6s cubic-bezier(0.4, 0, 0.2, 1) infinite;
  animation-play-state: paused;
}

.animation-svg.animate .travel-dot {
  animation-play-state: running;
}

@keyframes dotTravel {
  0%, 40% {
    cx: 30;
    opacity: 0;
    r: 4;
  }
  43% {
    opacity: 1;
    r: 6;
  }
  80% {
    opacity: 1;
    r: 6;
  }
  83%, 100% {
    cx: 250;
    opacity: 0;
    r: 4;
  }
}
</style>
