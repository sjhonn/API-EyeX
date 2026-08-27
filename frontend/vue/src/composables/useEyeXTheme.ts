import { computed, onMounted, ref } from 'vue';

export type ThemeType = 'normal' | 'protanopia' | 'deuteranopia' | 'tritanopia' | 'achromatopsia' | 'low_vision';
export type Severity = 'mild' | 'moderate' | 'severe';
export type ThemeMode = 'dark' | 'light';
export interface Palette { background:string; surface:string; text:string; primary:string; secondary:string; error:string; success:string; }
interface ThemeResponse { type:ThemeType; palette:Palette; contrast_ok:boolean; }
interface TypesResponse { types:ThemeType[]; }
const paletteKeys: Array<keyof Palette> = ['background','surface','text','primary','secondary','error','success'];
function apiBase(): string { const fromQuery=new URLSearchParams(window.location.search).get('api'); if(fromQuery)return fromQuery.replace(/\/$/,''); return window.location.port==='5173'?'http://localhost:8080':''; }

export function useEyeXTheme() {
  const types=ref<ThemeType[]>([]), selected=ref<ThemeType>('normal'), severity=ref<Severity>('moderate'), mode=ref<ThemeMode>('dark'), highContrast=ref(false);
  const palette=ref<Palette|null>(null), contrastOk=ref(false), loading=ref(false), error=ref(''); const base=apiBase();
  const status=computed(()=>loading.value?'Cargando tema...':error.value?error.value:palette.value?`Tema activo: ${selected.value}. Contraste AA: ${contrastOk.value?'OK':'revisar'}`:'Sin tema cargado');
  async function getJSON<T>(path:string):Promise<T>{const response=await fetch(`${base}${path}`,{headers:{Accept:'application/json'}});const payload=await response.json();if(!response.ok)throw new Error(payload.message||payload.error||`HTTP ${response.status}`);return payload as T;}
  function applyPalette(next:Palette):void{paletteKeys.forEach(key=>document.documentElement.style.setProperty(`--eyex-${key}`,next[key]));}
  async function loadTheme(type:ThemeType=selected.value):Promise<void>{loading.value=true;error.value='';try{const q=new URLSearchParams({severity:severity.value,mode:mode.value,high_contrast:String(highContrast.value)});const data=await getJSON<ThemeResponse>(`/api/v1/theme/${encodeURIComponent(type)}?${q}`);selected.value=data.type;palette.value=data.palette;contrastOk.value=data.contrast_ok;applyPalette(data.palette);localStorage.setItem('eyex-theme',data.type);}catch(cause){error.value=cause instanceof Error?cause.message:'No se pudo cargar EyeX';}finally{loading.value=false;}}
  async function initialize():Promise<void>{loading.value=true;try{const data=await getJSON<TypesResponse>('/api/v1/theme/types');types.value=data.types;const saved=localStorage.getItem('eyex-theme') as ThemeType|null;selected.value=saved&&data.types.includes(saved)?saved:'normal';mode.value=selected.value==='normal'?'light':'dark';await loadTheme(selected.value);}catch(cause){error.value=cause instanceof Error?cause.message:'No se pudo consultar EyeX';loading.value=false;}}
  onMounted(()=>void initialize());
  return {types,selected,severity,mode,highContrast,palette,contrastOk,loading,status,loadTheme};
}
