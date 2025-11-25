
<template>
   <UApp>
      <SideBar />
      <RouterView />
    </UApp>
</template>

<script setup lang="ts">
import { RouterView } from 'vue-router'

import SideBar from '@/components/SideBar.vue';
</script>
