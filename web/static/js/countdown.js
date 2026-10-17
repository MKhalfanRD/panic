(function () {
  var root = document.getElementById("countdown");
  if (!root) return;
  var end = Date.parse(root.getAttribute("data-offer-ends"));
  if (Number.isNaN(end)) return;
  var left = root.querySelector("[data-left]");
  var label = root.querySelector("[data-label]");
  function tick() {
    var diff = end - Date.now();
    if (diff <= 0) {
      label.textContent = "Harga perkenalan sudah berakhir";
      left.textContent = "";
      root.setAttribute("data-done", "true");
      return false;
    }
    var d = Math.floor(diff / 86400000);
    var h = Math.floor((diff % 86400000) / 3600000);
    var m = Math.floor((diff % 3600000) / 60000);
    left.textContent = d > 0 ? d + " hari " + h + " jam" : h > 0 ? h + " jam " + m + " menit" : Math.max(m, 1) + " menit";
    return true;
  }
  if (tick()) {
    var timer = setInterval(function () { if (!tick()) clearInterval(timer); }, 30000);
  }
})();
