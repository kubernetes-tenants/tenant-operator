<template>
  <div class="animation-container">
    <svg viewBox="0 0 280 200" class="animation-svg" :class="{ 'animate': inView }">
      <g opacity="0.35">
        <text x="10" y="22" fill="var(--vp-c-text-3)" font-size="11" font-weight="600">
          Manual Provisioning
        </text>
        <circle class="slow-node" cx="50" cy="60" r="8" fill="#6b7280" style="--delay: 0s"/>
        <circle class="slow-node" cx="95" cy="60" r="8" fill="#6b7280" style="--delay: 0.7s"/>
        <circle class="slow-node" cx="140" cy="60" r="8" fill="#6b7280" style="--delay: 1.4s"/>
        <circle class="slow-node" cx="185" cy="60" r="8" fill="#6b7280" style="--delay: 2.1s"/>
        <circle class="slow-node" cx="230" cy="60" r="8" fill="#6b7280" style="--delay: 2.8s"/>
        <text x="140" y="88" fill="var(--vp-c-text-3)" font-size="10" text-anchor="middle">
          ~30 minutes
        </text>
      </g>

      <line x1="20" y1="110" x2="260" y2="110" stroke="var(--vp-c-divider)" stroke-width="1.5" stroke-dasharray="6 4"/>

      <g>
        <text x="10" y="132" fill="#10b981" font-size="11" font-weight="600">
          Lynq Automation
        </text>
        <circle class="fast-glow" cx="50" cy="165" r="8" fill="none" stroke="#10b981" stroke-width="2" style="--delay: 0s"/>
        <circle class="fast-glow" cx="95" cy="165" r="8" fill="none" stroke="#10b981" stroke-width="2" style="--delay: 0.08s"/>
        <circle class="fast-glow" cx="140" cy="165" r="8" fill="none" stroke="#10b981" stroke-width="2" style="--delay: 0.16s"/>
        <circle class="fast-glow" cx="185" cy="165" r="8" fill="none" stroke="#10b981" stroke-width="2" style="--delay: 0.24s"/>
        <circle class="fast-glow" cx="230" cy="165" r="8" fill="none" stroke="#10b981" stroke-width="2" style="--delay: 0.32s"/>
        <circle class="fast-node" cx="50" cy="165" r="8" fill="#10b981" style="--delay: 0s"/>
        <circle class="fast-node" cx="95" cy="165" r="8" fill="#10b981" style="--delay: 0.08s"/>
        <circle class="fast-node" cx="140" cy="165" r="8" fill="#10b981" style="--delay: 0.16s"/>
        <circle class="fast-node" cx="185" cy="165" r="8" fill="#10b981" style="--delay: 0.24s"/>
        <circle class="fast-node" cx="230" cy="165" r="8" fill="#10b981" style="--delay: 0.32s"/>
        <text class="fast-text" x="140" y="188" fill="#10b981" font-size="10" text-anchor="middle" font-weight="600">
          ~30 seconds
        </text>
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

.slow-node {
  opacity: 0;
  transform: scale(0);
  transform-origin: center;
  animation: nodeSlowAppear 8s cubic-bezier(0.4, 0, 0.2, 1) infinite;
  animation-delay: var(--delay);
  animation-play-state: paused;
}

.animation-svg.animate .slow-node {
  animation-play-state: running;
}

@keyframes nodeSlowAppear {
  0% {
    opacity: 0;
    transform: scale(0) translateY(8px);
  }
  8% {
    opacity: 0;
    transform: scale(0) translateY(8px);
  }
  12% {
    opacity: 1;
    transform: scale(1.15) translateY(-2px);
  }
  15% {
    opacity: 1;
    transform: scale(0.95) translateY(1px);
  }
  18% {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
  100% {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.fast-node {
  opacity: 0;
  transform: scale(0);
  transform-origin: center;
  animation: nodeFastAppear 8s cubic-bezier(0.34, 1.56, 0.64, 1) infinite;
  animation-delay: var(--delay);
  animation-play-state: paused;
}

.animation-svg.animate .fast-node {
  animation-play-state: running;
}

@keyframes nodeFastAppear {
  0%, 25% {
    opacity: 0;
    transform: scale(0);
  }
  28% {
    opacity: 1;
    transform: scale(1.4);
  }
  30% {
    opacity: 1;
    transform: scale(0.9);
  }
  32% {
    opacity: 1;
    transform: scale(1.05);
  }
  34% {
    opacity: 1;
    transform: scale(1);
  }
  100% {
    opacity: 1;
    transform: scale(1);
  }
}

.fast-glow {
  opacity: 0;
  transform: scale(1);
  transform-origin: center;
  animation: glowPulse 8s ease-out infinite;
  animation-delay: var(--delay);
  animation-play-state: paused;
}

.animation-svg.animate .fast-glow {
  animation-play-state: running;
}

@keyframes glowPulse {
  0%, 25% {
    opacity: 0;
    r: 8;
  }
  27% {
    opacity: 0.8;
    r: 18;
  }
  29% {
    opacity: 0.4;
    r: 22;
  }
  32% {
    opacity: 0;
    r: 8;
  }
  100% {
    opacity: 0;
    r: 8;
  }
}

.fast-text {
  opacity: 0;
  animation: textFade 8s ease-out infinite;
  animation-play-state: paused;
}

.animation-svg.animate .fast-text {
  animation-play-state: running;
}

@keyframes textFade {
  0%, 34% {
    opacity: 0;
  }
  37% {
    opacity: 1;
  }
  100% {
    opacity: 1;
  }
}
</style>
