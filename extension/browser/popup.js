(function () {
  'use strict';
  var type = document.getElementById('type');
  var mode = document.getElementById('mode');
  var enabled = document.getElementById('enabled');

  function storageGet(defaults, callback) {
    if (typeof browser !== 'undefined') {
      browser.storage.sync.get(defaults).then(callback);
    } else {
      chrome.storage.sync.get(defaults, callback);
    }
  }

  function storageSet(values, callback) {
    if (typeof browser !== 'undefined') {
      browser.storage.sync.set(values).then(callback);
    } else {
      chrome.storage.sync.set(values, callback);
    }
  }

  storageGet({ eyexType: 'deuteranopia', eyexMode: 'dark', eyexEnabled: false }, function (data) {
    type.value = data.eyexType;
    mode.value = data.eyexMode;
    enabled.checked = data.eyexEnabled;
  });

  document.getElementById('save').addEventListener('click', function () {
    storageSet({ eyexType: type.value, eyexMode: mode.value, eyexEnabled: enabled.checked }, function () {
      window.close();
    });
  });
}());
