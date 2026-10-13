(function () {
  document.querySelectorAll("[data-repeater]").forEach(function (box) {
    var rows = box.querySelector(".repeater-rows");
    var tpl = box.querySelector("template");
    var max = Number(box.getAttribute("data-max") || "12");
    var add = box.querySelector("[data-add]");
    if (!rows || !tpl || !add) return;
    add.addEventListener("click", function () {
      if (rows.children.length >= max) return;
      rows.appendChild(tpl.content.cloneNode(true));
    });
    rows.addEventListener("click", function (event) {
      var button = event.target.closest("[data-remove]");
      if (!button || !rows.contains(button)) return;
      var row = button.closest(".repeater-row");
      if (row) row.remove();
    });
  });

  document.querySelectorAll("form[data-confirm]").forEach(function (form) {
    form.addEventListener("submit", function (event) {
      if (!window.confirm(form.getAttribute("data-confirm"))) event.preventDefault();
    });
  });

  var name = document.getElementById("name");
  var slug = document.getElementById("slug");
  if (!name || !slug) return;
  if (slug.value) slug.dataset.touched = "1";
  name.addEventListener("input", function () {
    if (slug.dataset.touched) return;
    slug.value = name.value.toLowerCase().trim().replace(/[^a-z0-9]+/g, "-").replace(/^-+|-+$/g, "").slice(0, 80);
  });
  slug.addEventListener("input", function () { slug.dataset.touched = "1"; });
})();
