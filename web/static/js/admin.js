(function () {
  document.querySelectorAll("[data-repeater]").forEach(function (box) {
    var rows = box.querySelector(".rep-rows");
    var tpl = box.querySelector("template");
    var add = box.querySelector("[data-add]");
    var max = Number(box.getAttribute("data-max") || "12");
    if (!rows || !tpl || !add) return;
    function sync() { add.disabled = rows.children.length >= max; }
    add.addEventListener("click", function () {
      if (rows.children.length >= max) return;
      rows.appendChild(tpl.content.cloneNode(true));
      var first = rows.lastElementChild && rows.lastElementChild.querySelector("input, textarea");
      if (first) first.focus();
      sync();
      markDirty();
    });
    rows.addEventListener("click", function (event) {
      var button = event.target.closest("[data-remove]");
      if (!button || !rows.contains(button)) return;
      var row = button.closest(".rep-row");
      if (row) row.remove();
      sync();
      markDirty();
    });
    sync();
  });

  // Unsaved-changes guard for the product form.
  var form = document.getElementById("product-form");
  var dirtyLabel = document.querySelector("[data-dirty-label]");
  var dirty = false;
  function markDirty() {
    if (!form || dirty) return;
    dirty = true;
    if (dirtyLabel) {
      dirtyLabel.textContent = "Ada perubahan belum disimpan";
      dirtyLabel.setAttribute("data-dirty", "1");
    }
  }
  if (form) {
    form.addEventListener("input", markDirty);
    form.addEventListener("change", markDirty);
    form.addEventListener("submit", function () { dirty = false; });
    window.addEventListener("beforeunload", function (event) {
      if (dirty) event.preventDefault();
    });
  }

  var name = document.getElementById("name");
  var slug = document.getElementById("slug");
  if (name && slug) {
    if (slug.value) slug.dataset.touched = "1";
    name.addEventListener("input", function () {
      if (slug.dataset.touched) return;
      slug.value = name.value.toLowerCase().trim().replace(/[^a-z0-9]+/g, "-").replace(/^-+|-+$/g, "").slice(0, 80);
    });
    slug.addEventListener("input", function () { slug.dataset.touched = "1"; });
  }

  // Rupiah inputs: thousand separators while typing (the server strips them), plus the discount line.
  var idr = new Intl.NumberFormat("id-ID");
  function digits(v) { return parseInt(String(v).replace(/\D/g, ""), 10) || 0; }
  var price = document.getElementById("price");
  var original = document.getElementById("original_price");
  var saving = document.querySelector("[data-saving]");
  function showSaving() {
    if (!saving || !price || !original) return;
    var p = digits(price.value), o = digits(original.value);
    saving.classList.remove("bad");
    if (o && p && o > p) saving.textContent = "Pembeli hemat Rp " + idr.format(o - p) + " (" + Math.round((o - p) / o * 100) + "%)";
    else if (o && p) { saving.textContent = "Harga coret harus lebih tinggi dari harga jual, atau kosongkan."; saving.classList.add("bad"); }
    else saving.textContent = "";
  }
  document.querySelectorAll("[data-idr]").forEach(function (input) {
    function fmt() { var n = digits(input.value); input.value = n ? idr.format(n) : ""; }
    fmt();
    input.addEventListener("input", function () { fmt(); showSaving(); });
  });
  showSaving();

  document.querySelectorAll("[data-count]").forEach(function (field) {
    var out = document.getElementById(field.getAttribute("aria-describedby"));
    var limit = Number(field.getAttribute("data-count"));
    function count() { out.textContent = field.value.length + "/" + limit + " karakter"; }
    if (out) { count(); field.addEventListener("input", count); }
  });

  var ask = document.querySelector("[data-ask-delete]");
  var del = document.querySelector("form[data-delete]");
  if (ask && del) {
    ask.addEventListener("click", function () {
      del.hidden = false;
      del.querySelector("button[type=submit]").focus();
    });
    del.querySelector("[data-cancel-delete]").addEventListener("click", function () {
      del.hidden = true;
      ask.focus();
    });
    del.addEventListener("submit", function () { dirty = false; });
  }

  // Orders: filter by status group, with counts.
  var filters = document.querySelector("[data-filters]");
  if (filters) {
    var items = document.querySelectorAll("[data-k]");
    var none = document.querySelector("[data-none]");
    var counts = { all: 0, pending: 0, paid: 0, closed: 0 };
    document.querySelectorAll("tbody [data-k]").forEach(function (row) {
      counts.all++;
      counts[row.getAttribute("data-k")]++;
    });
    filters.querySelectorAll("button").forEach(function (btn) {
      var key = btn.getAttribute("data-f");
      btn.querySelector("span").textContent = counts[key];
      btn.addEventListener("click", function () {
        filters.querySelectorAll("button").forEach(function (b) { b.setAttribute("aria-pressed", b === btn ? "true" : "false"); });
        items.forEach(function (el) { el.hidden = key !== "all" && el.getAttribute("data-k") !== key; });
        if (none) none.hidden = counts[key] > 0;
      });
    });
  }
})();
