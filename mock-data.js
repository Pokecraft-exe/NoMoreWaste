const MockData = (function () {
  const STORAGE_KEY = "nmw_mock_v7";

  const CATEGORIES_ANNONCE = [
    { value: "covoiturage", label: "Covoiturage" },
    { value: "reparation", label: "Reparation" },
    { value: "gardiennage", label: "Gardiennage" },
    { value: "location", label: "Location" },
    { value: "vente", label: "Vente" },
    { value: "don", label: "Don" },
  ];

  const ETATS_ANNONCE = [
    { value: "disponible", label: "Disponible", tone: "success" },
    { value: "en_negociation", label: "En negociation", tone: "warning" },
    { value: "reservee", label: "Reservee", tone: "warning" },
    { value: "terminee", label: "Terminee", tone: "muted" },
  ];

  const UNITES_PRODUIT = ["kg", "litres", "unites", "boites", "paquets"];

  function isoDate(now, offsetDays) {
    return new Date(now + offsetDays * 86400000).toISOString().slice(0, 10);
  }

  function normalizeProduitNom(nom) {
    return (nom || "")
      .trim()
      .replace(/\s+/g, " ")
      .split(" ")
      .map(function (word) {
        return word ? word.charAt(0).toUpperCase() + word.slice(1).toLowerCase() : word;
      })
      .join(" ");
  }

  function seed() {
    const now = Date.now();
    const day = 24 * 60 * 60 * 1000;

    return {
      currentUserId: 2,
      users: [
        { id: 1, prenom: "Camille", nom: "Petit", email: "camille.petit@example.com", type: "visiteur", telephone: "+33611111111", adresse: "12 rue des Lilas", codePostal: "75011", ville: "Paris", actif: true },
        { id: 2, prenom: "Julie", nom: "Martin", email: "julie.martin@example.com", type: "adherent", telephone: "+33622222222", adresse: "4 avenue Victor Hugo", codePostal: "75016", ville: "Paris", actif: true },
        { id: 3, prenom: "Karim", nom: "Haddad", email: "karim.haddad@example.com", type: "benevole", telephone: "+33633333333", adresse: "8 rue de la Paix", codePostal: "44000", ville: "Nantes", actif: true },
        { id: 4, prenom: "Boulangerie", nom: "du Coin", email: "contact@boulangerieducoin.example.com", type: "commercant", telephone: "+33644444444", adresse: "2 place du Marche", codePostal: "69001", ville: "Lyon", actif: true },
        { id: 5, prenom: "Lea", nom: "Dubois", email: "lea.dubois@nomorewaste.local", type: "administrateur", telephone: "+33655555555", adresse: "Siege NO MORE WASTE", codePostal: "75001", ville: "Paris", actif: true },
        { id: 6, prenom: "Admin", nom: "NoMoreWaste", email: "admin@nomorewaste.local", type: "administrateur", telephone: "+33600000000", adresse: "Siege NO MORE WASTE", codePostal: "75001", ville: "Paris", actif: true },
      ],

      adhesions: {
        2: { statut: "active", dateFin: "2026-11-30" },
      },

      collectes: [
        { id: 1, lieu: "Marche de Belleville, Paris", date: isoDate(now, 12), heure: "08:30", partenaire: "Croix-Rouge", commercantId: null, statut: "planifiee", description: "Collecte hebdomadaire de fruits et legumes invendus.", denrees: [], benevolesAffectes: [], stockMisAJour: false },
        { id: 2, lieu: "Boulangerie du Coin, Lyon", date: isoDate(now, -4), heure: "18:00", partenaire: "Restos du Coeur", commercantId: 4, statut: "terminee", description: "Recuperation des invendus de fin de journee.", denrees: [{ id: 1, produitId: 1, quantite: 12, dlc: null, nonPerissable: false, proposePar: 4, proposeParType: "commercant", confirmee: true, dateAjout: now - 5 * day }], benevolesAffectes: [{ benevoleId: 3, role: "Collecte et tri" }], stockMisAJour: true },
        { id: 3, lieu: "Marche Wazemmes, Lille", date: isoDate(now, 19), heure: "07:00", partenaire: "Secours Populaire", commercantId: null, statut: "confirmee", description: "", denrees: [], benevolesAffectes: [], stockMisAJour: false },
        { id: 4, lieu: "Supermarche Bio, Nantes", date: isoDate(now, -9), heure: "17:30", partenaire: "Banque Alimentaire", commercantId: null, statut: "terminee", description: "", denrees: [], benevolesAffectes: [], stockMisAJour: false },
        { id: 5, lieu: "Marche de Belleville, Paris", date: isoDate(now, 26), heure: "08:30", partenaire: "Croix-Rouge", commercantId: null, statut: "planifiee", description: "", denrees: [], benevolesAffectes: [], stockMisAJour: false },
        { id: 6, lieu: "Place Bellecour, Lyon", date: isoDate(now, -14), heure: "16:00", partenaire: "Restos du Coeur", commercantId: null, statut: "annulee", description: "Annulee pour cause meteo.", denrees: [], benevolesAffectes: [], stockMisAJour: false },
        { id: 7, lieu: "Marche des Capucins, Bordeaux", date: isoDate(now, -2), heure: "09:00", partenaire: "Croix-Rouge", commercantId: null, statut: "terminee", description: "", denrees: [], benevolesAffectes: [], stockMisAJour: false },
        {
          id: 8, lieu: "Boulangerie du Coin, Lyon", date: isoDate(now, 5), heure: "17:30", partenaire: null, commercantId: 4, statut: "planifiee",
          description: "Collecte des invendus de la boulangerie avant fermeture.",
          denrees: [
            { id: 2, produitId: 2, quantite: 8, dlc: isoDate(now, 3), nonPerissable: false, proposePar: 4, proposeParType: "commercant", confirmee: false, dateAjout: now - 0.5 * day },
          ],
          benevolesAffectes: [{ benevoleId: 3, role: "Collecte et tri" }],
          stockMisAJour: false,
        },
      ],
      nextCollecteId: 9,
      nextCollecteDenreeId: 3,

      candidaturesBenevole: [
        { id: 1, personneId: 1, statut: "recue", date: now - 1 * day, motivation: "Je souhaite aider ponctuellement le week-end.", disponibilite: "Samedis", commentaire: null },
        { id: 2, personneId: 1, statut: "en_etude", date: now - 6 * day, motivation: "Interessee par le tri et la distribution.", disponibilite: "Semaine soir", commentaire: "Entretien telephonique prevu." },
      ],
      nextCandidatureId: 3,

      participationsCollecte: [
        { userId: 3, collecteId: 1, date: now - 2 * day, fields: { role: "chauffeur", commentaire: "" } },
        { userId: 3, collecteId: 2, date: now - 6 * day, fields: { role: "tri", commentaire: "" } },
        { userId: 3, collecteId: 8, date: now - 1 * day, fields: { role: "chauffeur", commentaire: "" } },
      ],

      stocks: [
        { id: 1, nom: "Conserves De Légumes", unite: "boites", quantiteDisponible: 240 },
        { id: 2, nom: "Pâtes", unite: "kg", quantiteDisponible: 180 },
        { id: 3, nom: "Riz", unite: "kg", quantiteDisponible: 150 },
        { id: 4, nom: "Fruits Frais", unite: "kg", quantiteDisponible: 90 },
        { id: 5, nom: "Produits D'hygiène", unite: "unites", quantiteDisponible: 60 },
        { id: 6, nom: "Lait", unite: "litres", quantiteDisponible: 120 },
        { id: 7, nom: "Miel", unite: "kg", quantiteDisponible: 25 },
      ],

      distributions: [
        {
          id: 1, lieu: "Centre social de Belleville, Paris", date: isoDate(now, 8), heure: "14:00", statut: "planifiee",
          denrees: [{ produitId: 1, quantite: 60 }, { produitId: 2, quantite: 40 }, { produitId: 4, quantite: 30 }],
          quotaParAdherent: 3,
          benevolesParticipants: [3],
          benevolesAffectes: [{ benevoleId: 3, role: "Accueil des beneficiaires" }],
          reservations: [{ id: 1, adherentId: 2, produitId: 1, quantite: 2, date: now - 1 * day }],
        },
        {
          id: 2, lieu: "Salle des fetes, Nantes", date: isoDate(now, -3), heure: "17:00", statut: "terminee",
          denrees: [{ produitId: 3, quantite: 50 }, { produitId: 6, quantite: 40 }],
          quotaParAdherent: 2,
          benevolesParticipants: [3],
          benevolesAffectes: [{ benevoleId: 3, role: "Gestion des stocks" }],
          reservations: [],
        },
        {
          id: 3, lieu: "Marche couvert, Lille", date: isoDate(now, 15), heure: "10:00", statut: "planifiee",
          denrees: [{ produitId: 1, quantite: 30 }, { produitId: 5, quantite: 20 }],
          quotaParAdherent: 4,
          benevolesParticipants: [],
          benevolesAffectes: [],
          reservations: [],
        },
      ],
      nextDistributionId: 4,
      nextReservationId: 2,

      forums: [
        {
          id: 1, titre: "Astuces pour conserver les legumes plus longtemps",
          auteurId: 2, message: "Je met mes carottes dans un torchon humide au frigo, elles tiennent 3 semaines de plus !",
          vues: 340, createdAt: now - 2 * day,
          messages: [
            { id: 1, auteurId: 3, message: "Super astuce, je vais essayer avec les radis aussi.", date: now - 1.5 * day },
            { id: 2, auteurId: 5, message: "On peut aussi ajouter ca a la page cours de cuisine !", date: now - 1 * day },
          ],
        },
        {
          id: 2, titre: "Que faire avec du pain dur ?",
          auteurId: 3, message: "Chapelure maison, pain perdu, croutons... des idees a partager ?",
          vues: 512, createdAt: now - 5 * day,
          messages: [
            { id: 1, auteurId: 2, message: "Le pudding au pain perdu marche tres bien !", date: now - 4 * day },
            { id: 2, auteurId: 4, message: "On donne le notre invendu a l'association chaque soir.", date: now - 3 * day },
            { id: 3, auteurId: 6, message: "Merci pour ce partage, tres utile pour le forum conseil.", date: now - 2 * day },
          ],
        },
        {
          id: 3, titre: "Compost en appartement, vos retours ?",
          auteurId: 2, message: "Je debute le lombricompostage, des conseils pour eviter les odeurs ?",
          vues: 180, createdAt: now - 1 * day,
          messages: [
            { id: 1, auteurId: 3, message: "Bien equilibrer carbone/azote, et aerer regulierement.", date: now - 0.5 * day },
          ],
        },
        {
          id: 4, titre: "Date limite de consommation vs date limite d'utilisation optimale",
          auteurId: 5, message: "Beaucoup de produits DLUO sont encore parfaitement consommables apres la date.",
          vues: 760, createdAt: now - 10 * day,
          messages: [
            { id: 1, auteurId: 2, message: "Ca m'a evite de jeter plein de conserves, merci !", date: now - 9 * day },
          ],
        },
        {
          id: 5, titre: "Recuperer les fanes de radis et carottes",
          auteurId: 3, message: "Pesto de fanes, soupe... je fais tout avec !",
          vues: 245, createdAt: now - 3 * day,
          messages: [],
        },
        {
          id: 6, titre: "Congelation : les erreurs a eviter",
          auteurId: 6, message: "Ne jamais recongeler un produit decongele cru, sauf apres cuisson.",
          vues: 410, createdAt: now - 6 * day,
          messages: [
            { id: 1, auteurId: 4, message: "On etiquette systematiquement nos invendus congeles.", date: now - 5 * day },
          ],
        },
        {
          id: 7, titre: "Batch cooking anti-gaspi du dimanche",
          auteurId: 2, message: "Qui fait du batch cooking pour eviter le gaspillage en semaine ?",
          vues: 95, createdAt: now - 0.5 * day,
          messages: [],
        },
      ],

      recettes: [
        {
          id: 1, titre: "Soupe de fanes de carottes",
          ingredients: ["fanes de carottes", "pomme de terre", "oignon", "bouillon de legumes", "creme"],
          outils: ["mixeur plongeant", "casserole"],
          contenu: "Faites revenir l'oignon, ajoutez les fanes lavees et la pomme de terre, couvrez de bouillon, cuisez 20 min puis mixez.",
          video: null,
        },
        {
          id: 2, titre: "Pain perdu sale ou sucre",
          ingredients: ["pain rassis", "oeuf", "lait", "sucre ou fromage rape"],
          outils: ["poele", "saladier"],
          contenu: "Trempez le pain dans le melange oeuf/lait, dorez a la poele quelques minutes de chaque cote.",
          video: null,
        },
        {
          id: 3, titre: "Chips de peaux de legumes au four",
          ingredients: ["epluchures de pomme de terre", "epluchures de carottes", "huile d'olive", "sel"],
          outils: ["four", "plaque de cuisson"],
          contenu: "Melangez les epluchures avec un filet d'huile et du sel, etalez sur une plaque, cuisson 15 min a 180°C.",
          video: null,
        },
        {
          id: 4, titre: "Confiture de fruits trop murs",
          ingredients: ["fruits murs", "sucre", "citron"],
          outils: ["grande casserole", "bocaux"],
          contenu: "Coupez les fruits, ajoutez sucre et jus de citron, cuisez 30 a 40 min en remuant puis mettez en bocaux.",
          video: null,
        },
        {
          id: 5, titre: "Riz de la veille saute aux legumes",
          ingredients: ["riz cuit", "legumes restants", "oeuf", "sauce soja"],
          outils: ["wok", "spatule"],
          contenu: "Faites revenir les legumes, ajoutez le riz effrite, l'oeuf battu et la sauce soja, sautez 5 min a feu vif.",
          video: null,
        },
        {
          id: 6, titre: "Pesto de fanes de radis",
          ingredients: ["fanes de radis", "ail", "parmesan", "huile d'olive", "amandes"],
          outils: ["mixeur"],
          contenu: "Mixez tous les ingredients ensemble jusqu'a obtenir une texture homogene. Ajustez l'huile selon la consistance.",
          video: null,
        },
        {
          id: 7, titre: "Bouillon maison avec des epluchures",
          ingredients: ["epluchures variees", "queues d'herbes", "eau", "sel"],
          outils: ["grande casserole", "passoire"],
          contenu: "Faites bouillir les epluchures conservees au congelateur 45 min, filtrez et assaisonnez.",
          video: null,
        },
        {
          id: 8, titre: "Yaourts a boire aux fruits abimes",
          ingredients: ["fruits mous", "yaourt nature", "miel"],
          outils: ["mixeur"],
          contenu: "Retirez les parties abimees des fruits, mixez avec le yaourt et un filet de miel.",
          video: null,
        },
      ],

      annonces: [
        {
          id: 1, categorie: "covoiturage", auteurId: 2,
          titre: "Trajet Paris -> Nantes vendredi soir",
          description: "2 places disponibles, depart 18h30 depuis Porte d'Orleans.",
          prix: 15, etat: "disponible", createdAt: Date.now() - 2 * 86400000,
          messages: [],
        },
        {
          id: 2, categorie: "reparation", auteurId: 3,
          titre: "Je repare petits electromenagers",
          description: "Grille-pain, bouilloire, mixeur... contactez-moi avant de jeter !",
          prix: null, etat: "disponible", createdAt: Date.now() - 4 * 86400000,
          messages: [],
        },
        {
          id: 3, categorie: "gardiennage", auteurId: 2,
          titre: "Recherche gardiennage chat 1 semaine",
          description: "Depart en vacances du 10 au 17, chat calme et independant.",
          prix: 8, etat: "en_negociation", createdAt: Date.now() - 1 * 86400000,
          messages: [
            { id: 1, auteurId: 4, message: "Bonjour, je suis disponible sur ces dates.", prixPropose: null, date: Date.now() - 20 * 3600000 },
            { id: 2, auteurId: 2, message: "Parfait, seriez-vous ok pour 6€/jour ?", prixPropose: 6, date: Date.now() - 18 * 3600000 },
          ],
        },
        {
          id: 4, categorie: "location", auteurId: 4,
          titre: "Location table pliante + chaises",
          description: "Ideal pour un evenement associatif, disponible le week-end.",
          prix: 12, etat: "disponible", createdAt: Date.now() - 6 * 86400000,
          messages: [],
        },
        {
          id: 5, categorie: "vente", auteurId: 2,
          titre: "Lot de bocaux en verre (x15)",
          description: "Parfaits pour conserves et confitures maison, tres bon etat.",
          prix: 10, etat: "reservee", createdAt: Date.now() - 3 * 86400000,
          messages: [
            { id: 1, auteurId: 3, message: "Toujours disponible ?", prixPropose: null, date: Date.now() - 2 * 86400000 },
          ],
        },
        {
          id: 6, categorie: "don", auteurId: 3,
          titre: "Cagette de legumes invendus a donner",
          description: "Recuperee ce matin aupres d'un(e) commerçant(e) partenaire, a venir chercher vite.",
          prix: null, etat: "disponible", createdAt: Date.now() - 5 * 3600000,
          messages: [],
        },
        {
          id: 7, categorie: "covoiturage", auteurId: 4,
          titre: "Covoiturage collecte du samedi matin",
          description: "Je pars du centre-ville vers le site de collecte, 3 places.",
          prix: null, etat: "disponible", createdAt: Date.now() - 8 * 3600000,
          messages: [],
        },
        {
          id: 8, categorie: "vente", auteurId: 4,
          titre: "Pain de la veille a prix reduit",
          description: "Invendus du jour, encore tres bons, a prix casse en fin de journee.",
          prix: 2, etat: "disponible", createdAt: Date.now() - 12 * 3600000,
          messages: [],
        },
        {
          id: 9, categorie: "reparation", auteurId: 2,
          titre: "Cherche quelqu'un pour recoudre une housse",
          description: "Petite reparation de couture sur une housse de canape.",
          prix: null, etat: "terminee", createdAt: Date.now() - 15 * 86400000,
          messages: [],
        },
        {
          id: 10, categorie: "gardiennage", auteurId: 3,
          titre: "Disponible pour arroser vos plantes",
          description: "Bénévole dispo en semaine pour l'entretien de plantes d'appartement.",
          prix: null, etat: "disponible", createdAt: Date.now() - 1 * 3600000,
          messages: [],
        },
        {
          id: 11, categorie: "don", auteurId: 2,
          titre: "Livres de cuisine anti-gaspi a donner",
          description: "Une dizaine de livres, bon etat, a venir recuperer.",
          prix: null, etat: "disponible", createdAt: Date.now() - 30 * 3600000,
          messages: [],
        },
        {
          id: 12, categorie: "location", auteurId: 3,
          titre: "Pret de glaciere pour transport de collecte",
          description: "Grande glaciere 60L, disponible le week-end pour les tournees.",
          prix: 5, etat: "disponible", createdAt: Date.now() - 20 * 3600000,
          messages: [],
        },
      ],

      nextForumId: 8,
      nextAnnonceId: 13,
      nextStockId: 8,

      tickets: [
        { id: 1, auteurId: 2, sujet: "Probleme pour reserver une distribution", message: "Le quota affiche ne correspond pas a ce que j'ai reserve, pouvez-vous verifier ?", statut: "ouvert", reponse: null, date: now - 1 * day, traitePar: null, traiteAt: null },
        { id: 2, auteurId: 1, sujet: "Comment devenir benevole ?", message: "Je voudrais participer aux collectes, quelle est la marche a suivre ?", statut: "traite", reponse: "Bonjour, vous pouvez deposer une candidature benevole directement depuis la page d'une collecte ou d'une distribution. A bientot !", date: now - 5 * day, traitePar: 5, traiteAt: now - 4 * day },
      ],
      nextTicketId: 3,

      signalements: [
        { id: 1, type: "forum_message", forumId: 2, messageId: 2, annonceId: null, signalePar: 2, motif: "Message hors sujet.", date: now - 3 * day, statut: "ouvert", commentaire: null, traitePar: null, traiteAt: null },
      ],
      nextSignalementId: 2,

      messagesArchives: [],
      nextMessageArchiveId: 1,

      notifications: [],
      nextNotificationId: 1,
    };
  }

  function isValidState(candidate) {
    return !!candidate
      && Array.isArray(candidate.users)
      && Array.isArray(candidate.collectes)
      && Array.isArray(candidate.distributions)
      && Array.isArray(candidate.stocks)
      && Array.isArray(candidate.forums)
      && Array.isArray(candidate.recettes)
      && Array.isArray(candidate.annonces)
      && Array.isArray(candidate.candidaturesBenevole)
      && Array.isArray(candidate.tickets)
      && Array.isArray(candidate.signalements)
      && Array.isArray(candidate.messagesArchives)
      && Array.isArray(candidate.notifications);
  }

  function normalizeLoadedStocks(candidate) {
    let changed = false;
    candidate.stocks.forEach(function (s) {
      const normalized = normalizeProduitNom(s.nom);
      if (normalized !== s.nom) {
        s.nom = normalized;
        changed = true;
      }
    });
    return changed;
  }

  const state = {};

  function applyState(next) {
    Object.keys(state).forEach(function (k) { delete state[k]; });
    Object.assign(state, next);
  }

  function persist(next) {
    applyState(next);
    localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
  }

  function load() {
    try {
      const raw = localStorage.getItem(STORAGE_KEY);
      if (raw) {
        const parsed = JSON.parse(raw);
        if (isValidState(parsed)) {
          if (normalizeLoadedStocks(parsed)) {
            persist(parsed);
          } else {
            applyState(parsed);
          }
          return;
        }
      }
    } catch (e) {}
    persist(seed());
  }

  load();

  function save() {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
  }

  function resetDemo() {
    persist(seed());
  }

  function getState() {
    return state;
  }

  return {
    getState: getState,
    save: save,
    resetDemo: resetDemo,
    isoDate: isoDate,
    normalizeProduitNom: normalizeProduitNom,
    CATEGORIES_ANNONCE: CATEGORIES_ANNONCE,
    ETATS_ANNONCE: ETATS_ANNONCE,
    UNITES_PRODUIT: UNITES_PRODUIT,
  };
})();
