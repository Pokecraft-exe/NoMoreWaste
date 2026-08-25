const Mock = (function () {
  const state = MockData.getState();

  function save() {
    MockData.save();
  }

  function resetDemo() {
    MockData.resetDemo();
  }

  function getUsers() {
    return state.users;
  }

  function getUser(id) {
    return state.users.find(function (u) { return u.id === id; }) || null;
  }

  function getCurrentUser() {
    return state.currentUserId ? getUser(state.currentUserId) : null;
  }

  function setCurrentUser(id) {
    state.currentUserId = id ? Number(id) : null;
    save();
  }

  function isStaff(user) {
    return !!user && user.type === "administrateur";
  }

  function isAdmin(user) {
    return !!user && user.type === "administrateur";
  }

  function isBenevole(user) {
    return !!user && user.type === "benevole";
  }

  function isAdherent(user) {
    return !!user && user.type === "adherent";
  }

  function userLabel(user) {
    if (!user) return "Inconnu";
    return user.prenom + " " + user.nom;
  }

  function updateProfil(userId, fields) {
    const user = getUser(userId);
    if (!user) return null;
    Object.assign(user, fields);
    save();
    return user;
  }

  function getAdhesion(userId) {
    return state.adhesions[userId] || null;
  }

  function renewAdhesion(userId) {
    const current = state.adhesions[userId];
    const base = current && current.dateFin && new Date(current.dateFin) > new Date()
      ? new Date(current.dateFin)
      : new Date();
    base.setFullYear(base.getFullYear() + 1);
    state.adhesions[userId] = { statut: "active", dateFin: base.toISOString().slice(0, 10) };
    save();
    return state.adhesions[userId];
  }

  function trendScore(forum) {
    const ageDays = (Date.now() - forum.createdAt) / 86400000;
    const recency = Math.max(0, 14 - ageDays);
    return forum.vues + forum.messages.length * 25 + recency * 10;
  }

  function listForums() {
    return state.forums.slice().sort(function (a, b) { return b.createdAt - a.createdAt; });
  }

  function topTrending(n) {
    return state.forums.slice()
      .sort(function (a, b) { return trendScore(b) - trendScore(a); })
      .slice(0, n || 10);
  }

  function searchForums(query) {
    const q = (query || "").trim().toLowerCase();
    if (!q) return listForums();
    return listForums().filter(function (f) { return f.titre.toLowerCase().includes(q); });
  }

  function myForums(userId) {
    return state.forums.filter(function (f) { return f.auteurId === userId; });
  }

  function getForum(id) {
    return state.forums.find(function (f) { return f.id === Number(id); }) || null;
  }

  function recordForumView(id) {
    const forum = getForum(id);
    if (forum) { forum.vues += 1; save(); }
  }

  function canManageForum(forum, user) {
    return !!forum && !!user && (isStaff(user) || forum.auteurId === user.id);
  }

  function canDeleteForumMessage(message, user) {
    return !!message && !!user && (isAdmin(user) || message.auteurId === user.id);
  }

  function createForum(auteurId, titre, message) {
    const forum = {
      id: state.nextForumId++, titre: titre, message: message,
      auteurId: auteurId, vues: 1, createdAt: Date.now(), messages: [],
    };
    state.forums.unshift(forum);
    save();
    return forum;
  }

  function updateForum(id, fields) {
    const forum = getForum(id);
    if (!forum) return null;
    Object.assign(forum, fields);
    save();
    return forum;
  }

  function deleteForum(id) {
    state.forums = state.forums.filter(function (f) { return f.id !== Number(id); });
    save();
  }

  function addForumMessage(forumId, auteurId, message) {
    const forum = getForum(forumId);
    if (!forum) return null;
    const nextId = forum.messages.reduce(function (max, m) { return Math.max(max, m.id); }, 0) + 1;
    const entry = { id: nextId, auteurId: auteurId, message: message, date: Date.now() };
    forum.messages.push(entry);
    forum.vues += 2;
    save();
    return entry;
  }

  function deleteForumMessage(forumId, messageId) {
    const forum = getForum(forumId);
    if (!forum) return;
    forum.messages = forum.messages.filter(function (m) { return m.id !== Number(messageId); });
    save();
  }

  function listRecettes() {
    return state.recettes;
  }

  function getRecette(id) {
    return state.recettes.find(function (r) { return r.id === Number(id); }) || null;
  }

  function searchRecettes(filters) {
    filters = filters || {};
    const ing = (filters.ingredient || "").trim().toLowerCase();
    const outil = (filters.outil || "").trim().toLowerCase();
    const nom = (filters.nom || "").trim().toLowerCase();

    return state.recettes.filter(function (r) {
      const matchIng = !ing || r.ingredients.some(function (i) { return i.toLowerCase().includes(ing); });
      const matchOutil = !outil || r.outils.some(function (o) { return o.toLowerCase().includes(outil); });
      const matchNom = !nom || r.titre.toLowerCase().includes(nom);
      return matchIng && matchOutil && matchNom;
    });
  }

  function listAnnonces(filters) {
    filters = filters || {};
    let list = state.annonces.slice().sort(function (a, b) { return b.createdAt - a.createdAt; });
    if (filters.categorie) {
      list = list.filter(function (a) { return a.categorie === filters.categorie; });
    }
    if (filters.query) {
      const q = filters.query.trim().toLowerCase();
      list = list.filter(function (a) { return a.titre.toLowerCase().includes(q) || a.description.toLowerCase().includes(q); });
    }
    return list;
  }

  function myAnnonces(userId) {
    return state.annonces.filter(function (a) { return a.auteurId === userId; });
  }

  function getAnnonce(id) {
    return state.annonces.find(function (a) { return a.id === Number(id); }) || null;
  }

  function createAnnonce(auteurId, fields) {
    const annonce = {
      id: state.nextAnnonceId++, auteurId: auteurId, etat: "disponible",
      createdAt: Date.now(), messages: [],
      categorie: fields.categorie, titre: fields.titre,
      description: fields.description, prix: fields.prix,
    };
    state.annonces.unshift(annonce);
    save();
    return annonce;
  }

  function updateAnnonce(id, fields) {
    const annonce = getAnnonce(id);
    if (!annonce) return null;
    Object.assign(annonce, fields);
    save();
    return annonce;
  }

  function deleteAnnonce(id) {
    state.annonces = state.annonces.filter(function (a) { return a.id !== Number(id); });
    save();
  }

  function sendAnnonceMessage(annonceId, expediteurId, message, prixPropose) {
    const annonce = getAnnonce(annonceId);
    if (!annonce) return null;
    const nextId = annonce.messages.reduce(function (max, m) { return Math.max(max, m.id); }, 0) + 1;
    const entry = { id: nextId, auteurId: expediteurId, message: message, prixPropose: prixPropose || null, date: Date.now() };
    annonce.messages.push(entry);
    if (prixPropose && annonce.etat === "disponible") annonce.etat = "en_negociation";
    save();
    return entry;
  }

  function listCollectes() {
    return state.collectes.slice().sort(function (a, b) { return a.date < b.date ? 1 : -1; });
  }

  function recentCollectes(n) {
    return listCollectes().slice(0, n || 5);
  }

  function getNextCollecte() {
    const today = new Date().toISOString().slice(0, 10);
    const upcoming = state.collectes
      .filter(function (c) { return c.date >= today && c.statut !== "annulee"; })
      .sort(function (a, b) { return a.date < b.date ? -1 : 1; });
    return upcoming[0] || listCollectes()[0] || null;
  }

  function getCollecte(id) {
    return state.collectes.find(function (c) { return c.id === Number(id); }) || null;
  }

  function searchCollectes(filters) {
    filters = filters || {};
    const q = (filters.q || "").trim().toLowerCase();
    return listCollectes().filter(function (c) {
      const matchQ = !q || c.lieu.toLowerCase().includes(q) || (c.partenaire || "").toLowerCase().includes(q);
      const matchStatut = !filters.statut || c.statut === filters.statut;
      const matchDate = !filters.date || c.date === filters.date;
      return matchQ && matchStatut && matchDate;
    });
  }

  function createCollecte(fields) {
    const collecte = {
      id: state.nextCollecteId++,
      lieu: fields.lieu, date: fields.date, heure: fields.heure || "",
      partenaire: fields.partenaire || null,
      commercantId: fields.commercantId || null,
      statut: fields.statut || "planifiee",
      description: fields.description || "",
      denrees: [], benevolesAffectes: [], stockMisAJour: false,
    };
    state.collectes.push(collecte);
    save();
    return collecte;
  }

  function updateCollecte(id, fields) {
    const collecte = getCollecte(id);
    if (!collecte) return null;
    Object.assign(collecte, fields);
    save();
    return collecte;
  }

  function deleteCollecte(id) {
    state.collectes = state.collectes.filter(function (c) { return c.id !== Number(id); });
    save();
  }

  function countCollectesLast7Days() {
    const today = new Date().toISOString().slice(0, 10);
    const weekAgo = MockData.isoDate(Date.now(), -7);
    return state.collectes.filter(function (c) { return c.date >= weekAgo && c.date <= today; }).length;
  }

  function setCollecteAffectations(id, affectations) {
    return updateCollecte(id, { benevolesAffectes: affectations });
  }

  function searchStockProduits(query) {
    const q = (query || "").trim().toLowerCase();
    const list = state.stocks.slice().sort(function (a, b) { return a.nom.localeCompare(b.nom); });
    if (!q) return list;
    return list.filter(function (s) { return s.nom.toLowerCase().includes(q); });
  }

  function levenshtein(a, b) {
    a = (a || "").toLowerCase();
    b = (b || "").toLowerCase();
    const m = a.length;
    const n = b.length;
    const dp = [];
    for (let i = 0; i <= m; i++) dp.push([i].concat(new Array(n).fill(0)));
    for (let j = 0; j <= n; j++) dp[0][j] = j;
    for (let i = 1; i <= m; i++) {
      for (let j = 1; j <= n; j++) {
        const cost = a[i - 1] === b[j - 1] ? 0 : 1;
        dp[i][j] = Math.min(
          dp[i - 1][j] + 1,
          dp[i][j - 1] + 1,
          dp[i - 1][j - 1] + cost
        );
        if (i > 1 && j > 1 && a[i - 1] === b[j - 2] && a[i - 2] === b[j - 1]) {
          dp[i][j] = Math.min(dp[i][j], dp[i - 2][j - 2] + 1);
        }
      }
    }
    return dp[m][n];
  }

  function findSimilarStockProduit(nom) {
    nom = (nom || "").trim();
    if (!nom) return null;
    let best = null;
    let bestDist = Infinity;
    state.stocks.forEach(function (s) {
      const dist = levenshtein(s.nom, nom);
      if (dist < bestDist) { bestDist = dist; best = s; }
    });
    const threshold = Math.max(1, Math.floor(nom.length * 0.3));
    return best && bestDist > 0 && bestDist <= threshold ? best : null;
  }

  function createStockProduit(nom, unite) {
    nom = MockData.normalizeProduitNom(nom);
    if (!nom) return null;
    const existing = state.stocks.find(function (s) { return s.nom.toLowerCase() === nom.toLowerCase(); });
    if (existing) return existing;
    const produit = { id: state.nextStockId++, nom: nom, unite: unite || "unites", quantiteDisponible: 0 };
    state.stocks.push(produit);
    save();
    return produit;
  }

  function isBenevoleConfirmeCollecte(user, collecte) {
    return !!user && !!collecte && user.type === "benevole" &&
      collecte.benevolesAffectes.some(function (a) { return a.benevoleId === user.id; });
  }

  function canManageCollecteDenrees(user, collecte) {
    return !!user && !!collecte && (isStaff(user) || isBenevoleConfirmeCollecte(user, collecte));
  }

  function canProposeCollecteDenrees(user, collecte) {
    return !!user && !!collecte && user.type === "commercant" && collecte.commercantId === user.id;
  }

  function ajouterDenreeCollecte(collecteId, compteId, proposeParType, fields) {
    const collecte = getCollecte(collecteId);
    if (!collecte) return null;
    const entry = {
      id: state.nextCollecteDenreeId++,
      produitId: Number(fields.produitId),
      quantite: Number(fields.quantite),
      nonPerissable: !!fields.nonPerissable,
      dlc: fields.nonPerissable ? null : (fields.dlc || null),
      proposePar: compteId,
      proposeParType: proposeParType,
      confirmee: proposeParType === "staff",
      dateAjout: Date.now(),
    };
    collecte.denrees.push(entry);
    save();
    return entry;
  }

  function toggleDenreeConfirmee(collecteId, denreeId, confirmee) {
    const collecte = getCollecte(collecteId);
    if (!collecte) return;
    const denree = collecte.denrees.find(function (d) { return d.id === Number(denreeId); });
    if (denree) { denree.confirmee = !!confirmee; save(); }
  }

  function supprimerDenreeCollecte(collecteId, denreeId) {
    const collecte = getCollecte(collecteId);
    if (!collecte) return;
    collecte.denrees = collecte.denrees.filter(function (d) { return d.id !== Number(denreeId); });
    save();
  }

  function confirmerCollecte(collecteId) {
    const collecte = getCollecte(collecteId);
    if (!collecte) return { ok: false, error: "Collecte introuvable." };
    if (collecte.stockMisAJour) return { ok: false, error: "Cette collecte a deja ete confirmee." };

    let updated = 0;
    collecte.denrees.forEach(function (d) {
      if (!d.confirmee) return;
      const produit = getStockProduit(d.produitId);
      if (produit) {
        produit.quantiteDisponible += d.quantite;
        updated++;
      }
    });
    collecte.statut = "terminee";
    collecte.stockMisAJour = true;
    save();
    return { ok: true, updated: updated };
  }

  function listDistributions() {
    return state.distributions.slice().sort(function (a, b) { return a.date < b.date ? 1 : -1; });
  }

  function recentDistributions(n) {
    return listDistributions().slice(0, n || 5);
  }

  function getNextDistribution() {
    const today = new Date().toISOString().slice(0, 10);
    const upcoming = state.distributions
      .filter(function (d) { return d.date >= today && d.statut !== "annulee"; })
      .sort(function (a, b) { return a.date < b.date ? -1 : 1; });
    return upcoming[0] || listDistributions()[0] || null;
  }

  function getDistribution(id) {
    return state.distributions.find(function (d) { return d.id === Number(id); }) || null;
  }

  function searchDistributions(filters) {
    filters = filters || {};
    const q = (filters.q || "").trim().toLowerCase();
    return listDistributions().filter(function (d) {
      const matchQ = !q || d.lieu.toLowerCase().includes(q);
      const matchStatut = !filters.statut || d.statut === filters.statut;
      return matchQ && matchStatut;
    });
  }

  function createDistribution(fields) {
    const distribution = {
      id: state.nextDistributionId++,
      lieu: fields.lieu, date: fields.date, heure: fields.heure || "",
      statut: fields.statut || "planifiee",
      denrees: [], quotaParAdherent: Number(fields.quotaParAdherent) || 0,
      benevolesParticipants: [], benevolesAffectes: [], reservations: [],
    };
    state.distributions.push(distribution);
    save();
    return distribution;
  }

  function updateDistribution(id, fields) {
    const distribution = getDistribution(id);
    if (!distribution) return null;
    Object.assign(distribution, fields);
    save();
    return distribution;
  }

  function deleteDistribution(id) {
    state.distributions = state.distributions.filter(function (d) { return d.id !== Number(id); });
    save();
  }

  function countDistributionsLast7Days() {
    const today = new Date().toISOString().slice(0, 10);
    const weekAgo = MockData.isoDate(Date.now(), -7);
    return state.distributions.filter(function (d) { return d.date >= weekAgo && d.date <= today; }).length;
  }

  function setDistributionDenrees(id, denrees) {
    return updateDistribution(id, { denrees: denrees });
  }

  function setDistributionQuota(id, quota) {
    return updateDistribution(id, { quotaParAdherent: Number(quota) });
  }

  function setDistributionAffectations(id, affectations) {
    return updateDistribution(id, { benevolesAffectes: affectations });
  }

  function participerDistribution(benevoleId, distributionId) {
    const distribution = getDistribution(distributionId);
    if (!distribution) return;
    if (distribution.benevolesParticipants.indexOf(benevoleId) === -1) {
      distribution.benevolesParticipants.push(benevoleId);
      save();
    }
  }

  function hasParticipationDistribution(benevoleId, distributionId) {
    const distribution = getDistribution(distributionId);
    return !!distribution && distribution.benevolesParticipants.indexOf(benevoleId) !== -1;
  }

  function getDistributionReservedQty(distribution, produitId) {
    return distribution.reservations
      .filter(function (r) { return r.produitId === produitId; })
      .reduce(function (sum, r) { return sum + r.quantite; }, 0);
  }

  function getDistributionRemainingQty(distribution, produitId) {
    const denree = distribution.denrees.find(function (d) { return d.produitId === produitId; });
    if (!denree) return 0;
    return Math.max(0, denree.quantite - getDistributionReservedQty(distribution, produitId));
  }

  function getAdherentReservedTotal(distribution, adherentId) {
    return distribution.reservations
      .filter(function (r) { return r.adherentId === adherentId; })
      .reduce(function (sum, r) { return sum + r.quantite; }, 0);
  }

  function reserverStock(adherentId, distributionId, produitId, quantite) {
    const distribution = getDistribution(distributionId);
    if (!distribution) return { ok: false, error: "Distribution introuvable." };

    quantite = Number(quantite);
    if (!quantite || quantite <= 0) return { ok: false, error: "Quantite invalide." };

    const remaining = getDistributionRemainingQty(distribution, produitId);
    if (quantite > remaining) return { ok: false, error: "Il ne reste que " + remaining + " unite(s) disponible(s) pour ce produit." };

    const dejaReserve = getAdherentReservedTotal(distribution, adherentId);
    const quotaRestant = distribution.quotaParAdherent - dejaReserve;
    if (quantite > quotaRestant) return { ok: false, error: "Quota depasse : il vous reste " + quotaRestant + " unite(s) sur votre quota pour cette distribution." };

    distribution.reservations.push({ id: state.nextReservationId++, adherentId: adherentId, produitId: produitId, quantite: quantite, date: Date.now() });
    save();
    return { ok: true };
  }

  function listStocks() {
    return state.stocks;
  }

  function getStockProduit(id) {
    return state.stocks.find(function (s) { return s.id === Number(id); }) || null;
  }

  function listCandidatures() {
    return state.candidaturesBenevole.slice().sort(function (a, b) { return b.date - a.date; });
  }

  function getCandidature(id) {
    return state.candidaturesBenevole.find(function (c) { return c.id === Number(id); }) || null;
  }

  function searchCandidatures(filters) {
    filters = filters || {};
    return listCandidatures().filter(function (c) {
      const personne = getUser(c.personneId);
      const q = (filters.q || "").trim().toLowerCase();
      const matchQ = !q || !personne || userLabel(personne).toLowerCase().includes(q);
      const matchStatut = !filters.statut || c.statut === filters.statut;
      return matchQ && matchStatut;
    });
  }

  function pendingCandidaturesCount() {
    return state.candidaturesBenevole.filter(function (c) { return c.statut === "recue" || c.statut === "en_etude"; }).length;
  }

  function decideCandidature(id, statut, commentaire) {
    const candidature = getCandidature(id);
    if (!candidature) return null;
    candidature.statut = statut;
    if (commentaire) candidature.commentaire = commentaire;
    if (statut === "validee") {
      const user = getUser(candidature.personneId);
      if (user) user.type = "benevole";
    }
    save();
    return candidature;
  }

  function deleteCandidature(id) {
    state.candidaturesBenevole = state.candidaturesBenevole.filter(function (c) { return c.id !== Number(id); });
    save();
  }

  function createCandidatureAdmin(personneId, fields) {
    const candidature = {
      id: state.nextCandidatureId++, personneId: personneId, statut: "recue",
      date: Date.now(), motivation: fields.motivation || "", disponibilite: fields.disponibilite || "", commentaire: null,
    };
    state.candidaturesBenevole.unshift(candidature);
    save();
    return candidature;
  }

  function candidaterBenevole(userId, fields) {
    const candidature = {
      id: state.nextCandidatureId++, personneId: userId, statut: "recue",
      date: Date.now(), motivation: fields.motivation, disponibilite: fields.disponibilite, commentaire: null,
    };
    state.candidaturesBenevole.unshift(candidature);
    const user = getUser(userId);
    if (user) user.type = "benevole";
    save();
    return candidature;
  }

  function participerCollecte(userId, collecteId, fields) {
    state.participationsCollecte.push({ userId: userId, collecteId: collecteId, date: Date.now(), fields: fields });
    save();
  }

  function hasParticipation(userId, collecteId) {
    return state.participationsCollecte.some(function (p) { return p.userId === userId && p.collecteId === collecteId; });
  }

  function getCollecteParticipants(collecteId) {
    return state.participationsCollecte
      .filter(function (p) { return p.collecteId === Number(collecteId); })
      .map(function (p) { return p.userId; });
  }

  function searchComptes(filters) {
    filters = filters || {};
    const q = (filters.q || "").trim().toLowerCase();
    return state.users.filter(function (u) {
      const matchQ = !q || userLabel(u).toLowerCase().includes(q) || u.email.toLowerCase().includes(q);
      const matchType = !filters.type || u.type === filters.type;
      return matchQ && matchType;
    });
  }

  function createCompte(fields) {
    const nextId = state.users.reduce(function (max, u) { return Math.max(max, u.id); }, 0) + 1;
    const user = {
      id: nextId,
      prenom: fields.prenom, nom: fields.nom, email: fields.email,
      type: fields.type || "visiteur",
      telephone: fields.telephone || "", adresse: fields.adresse || "",
      codePostal: fields.codePostal || "", ville: fields.ville || "",
      actif: true,
    };
    state.users.push(user);
    save();
    return user;
  }

  function updateCompteAdmin(id, fields) {
    const user = getUser(Number(id));
    if (!user) return null;
    Object.assign(user, fields);
    save();
    return user;
  }

  function setCompteActif(id, actif) {
    return updateCompteAdmin(id, { actif: actif });
  }

  function accountsByType() {
    const counts = {};
    state.users.forEach(function (u) { counts[u.type] = (counts[u.type] || 0) + 1; });
    return counts;
  }

  function adhesionsExpiringSoon(days) {
    const limit = Date.now() + (days || 30) * 86400000;
    const result = [];
    Object.keys(state.adhesions).forEach(function (userId) {
      const adhesion = state.adhesions[userId];
      const dateFin = new Date(adhesion.dateFin).getTime();
      if (dateFin <= limit) {
        result.push({ user: getUser(Number(userId)), adhesion: adhesion });
      }
    });
    return result.sort(function (a, b) { return a.adhesion.dateFin < b.adhesion.dateFin ? -1 : 1; });
  }

  function createRecette(fields) {
    const nextId = state.recettes.reduce(function (max, r) { return Math.max(max, r.id); }, 0) + 1;
    const recette = {
      id: nextId, titre: fields.titre,
      ingredients: fields.ingredients || [], outils: fields.outils || [],
      contenu: fields.contenu || "", video: fields.video || null,
    };
    state.recettes.push(recette);
    save();
    return recette;
  }

  function updateRecette(id, fields) {
    const recette = getRecette(id);
    if (!recette) return null;
    Object.assign(recette, fields);
    save();
    return recette;
  }

  function deleteRecette(id) {
    state.recettes = state.recettes.filter(function (r) { return r.id !== Number(id); });
    save();
  }

  function countForumsLast7Days() {
    const weekAgo = Date.now() - 7 * 86400000;
    return state.forums.filter(function (f) { return f.createdAt >= weekAgo; }).length;
  }

  function countAnnoncesLast7Days() {
    const weekAgo = Date.now() - 7 * 86400000;
    return state.annonces.filter(function (a) { return a.createdAt >= weekAgo; }).length;
  }

  function pushNotification(userId, fields) {
    const notif = {
      id: state.nextNotificationId++, userId: userId,
      type: fields.type, message: fields.message, lien: fields.lien || null,
      date: Date.now(), lu: false,
    };
    state.notifications.push(notif);
    return notif;
  }

  function getNotifications(userId) {
    return state.notifications
      .filter(function (n) { return n.userId === userId; })
      .sort(function (a, b) { return b.date - a.date; });
  }

  function unreadNotificationsCount(userId) {
    return state.notifications.filter(function (n) { return n.userId === userId && !n.lu; }).length;
  }

  function markNotificationsRead(userId) {
    state.notifications.forEach(function (n) { if (n.userId === userId) n.lu = true; });
    save();
  }

  function createTicket(auteurId, fields) {
    const ticket = {
      id: state.nextTicketId++, auteurId: auteurId || null,
      contactNom: fields.contactNom || null, contactEmail: fields.contactEmail || null,
      sujet: fields.sujet, message: fields.message,
      statut: "ouvert", reponse: null, date: Date.now(),
      traitePar: null, traiteAt: null,
    };
    state.tickets.push(ticket);
    save();
    return ticket;
  }

  function listTickets() {
    return state.tickets.slice().sort(function (a, b) { return b.date - a.date; });
  }

  function getTicket(id) {
    return state.tickets.find(function (t) { return t.id === Number(id); }) || null;
  }

  function searchTickets(filters) {
    filters = filters || {};
    const q = (filters.q || "").trim().toLowerCase();
    return listTickets().filter(function (t) {
      const auteur = getUser(t.auteurId);
      const matchQ = !q || t.sujet.toLowerCase().includes(q) || t.message.toLowerCase().includes(q) ||
        (auteur && userLabel(auteur).toLowerCase().includes(q));
      const matchStatut = !filters.statut || t.statut === filters.statut;
      return matchQ && matchStatut;
    });
  }

  function pendingTicketsCount() {
    return state.tickets.filter(function (t) { return t.statut === "ouvert"; }).length;
  }

  function repondreTicket(id, adminId, reponse) {
    const ticket = getTicket(id);
    if (!ticket) return null;
    ticket.statut = "traite";
    ticket.reponse = reponse;
    ticket.traitePar = adminId;
    ticket.traiteAt = Date.now();
    if (ticket.auteurId) {
      pushNotification(ticket.auteurId, {
        type: "ticket_traite",
        message: "Reponse du staff a votre message \"" + ticket.sujet + "\" : " + reponse,
      });
    }
    save();
    return ticket;
  }

  function createSignalement(type, refs, signalePar, motif) {
    const signalement = {
      id: state.nextSignalementId++, type: type,
      forumId: refs.forumId != null ? refs.forumId : null,
      messageId: refs.messageId != null ? refs.messageId : null,
      annonceId: refs.annonceId != null ? refs.annonceId : null,
      signalePar: signalePar, motif: motif || "",
      date: Date.now(), statut: "ouvert",
      commentaire: null, traitePar: null, traiteAt: null,
    };
    state.signalements.push(signalement);
    save();
    return signalement;
  }

  function listSignalements() {
    return state.signalements.slice().sort(function (a, b) { return b.date - a.date; });
  }

  function getSignalement(id) {
    return state.signalements.find(function (s) { return s.id === Number(id); }) || null;
  }

  function searchSignalements(filters) {
    filters = filters || {};
    return listSignalements().filter(function (s) {
      return !filters.statut || s.statut === filters.statut;
    });
  }

  function pendingSignalementsCount() {
    return state.signalements.filter(function (s) { return s.statut === "ouvert"; }).length;
  }

  function getSignalementDiscussion(signalement) {
    if (signalement.type === "annonce_message") {
      const annonce = getAnnonce(signalement.annonceId);
      return { annonce: annonce, messages: annonce ? annonce.messages : [] };
    }
    const forum = getForum(signalement.forumId);
    return { forum: forum, messages: forum ? forum.messages : [] };
  }

  function getSignalementFlaggedContent(signalement) {
    if (signalement.type === "annonce_message") {
      const annonce = getAnnonce(signalement.annonceId);
      const m = annonce ? annonce.messages.find(function (mm) { return mm.id === signalement.messageId; }) : null;
      return m ? { auteurId: m.auteurId, message: m.message, date: m.date } : null;
    }
    if (signalement.type === "forum_message") {
      const forum = getForum(signalement.forumId);
      const m = forum ? forum.messages.find(function (mm) { return mm.id === signalement.messageId; }) : null;
      return m ? { auteurId: m.auteurId, message: m.message, date: m.date } : null;
    }
    const forum = getForum(signalement.forumId);
    return forum ? { auteurId: forum.auteurId, message: forum.message, date: forum.createdAt } : null;
  }

  function archiveSignalementMessage(signalement) {
    const content = getSignalementFlaggedContent(signalement);
    if (!content) return null;
    const archive = {
      id: state.nextMessageArchiveId++, signalementId: signalement.id,
      auteurId: content.auteurId, message: content.message, date: content.date,
      dateArchivage: Date.now(),
    };
    state.messagesArchives.push(archive);
    return archive;
  }

  function getArchivedMessage(signalementId) {
    return state.messagesArchives.find(function (a) { return a.signalementId === Number(signalementId); }) || null;
  }

  function resoudreSignalement(id, adminId, commentaire) {
    const signalement = getSignalement(id);
    if (!signalement) return null;
    const wasAlreadyTraite = signalement.statut === "traite";
    signalement.statut = "traite";
    signalement.commentaire = commentaire;
    signalement.traitePar = adminId;
    signalement.traiteAt = Date.now();
    if (!wasAlreadyTraite) {
      archiveSignalementMessage(signalement);
      pushNotification(signalement.signalePar, {
        type: "signalement_traite",
        message: "Votre signalement a ete traite : " + commentaire,
      });
    }
    save();
    return signalement;
  }

  return {
    CATEGORIES_ANNONCE: MockData.CATEGORIES_ANNONCE,
    ETATS_ANNONCE: MockData.ETATS_ANNONCE,
    UNITES_PRODUIT: MockData.UNITES_PRODUIT,
    resetDemo: resetDemo,

    getUsers: getUsers, getUser: getUser, getCurrentUser: getCurrentUser, setCurrentUser: setCurrentUser,
    isStaff: isStaff, isAdmin: isAdmin, isBenevole: isBenevole, isAdherent: isAdherent, userLabel: userLabel,

    updateProfil: updateProfil, getAdhesion: getAdhesion, renewAdhesion: renewAdhesion,

    listForums: listForums, topTrending: topTrending, searchForums: searchForums, myForums: myForums,
    getForum: getForum, recordForumView: recordForumView, canManageForum: canManageForum,
    canDeleteForumMessage: canDeleteForumMessage,
    createForum: createForum, updateForum: updateForum, deleteForum: deleteForum,
    addForumMessage: addForumMessage, deleteForumMessage: deleteForumMessage,

    listRecettes: listRecettes, getRecette: getRecette, searchRecettes: searchRecettes,

    listAnnonces: listAnnonces, myAnnonces: myAnnonces, getAnnonce: getAnnonce,
    createAnnonce: createAnnonce, updateAnnonce: updateAnnonce, deleteAnnonce: deleteAnnonce,
    sendAnnonceMessage: sendAnnonceMessage,

    listCollectes: listCollectes, recentCollectes: recentCollectes, getNextCollecte: getNextCollecte,
    getCollecte: getCollecte, searchCollectes: searchCollectes, createCollecte: createCollecte,
    updateCollecte: updateCollecte,
    deleteCollecte: deleteCollecte, countCollectesLast7Days: countCollectesLast7Days,
    setCollecteAffectations: setCollecteAffectations,

    searchStockProduits: searchStockProduits, createStockProduit: createStockProduit,
    findSimilarStockProduit: findSimilarStockProduit,

    isBenevoleConfirmeCollecte: isBenevoleConfirmeCollecte, canManageCollecteDenrees: canManageCollecteDenrees,
    canProposeCollecteDenrees: canProposeCollecteDenrees, ajouterDenreeCollecte: ajouterDenreeCollecte,
    toggleDenreeConfirmee: toggleDenreeConfirmee, supprimerDenreeCollecte: supprimerDenreeCollecte,
    confirmerCollecte: confirmerCollecte,

    listDistributions: listDistributions, recentDistributions: recentDistributions,
    getNextDistribution: getNextDistribution, getDistribution: getDistribution,
    searchDistributions: searchDistributions, createDistribution: createDistribution,
    updateDistribution: updateDistribution,
    deleteDistribution: deleteDistribution, countDistributionsLast7Days: countDistributionsLast7Days,
    setDistributionDenrees: setDistributionDenrees, setDistributionQuota: setDistributionQuota,
    setDistributionAffectations: setDistributionAffectations,
    participerDistribution: participerDistribution, hasParticipationDistribution: hasParticipationDistribution,
    getDistributionReservedQty: getDistributionReservedQty, getDistributionRemainingQty: getDistributionRemainingQty,
    getAdherentReservedTotal: getAdherentReservedTotal, reserverStock: reserverStock,
    listStocks: listStocks, getStockProduit: getStockProduit,

    listCandidatures: listCandidatures, getCandidature: getCandidature, searchCandidatures: searchCandidatures,
    pendingCandidaturesCount: pendingCandidaturesCount, decideCandidature: decideCandidature,
    deleteCandidature: deleteCandidature, candidaterBenevole: candidaterBenevole,
    createCandidatureAdmin: createCandidatureAdmin,

    participerCollecte: participerCollecte, hasParticipation: hasParticipation,
    getCollecteParticipants: getCollecteParticipants,

    searchComptes: searchComptes, createCompte: createCompte, updateCompteAdmin: updateCompteAdmin,
    setCompteActif: setCompteActif,
    accountsByType: accountsByType, adhesionsExpiringSoon: adhesionsExpiringSoon,
    createRecette: createRecette, updateRecette: updateRecette, deleteRecette: deleteRecette,
    countForumsLast7Days: countForumsLast7Days, countAnnoncesLast7Days: countAnnoncesLast7Days,

    getNotifications: getNotifications, unreadNotificationsCount: unreadNotificationsCount,
    markNotificationsRead: markNotificationsRead,

    createTicket: createTicket, listTickets: listTickets, getTicket: getTicket,
    searchTickets: searchTickets, pendingTicketsCount: pendingTicketsCount, repondreTicket: repondreTicket,

    createSignalement: createSignalement, listSignalements: listSignalements,
    getSignalement: getSignalement, searchSignalements: searchSignalements,
    pendingSignalementsCount: pendingSignalementsCount, getSignalementDiscussion: getSignalementDiscussion,
    getSignalementFlaggedContent: getSignalementFlaggedContent, getArchivedMessage: getArchivedMessage,
    resoudreSignalement: resoudreSignalement,
  };
})();
