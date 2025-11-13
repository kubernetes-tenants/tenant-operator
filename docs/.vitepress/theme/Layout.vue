<template>
  <DefaultTheme.Layout>
    <template #layout-top>
      <AnnouncementBanner />
    </template>
  </DefaultTheme.Layout>
</template>

<script setup>
import DefaultTheme from 'vitepress/theme';
import AnnouncementBanner from '../components/AnnouncementBanner.vue';
</script>
