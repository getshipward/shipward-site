// The light/dark toggle: shown only with JavaScript, saved in localStorage (theme.js reads it before the next paint).
(function () {
  var root = document.documentElement;
  var button = document.querySelector('[data-theme-toggle]');
  if (!button) return;
  button.hidden = false;
  button.addEventListener('click', function () {
    var dark = root.dataset.theme === 'dark' ||
      (root.dataset.theme !== 'light' && window.matchMedia('(prefers-color-scheme: dark)').matches);
    root.dataset.theme = dark ? 'light' : 'dark';
    try {
      localStorage.setItem('shipward-theme', root.dataset.theme);
    } catch (e) {
      // Storage blocked: the choice lasts for this page only.
    }
  });
})();
