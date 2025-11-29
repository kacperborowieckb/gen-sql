<template>
  <main class="p-4 bg-secondary-100 w-full gap-4 h-full overflow-y-auto">
    <div class="flex flex-col gap-4">
      <GenerationParameters />
      <ProjectsList />
      <GenerationTable />
    </div>
  </main>
</template>

<script setup lang="ts">
import GenerationParameters from "./components/GenerationParameters.vue";
import GenerationTable from "./components/GenerationTable.vue";
import ProjectsList from "./components/ProjectsList.vue";
</script>
