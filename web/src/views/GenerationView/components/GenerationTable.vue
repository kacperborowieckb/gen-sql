<template>
  <UCard variant="subtle" class="bg-white">
    <div v-if="isLoadingProjectData">
      Loading..
    </div>
    <div v-else-if="!currentProjectData">
      Generation in progress.
    </div>
    <div v-else class="flex flex-col gap-4">
      <div class="flex items-center">
        <p>Generated data:</p>
        <div class="flex gap-2 items-center ml-auto">
          <p>Currently selected table: </p>
          <USelect class="w-48" v-model="selectedTable" :items="currentProjectTables" />
        </div>
      </div>
      <UTable :data="currentProjectData[selectedTable ?? '']" class="flex-1 h-120" sticky />
    </div>
  </UCard>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia';

import { useProjectsStore } from '@/stores/projectsStore';

const projectStore = useProjectsStore()
const { currentProjectData, currentProjectTables, selectedTable, isLoadingProjectData } = storeToRefs(projectStore)
</script>
