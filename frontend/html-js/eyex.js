(function () {
  'use strict';
  if (window.EyeX && window.EyeX.__widgetLoaded) return;

  var script = document.currentScript;
  var src = script && script.src ? new URL(script.src, window.location.href) : new URL(window.location.href);
  var apiBase = (script && script.dataset.eyexApi ? script.dataset.eyexApi : src.origin).replace(/\/$/, '');
  var keys = ['background', 'surface', 'text', 'primary', 'secondary', 'error', 'success'];
  var labels = { normal: 'Normal', protanopia: 'Protanopia', deuteranopia: 'Deuteranopia', tritanopia: 'Tritanopia', achromatopsia: 'Acromatopsia', low_vision: 'Baja visión' };

  function setPalette(data) {
    keys.forEach(function (key) { document.documentElement.style.setProperty('--eyex-' + key, data.palette[key]); });
    document.documentElement.dataset.eyexTheme = data.type;
    document.body.style.backgroundColor = 'var(--eyex-background)';
    document.body.style.color = 'var(--eyex-text)';
  }

  async function apply(type, options) {
    options = options || {};
    var params = new URLSearchParams();
    if (options.severity) params.set('severity', options.severity);
    if (options.mode) params.set('mode', options.mode);
    if (typeof options.highContrast === 'boolean') params.set('high_contrast', String(options.highContrast));
    var response = await fetch(apiBase + '/api/v1/theme/' + encodeURIComponent(type) + (params.toString() ? '?' + params : ''), { headers: { Accept: 'application/json' } });
    var data = await response.json();
    if (!response.ok) throw new Error(data.message || data.error || 'EyeX request failed');
    setPalette(data);
    localStorage.setItem('eyex-widget-theme', data.type);
    return data;
  }

  function injectBaseStyles() {
    var style = document.createElement('style');
    style.id = 'eyex-widget-style';
    style.textContent = 'body{transition:background-color .15s,color .15s}a{color:var(--eyex-primary)!important}button,input,select,textarea{border-color:var(--eyex-primary)!important}.eyex-surface,[data-eyex-surface]{background:var(--eyex-surface)!important;color:var(--eyex-text)!important}.eyex-primary,[data-eyex-primary]{background:var(--eyex-primary)!important;color:#fff!important}#eyex-widget{position:fixed;right:18px;bottom:18px;z-index:2147483647;font-family:Arial,sans-serif}#eyex-widget button,#eyex-widget select{font:13px Arial,sans-serif}#eyex-toggle{width:48px;height:48px;border:0;border-radius:24px;background:#222831;color:#fff;box-shadow:0 2px 9px #0005;cursor:pointer;font-weight:700}#eyex-panel{display:none;position:absolute;right:0;bottom:58px;width:230px;padding:12px;background:#fff;color:#20252B;border:1px solid #8b949e;box-shadow:0 4px 18px #0004}#eyex-panel.open{display:grid;gap:9px}#eyex-panel label{display:grid;gap:4px;font-size:11px;font-weight:700}#eyex-panel select{height:32px;width:100%}#eyex-apply{padding:8px;background:#2E6DA4;color:#fff;border:0;cursor:pointer}';
    document.head.appendChild(style);
  }

  function makeWidget(types) {
    var host = document.createElement('div'); host.id = 'eyex-widget';
    var panel = document.createElement('div'); panel.id = 'eyex-panel'; panel.setAttribute('role', 'dialog'); panel.setAttribute('aria-label', 'Opciones EyeX');
    var typeLabel = document.createElement('label'); typeLabel.textContent = 'Modo'; var typeSelect = document.createElement('select');
    types.forEach(function (type) { var option = document.createElement('option'); option.value = type; option.textContent = labels[type] || type; typeSelect.appendChild(option); }); typeLabel.appendChild(typeSelect);
    var severityLabel = document.createElement('label'); severityLabel.textContent = 'Intensidad'; var severity = document.createElement('select');
    [['mild','Suave'],['moderate','Moderada'],['severe','Severa']].forEach(function (item) { var o=document.createElement('option');o.value=item[0];o.textContent=item[1];if(item[0]==='moderate')o.selected=true;severity.appendChild(o); }); severityLabel.appendChild(severity);
    var modeLabel = document.createElement('label'); modeLabel.textContent = 'Tema'; var mode = document.createElement('select');
    [['dark','Oscuro'],['light','Claro']].forEach(function (item) { var o=document.createElement('option');o.value=item[0];o.textContent=item[1];mode.appendChild(o); }); modeLabel.appendChild(mode);
    var hcLabel = document.createElement('label'); var hc = document.createElement('input'); hc.type = 'checkbox'; hcLabel.textContent = 'Alto contraste '; hcLabel.appendChild(hc);
    var applyButton = document.createElement('button'); applyButton.id = 'eyex-apply'; applyButton.type = 'button'; applyButton.textContent = 'Aplicar';
    panel.appendChild(typeLabel); panel.appendChild(severityLabel); panel.appendChild(modeLabel); panel.appendChild(hcLabel); panel.appendChild(applyButton);
    var toggle = document.createElement('button'); toggle.id = 'eyex-toggle'; toggle.type = 'button'; toggle.textContent = 'EyeX'; toggle.setAttribute('aria-label', 'Abrir opciones de accesibilidad EyeX');
    toggle.addEventListener('click', function () { panel.classList.toggle('open'); });
    applyButton.addEventListener('click', function () { apply(typeSelect.value, { severity: severity.value, mode: mode.value, highContrast: hc.checked }).then(function () { panel.classList.remove('open'); }).catch(function (error) { applyButton.textContent = error.message; }); });
    host.appendChild(panel); host.appendChild(toggle); document.body.appendChild(host);
    var saved = localStorage.getItem('eyex-widget-theme'); if (saved && types.indexOf(saved) >= 0) typeSelect.value = saved;
  }

  async function init() {
    injectBaseStyles();
    try {
      var response = await fetch(apiBase + '/api/v1/theme/types'); var data = await response.json();
      if (!response.ok) throw new Error('No se pudo cargar EyeX'); makeWidget(data.types);
    } catch (error) { console.warn('[EyeX]', error); }
  }

  window.EyeX = { __widgetLoaded: true, apply: apply, apiBase: apiBase };
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', init); else init();
}());
