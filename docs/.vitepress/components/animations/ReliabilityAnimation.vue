<template>
  <div class="animation-container">
    <svg viewBox="0 0 280 200" class="animation-svg" :class="{ 'animate': inView }">
      <g opacity="0.35">
        <text x="10" y="22" fill="var(--vp-c-text-3)" font-size="11" font-weight="600">
          Manual Scripts
        </text>
        <g style="--manual-delay: 0s">
          <circle class="manual-error-circle" cx="70" cy="60" r="10" fill="none" stroke="#ef4444" stroke-width="2.5"/>
          <line class="manual-error-line" x1="65" y1="55" x2="75" y2="65" stroke="#ef4444" stroke-width="2.5" stroke-linecap="round"/>
          <line class="manual-error-line" x1="75" y1="55" x2="65" y2="65" stroke="#ef4444" stroke-width="2.5" stroke-linecap="round"/>
        </g>
        <text x="70" y="83" fill="var(--vp-c-text-3)" font-size="9" text-anchor="middle">Drift</text>
        <g style="--manual-delay: 0.3s">
          <circle class="manual-error-circle" cx="140" cy="60" r="10" fill="none" stroke="#f59e0b" stroke-width="2.5"/>
          <line class="manual-error-line" x1="135" y1="55" x2="145" y2="65" stroke="#f59e0b" stroke-width="2.5" stroke-linecap="round"/>
          <line class="manual-error-line" x1="145" y1="55" x2="135" y2="65" stroke="#f59e0b" stroke-width="2.5" stroke-linecap="round"/>
        </g>
        <text x="140" y="83" fill="var(--vp-c-text-3)" font-size="9" text-anchor="middle">Conflict</text>
        <g style="--manual-delay: 0.6s">
          <circle class="manual-error-circle" cx="210" cy="60" r="10" fill="none" stroke="#ef4444" stroke-width="2.5"/>
          <line class="manual-error-line" x1="205" y1="55" x2="215" y2="65" stroke="#ef4444" stroke-width="2.5" stroke-linecap="round"/>
          <line class="manual-error-line" x1="215" y1="55" x2="205" y2="65" stroke="#ef4444" stroke-width="2.5" stroke-linecap="round"/>
        </g>
        <text x="210" y="83" fill="var(--vp-c-text-3)" font-size="9" text-anchor="middle">Deps</text>
      </g>

      <line x1="20" y1="110" x2="260" y2="110" stroke="var(--vp-c-divider)" stroke-width="1.5" stroke-dasharray="6 4"/>

      <g>
        <text x="10" y="132" fill="#8b5cf6" font-size="11" font-weight="600">
          Lynq Automation
        </text>
        <g class="success-direct" transform="translate(70, 165)" style="--delay: 0s">
          <circle class="check-circle-direct" cx="0" cy="0" r="10" fill="none" stroke="#10b981" stroke-width="2.5"/>
          <path class="check-mark-direct" d="M -4 0 L -1 4 L 5 -4" stroke="#10b981" stroke-width="2.5" fill="none" stroke-linecap="round" stroke-linejoin="round"/>
        </g>
        <g class="auto-recover" transform="translate(140, 165)" style="--delay: 0.15s">
          <circle class="conflict-circle" cx="0" cy="0" r="10" fill="none" stroke="#f59e0b" stroke-width="2.5"/>
          <line class="conflict-line" x1="-5" y1="-5" x2="5" y2="5" stroke="#f59e0b" stroke-width="2.5" stroke-linecap="round"/>
          <line class="conflict-line" x1="5" y1="-5" x2="-5" y2="5" stroke="#f59e0b" stroke-width="2.5" stroke-linecap="round"/>
          <circle class="check-circle" cx="0" cy="0" r="10" fill="none" stroke="#10b981" stroke-width="2.5"/>
          <path class="check-mark" d="M -4 0 L -1 4 L 5 -4" stroke="#10b981" stroke-width="2.5" fill="none" stroke-linecap="round" stroke-linejoin="round"/>
        </g>
        <g class="success-direct" transform="translate(210, 165)" style="--delay: 0.3s">
          <circle class="check-circle-direct" cx="0" cy="0" r="10" fill="none" stroke="#10b981" stroke-width="2.5"/>
          <path class="check-mark-direct" d="M -4 0 L -1 4 L 5 -4" stroke="#10b981" stroke-width="2.5" fill="none" stroke-linecap="round" stroke-linejoin="round"/>
        </g>
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

