type EyeXType =
  | "protanopia"
  | "protanomaly"
  | "deuteranopia"
  | "deuteranomaly"
  | "tritanopia"
  | "tritanomaly"
  | "acromatopsia";

interface SimulateResponse { original: string; simulated: string }
interface CorrectResponse { original: string; corrected: string; contrast_ratio: number }
interface PaletteResponse { background: string; suggested_variants: string[] }

const colorInput = document.querySelector<HTMLInputElement>("#color")!;
const hexInput = document.querySelector<HTMLInputElement>("#hex")!;
const typeInput = document.querySelector<HTMLSelectElement>("#type")!;
const statusEl = document.querySelector<HTMLElement>("#status")!;
const palette = document.querySelector<HTMLElement>("#palette")!;

const original = document.querySelector<HTMLElement>("#original")!;
const simulated = document.querySelector<HTMLElement>("#simulated")!;
const corrected = document.querySelector<HTMLElement>("#corrected")!;
const originalValue = document.querySelector<HTMLElement>("#original-value")!;
const simulatedValue = document.querySelector<HTMLElement>("#simulated-value")!;
const correctedValue = document.querySelector<HTMLElement>("#corrected-value")!;
const ratioValue = document.querySelector<HTMLElement>("#ratio-value")!;

let timer: number | undefined;
const cache = new Map<string, { sim: SimulateResponse; corr: CorrectResponse; pal: PaletteResponse }>();

function normalizeHex(value: string): string | null {
  const normalized = value.trim().toUpperCase();
  return /^#[0-9A-F]{6}$/.test(normalized) ? normalized : null;
}

async function post<T>(path: string, body: unknown): Promise<T> {
  const response = await fetch(path, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  const payload = await response.json();
  if (!response.ok) throw new Error(payload.details || payload.error || `HTTP ${response.status}`);
  return payload as T;
}

async function refresh(): Promise<void> {
  const hex = normalizeHex(hexInput.value);
  if (!hex) {
    statusEl.textContent = "Usa un color hexadecimal con formato #RRGGBB.";
    return;
  }

  colorInput.value = hex;
  const type = typeInput.value as EyeXType;
  const key = `${hex}|${type}`;
  statusEl.textContent = "Calculando…";

  try {
    let data = cache.get(key);
    if (!data) {
      const [sim, corr, pal] = await Promise.all([
        post<SimulateResponse>("/api/v1/simulate", { hex, type }),
        post<CorrectResponse>("/api/v1/correct", { hex, type }),
        post<PaletteResponse>("/api/v1/palette", { base_hex: hex, type }),
      ]);
      data = { sim, corr, pal };
      cache.set(key, data);
    }

    original.style.background = data.sim.original;
    simulated.style.background = data.sim.simulated;
    corrected.style.background = data.corr.corrected;
    originalValue.textContent = data.sim.original;
    simulatedValue.textContent = data.sim.simulated;
    correctedValue.textContent = data.corr.corrected;
    ratioValue.textContent = data.corr.contrast_ratio.toFixed(2);

    palette.replaceChildren(...data.pal.suggested_variants.map((value) => {
      const el = document.createElement("div");
      el.className = "palette-item";
      el.style.background = value;
      el.textContent = value;
      return el;
    }));
    statusEl.textContent = "Resultado actualizado.";
  } catch (error) {
    statusEl.textContent = error instanceof Error ? error.message : "No se pudo consultar EyeX.";
  }
}

function scheduleRefresh(): void {
  window.clearTimeout(timer);
  timer = window.setTimeout(() => void refresh(), 180);
}

colorInput.addEventListener("input", () => {
  hexInput.value = colorInput.value.toUpperCase();
  scheduleRefresh();
});
hexInput.addEventListener("input", scheduleRefresh);
typeInput.addEventListener("change", scheduleRefresh);

void refresh();
