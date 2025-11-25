<template>
  <UDashboardSidebar>
    <template #header>
      Gen SQL
    </template>
    <UNavigationMenu :items="items" orientation="vertical" />
  </UDashboardSidebar>
</template>

<script setup lang="ts">
import { ROUTES } from "@/constants";
import type { NavigationMenuItem } from "@nuxt/ui";

const { HOME, TALK_TO_DATA } = ROUTES

const items: NavigationMenuItem[] = [
  {
    label: "Data Generation",
    icon: "i-lucide-lightbulb",
    to: HOME.path
  },
  {
    label: "Talk To Your Data",
    icon: "i-lucide-message-circle-code",
    to: TALK_TO_DATA.path
  },
];
</script>
