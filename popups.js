const Popups = (function () {
  let overlay = null;
  let toastEl = null;
  let toastTimer = null;

  function onKeydown(evt) {
    if (evt.key === "Escape") close();
  }

  function close() {
    if (!overlay) return;
    overlay.remove();
    overlay = null;
    document.removeEventListener("keydown", onKeydown);
  }

  function open(innerHTML, options) {
    options = options || {};
    close();

    overlay = document.createElement("div");
    overlay.className = "popup-overlay";
    overlay.addEventListener("click", function (evt) {
      if (evt.target === overlay) close();
    });

    const box = document.createElement("div");
    box.className = "popup-box" + (options.wide ? " popup-wide" : "");
    box.innerHTML =
      '<button type="button" class="popup-close" data-popup-close aria-label="Fermer">✕</button>' +
      innerHTML;

    overlay.appendChild(box);
    document.body.appendChild(overlay);

    box.querySelectorAll("[data-popup-close]").forEach(function (btn) {
      btn.addEventListener("click", close);
    });

    document.addEventListener("keydown", onKeydown);

    const firstField = box.querySelector("input, textarea, select");
    if (firstField) firstField.focus();

    return box;
  }

  function toast(message) {
    if (!toastEl) {
      toastEl = document.createElement("div");
      toastEl.className = "toast";
      document.body.appendChild(toastEl);
    }
    toastEl.textContent = message;
    requestAnimationFrame(function () {
      toastEl.classList.add("show");
    });
    clearTimeout(toastTimer);
    toastTimer = setTimeout(function () {
      toastEl.classList.remove("show");
    }, 2600);
  }

  function confirmAction(message, onConfirm, options) {
    options = options || {};
    const box = open(
      '<h3 class="popup-title">' + (options.title || "Confirmation") + "</h3>" +
      "<p>" + message + "</p>" +
      '<div class="actions">' +
      '<button type="button" class="btn-danger" data-confirm-yes>' + (options.confirmLabel || "Confirmer") + "</button>" +
      '<button type="button" class="btn-secondary" data-popup-close>Annuler</button>' +
      "</div>"
    );
    box.querySelector("[data-confirm-yes]").addEventListener("click", function () {
      close();
      onConfirm();
    });
  }

  function entitySearchField(prefix, options) {
    options = options || {};
    const idValue = options.initialId != null ? options.initialId : "";
    const labelValue = options.initialLabel || "";
    return (
      '<div class="field">' +
      '<label for="' + prefix + '-search">' + options.label + "</label>" +
      '<input id="' + prefix + '-search" type="text" autocomplete="off" placeholder="' +
      (options.placeholder || "Rechercher...") + '" value="' + labelValue + '">' +
      '<div id="' + prefix + '-suggestions" class="suggestions" style="display:none;"></div>' +
      '<input type="hidden" id="' + prefix + '-id" value="' + idValue + '"' + (options.required ? " required" : "") + ">" +
      "</div>"
    );
  }

  function wireEntitySearch(prefix, searchFn, renderItem) {
    const searchInput = document.getElementById(prefix + "-search");
    const suggestionsBox = document.getElementById(prefix + "-suggestions");
    const hiddenId = document.getElementById(prefix + "-id");
    if (!searchInput || !suggestionsBox || !hiddenId) return;

    function renderSuggestions() {
      const query = searchInput.value;
      const matches = searchFn(query).slice(0, 8);
      if (!query.trim() || matches.length === 0) {
        suggestionsBox.style.display = "none";
        suggestionsBox.innerHTML = "";
        return;
      }
      suggestionsBox.innerHTML = matches.map(function (item) {
        const rendered = renderItem(item);
        return '<button type="button" data-id="' + item.id + '" data-label="' +
          escapeHtml(rendered.label) + '">' + rendered.html + "</button>";
      }).join("");
      suggestionsBox.style.display = "block";
      suggestionsBox.querySelectorAll("[data-id]").forEach(function (btn) {
        btn.addEventListener("click", function () {
          hiddenId.value = btn.dataset.id;
          searchInput.value = btn.dataset.label;
          suggestionsBox.style.display = "none";
        });
      });
    }

    searchInput.addEventListener("input", function () {
      hiddenId.value = "";
      renderSuggestions();
    });
    searchInput.addEventListener("focus", renderSuggestions);
    searchInput.addEventListener("blur", function () {
      setTimeout(function () { suggestionsBox.style.display = "none"; }, 150);
    });
  }

  function openReportForm(title, onSubmit) {
    const box = open(
      '<h3 class="popup-title">' + (title || "Signaler ce contenu") + "</h3>" +
      '<p class="muted">Expliquez en quelques mots ce qui pose probleme. Un(e) administrateur(rice) examinera votre signalement.</p>' +
      '<form id="report-form">' +
      '<div class="field"><label for="report-motif">Motif</label><textarea id="report-motif" required placeholder="Decrivez le probleme..."></textarea></div>' +
      '<div class="actions"><button type="submit" class="btn-danger">Envoyer le signalement</button><button type="button" class="btn-secondary" data-popup-close>Annuler</button></div>' +
      "</form>"
    );
    box.querySelector("#report-form").addEventListener("submit", function (evt) {
      evt.preventDefault();
      const motif = document.getElementById("report-motif").value.trim();
      if (!motif) return;
      onSubmit(motif);
      close();
      toast("Signalement envoye, merci.");
    });
  }

  return {
    open: open, close: close, toast: toast, confirmAction: confirmAction,
    entitySearchField: entitySearchField, wireEntitySearch: wireEntitySearch,
    openReportForm: openReportForm,
  };
})();
