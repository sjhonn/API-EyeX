<script setup lang="ts">
import { useEyeXTheme, type ThemeType } from './composables/useEyeXTheme';
const { types, selected, severity, mode, highContrast, palette, loading, status, loadTheme } = useEyeXTheme();
const labels:Record<ThemeType,string>={normal:'Normal',protanopia:'Protanopia',deuteranopia:'Deuteranopia',tritanopia:'Tritanopia',achromatopsia:'Acromatopsia',low_vision:'Baja visión'};
function refresh(){void loadTheme(selected.value);}
</script>
<template>
  <header class="bar"><div class="wrap"><strong>EyeX Vue</strong><span>Cliente de demostración</span></div></header>
  <main class="wrap main">
    <section class="intro"><div><h1>Temas de pantalla accesibles</h1><p>Este cliente Vue consume la API y aplica la paleta completa como variables CSS globales.</p></div></section>
    <section class="filters">
      <label><span>Modo</span><select v-model="selected" :disabled="loading" @change="refresh"><option v-for="type in types" :key="type" :value="type">{{ labels[type] }}</option></select></label>
      <label><span>Intensidad</span><select v-model="severity" @change="refresh"><option value="mild">Suave</option><option value="moderate">Moderada</option><option value="severe">Severa</option></select></label>
      <label><span>Tema</span><select v-model="mode" @change="refresh"><option value="light">Claro</option><option value="dark">Oscuro</option></select></label>
      <label class="check"><input v-model="highContrast" type="checkbox" @change="refresh"> Alto contraste</label>
    </section>
    <p class="status">{{ status }}</p>
    <section class="panel"><h2>Interfaz de ejemplo</h2><p>Todo este bloque cambia con la paleta devuelta por EyeX.</p><div class="actions"><button class="primary">Principal</button><button class="secondary">Secundaria</button></div><div class="alerts"><div class="error">Error</div><div class="success">Correcto</div></div></section>
    <section v-if="palette" class="panel"><h2>Paleta</h2><div v-for="(value,key) in palette" :key="key" class="row"><span class="chip" :style="{backgroundColor:value}"></span><code>{{ key }}</code><code>{{ value }}</code></div></section>
  </main>
</template>
