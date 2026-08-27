(function () {
  'use strict';
  var palettes = {
    normal: { dark: ['#181A1D','#24272B','#F5F7FA','#5CA9E6','#AAB4BE'], light: ['#F4F5F7','#FFFFFF','#20252B','#2E6DA4','#6B7785'] },
    protanopia: { dark: ['#1E1E1E','#2A2A2A','#F5F5F5','#3F8FD2','#E3B341'], light: ['#F7F8FA','#FFFFFF','#1D2329','#256EA6','#916B00'] },
    deuteranopia: { dark: ['#1E1E1E','#2A2A2A','#F5F5F5','#4A90D9','#D9A24A'], light: ['#F7F8FA','#FFFFFF','#1D2329','#236FAE','#8A6200'] },
    tritanopia: { dark: ['#202124','#2D2F33','#F5F5F5','#D65DB1','#4CC9A7'], light: ['#F7F7F8','#FFFFFF','#202124','#9B3F80','#167A65'] },
    achromatopsia: { dark: ['#202020','#303030','#F2F2F2','#D0D0D0','#A8A8A8'], light: ['#FAFAFA','#FFFFFF','#181818','#4A4A4A','#666666'] },
    low_vision: { dark: ['#000000','#121212','#FFFFFF','#66B2FF','#FFD166'], light: ['#FFFFFF','#F2F2F2','#000000','#005FCC','#6D4C00'] }
  };

  function storageGet(defaults, callback) {
    if (typeof browser !== 'undefined') browser.storage.sync.get(defaults).then(callback);
    else chrome.storage.sync.get(defaults, callback);
  }

  function removeTheme() {
    var old = document.getElementById('eyex-extension-style');
    if (old) old.remove();
  }

  function apply(settings) {
    removeTheme();
    if (!settings.eyexEnabled) return;
    var typePalettes = palettes[settings.eyexType] || palettes.deuteranopia;
    var p = typePalettes[settings.eyexMode] || typePalettes.dark;
    var style = document.createElement('style');
    style.id = 'eyex-extension-style';
    style.textContent = [
      'html,body{background:' + p[0] + '!important;color:' + p[2] + '!important}',
      'main,article,section,aside,nav,header,footer,dialog{border-color:' + p[4] + '!important}',
      'a{color:' + p[3] + '!important}',
      'button,input,select,textarea{background:' + p[1] + '!important;color:' + p[2] + '!important;border-color:' + p[3] + '!important}',
      'table,td,th,pre,code{border-color:' + p[4] + '!important}'
    ].join('');
    document.documentElement.appendChild(style);
  }

  function read() {
    storageGet({ eyexType: 'deuteranopia', eyexMode: 'dark', eyexEnabled: false }, apply);
  }

  read();
  var storage = typeof browser !== 'undefined' ? browser.storage : chrome.storage;
  storage.onChanged.addListener(read);
}());
