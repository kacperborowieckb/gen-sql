<template>
  <UCard variant="subtle" class="bg-white">
    <UForm :validate="validate" :state="state" class="space-y-4" @submit="onSubmit">
      <div class="flex gap-4">
        <UFormField class="w-3/4" label="Instructions" name="instructions">
          <UTextarea class="w-full" v-model="state.instructions" :rows="5" />
        </UFormField>

        <UFormField class="w-lg" label="Database schema" name="schemaFile">
          <UFileUpload
            v-model="state.schemaFile"
            accept="image/*"
            class="min-h-24 w-lg"
            description="Upload .ddl file"
          />
        </UFormField>
      </div>
      <USeparator />
      <p class="font-semibold">Advanced Parameters</p>
      <div class="flex gap-4">
        <UFormField class="grow" :label="`Temperature ${state.temperature}`" name="temperature">
          <USlider class="items-center" :min="0" :max="100" :default-value="50" v-model="state.temperature" />
        </UFormField>
        <UFormField class="w-md" label="Rows to generate" name="rowsToGenerate">
          <UInputNumber class="w-full" v-model="state.rowsToGenerate" orientation="vertical" />
        </UFormField>
      </div>
      <UButton type="submit"> Generate </UButton>
    </UForm>
  </UCard>
</template>

<script setup lang="ts">
import { reactive } from "vue";

import type { FormError, FormSubmitEvent } from "@nuxt/ui";

interface Parameters {
  instructions: string;
  schemaFile: File | null;
  temperature: number;
  rowsToGenerate: number;
}

const state = reactive<Parameters>({
  instructions: "",
  rowsToGenerate: 10,
  schemaFile: null,
  temperature: 50,
});

type Schema = typeof state;

function validate(state: Partial<Schema>): FormError[] {
  const errors = [];

  const { instructions, rowsToGenerate, schemaFile, temperature } = state;

  if (!instructions?.length || instructions?.length > 100) {
    errors.push({ name: "instructions", message: "Required and less then 250 chars" });
  }

  if (!rowsToGenerate || rowsToGenerate < 1 || rowsToGenerate > 100) {
    errors.push({ name: "rowsToGenerate", message: "Required and between 1-100" });
  }

  if (!schemaFile) {
    errors.push({ name: "schemaFile", message: "Schema file is required" });
  }

  if (!temperature || temperature < 1 || temperature > 100) {
    errors.push({ name: "temperature", message: "Required and between 1-100" });
  }

  return errors;
}

const toast = useToast();

async function onSubmit(event: FormSubmitEvent<Schema>) {
  toast.add({ title: "Success", description: "Generation started", color: "success" });
  console.log(event.data);
}
</script>
