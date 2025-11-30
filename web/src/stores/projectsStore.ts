import { computed, ref, watch } from "vue";
import { defineStore } from "pinia";

export const useProjectsStore = defineStore("projects", () => {
  const isLoadingProjectIds = ref(false);
  const allProjectsIds = ref(new Set<string>([]));
  const selectedProjectId = ref<string | null>(null);

  async function fetchProjects() {
    isLoadingProjectIds.value = true;
    try {
      console.info('fetching projects')
      const res = await fetch("http://localhost:8080/projects");

      const data = await res.json();
      const projectsIds = data.projectIds;

      allProjectsIds.value = new Set(projectsIds);
      if (!selectedProjectId.value) {
        selectedProjectId.value = projectsIds[0];
      }
    } finally {
      isLoadingProjectIds.value = false;
    }
  }

  const isLoadingProjectData = ref(false);
  const currentProjectData = ref<Record<string, unknown[]> | null>(null);
  const selectedTable = ref<string>('')

  const currentProjectTables = computed(() => {
    if (!currentProjectData.value) return [];

    return Object.keys(currentProjectData.value);
  });

  async function fetchProjectData() {
    isLoadingProjectData.value = true;
    try {
      console.info('fetching project data')
      const res = await fetch(`http://localhost:8080/projects/${selectedProjectId.value}`);

      const data = await res.json();

      if (Object.keys(data).length === 0) {
        currentProjectData.value = null
      } else {
        currentProjectData.value = data
      }

      selectedTable.value = Object.keys(data)?.[0] ?? ''
    } finally {
      isLoadingProjectData.value = false;
    }
  }

  watch(selectedProjectId, () => {
    fetchProjectData();
  });

  return {
    // /projects
    isLoadingProjectIds,
    allProjectsIds,
    selectedProjectId,
    fetchProjects,

    // /projects/:id
    isLoadingProjectData,
    currentProjectData,
    currentProjectTables,
    selectedTable,
    fetchProjectData,
  };
});
