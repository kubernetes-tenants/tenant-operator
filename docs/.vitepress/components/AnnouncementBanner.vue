<template>
  <Teleport to="body">
    <Transition name="slide-down">
      <div v-if="isVisible" class="announcement-banner">
        <div class="announcement-banner-content">
          <span class="announcement-icon">📢</span>
          <span class="announcement-text">
            <strong>Official Announcement:</strong> The project formerly known as "tenant-operator" is now <strong>Lynq</strong>
          </span>
          <a href="https://github.com/k8s-lynq/lynq" class="announcement-link" target="_blank" rel="noopener noreferrer">
            View on GitHub →
          </a>
          <button @click="closeBanner" class="close-button" aria-label="Close announcement">
            ✕
          </button>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue';

const STORAGE_KEY = 'lynq-announcement-banner-closed';
const isVisible = ref(false);

onMounted(() => {
  // Check if banner was closed in this session
  const wasClosed = sessionStorage.getItem(STORAGE_KEY);
  isVisible.value = !wasClosed;

  // Apply layout offset if banner is visible
  if (isVisible.value) {
    setTimeout(() => {
      updateLayoutOffset();
    }, 100);
  }

  // Handle window resize to adjust banner height
  window.addEventListener('resize', handleResize);
});

onUnmounted(() => {
  // Clean up layout offset when component unmounts
  removeLayoutOffset();
  window.removeEventListener('resize', handleResize);
});

watch(isVisible, (visible) => {
  if (visible) {
    updateLayoutOffset();
  } else {
    removeLayoutOffset();
  }
});

function handleResize() {
  if (isVisible.value) {
    updateLayoutOffset();
  }
}

function closeBanner() {
  isVisible.value = false;
  sessionStorage.setItem(STORAGE_KEY, 'true');
  removeLayoutOffset();
}

function updateLayoutOffset() {
  // Get the actual rendered height of the banner
  const banner = document.querySelector('.announcement-banner');
  if (banner) {
    const rect = banner.getBoundingClientRect();
    const bannerHeight = `${rect.height}px`;
    // Use VitePress's built-in CSS variable for layout top offset
    document.documentElement.style.setProperty('--vp-layout-top-height', bannerHeight);
  }
}

function removeLayoutOffset() {
  // Remove the offset - layout returns to default
  document.documentElement.style.removeProperty('--vp-layout-top-height');
}
</script>

<style scoped>
.announcement-banner {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  width: 100%;
  z-index: 1000;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 0.75rem 1.5rem;
  text-align: center;
  font-size: 0.95rem;
  font-weight: 500;
  box-shadow: 0 2px 12px rgba(102, 126, 234, 0.3);
}

.announcement-banner-content {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  flex-wrap: wrap;
  position: relative;
}

.announcement-icon {
  font-size: 1.2rem;
  animation: pulse 2s ease-in-out infinite;
}

.announcement-text {
  line-height: 1.4;
}

.announcement-text strong {
  font-weight: 700;
  text-decoration: underline;
  text-decoration-thickness: 2px;
  text-underline-offset: 3px;
}

.announcement-link {
  color: white;
  text-decoration: none;
  padding: 0.25rem 0.75rem;
  border: 1.5px solid rgba(255, 255, 255, 0.5);
  border-radius: 6px;
  font-weight: 600;
  transition: all 0.3s ease;
  white-space: nowrap;
}

.announcement-link:hover {
  background: rgba(255, 255, 255, 0.2);
  border-color: white;
  transform: translateY(-1px);
}

.close-button {
  position: absolute;
  right: 0;
  top: 50%;
  transform: translateY(-50%);
  background: transparent;
  border: none;
  color: white;
  font-size: 1.5rem;
  cursor: pointer;
  padding: 0.25rem 0.5rem;
  line-height: 1;
  opacity: 0.7;
  transition: all 0.2s ease;
  border-radius: 4px;
}

.close-button:hover {
  opacity: 1;
  background: rgba(255, 255, 255, 0.1);
}

.close-button:active {
  transform: translateY(-50%) scale(0.95);
}

@keyframes pulse {
  0%, 100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.1);
  }
}

/* Transition animations */
.slide-down-enter-active {
  animation: slideDown 0.5s ease-out;
}

.slide-down-leave-active {
  animation: slideUp 0.3s ease-in;
}

@keyframes slideDown {
  from {
    transform: translateY(-100%);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

@keyframes slideUp {
  from {
    transform: translateY(0);
    opacity: 1;
  }
  to {
    transform: translateY(-100%);
    opacity: 0;
  }
}

@media (max-width: 768px) {
  .announcement-banner {
    padding: 0.6rem 1rem;
    font-size: 0.85rem;
  }

  .announcement-banner-content {
    gap: 0.5rem;
    padding-right: 2rem;
  }

  .announcement-icon {
    display: none;
  }

  .close-button {
    font-size: 1.25rem;
    padding: 0.25rem;
  }
}
</style>
