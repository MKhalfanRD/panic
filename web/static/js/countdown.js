(function () {
  var root = document.getElementById("countdown");
  if (!root) return;
  var end = Date.parse(root.getAttribute("data-offer-ends"));
  if (Number.isNaN(end)) return;
  var units = {
    days: root.querySelector("[data-unit=days]"),
    hours: root.querySelector("[data-unit=hours]"),
    minutes: root.querySelector("[data-unit=minutes]"),
    seconds: root.querySelector("[data-unit=seconds]")
  };
  function pad(n) { return String(n).padStart(2, "0"); }
  function tick() {
    var diff = Math.max(0, end - Date.now());
    units.days.textContent = pad(Math.floor(diff / 86400000));
    units.hours.textContent = pad(Math.floor((diff % 86400000) / 3600000));
    units.minutes.textContent = pad(Math.floor((diff % 3600000) / 60000));
    units.seconds.textContent = pad(Math.floor((diff % 60000) / 1000));
    if (diff === 0) root.setAttribute("data-done", "true");
  }
  tick();
  setInterval(tick, 1000);
})();
