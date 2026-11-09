// Runs in <head>, before the page paints: the saved theme ("light" or "dark"; nothing means the system's) and the
// "js" class, which lets the CSS keep things hidden that site.js sets up.
(function () {
  var root = document.documentElement;
  root.classList.add('js');
  try {
    var theme = localStorage.getItem('shipward-theme');
    if (theme === 'light' || theme === 'dark') {
      root.dataset.theme = theme;
    }
  } catch (e) {
    // Storage blocked (private mode, settings): the system theme it is.
  }
})();
