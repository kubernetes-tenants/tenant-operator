<template>
  <DefaultTheme.Layout>
    <template #layout-top>
      <AnnouncementBanner />
    </template>
    <template #nav-bar-content-after>
      <GitHubStarBadge />
    </template>
  </DefaultTheme.Layout>
</template>

<script setup>
import DefaultTheme from 'vitepress/theme';
import AnnouncementBanner from '../components/AnnouncementBanner.vue';
import GitHubStarBadge from '../components/GitHubStarBadge.vue';
</script>
