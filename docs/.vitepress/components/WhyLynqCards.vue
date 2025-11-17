<template>
  <div ref="cardsContainer" class="why-lynq-cards" :class="{ 'is-visible': isInView }">
    <div
      v-for="(card, index) in cards"
      :key="card.id"
      class="lynq-card"
      :style="{
        '--card-accent': card.accent,
        '--card-accent-light': card.accentLight,
        '--card-accent-alpha': card.accentAlpha,
        '--index': index
      }"
    >
      <div class="card-header">
        <span class="card-icon">{{ card.icon }}</span>
        <h3 class="card-title">{{ card.title }}</h3>
        <p class="card-subtitle">{{ card.subtitle }}</p>
      </div>

      <component
        :is="card.component"
        :in-view="isInView"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import ScaleAnimation from './animations/ScaleAnimation.vue'
import SpeedAnimation from './animations/SpeedAnimation.vue'
import ReliabilityAnimation from './animations/ReliabilityAnimation.vue'

const cardsContainer = ref(null)
const isInView = ref(false)
let observer = null

const cards = [
  {
    id: 'scale',
    icon: '⚡',
    title: 'Instant Scale',
    subtitle: 'One database row, unlimited automation',
    accent: '#10b981',
    accentLight: '#6ee7b7',
    accentAlpha: 'rgba(16, 185, 129, 0.2)',
    component: ScaleAnimation
  },
  {
    id: 'speed',
    icon: '🚀',
    title: 'Direct Path',
    subtitle: 'Database to Kubernetes, no CI/CD delays',
    accent: '#06b6d4',
    accentLight: '#67e8f9',
    accentAlpha: 'rgba(6, 182, 212, 0.2)',
    component: SpeedAnimation
  },
  {
    id: 'reliability',
    icon: '🛡️',
    title: 'Built-in Reliability',
    subtitle: 'Automatic drift correction and conflict resolution',
    accent: '#8b5cf6',
    accentLight: '#b794f6',
    accentAlpha: 'rgba(139, 92, 246, 0.2)',
    component: ReliabilityAnimation
  }
]

onMounted(() => {
  observer = new IntersectionObserver(
    (entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          isInView.value = true
        }
      })
    },
    { threshold: 0.2 }
  )

  if (cardsContainer.value) {
    observer.observe(cardsContainer.value)
  }
})

onUnmounted(() => {
  if (observer) {
    observer.disconnect()
  }
})
</script>

<style scoped>
.why-lynq-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 2rem;
  margin: 3rem 0;
}

.lynq-card {
  position: relative;
  background: var(--vp-c-bg-soft);
  border: 1px solid var(--vp-c-divider);
  border-radius: 12px;
  padding: 2rem;
  min-height: 480px;
  display: flex;
  flex-direction: column;
  overflow: visible;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  opacity: 0;
  transform: translateY(30px);
  animation: cardFadeIn 0.6s ease-out forwards;
  animation-delay: calc(var(--index) * 0.2s);
}

@keyframes cardFadeIn {
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.lynq-card::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: 12px;
  padding: 1px;
  background: linear-gradient(135deg, var(--card-accent), transparent 50%, var(--card-accent-light));
  -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
  mask-composite: exclude;
  opacity: 0;
  transition: opacity 0.3s ease;
  pointer-events: none;
}

.lynq-card:hover {
  border-color: transparent;
  box-shadow:
    0 8px 32px rgba(0, 0, 0, 0.08),
    0 0 0 1px var(--card-accent-alpha);
  transform: translateY(-4px);
  background: var(--vp-c-bg);
}

.lynq-card:hover::before {
  opacity: 1;
}

.card-header {
  margin-bottom: 2rem;
}

.card-icon {
  font-size: 2.5rem;
  display: block;
  margin-bottom: 1rem;
  filter: grayscale(0.2);
  transition: all 0.3s ease;
}

.lynq-card:hover .card-icon {
  filter: grayscale(0);
  transform: scale(1.1);
}

.card-title {
  font-size: 1.4rem;
  font-weight: 700;
  color: var(--vp-c-text-1);
  margin: 0 0 0.5rem 0;
}

.card-subtitle {
  font-size: 0.95rem;
  color: var(--vp-c-text-2);
  margin: 0;
  font-weight: 500;
  line-height: 1.5;
}

@media (max-width: 768px) {
  .why-lynq-cards {
    grid-template-columns: 1fr;
    gap: 1.5rem;
  }

  .lynq-card {
    min-height: 420px;
    padding: 1.5rem;
  }

  .card-icon {
    font-size: 2rem;
  }

  .card-title {
    font-size: 1.2rem;
  }

  .card-subtitle {
    font-size: 0.9rem;
  }
}
</style>
