<template>
  <UCard variant="subtle" class="bg-white">
    <div v-if="isLoadingProjectData">Loading..</div>
    <div v-else-if="!currentProjectData">Generation in progress.</div>
    <div v-else class="flex flex-col gap-4">
      <div class="flex items-center">
        <p>Generated data:</p>
        <div class="flex gap-2 items-center ml-auto">
          <p>Currently selected table:</p>
          <USelect class="w-48" v-model="selectedTable" :items="currentProjectTables" />
        </div>
      </div>
      <UTable :data="currentProjectData[selectedTable ?? '']" class="flex-1 h-120" sticky />
      <UButton class="ml-auto" color="error" @click="deleteProject">
        {{ isDeletingProject ? 'Loading..' : 'Delete Project' }}
      </UButton>
    </div>
  </UCard>
</template>

<script setup lang="ts">
import { storeToRefs } from "pinia";

import { useProjectsStore } from "@/stores/projectsStore";
import { ref } from "vue";

const toast = useToast();
const projectStore = useProjectsStore();
const { selectedProjectId, currentProjectData, currentProjectTables, selectedTable, isLoadingProjectData } =
  storeToRefs(projectStore);

const isDeletingProject = ref(false);

async function deleteProject() {
  isDeletingProject.value = true;

  try {
    await fetch(`http://localhost:8080/projects/${selectedProjectId.value}`, {
      method: "DELETE"
    });
    selectedProjectId.value = null

    projectStore.fetchProjects()
    toast.add({ title: "Project Deleted", color: "success" });
  } catch {
    toast.add({ title: "Error", description: "Failed to delete project", color: "error" });
  } finally {
    isDeletingProject.value = false;
  }
}
</script>
