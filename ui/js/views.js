// views.js — 單頁應用的畫面切換（U16）
(function () {
  const VIEWS = ["settings", "menu", "new", "loading", "game", "ending"];
  function show(name) {
    VIEWS.forEach((v) => {
      const el = document.getElementById("view-" + v);
      if (el) el.classList.toggle("active", v === name);
    });
    window.NA && (NA._view = name);
  }
  window.Views = { show };
})();
