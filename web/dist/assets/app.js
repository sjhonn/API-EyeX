"use strict";
const colorInput = document.querySelector('#color');
const hexInput = document.querySelector('#hex');
const typeInput = document.querySelector('#type');
const severityInput = document.querySelector('#severity');
const severityValue = document.querySelector('#severity-value');
const statusEl = document.querySelector('#status');
const original = document.querySelector('#original');
const simulated = document.querySelector('#simulated');
const originalValue = document.querySelector('#original-value');
const simulatedValue = document.querySelector('#simulated-value');
let timer;
function normalizeHex(value) { const normalized = value.trim().toUpperCase(); return /^#[0-9A-F]{6}$/.test(normalized) ? normalized : null; }
async function refresh() {
    const hex = normalizeHex(hexInput.value);
    if (!hex) {
        statusEl.textContent = 'Usa un color hexadecimal con formato #RRGGBB.';
        return;
    }
    const severity = Number(severityInput.value);
    statusEl.textContent = 'Simulando...';
    try {
        const response = await fetch('/api/v1/simulate', { method: 'POST', headers: { Accept: 'application/json', 'Content-Type': 'application/json' }, body: JSON.stringify({ hex, type: typeInput.value, severity }) });
        const data = await response.json();
        if (!response.ok)
            throw new Error(data.message || data.error || `HTTP ${response.status}`);
        colorInput.value = data.original;
        hexInput.value = data.original;
        original.style.background = data.original;
        simulated.style.background = data.simulated;
        originalValue.textContent = data.original;
        simulatedValue.textContent = data.simulated;
        statusEl.textContent = `${data.model} · severidad ${data.severity.toFixed(2)}`;
    }
    catch (error) {
        statusEl.textContent = error instanceof Error ? error.message : 'No se pudo consultar EyeX.';
    }
}
function schedule() { window.clearTimeout(timer); timer = window.setTimeout(() => void refresh(), 180); }
colorInput.addEventListener('input', () => { hexInput.value = colorInput.value.toUpperCase(); schedule(); });
hexInput.addEventListener('input', schedule);
typeInput.addEventListener('change', () => void refresh());
severityInput.addEventListener('input', () => { severityValue.textContent = Number(severityInput.value).toFixed(2); schedule(); });
void refresh();
