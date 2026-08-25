(function () {
  'use strict';

  var select = document.getElementById('theme-type');
  var status = document.getElementById('theme-status');
  var paletteList = document.getElementById('palette-list');
  var paletteKeys = ['background', 'surface', 'text', 'primary', 'secondary', 'error', 'success'];

  function labelFor(type) {
    var labels = {
      normal: 'Normal',
      protanopia: 'Protanopia',
      deuteranopia: 'Deuteranopia',
      tritanopia: 'Tritanopia',
      achromatopsia: 'Acromatopsia'
    };
    return labels[type] || type;
  }

  function applyPalette(palette) {
    paletteKeys.forEach(function (key) {
      document.documentElement.style.setProperty('--eyex-' + key, palette[key]);
    });
  }

  function renderPalette(palette) {
    paletteList.innerHTML = '';
    paletteKeys.forEach(function (key) {
      var row = document.createElement('div');
      var chip = document.createElement('span');
      var term = document.createElement('dt');
      var value = document.createElement('dd');

      row.className = 'palette-row';
      chip.className = 'palette-chip';
      chip.style.backgroundColor = palette[key];
      term.textContent = key;
      value.textContent = palette[key];

      row.appendChild(chip);
      row.appendChild(term);
      row.appendChild(value);
      paletteList.appendChild(row);
    });
  }

  async function getJSON(path) {
    var response = await fetch(path, { headers: { Accept: 'application/json' } });
    var payload = await response.json();
    if (!response.ok) {
      throw new Error(payload.message || payload.error || ('HTTP ' + response.status));
    }
    return payload;
  }

  async function loadTheme(type) {
    status.textContent = 'Cargando tema ' + labelFor(type) + '...';
    try {
      var data = await getJSON('/api/v1/theme/' + encodeURIComponent(type));
      applyPalette(data.palette);
      renderPalette(data.palette);
      status.textContent = 'Tema activo: ' + labelFor(data.type) + '.';
      localStorage.setItem('eyex-theme', data.type);
    } catch (error) {
      status.textContent = 'No se pudo aplicar el tema: ' + error.message;
    }
  }

  async function init() {
    try {
      var data = await getJSON('/api/v1/theme/types');
      select.innerHTML = '';
      data.types.forEach(function (type) {
        var option = document.createElement('option');
        option.value = type;
        option.textContent = labelFor(type);
        select.appendChild(option);
      });

      var saved = localStorage.getItem('eyex-theme');
      var initial = data.types.indexOf(saved) >= 0 ? saved : 'normal';
      select.value = initial;
      await loadTheme(initial);
    } catch (error) {
      status.textContent = 'No se pudo consultar EyeX: ' + error.message;
    }
  }

  select.addEventListener('change', function () {
    loadTheme(select.value);
  });

  init();
}());
