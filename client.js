const USER_TYPE_LABELS = {
  visiteur: "Visiteur(se)",
  adherent: "Adhérent(e)",
  benevole: "Bénévole",
  commercant: "Commerçant(e)",
  administrateur: "Administrateur(rice)",
};

function userTypeLabel(type) {
  return USER_TYPE_LABELS[type] || type;
}

function renderNavShell(brandHref, links) {
  const root = document.getElementById("topnav-root");
  if (!root) return;

  const user = Mock.getCurrentUser();
  const users = Mock.getUsers();
  const badgeLabel = user ? escapeHtml(user.prenom) + " " + escapeHtml(user.nom) + " &middot; " + userTypeLabel(user.type) : "Invité(e) (non connecté(e))";

  root.innerHTML =
    '<header class="topnav">' +
    '<a class="brand" href="' + brandHref + '"><span class="brand-mark">N</span> NO MORE WASTE</a>' +
    "<nav>" +
    links.map(function (l) {
      return '<a href="' + l.href + '" class="' + (l.active ? "active" : "") + '">' + l.label + "</a>";
    }).join("") +
    "</nav>" +
    '<div class="who">' +
    (user ? renderNotifBellHtml(user) : "") +
    '<span class="badge">' + badgeLabel + "</span>" +
    '<select id="demo-user-switch" title="Changer de profil de demonstration (mockup)">' +
    '<option value=""' + (user ? "" : " selected") + ">Invité(e) (non connecté(e))</option>" +
    users.map(function (u) {
      return '<option value="' + u.id + '"' + (user && u.id === user.id ? " selected" : "") + ">" + escapeHtml(u.prenom) + " " + escapeHtml(u.nom) + " (" + userTypeLabel(u.type) + ")</option>";
    }).join("") +
    "</select>" +
    "</div>" +
    "</header>";

  document.getElementById("demo-user-switch").addEventListener("change", function (evt) {
    Mock.setCurrentUser(evt.target.value);
    location.reload();
  });

  if (user) wireNotifBell(user);
  renderHelpFab();
}

function renderNotifBellHtml(user) {
  const count = Mock.unreadNotificationsCount(user.id);
  return (
    '<button type="button" id="notif-bell" class="notif-bell" title="Notifications" aria-label="Notifications">' +
    "🔔" +
    (count > 0 ? '<span class="notif-count">' + (count > 9 ? "9+" : count) + "</span>" : "") +
    "</button>"
  );
}

function wireNotifBell(user) {
  const btn = document.getElementById("notif-bell");
  if (!btn) return;
  btn.addEventListener("click", function () {
    const notifs = Mock.getNotifications(user.id);
    Popups.open(
      '<h3 class="popup-title">Notifications</h3>' +
      (notifs.length === 0
        ? '<div class="empty">Aucune notification pour le moment.</div>'
        : notifs.map(function (n) {
            return '<div class="thread-message">' +
              '<div class="bubble">' +
              '<div class="msg-meta"><span>' + timeAgo(n.date) + "</span></div>" +
              "<div>" + escapeHtml(n.message) + "</div>" +
              "</div>" +
              "</div>";
          }).join(""))
    );
    if (Mock.unreadNotificationsCount(user.id) > 0) {
      Mock.markNotificationsRead(user.id);
      btn.querySelector(".notif-count").remove();
    }
  });
}

function renderHelpFab() {
  if (document.getElementById("help-fab")) return;

  const fab = document.createElement("button");
  fab.type = "button";
  fab.id = "help-fab";
  fab.className = "help-fab";
  fab.title = "Contacter le staff";
  fab.setAttribute("aria-label", "Aide");
  fab.textContent = "? Aide";
  document.body.appendChild(fab);

  fab.addEventListener("click", openHelpForm);
}

