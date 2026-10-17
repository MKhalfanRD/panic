(function () {
  // Busy label on submit, so a slow server does not invite a second click.
  document.querySelectorAll("form[data-busy]").forEach(function (form) {
    form.addEventListener("submit", function () {
      var btn = form.querySelector("button[type=submit]");
      if (!btn) return;
      btn.dataset.label = btn.textContent;
      btn.disabled = true;
      btn.textContent = form.getAttribute("data-busy");
    });
  });
  // Coming back with the Back button restores the page from cache with the button still busy.
  window.addEventListener("pageshow", function () {
    document.querySelectorAll("form[data-busy] button[data-label]").forEach(function (btn) {
      btn.disabled = false;
      btn.textContent = btn.dataset.label;
    });
  });

  document.querySelectorAll("[data-reveal]").forEach(function (btn) {
    var input = document.getElementById(btn.getAttribute("data-reveal"));
    if (!input) return;
    btn.addEventListener("click", function () {
      var show = input.type === "password";
      input.type = show ? "text" : "password";
      btn.textContent = show ? "Sembunyikan" : "Tampilkan";
    });
  });

  document.querySelectorAll("[data-copy]").forEach(function (btn) {
    var source = document.getElementById(btn.getAttribute("data-copy"));
    if (!source) return;
    btn.addEventListener("click", function () {
      function done() {
        btn.textContent = "Tersalin";
        setTimeout(function () { btn.textContent = "Salin"; }, 1500);
      }
      function fallback() {
        var range = document.createRange();
        range.selectNodeContents(source);
        var sel = window.getSelection();
        sel.removeAllRanges();
        sel.addRange(range);
      }
      if (navigator.clipboard) navigator.clipboard.writeText(source.textContent.trim()).then(done, fallback);
      else fallback();
    });
  });
})();
