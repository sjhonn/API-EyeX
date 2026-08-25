<script setup lang="ts">
import { useEyeXTheme, type ThemeType } from './composables/useEyeXTheme';

const { types, selected, palette, loading, status, loadTheme } = useEyeXTheme();

const labels: Record<ThemeType, string> = {
  normal: 'Normal',
  protanopia: 'Protanopia',
  deuteranopia: 'Deuteranopia',
  tritanopia: 'Tritanopia',
  achromatopsia: 'Acromatopsia',
};

function onChange(event: Event) {
  const value = (event.target as HTMLSelectElement).value as ThemeType;
  void loadTheme(value);
}
</script>

<template>
  <header class="bar">
    <div class="wrap"><strong>EyeX Vue</strong><span>Cliente de demostración</span></div>
  </header>
  <main class="wrap main">
    <section class="intro">
      <div>
        <h1>Modo daltónico por tema</h1>
        <p>Este cliente Vue consume la misma API GET y aplica la paleta completa como variables CSS globales.</p>
      </div>
      <label>
        <span>Modo</span>
        <select :value="selected" :disabled="loading" @change="onChange">
          <option v-for="type in types" :key="type" :value="type">{{ labels[type] }}</option>
        </select>
      </label>
    </section>

    <p class="status">{{ status }}</p>

    <section class="panel">
      <h2>Interfaz de ejemplo</h2>
      <p>Todo este bloque cambia con la paleta devuelta por EyeX.</p>
      <div class="actions">
        <button class="primary">Principal</button>
        <button class="secondary">Secundaria</button>
      </div>
      <div class="alerts">
        <div class="error">Error</div>
        <div class="success">Correcto</div>
      </div>
    </section>

    <section v-if="palette" class="panel">
      <h2>Paleta</h2>
      <div v-for="(value, key) in palette" :key="key" class="row">
        <span class="chip" :style="{ backgroundColor: value }"></span>
        <code>{{ key }}</code>
        <code>{{ value }}</code>
      </div>
    </section>
  </main>
</template>
