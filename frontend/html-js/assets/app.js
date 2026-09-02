(function () {
  'use strict';

  var paletteKeys = ['background', 'surface', 'text', 'primary', 'secondary', 'error', 'success'];
  var defaultPalette = { background: '#F4F5F7', surface: '#FFFFFF', text: '#20252B', primary: '#2E6DA4', secondary: '#6B7785', error: '#C94C4C', success: '#3C8D5A' };
  var labels = { normal: 'Normal', protanopia: 'Protanopia', deuteranopia: 'Deuteranopia', tritanopia: 'Tritanopia', achromatopsia: 'Acromatopsia', low_vision: 'Baja visión' };
  var typeSelect = document.getElementById('theme-type');
  var severitySelect = document.getElementById('severity');
  var modeSelect = document.getElementById('mode');
  var highContrast = document.getElementById('high-contrast');
  var status = document.getElementById('theme-status');
  var contrastBadge = document.getElementById('contrast-badge');
  var paletteList = document.getElementById('palette-list');

  function applyPalette(palette) {
    paletteKeys.forEach(function (key) { document.documentElement.style.setProperty('--eyex-' + key, palette[key]); });
  }

  function renderPalette(palette) {
    paletteList.innerHTML = '';
    paletteKeys.forEach(function (key) {
      var row = document.createElement('div');
      var chip = document.createElement('span');
      var term = document.createElement('dt');
      var value = document.createElement('dd');
      row.className = 'palette-row'; chip.className = 'palette-chip'; chip.style.backgroundColor = palette[key]; term.textContent = key; value.textContent = palette[key];
      row.appendChild(chip); row.appendChild(term); row.appendChild(value); paletteList.appendChild(row);
    });
  }

  async function request(path, options) {
    var response = await fetch(path, Object.assign({ headers: { Accept: 'application/json' } }, options || {}));
    var payload = await response.json();
    if (!response.ok) throw new Error(payload.message || payload.error || ('HTTP ' + response.status));
    return payload;
  }

  function themePath() {
    var params = new URLSearchParams();
    params.set('severity', severitySelect.value);
    params.set('mode', modeSelect.value);
    params.set('high_contrast', highContrast.checked ? 'true' : 'false');
    return '/api/v1/theme/' + encodeURIComponent(typeSelect.value) + '?' + params.toString();
  }

  async function loadTheme() {
    status.textContent = 'Aplicando tema...';
    try {
      var data = await request(themePath());
      applyPalette(data.palette); renderPalette(data.palette);
      contrastBadge.textContent = data.contrast_ok ? 'WCAG AA: OK' : 'WCAG AA: revisar';
      contrastBadge.className = data.contrast_ok ? 'ok' : 'bad';
      status.textContent = 'Tema activo: ' + (labels[data.type] || data.type) + '.';
      localStorage.setItem('eyex-theme', data.type);
      localStorage.setItem('eyex-severity', severitySelect.value);
      localStorage.setItem('eyex-mode', modeSelect.value);
      localStorage.setItem('eyex-high-contrast', highContrast.checked ? 'true' : 'false');
    } catch (error) {
      status.textContent = 'No se pudo aplicar el tema: ' + error.message;
    }
  }

  function createCustomInputs() {
    var host = document.getElementById('custom-colors');
    paletteKeys.forEach(function (key) {
      var label = document.createElement('label');
      var text = document.createElement('span');
      var input = document.createElement('input');
      text.textContent = key; input.type = 'color'; input.name = key; input.value = defaultPalette[key];
      label.className = 'color-field'; label.appendChild(text); label.appendChild(input); host.appendChild(label);
    });
  }

  async function submitCustom(event) {
    event.preventDefault();
    var form = event.currentTarget;
    var customStatus = document.getElementById('custom-status');
    var palette = {};
    paletteKeys.forEach(function (key) { palette[key] = form.elements[key].value.toUpperCase(); });
    customStatus.textContent = 'Adaptando...';
    try {
      var data = await request('/api/v1/theme/custom', {
        method: 'POST',
        headers: { Accept: 'application/json', 'Content-Type': 'application/json' },
        body: JSON.stringify({ type: typeSelect.value, severity: severitySelect.value, mode: modeSelect.value, high_contrast: highContrast.checked, palette: palette })
      });
      applyPalette(data.palette); renderPalette(data.palette);
      contrastBadge.textContent = data.contrast_ok ? 'WCAG AA: OK' : 'WCAG AA: revisar';
      customStatus.textContent = 'Paleta adaptada y aplicada.';
    } catch (error) { customStatus.textContent = error.message; }
  }

  function normalizeSimulationHex(value) {
    var normalized = String(value || '').trim().toUpperCase();
    return /^#[0-9A-F]{6}$/.test(normalized) ? normalized : null;
  }

  async function submitSimulation(event) {
    event.preventDefault();
    var hexInput = document.getElementById('simulation-hex');
    var colorInput = document.getElementById('simulation-color');
    var typeInput = document.getElementById('simulation-type');
    var severityInput = document.getElementById('simulation-severity');
    var simulationStatus = document.getElementById('simulation-status');
    var hex = normalizeSimulationHex(hexInput.value);
    if (!hex) {
      simulationStatus.textContent = 'El color debe usar formato #RRGGBB.';
      return;
    }
    var severity = Number(severityInput.value);
    simulationStatus.textContent = 'Simulando...';
    try {
      var data = await request('/api/v1/simulate', {
        method: 'POST',
        headers: { Accept: 'application/json', 'Content-Type': 'application/json' },
        body: JSON.stringify({ hex: hex, type: typeInput.value, severity: severity })
      });
      hexInput.value = data.original;
      colorInput.value = data.original;
      document.getElementById('simulation-original').style.backgroundColor = data.original;
      document.getElementById('simulation-output').style.backgroundColor = data.simulated;
      document.getElementById('simulation-original-value').textContent = data.original;
      document.getElementById('simulation-output-value').textContent = data.simulated;
      simulationStatus.textContent = 'Modelo ' + data.model + ' · severidad ' + Number(data.severity).toFixed(2) + '.';
    } catch (error) {
      simulationStatus.textContent = 'No se pudo simular: ' + error.message;
    }
  }

  function initSimulation() {
    var form = document.getElementById('simulation-form');
    var hexInput = document.getElementById('simulation-hex');
    var colorInput = document.getElementById('simulation-color');
    var severityInput = document.getElementById('simulation-severity');
    var severityValue = document.getElementById('simulation-severity-value');
    colorInput.addEventListener('input', function () { hexInput.value = colorInput.value.toUpperCase(); });
    hexInput.addEventListener('input', function () {
      var normalized = normalizeSimulationHex(hexInput.value);
      if (normalized) colorInput.value = normalized;
    });
    severityInput.addEventListener('input', function () { severityValue.textContent = Number(severityInput.value).toFixed(2); });
    form.addEventListener('submit', submitSimulation);
    document.getElementById('simulation-original').style.backgroundColor = '#FF0000';
    void submitSimulation({ preventDefault: function () {} });
  }

  async function submitTest(event) {
    event.preventDefault();
    var form = event.currentTarget;
    var answers = {};
    ['reds_look_darker', 'green_brown_confusion', 'blue_yellow_confusion', 'colors_look_gray'].forEach(function (name) { answers[name] = form.elements[name].checked; });
    var result = document.getElementById('test-result'); result.textContent = 'Calculando sugerencia...';
    try {
      var data = await request('/api/v1/test/suggest', { method: 'POST', headers: { Accept: 'application/json', 'Content-Type': 'application/json' }, body: JSON.stringify({ answers: answers }) });
      result.textContent = 'Sugerencia: ' + (labels[data.suggested_type] || data.suggested_type) + '. ' + data.disclaimer;
    } catch (error) { result.textContent = 'No se pudo obtener una sugerencia: ' + error.message; }
  }

  async function init() {
    createCustomInputs();
    try {
      var data = await request('/api/v1/theme/types');
      typeSelect.innerHTML = '';
      data.types.forEach(function (type) { var option = document.createElement('option'); option.value = type; option.textContent = labels[type] || type; typeSelect.appendChild(option); });
      var saved = localStorage.getItem('eyex-theme'); typeSelect.value = data.types.indexOf(saved) >= 0 ? saved : 'normal';
      severitySelect.value = localStorage.getItem('eyex-severity') || 'moderate';
      modeSelect.value = localStorage.getItem('eyex-mode') || (typeSelect.value === 'normal' ? 'light' : 'dark');
      highContrast.checked = localStorage.getItem('eyex-high-contrast') === 'true';
      await loadTheme();
    } catch (error) { status.textContent = 'No se pudo consultar EyeX: ' + error.message; }
  }

  [typeSelect, severitySelect, modeSelect, highContrast].forEach(function (control) { control.addEventListener('change', loadTheme); });
  document.getElementById('custom-form').addEventListener('submit', submitCustom);
  document.getElementById('quick-test').addEventListener('submit', submitTest);
  initSimulation();
  init();
}());
