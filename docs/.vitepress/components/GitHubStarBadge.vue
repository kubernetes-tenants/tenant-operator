<template>
  <a
    href="https://github.com/k8s-lynq/lynq"
    target="_blank"
    rel="noopener noreferrer"
    class="github-star-badge"
    aria-label="Star k8s-lynq/lynq on GitHub"
  >
    <img
      src="https://img.shields.io/github/stars/k8s-lynq/lynq?style=social"
      alt="GitHub stars"
      class="badge-img"
    />
  </a>
</template>

<style scoped>
.github-star-badge {
  display: inline-flex;
  align-items: center;
  margin-left: 8px;
  transition: opacity 0.2s ease;
}

.github-star-badge:hover {
  opacity: 0.8;
}

.badge-img {
  height: 20px;
  vertical-align: middle;
}

@media (max-width: 768px) {
  .badge-img {
    height: 18px;
  }
}
</style>
