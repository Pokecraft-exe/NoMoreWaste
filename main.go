package main

import (
	"fmt"
	"net/http"
)

func main() {
	// Authentification (ex-service oauth, fusionne ici : meme processus,
	// meme port). "S'inscrire", "Se connecter".
	http.HandleFunc("/oauth/v3/inscription", createAccount)
	http.HandleFunc("/oauth/v3/token", token)
	http.HandleFunc("/oauth/v3/introspect", introspect)
	http.HandleFunc("/oauth/v3/type_utilisateur", typeUtilisateur)

	// Profil / adhesion (front-office - "Se connecter", "Consulter son
	// profil", "Modifier ses informations", "Renouveler son adhesion").
	http.HandleFunc("/api/v1/profil", profilHandler)
	http.HandleFunc("/api/v1/adhesion/renouveler", renouvelerAdhesionHandler)

	// Adherents (back-office - "rechercher un adherant", "modifier
	// l'adherant", "Envoyer un rappel de renouvellement").
	http.HandleFunc("/api/v1/adherents", adherentsHandler)
	http.HandleFunc("/api/v1/adherents/rappel", rappelAdhesionHandler)

	// Commercants ("gerer les adhesions des commercants").
	http.HandleFunc("/api/v1/commercants", commercantsHandler)
	http.HandleFunc("/api/v1/commercants/adhesion", adhesionCommercantHandler)

	// Collectes ("rechercher/modifier/creer une collecte", "voir/participer
	// a la prochaine collecte").
	http.HandleFunc("/api/v1/collectes", collectesHandler)
	http.HandleFunc("/api/v1/collectes/denrees", collecteDenreesHandler)
	http.HandleFunc("/api/v1/collectes/confirmer", confirmerCollecteHandler)
	http.HandleFunc("/api/v1/collectes/benevoles", collecteBenevolesHandler)
	http.HandleFunc("/api/v1/collectes/participation", participationCollecteHandler)

	// Distributions aux adherents ("rechercher/modifier/creer une
	// distribution", quotas, benevoles, reservations).
	http.HandleFunc("/api/v1/distributions", distributionsHandler)
	http.HandleFunc("/api/v1/distributions/denrees", distributionDenreesHandler)
	http.HandleFunc("/api/v1/distributions/benevoles", distributionBenevolesHandler)
	http.HandleFunc("/api/v1/distributions/participation", participationDistributionHandler)
	http.HandleFunc("/api/v1/reservations", reservationHandler)

	// Tournees de distribution ("rechercher/modifier/creer une
	// distribution").
	http.HandleFunc("/api/v1/tournees", tourneesHandler)
	http.HandleFunc("/api/v1/tournees/etapes", tourneeEtapesHandler)

	// Stock et produits ("gerer les stocks", chaque produit reference par
	// code barre).
	http.HandleFunc("/api/v1/stocks", stocksHandler)
	http.HandleFunc("/api/v1/produits", produitsHandler)
	http.HandleFunc("/api/v1/produits/mouvements", mouvementsStockHandler)

	// Catalogue de produits par type, utilise pour les quantites de denrees
	// suivies sur une collecte ou une distribution.
	http.HandleFunc("/api/v1/stock-produits", stockProduitsHandler)

	// Benevoles ("gerer le suivi des benevoles, depuis leur candidature
	// jusqu'a leur affectation a un service donne").
	http.HandleFunc("/api/v1/benevoles/candidatures", candidaturesBenevoleHandler)
	http.HandleFunc("/api/v1/benevoles", benevolesHandler)
	http.HandleFunc("/api/v1/benevoles/competences", benevoleCompetencesHandler)
	http.HandleFunc("/api/v1/benevoles/affectations", affectationsBenevoleHandler)

	// Services ("gestion des services : propositions, plannings,
	// inscriptions"; "Gerer le planning", "Ajouter une date", "Ajouter un
	// benevole", "affecter le benevole").
	http.HandleFunc("/api/v1/services", servicesHandler)
	http.HandleFunc("/api/v1/services/planning", planningServiceHandler)
	http.HandleFunc("/api/v1/services/inscriptions", inscriptionsServiceHandler)

	// Forum de conseil ("Acceder a forum de conseil", "Creer un forum").
	http.HandleFunc("/api/v1/forum", forumHandler)
	http.HandleFunc("/api/v1/forum/messages", forumMessagesHandler)

	// Recettes / cours de cuisine.
	http.HandleFunc("/api/v1/ressources-cuisine", ressourcesCuisineHandler)

	// Annonces PAP - covoiturage, reparation, gardiennage, echange de
	// services entre particuliers.
	http.HandleFunc("/api/v1/annonces", annoncesHandler)
	http.HandleFunc("/api/v1/annonces/messages", messagesAnnonceHandler)

	// Documents generes (recapitulatif PDF d'une collecte, etc.).
	http.HandleFunc("/api/v1/documents", documentsHandler)

	// Administration - creation des comptes responsable/administrateur
	// (jamais via l'inscription publique) et changement du type
	// d'utilisateur d'un compte existant.
	http.HandleFunc("/api/v1/admin/comptes", adminComptesHandler)

	// Notifications (cloche a cote du profil) et signalements/tickets de
	// moderation traites par le back-office.
	http.HandleFunc("/api/v1/notifications", notificationsHandler)
	http.HandleFunc("/api/v1/signalements", signalementsHandler)
	http.HandleFunc("/api/v1/tickets", ticketsHandler)

	port := getEnv("PORT", "8080") // oauth and api are now a single backend service
	certFile := getEnv("TLS_CERT_FILE", "/certs/api.crt")
	keyFile := getEnv("TLS_KEY_FILE", "/certs/api.key")
	fmt.Println("Listening on :" + port + " (TLS)")

	err := http.ListenAndServeTLS(":"+port, certFile, keyFile, nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