function openHelpForm() {
  const user = Mock.getCurrentUser();
  const box = Popups.open(
    '<h3 class="popup-title">Contacter le staff</h3>' +
    '<p class="muted">Une question, un probleme sur le site ? Ecrivez-nous, un(e) administrateur(rice) vous repondra.</p>' +
    '<form id="help-form">' +
    (user
      ? ""
      : '<div class="grid grid-2">' +
        '<div class="field"><label for="help-nom">Votre nom</label><input id="help-nom" required></div>' +
        '<div class="field"><label for="help-email">Votre email</label><input id="help-email" type="email" required></div>' +
        "</div>") +
    '<div class="field"><label for="help-sujet">Sujet</label><input id="help-sujet" required></div>' +
    '<div class="field"><label for="help-message">Message</label><textarea id="help-message" required></textarea></div>' +
    '<div class="actions"><button type="submit">Envoyer</button><button type="button" class="btn-secondary" data-popup-close>Annuler</button></div>' +
    "</form>"
  );

  box.querySelector("#help-form").addEventListener("submit", function (evt) {
    evt.preventDefault();
    Mock.createTicket(user ? user.id : null, {
      sujet: document.getElementById("help-sujet").value,
      message: document.getElementById("help-message").value,
      contactNom: user ? null : document.getElementById("help-nom").value,
      contactEmail: user ? null : document.getElementById("help-email").value,
    });
    Popups.close();
    Popups.toast("Message envoye au staff, merci !");
  });
}

function renderTopNav(active) {
  const user = Mock.getCurrentUser();
  const links = [
    { href: "index.html", label: "Accueil", key: "accueil" },
    { href: "forums.html", label: "Forums", key: "forums" },
    { href: "cuisine.html", label: "Cuisine", key: "cuisine" },
    { href: "annonces.html", label: "Annonces", key: "annonces" },
    { href: "profil.html", label: "Mon profil", key: "profil" },
  ];
  if (Mock.isAdmin(user)) {
    links.push({ href: "../backoffice/index.html", label: "Back-office", key: "backoffice" });
  }
  renderNavShell("index.html", links.map(function (l) { return { href: l.href, label: l.label, active: l.key === active }; }));
}

function renderBackofficeNav(active) {
  const links = [
    { href: "index.html", label: "Tableau de bord", key: "dashboard" },
    { href: "comptes.html", label: "Comptes", key: "comptes" },
    { href: "forums.html", label: "Forums", key: "forums" },
    { href: "annonces.html", label: "Annonces", key: "annonces" },
    { href: "collectes.html", label: "Collectes", key: "collectes" },
    { href: "distributions.html", label: "Distributions", key: "distributions" },
    { href: "candidatures.html", label: "Candidatures", key: "candidatures" },
    { href: "recettes.html", label: "Recettes", key: "recettes" },
    { href: "signalements.html", label: "Signalements", key: "signalements" },
    { href: "tickets.html", label: "Tickets", key: "tickets" },
    { href: "../frontoffice/index.html", label: "Retour au site", key: "front" },
  ];
  renderNavShell("index.html", links.map(function (l) { return { href: l.href, label: l.label, active: l.key === active }; }));
}

function guardBackoffice(contentId) {
  const user = Mock.getCurrentUser();
  if (Mock.isAdmin(user)) return true;

  const el = document.getElementById(contentId);
  if (el) {
    const who = user ? escapeHtml(Mock.userLabel(user)) + "</strong> (" + userTypeLabel(user.type) + ")" : "Invité(e) (non connecté(e))</strong>";
    el.innerHTML =
      '<div class="access-denied">' +
      "<h2>Acces reserve</h2>" +
      '<p class="muted">Le back-office est reserve aux comptes administrateur(rice). Le profil de demonstration actuel, ' +
      "<strong>" + who + ", n'y a pas acces.</p>" +
      '<a class="btn" href="../frontoffice/index.html">Retour au site</a>' +
      "</div>";
  }
  return false;
}

function requireLogin(contentId) {
  const user = Mock.getCurrentUser();
  if (user) return true;

  const el = document.getElementById(contentId);
  if (el) {
    el.innerHTML =
      '<div class="access-denied">' +
      "<h2>Connexion requise</h2>" +
      '<p class="muted">Cette page est reservee aux comptes identifies. Choisissez un profil de demonstration dans la barre de navigation, ou <a href="connexion-inscription.html">connectez-vous / inscrivez-vous</a>.</p>' +
      "</div>";
  }
  return false;
}

function timeAgo(timestamp) {
  const diff = Date.now() - timestamp;
  const hours = Math.floor(diff / 3600000);
  if (hours < 1) return "a l'instant";
  if (hours < 24) return "il y a " + hours + "h";
  const days = Math.floor(hours / 24);
  return "il y a " + days + " j";
}

function escapeHtml(value) {
  const div = document.createElement("div");
  div.textContent = value == null ? "" : String(value);
  return div.innerHTML;
}
