<script setup>
const rowSlots = [
  { y: 60, label: "Row #1" },
  { y: 115, label: "Row #2" },
  { y: 170, label: "Row #3" },
];

const stackSlices = [
  { y: 50, label: "Ingress" },
  { y: 90, label: "DNS / Networking" },
  { y: 130, label: "Service" },
  { y: 170, label: "Deployment" },
  { y: 210, label: "Custom CRDs" },
];

const connectors = [
  { d: "M 170 70 C 260 35 350 45 440 65", delay: 0 },
  { d: "M 170 120 C 260 110 350 110 440 120", delay: 0.6 },
  { d: "M 170 175 C 260 210 350 205 440 170", delay: 1.2 },
];
</script>

<template>
  <div class="row-stack-banner">
    <svg
      class="banner-svg"
      viewBox="0 0 640 260"
      preserveAspectRatio="xMidYMid meet"
      aria-hidden="true"
    >
      <defs>
        <linearGradient id="banner-bg" x1="0%" y1="0%" x2="100%" y2="0%">
          <stop offset="0%" stop-color="rgba(0,1,7,0.98)" />
          <stop offset="100%" stop-color="rgba(3,7,18,0.92)" />
        </linearGradient>
        <linearGradient id="connector-gradient" x1="0%" y1="0%" x2="100%" y2="0%">
          <stop offset="0%" stop-color="#6EE7B7" stop-opacity="0" />
          <stop offset="30%" stop-color="#6EE7B7" stop-opacity="0.8" />
          <stop offset="70%" stop-color="#38BDF8" stop-opacity="0.8" />
          <stop offset="100%" stop-color="#38BDF8" stop-opacity="0" />
        </linearGradient>
        <linearGradient id="stack-fill" x1="0%" y1="0%" x2="0%" y2="100%">
          <stop offset="0%" stop-color="rgba(59,130,246,0.85)" />
          <stop offset="100%" stop-color="rgba(79,70,229,0.85)" />
        </linearGradient>
        <filter id="soft-glow">
          <feGaussianBlur stdDeviation="8" result="coloredBlur" />
          <feMerge>
            <feMergeNode in="coloredBlur" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>
      </defs>

      <rect
        x="0"
        y="0"
        width="640"
        height="260"
        fill="url(#banner-bg)"
        opacity="0.55"
        rx="28"
      />

      <!-- Database rows -->
      <g class="db-block">
        <rect x="40" y="40" width="110" height="180" rx="14" opacity="0.18" />
        <rect x="55" y="55" width="80" height="20" rx="6" opacity="0.45" />
        <g v-for="row in rowSlots" :key="row.y">
          <rect
            x="55"
            :y="row.y"
            width="70"
            height="22"
            rx="5"
            class="row-pill"
          />
          <circle :cx="130" :cy="row.y + 11" r="3" class="row-pulse">
            <animate
              attributeName="r"
              values="3;6;3"
              dur="3s"
              repeatCount="indefinite"
              :begin="`${row.y / 80}s`"
            />
            <animate
              attributeName="opacity"
              values="0.6;0;0.6"
              dur="3s"
              repeatCount="indefinite"
              :begin="`${row.y / 80}s`"
            />
          </circle>
        </g>
      </g>

      <!-- Connectors -->
      <g class="connector-group">
        <template v-for="(connector, index) in connectors" :key="connector.d">
          <path
            :d="connector.d"
            class="connector"
            :style="{ animationDelay: `${connector.delay}s` }"
          />
          <circle r="6" class="connector-dot">
            <animateMotion
              :dur="`${3 + index * 0.4}s`"
              repeatCount="indefinite"
              :path="connector.d"
              :begin="`${connector.delay}s`"
            />
          </circle>
        </template>
      </g>

      <!-- Stack -->
      <g class="stack-block" filter="url(#soft-glow)">
        <rect x="420" y="40" width="180" height="180" rx="20" opacity="0.2" />
        <g v-for="(layer, index) in stackSlices" :key="layer.label">
          <rect
            x="430"
            :y="layer.y"
            width="160"
            height="28"
            rx="8"
            class="stack-layer"
          />
          <text
            x="510"
            :y="layer.y + 18"
            text-anchor="middle"
            class="stack-label"
          >
            {{ layer.label }}
          </text>
          <rect
            x="592"
            :y="layer.y + 5"
            width="6"
            height="18"
            rx="3"
            class="stack-meter"
            :style="{ animationDelay: `${index * 0.3}s` }"
          />
        </g>
      </g>
    </svg>

    <div class="banner-content">
      <p class="eyebrow">Row-synchronized infrastructure</p>
      <h3>Every Row Drives a Full Application Slice</h3>
      <p>
        Insert a record in your datasource and Lynq provisions the entire
        application footprint—Deployments, Services, Ingress, DNS, and any
        custom resources—kept in sync automatically.
      </p>
    </div>
  </div>
