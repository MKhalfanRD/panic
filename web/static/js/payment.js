(function () {
  var btn = document.getElementById("pay-btn");
  if (!btn || !window.snap) return;
  btn.addEventListener("click", function () {
    btn.disabled = true;
    window.snap.pay(btn.getAttribute("data-token"), {
      onSuccess: function () { window.location.reload(); },
      onPending: function () { window.location.reload(); },
      onError: function () { btn.disabled = false; },
      onClose: function () { btn.disabled = false; }
    });
  });
})();
