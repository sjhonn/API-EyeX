type SimulationType = 'protanopia' | 'deuteranopia' | 'tritanopia';
interface SimulateResponse { original:string; simulated:string; type:SimulationType; severity:number; model:'machado-2009' }

const colorInput=document.querySelector<HTMLInputElement>('#color')!;
const hexInput=document.querySelector<HTMLInputElement>('#hex')!;
const typeInput=document.querySelector<HTMLSelectElement>('#type')!;
const severityInput=document.querySelector<HTMLInputElement>('#severity')!;
const severityValue=document.querySelector<HTMLElement>('#severity-value')!;
const statusEl=document.querySelector<HTMLElement>('#status')!;
const original=document.querySelector<HTMLElement>('#original')!;
const simulated=document.querySelector<HTMLElement>('#simulated')!;
const originalValue=document.querySelector<HTMLElement>('#original-value')!;
const simulatedValue=document.querySelector<HTMLElement>('#simulated-value')!;
let timer:number|undefined;

function normalizeHex(value:string):string|null{const normalized=value.trim().toUpperCase();return /^#[0-9A-F]{6}$/.test(normalized)?normalized:null;}

async function refresh():Promise<void>{
  const hex=normalizeHex(hexInput.value);
  if(!hex){statusEl.textContent='Usa un color hexadecimal con formato #RRGGBB.';return;}
  const severity=Number(severityInput.value);
  statusEl.textContent='Simulando...';
  try{
    const response=await fetch('/api/v1/simulate',{method:'POST',headers:{Accept:'application/json','Content-Type':'application/json'},body:JSON.stringify({hex,type:typeInput.value,severity})});
    const data=await response.json() as SimulateResponse & {error?:string;message?:string};
    if(!response.ok)throw new Error(data.message||data.error||`HTTP ${response.status}`);
    colorInput.value=data.original;hexInput.value=data.original;
    original.style.background=data.original;simulated.style.background=data.simulated;
    originalValue.textContent=data.original;simulatedValue.textContent=data.simulated;
    statusEl.textContent=`${data.model} · severidad ${data.severity.toFixed(2)}`;
  }catch(error){statusEl.textContent=error instanceof Error?error.message:'No se pudo consultar EyeX.';}
}
function schedule():void{window.clearTimeout(timer);timer=window.setTimeout(()=>void refresh(),180);}
colorInput.addEventListener('input',()=>{hexInput.value=colorInput.value.toUpperCase();schedule();});
hexInput.addEventListener('input',schedule);
typeInput.addEventListener('change',()=>void refresh());
severityInput.addEventListener('input',()=>{severityValue.textContent=Number(severityInput.value).toFixed(2);schedule();});
void refresh();