</template>

<style scoped>
.row-stack-banner {
  position: relative;
  overflow: hidden;
  border-radius: 18px;
  border: 1px solid rgba(15, 23, 42, 0.85);
  background: linear-gradient(145deg, #010314, #030919);
  box-shadow: 0 40px 70px rgba(0, 0, 0, 0.75);
}

.banner-svg {
  width: 100%;
  height: clamp(190px, 28vw, 280px);
  display: block;
  opacity: 0.9;
}

.banner-content {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 2rem 1.75rem;
  pointer-events: none;
  color: #f8fafc;
  text-shadow: 0 12px 36px rgba(0, 0, 0, 0.75);
}

.banner-content h3 {
  font-size: clamp(1.4rem, 4vw, 2rem);
  margin: 0.25rem 0 0.75rem;
  font-weight: 700;
  background: linear-gradient(90deg, #f8fafc, #e2e8f0);
  -webkit-background-clip: text;
  color: transparent;
}

.banner-content p {
  max-width: 580px;
  margin: 0 auto;
  line-height: 1.6;
  color: rgba(203, 213, 225, 0.92);
}

.banner-content .eyebrow {
  font-size: 0.85rem;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: rgba(59, 130, 246, 0.92);
  margin-bottom: 0.4rem;
}

.db-block rect {
  fill: rgba(0, 0, 0, 0.65);
  stroke: rgba(15, 118, 110, 0.4);
  stroke-width: 1.2;
}

.row-pill {
  fill: rgba(2, 132, 199, 0.32);
}

.row-pulse {
  fill: rgba(2, 132, 199, 0.25);
  opacity: 0.25;
}

.connector-group .connector {
  fill: none;
  stroke: url(#connector-gradient);
  stroke-width: 2.5;
  stroke-linecap: round;
  opacity: 0.22;
  stroke-dasharray: 6 18;
  animation: dash 4s linear infinite;
}

.connector-dot {
  fill: rgba(6, 182, 212, 0.45);
  filter: drop-shadow(0 0 4px rgba(6, 182, 212, 0.25));
}

.stack-block rect {
  fill: rgba(3, 7, 18, 0.8);
  stroke: rgba(15, 23, 42, 0.8);
  stroke-width: 1.2;
}

.stack-layer {
  fill: rgba(30, 64, 175, 0.35);
  stroke: rgba(248, 250, 252, 0.08);
}

.stack-label {
  font-size: 12px;
  fill: rgba(226, 232, 240, 0.5);
  font-weight: 600;
  letter-spacing: 0.02em;
}

.stack-meter {
  fill: rgba(16, 185, 129, 0.35);
  animation: pulse 2.8s ease-in-out infinite;
}

@keyframes dash {
  from {
    stroke-dashoffset: 0;
  }
  to {
    stroke-dashoffset: -120;
  }
}

@keyframes pulse {
  0%,
  100% {
    opacity: 0.4;
    transform: scaleY(0.65);
  }
  50% {
    opacity: 1;
    transform: scaleY(1);
  }
}

@media (max-width: 640px) {
  .banner-content {
    position: static;
  }

  .banner-svg {
    height: 200px;
  }

  .row-stack-banner {
    padding-bottom: 0;
  }

  .banner-content p {
    font-size: 0.95rem;
  }
}
</style>