.manual-error-circle,
.manual-error-line {
  opacity: 0;
  animation: manualErrorAppear 7s ease-in-out infinite;
  animation-delay: var(--manual-delay);
  animation-play-state: paused;
}

.animation-svg.animate .manual-error-circle,
.animation-svg.animate .manual-error-line {
  animation-play-state: running;
}

@keyframes manualErrorAppear {
  0% {
    opacity: 0;
  }
  5% {
    opacity: 0.8;
  }
  15%, 100% {
    opacity: 0.8;
  }
}

.conflict-circle,
.conflict-line {
  opacity: 0;
  animation: conflictAppear 7s ease-in-out infinite;
  animation-delay: var(--delay);
  animation-play-state: paused;
}

.animation-svg.animate .conflict-circle,
.animation-svg.animate .conflict-line {
  animation-play-state: running;
}

@keyframes conflictAppear {
  0% {
    opacity: 0;
  }
  5% {
    opacity: 1;
  }
  10%, 20% {
    opacity: 0.8;
  }
  25% {
    opacity: 0;
  }
  100% {
    opacity: 0;
  }
}

.check-circle {
  stroke-dasharray: 65;
  stroke-dashoffset: 65;
  opacity: 0;
  animation: circleAppear 7s ease-out infinite;
  animation-delay: var(--delay);
  animation-play-state: paused;
}

.animation-svg.animate .check-circle {
  animation-play-state: running;
}

@keyframes circleAppear {
  0%, 25% {
    opacity: 0;
    stroke-dashoffset: 65;
  }
  30% {
    opacity: 1;
    stroke-dashoffset: 45;
  }
  35% {
    opacity: 1;
    stroke-dashoffset: 20;
  }
  40%, 100% {
    opacity: 1;
    stroke-dashoffset: 0;
  }
}

.check-mark {
  stroke-dasharray: 20;
  stroke-dashoffset: 20;
  opacity: 0;
  animation: checkAppear 7s ease-out infinite;
  animation-delay: var(--delay);
  animation-play-state: paused;
}

.animation-svg.animate .check-mark {
  animation-play-state: running;
}

@keyframes checkAppear {
  0%, 40% {
    opacity: 0;
    stroke-dashoffset: 20;
  }
  42% {
    opacity: 1;
  }
  48%, 100% {
    opacity: 1;
    stroke-dashoffset: 0;
  }
}

.check-circle-direct {
  stroke-dasharray: 65;
  stroke-dashoffset: 65;
  opacity: 0;
  animation: circleDirectAppear 7s ease-out infinite;
  animation-delay: var(--delay);
  animation-play-state: paused;
}

.animation-svg.animate .check-circle-direct {
  animation-play-state: running;
}

@keyframes circleDirectAppear {
  0%, 5% {
    opacity: 0;
    stroke-dashoffset: 65;
  }
  10% {
    opacity: 1;
    stroke-dashoffset: 40;
  }
  15% {
    opacity: 1;
    stroke-dashoffset: 15;
  }
  20%, 100% {
    opacity: 1;
    stroke-dashoffset: 0;
  }
}

.check-mark-direct {
  stroke-dasharray: 20;
  stroke-dashoffset: 20;
  opacity: 0;
  animation: checkDirectAppear 7s ease-out infinite;
  animation-delay: var(--delay);
  animation-play-state: paused;
}

.animation-svg.animate .check-mark-direct {
  animation-play-state: running;
}

@keyframes checkDirectAppear {
  0%, 20% {
    opacity: 0;
    stroke-dashoffset: 20;
  }
  22% {
    opacity: 1;
  }
  28%, 100% {
    opacity: 1;
    stroke-dashoffset: 0;
  }
}
</style>
