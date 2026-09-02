<script setup lang="ts">
import { ref } from 'vue';
import { useEyeXTheme, type ThemeType } from './composables/useEyeXTheme';

const { types, selected, severity, mode, highContrast, palette, loading, status, loadTheme } = useEyeXTheme();
const labels:Record<ThemeType,string>={normal:'Normal',protanopia:'Protanopia',deuteranopia:'Deuteranopia',tritanopia:'Tritanopia',achromatopsia:'Acromatopsia',low_vision:'Baja visión'};
function refresh(){void loadTheme(selected.value);}

type SimulationType = 'protanopia' | 'deuteranopia' | 'tritanopia';
interface SimulationResponse { original:string; simulated:string; type:SimulationType; severity:number; model:'machado-2009' }
const simulationHex = ref('#FF0000');
const simulationType = ref<SimulationType>('protanopia');
const simulationSeverity = ref(1);
const simulationResult = ref<SimulationResponse | null>(null);
const simulationStatus = ref('Listo para simular.');
const simulationLoading = ref(false);

function normalizeHex(value:string):string|null{
  const normalized=value.trim().toUpperCase();
  return /^#[0-9A-F]{6}$/.test(normalized)?normalized:null;
}

async function simulate():Promise<void>{
  const hex=normalizeHex(simulationHex.value);
  if(!hex){simulationStatus.value='El color debe usar formato #RRGGBB.';return;}
  simulationLoading.value=true;
  simulationStatus.value='Simulando...';
  try{
    const response=await fetch('/api/v1/simulate',{method:'POST',headers:{Accept:'application/json','Content-Type':'application/json'},body:JSON.stringify({hex,type:simulationType.value,severity:simulationSeverity.value})});
    const data=await response.json() as SimulationResponse & {error?:string;message?:string};
    if(!response.ok)throw new Error(data.message||data.error||`HTTP ${response.status}`);
    simulationResult.value=data;
    simulationHex.value=data.original;
    simulationStatus.value=`Modelo ${data.model} · severidad ${data.severity.toFixed(2)}.`;
  }catch(cause){simulationStatus.value=cause instanceof Error?cause.message:'No se pudo simular el color';}
  finally{simulationLoading.value=false;}
}
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

    <section class="panel simulation-panel">
      <h2>Simulación Machado 2009</h2>
      <p>Consulta <code>POST /api/v1/simulate</code> con severidad continua entre 0 y 1.</p>
      <form class="simulation-form" @submit.prevent="simulate">
        <label><span>HEX</span><input v-model="simulationHex" maxlength="7" spellcheck="false"></label>
        <label><span>Tipo</span><select v-model="simulationType"><option value="protanopia">Protanopia</option><option value="deuteranopia">Deuteranopia</option><option value="tritanopia">Tritanopia</option></select></label>
        <label><span>Severidad {{ simulationSeverity.toFixed(2) }}</span><input v-model.number="simulationSeverity" type="range" min="0" max="1" step="0.05"></label>
        <button class="primary" type="submit" :disabled="simulationLoading">{{ simulationLoading ? 'Simulando...' : 'Simular' }}</button>
      </form>
      <p class="status">{{ simulationStatus }}</p>
      <div v-if="simulationResult" class="simulation-swatches">
        <div :style="{backgroundColor:simulationResult.original}"><span>Original {{ simulationResult.original }}</span></div>
        <div :style="{backgroundColor:simulationResult.simulated}"><span>Simulado {{ simulationResult.simulated }}</span></div>
      </div>
    </section>
  </main>
</template>
