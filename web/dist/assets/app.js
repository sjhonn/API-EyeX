"use strict";
const colorInput = document.querySelector("#color");
const hexInput = document.querySelector("#hex");
const typeInput = document.querySelector("#type");
const statusEl = document.querySelector("#status");
const palette = document.querySelector("#palette");
const original = document.querySelector("#original");
const simulated = document.querySelector("#simulated");
const corrected = document.querySelector("#corrected");
const originalValue = document.querySelector("#original-value");
const simulatedValue = document.querySelector("#simulated-value");
const correctedValue = document.querySelector("#corrected-value");
const ratioValue = document.querySelector("#ratio-value");
let timer;
const cache = new Map();
function normalizeHex(value) {
    const normalized = value.trim().toUpperCase();
    return /^#[0-9A-F]{6}$/.test(normalized) ? normalized : null;
}
async function post(path, body) {
    const response = await fetch(path, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
    });
    const payload = await response.json();
    if (!response.ok)
        throw new Error(payload.details || payload.error || `HTTP ${response.status}`);
    return payload;
}
async function refresh() {
    const hex = normalizeHex(hexInput.value);
    if (!hex) {
        statusEl.textContent = "Usa un color hexadecimal con formato #RRGGBB.";
        return;
    }
    colorInput.value = hex;
    const type = typeInput.value;
    const key = `${hex}|${type}`;
    statusEl.textContent = "Calculando…";
    try {
        let data = cache.get(key);
        if (!data) {
            const [sim, corr, pal] = await Promise.all([
                post("/api/v1/simulate", { hex, type }),
                post("/api/v1/correct", { hex, type }),
                post("/api/v1/palette", { base_hex: hex, type }),
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
    }
    catch (error) {
        statusEl.textContent = error instanceof Error ? error.message : "No se pudo consultar EyeX.";
    }
}
function scheduleRefresh() {
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
