<template>
  <main class="p-4 bg-secondary-100 w-full gap-4 h-full overflow-y-auto">
    <h1 class="font-bold text-lg">Talk to your data</h1>
    <div class="flex flex-col gap-4">
      <ProjectsList />
      <GenerationTable />
      <div class="flex gap-2">
        <UInput v-model="query" class="w-full" placeholder="Query data using natural language" />
        <UButton @click="sendQuery">Query</UButton>
      </div>
      <div>{{ rawSql }}</div>

      <UCard variant="subtle" class="bg-white">
        <div class="flex flex-col gap-4">
          <UTable v-if="queryData.length" :data="JSON.parse(queryData)" class="flex-1 h-120" sticky />
        </div>
      </UCard>
    </div>
  </main>
</template>

<script setup lang="ts">
import { ref } from "vue";

import { useProjectsStore } from "@/stores/projectsStore";

import GenerationTable from "./GenerationView/components/GenerationTable.vue";
import ProjectsList from "./GenerationView/components/ProjectsList.vue";

const query = ref("");
const rawSql = ref("");
const queryData = ref("");

const projectsStore = useProjectsStore();

async function sendQuery() {
  rawSql.value = "";
  queryData.value = "";

  const res = await fetch(
    `http://localhost:8080/projects/${projectsStore.selectedProjectId}/query`,
    {
      method: "POST",
      body: JSON.stringify({
        query: query.value,
      }),
      headers: {
        "Content-Type": "application/json",
      },
    }
  );

  const data = await res.json();

  rawSql.value = data.generated_query;
  queryData.value = data.json_data;
}
</script>
