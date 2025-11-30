<template>
  <div class="flex flex-col gap-2">
    Created projects:
    <div v-if="isLoadingProjectIds" class="flex gap-4">
      <USkeleton class="h-4 w-[200px] bg-primary" />
      <USkeleton class="h-4 w-[200px] bg-primary" />
    </div>
    <div v-else class="flex gap-4">
      <UBadge
      v-for="id in [...allProjectsIds]"
      :key="id"
      class="cursor-pointer"
      color="primary"
      :variant="selectedProjectId === id ? 'solid' : 'outline'"
      @click="selectedProjectId = id"
      >
      {{ id }}
    </UBadge>
  </div>
</div>
</template>

<script lang="ts" setup>
import { useProjectsStore } from "@/stores/projectsStore";
import { storeToRefs } from "pinia";
import { onMounted } from "vue";

const projectsStore = useProjectsStore();
const { isLoadingProjectIds, allProjectsIds, selectedProjectId } = storeToRefs(projectsStore);

onMounted(() => {
  projectsStore.fetchProjects();
});
</script>
