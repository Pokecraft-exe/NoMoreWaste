BEGIN;

-- pgcrypto fournit crypt()/gen_salt('bf') : utilise par seed.sql pour
-- hasher les mots de passe de demonstration (bcrypt, compatible avec
-- golang.org/x/crypto/bcrypt cote back/backend).
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS site (
	site_id SERIAL PRIMARY KEY,
	nom VARCHAR(64) NOT NULL UNIQUE,
	type_site VARCHAR(16) NOT NULL CHECK (type_site IN ('siege', 'agence', 'entrepot', 'point_relais', 'autre')),
	adresse VARCHAR(128) NOT NULL,
	complement_adresse VARCHAR(128),
	code_postal VARCHAR(10) NOT NULL,
	ville VARCHAR(64) NOT NULL,
	pays VARCHAR(64) NOT NULL,
	telephone VARCHAR(20)
);

CREATE TABLE IF NOT EXISTS personne (
	personne_id SERIAL PRIMARY KEY,
	nom VARCHAR(64) NOT NULL,
	prenom VARCHAR(64) NOT NULL,
	email VARCHAR(128) NOT NULL UNIQUE,
	telephone VARCHAR(20),
	adresse VARCHAR(128),
	complement_adresse VARCHAR(128),
	code_postal VARCHAR(10),
	ville VARCHAR(64),
	pays VARCHAR(64) DEFAULT 'France',
	date_creation TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- type_utilisateur remplace l'ancien systeme de roles (role_compte / compte_role) :
-- chaque compte porte directement son type d'utilisateur, en dur, selon les
-- profils du cahier des charges (visiteur, benevole, adherent, commercant,
-- responsable = salarie de l'association, administrateur).
CREATE TABLE IF NOT EXISTS compte (
	compte_id SERIAL PRIMARY KEY,
	personne_id INTEGER NOT NULL UNIQUE REFERENCES personne(personne_id) ON DELETE CASCADE,
	mot_de_passe VARCHAR(255) NOT NULL,
	type_utilisateur VARCHAR(16) NOT NULL DEFAULT 'visiteur' CHECK (type_utilisateur IN ('visiteur', 'benevole', 'adherent', 'commercant', 'responsable', 'administrateur')),
	actif CHAR(1) NOT NULL DEFAULT '1' CHECK (actif IN ('1', '0')),
	dernier_login TIMESTAMP,
	date_creation TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Jetons d'acces delivres par le service oauth. Le scope renvoye par
-- /oauth/v3/introspect est directement compte.type_utilisateur (voir back/oauth).
CREATE TABLE IF NOT EXISTS token (
	token_id SERIAL PRIMARY KEY,
	compte_id INTEGER NOT NULL REFERENCES compte(compte_id) ON DELETE CASCADE,
	token_value VARCHAR(64) NOT NULL UNIQUE,
	date_expiration TIMESTAMP NOT NULL,
	active CHAR(1) NOT NULL DEFAULT '1' CHECK (active IN ('1', '0')),
	revoked CHAR(1) NOT NULL DEFAULT '0' CHECK (revoked IN ('1', '0')),
	date_creation TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS commercant (
	commercant_id SERIAL PRIMARY KEY,
	personne_id INTEGER UNIQUE REFERENCES personne(personne_id) ON DELETE SET NULL,
	raison_sociale VARCHAR(128) NOT NULL UNIQUE,
	identifiant_legal VARCHAR(32) UNIQUE,
	email VARCHAR(128),
	telephone VARCHAR(20),
	adresse VARCHAR(128) NOT NULL,
	complement_adresse VARCHAR(128),
	code_postal VARCHAR(10) NOT NULL,
	ville VARCHAR(64) NOT NULL,
	pays VARCHAR(64) NOT NULL,
	actif CHAR(1) NOT NULL DEFAULT '1' CHECK (actif IN ('1', '0')),
	date_creation TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS forfait (
	forfait_id SERIAL PRIMARY KEY,
	libelle VARCHAR(64) NOT NULL UNIQUE,
	type_forfait VARCHAR(16) NOT NULL CHECK (type_forfait IN ('adhesion', 'commerce', 'service', 'autre')),
	periodicite VARCHAR(16) NOT NULL CHECK (periodicite IN ('mensuel', 'annuel', 'ponctuel', 'lifetime')),
	prix NUMERIC(10,2) NOT NULL CHECK (prix >= 0),
	actif CHAR(1) NOT NULL DEFAULT '1' CHECK (actif IN ('1', '0'))
);

CREATE TABLE IF NOT EXISTS adherent (
	adherent_id SERIAL PRIMARY KEY,
	personne_id INTEGER NOT NULL UNIQUE REFERENCES personne(personne_id) ON DELETE CASCADE,
	date_inscription DATE NOT NULL DEFAULT CURRENT_DATE,
	statut VARCHAR(16) NOT NULL DEFAULT 'actif' CHECK (statut IN ('actif', 'suspendu', 'radie', 'en_attente'))
);

CREATE TABLE IF NOT EXISTS adhesion_association (
	adhesion_association_id SERIAL PRIMARY KEY,
	adherent_id INTEGER NOT NULL REFERENCES adherent(adherent_id) ON DELETE CASCADE,
	forfait_id INTEGER NOT NULL REFERENCES forfait(forfait_id),
	date_debut DATE NOT NULL DEFAULT CURRENT_DATE,
	date_fin DATE,
	renouvellement_automatique CHAR(1) NOT NULL DEFAULT '1' CHECK (renouvellement_automatique IN ('1', '0')),
	statut VARCHAR(16) NOT NULL DEFAULT 'active' CHECK (statut IN ('active', 'expiree', 'suspendue', 'resiliee')),
	created_by INTEGER REFERENCES compte(compte_id),
	UNIQUE (adherent_id, forfait_id, date_debut)
);

CREATE TABLE IF NOT EXISTS rappel_adhesion (
	rappel_adhesion_id SERIAL PRIMARY KEY,
	adhesion_association_id INTEGER NOT NULL REFERENCES adhesion_association(adhesion_association_id) ON DELETE CASCADE,
	date_rappel TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	canal VARCHAR(16) NOT NULL CHECK (canal IN ('email', 'sms', 'notification', 'courrier')),
	statut VARCHAR(16) NOT NULL DEFAULT 'a_envoyer' CHECK (statut IN ('a_envoyer', 'envoye', 'lu', 'echoue')),
	message VARCHAR(512) NOT NULL
);

CREATE TABLE IF NOT EXISTS partenaire (
	partenaire_id SERIAL PRIMARY KEY,
	nom VARCHAR(128) NOT NULL UNIQUE,
	type_partenaire VARCHAR(20) NOT NULL CHECK (type_partenaire IN ('association', 'collectivite', 'entreprise', 'ecole', 'particulier', 'autre')),
	email VARCHAR(128),
	telephone VARCHAR(20),
	adresse VARCHAR(128),
	complement_adresse VARCHAR(128),
	code_postal VARCHAR(10),
	ville VARCHAR(64),
	pays VARCHAR(64),
	actif CHAR(1) NOT NULL DEFAULT '1' CHECK (actif IN ('1', '0')),
	date_creation TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS collecte (
	collecte_id SERIAL PRIMARY KEY,
	site_id INTEGER REFERENCES site(site_id) ON DELETE SET NULL,
	commercant_id INTEGER REFERENCES commercant(commercant_id) ON DELETE SET NULL,
	partenaire_id INTEGER REFERENCES partenaire(partenaire_id) ON DELETE SET NULL,
	lieu VARCHAR(128) NOT NULL,
	date_collecte DATE NOT NULL,
	heure_collecte TIME,
	statut VARCHAR(16) NOT NULL DEFAULT 'planifiee' CHECK (statut IN ('planifiee', 'confirmee', 'en_cours', 'terminee', 'annulee')),
	description VARCHAR(512),
	stock_mis_a_jour CHAR(1) NOT NULL DEFAULT '0' CHECK (stock_mis_a_jour IN ('1', '0')),
	created_by INTEGER REFERENCES compte(compte_id),
	date_creation TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS categorie_produit (
	categorie_produit_id SERIAL PRIMARY KEY,
	libelle VARCHAR(64) NOT NULL UNIQUE,
	description VARCHAR(256)
);

CREATE TABLE IF NOT EXISTS stock (
	stock_id SERIAL PRIMARY KEY,
	site_id INTEGER NOT NULL REFERENCES site(site_id) ON DELETE CASCADE,
	nom VARCHAR(64) NOT NULL,
	type_stock VARCHAR(16) NOT NULL CHECK (type_stock IN ('sec', 'frais', 'surgeles', 'materiel', 'autre')),
	capacite_max NUMERIC(12,3),
	actif CHAR(1) NOT NULL DEFAULT '1' CHECK (actif IN ('1', '0')),
	UNIQUE (site_id, nom)
);

CREATE TABLE IF NOT EXISTS produit (
	produit_id SERIAL PRIMARY KEY,
	collecte_id INTEGER REFERENCES collecte(collecte_id) ON DELETE SET NULL,
	categorie_produit_id INTEGER REFERENCES categorie_produit(categorie_produit_id),
	stock_id INTEGER REFERENCES stock(stock_id) ON DELETE SET NULL,
	barcode VARCHAR(64) NOT NULL UNIQUE,
	nom VARCHAR(128) NOT NULL,
	description VARCHAR(512),
	quantite INTEGER NOT NULL DEFAULT 1 CHECK (quantite > 0),
	poids_kg NUMERIC(10,3) CHECK (poids_kg >= 0),
	date_peremption DATE,
	etat VARCHAR(16) NOT NULL DEFAULT 'recu' CHECK (etat IN ('recu', 'controle', 'stocke', 'reserve', 'redistribue', 'jete')),
	date_reception TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS mouvement_stock (
	mouvement_stock_id SERIAL PRIMARY KEY,
	produit_id INTEGER NOT NULL REFERENCES produit(produit_id) ON DELETE CASCADE,
	stock_id INTEGER REFERENCES stock(stock_id) ON DELETE SET NULL,
	type_mouvement VARCHAR(16) NOT NULL CHECK (type_mouvement IN ('entree', 'sortie', 'transfert', 'ajustement')),
	quantite INTEGER NOT NULL DEFAULT 1 CHECK (quantite > 0),
	commentaire VARCHAR(512),
	date_mouvement TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tournee (
	tournee_id SERIAL PRIMARY KEY,
	site_id INTEGER REFERENCES site(site_id) ON DELETE SET NULL,
	date_tournee DATE NOT NULL,
	type_tournee VARCHAR(16) NOT NULL CHECK (type_tournee IN ('collecte', 'distribution', 'mixte')),
	statut VARCHAR(16) NOT NULL DEFAULT 'planifiee' CHECK (statut IN ('planifiee', 'en_cours', 'terminee', 'annulee')),
	vehicule VARCHAR(64),
	commentaire VARCHAR(512),
	created_by INTEGER REFERENCES compte(compte_id),
	date_creation TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tournee_etape (
	tournee_etape_id SERIAL PRIMARY KEY,
	tournee_id INTEGER NOT NULL REFERENCES tournee(tournee_id) ON DELETE CASCADE,
	ordre INTEGER NOT NULL CHECK (ordre > 0),
	collecte_id INTEGER REFERENCES collecte(collecte_id) ON DELETE SET NULL,
	partenaire_id INTEGER REFERENCES partenaire(partenaire_id) ON DELETE SET NULL,
	adresse VARCHAR(128),
	ville VARCHAR(64),
	heure_prevue TIME,
	heure_reelle TIME,
	commentaire VARCHAR(512),
	UNIQUE (tournee_id, ordre)
);

CREATE TABLE IF NOT EXISTS benevole (
	benevole_id SERIAL PRIMARY KEY,
	personne_id INTEGER NOT NULL UNIQUE REFERENCES personne(personne_id) ON DELETE CASCADE,
	date_inscription DATE NOT NULL DEFAULT CURRENT_DATE,
	statut VARCHAR(16) NOT NULL DEFAULT 'candidat' CHECK (statut IN ('candidat', 'actif', 'suspendu', 'refuse', 'inactif')),
	disponibilite VARCHAR(128),
	commentaire VARCHAR(512)
);

CREATE TABLE IF NOT EXISTS competence (
	competence_id SERIAL PRIMARY KEY,
	libelle VARCHAR(64) NOT NULL UNIQUE,
	description VARCHAR(256)
);

CREATE TABLE IF NOT EXISTS benevole_competence (
	benevole_id INTEGER NOT NULL REFERENCES benevole(benevole_id) ON DELETE CASCADE,
	competence_id INTEGER NOT NULL REFERENCES competence(competence_id) ON DELETE CASCADE,
	niveau INTEGER NOT NULL DEFAULT 1 CHECK (niveau BETWEEN 1 AND 5),
	PRIMARY KEY (benevole_id, competence_id)
);

CREATE TABLE IF NOT EXISTS candidature_benevole (
	candidature_benevole_id SERIAL PRIMARY KEY,
	personne_id INTEGER NOT NULL REFERENCES personne(personne_id) ON DELETE CASCADE,
	date_candidature TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	statut VARCHAR(16) NOT NULL DEFAULT 'recue' CHECK (statut IN ('recue', 'en_etude', 'validee', 'refusee', 'archivee')),
	motivation VARCHAR(512),
	disponibilite VARCHAR(128),
	commentaire VARCHAR(512),
	traite_par INTEGER REFERENCES compte(compte_id),
	date_decision TIMESTAMP
);

CREATE TABLE IF NOT EXISTS service (
	service_id SERIAL PRIMARY KEY,
	libelle VARCHAR(64) NOT NULL UNIQUE,
	categorie_service VARCHAR(20) NOT NULL CHECK (categorie_service IN ('conseil', 'cuisine', 'vehicule', 'echange', 'reparation', 'gardiennage', 'autre')),
	description VARCHAR(512) NOT NULL,
	tarif NUMERIC(10,2) NOT NULL DEFAULT 0 CHECK (tarif >= 0),
	actif CHAR(1) NOT NULL DEFAULT '1' CHECK (actif IN ('1', '0'))
);

CREATE TABLE IF NOT EXISTS planning_service (
	planning_service_id SERIAL PRIMARY KEY,
	service_id INTEGER NOT NULL REFERENCES service(service_id) ON DELETE CASCADE,
	site_id INTEGER REFERENCES site(site_id) ON DELETE SET NULL,
	date_service DATE NOT NULL,
	heure_debut TIME NOT NULL,
	heure_fin TIME NOT NULL,
	capacite INTEGER NOT NULL DEFAULT 0 CHECK (capacite >= 0),
	statut VARCHAR(16) NOT NULL DEFAULT 'planifie' CHECK (statut IN ('planifie', 'ouvert', 'complete', 'annule')),
	commentaire VARCHAR(512),
	created_by INTEGER REFERENCES compte(compte_id),
	date_creation TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS inscription_service (
	inscription_service_id SERIAL PRIMARY KEY,
	planning_service_id INTEGER NOT NULL REFERENCES planning_service(planning_service_id) ON DELETE CASCADE,
	adherent_id INTEGER NOT NULL REFERENCES adherent(adherent_id) ON DELETE CASCADE,
	statut VARCHAR(16) NOT NULL DEFAULT 'inscrit' CHECK (statut IN ('inscrit', 'attente', 'annule', 'present', 'absent')),
	date_inscription TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	commentaire VARCHAR(512),
	UNIQUE (planning_service_id, adherent_id)
);

CREATE TABLE IF NOT EXISTS affectation_benevole (
	affectation_benevole_id SERIAL PRIMARY KEY,
	benevole_id INTEGER NOT NULL REFERENCES benevole(benevole_id) ON DELETE CASCADE,
	tournee_id INTEGER REFERENCES tournee(tournee_id) ON DELETE CASCADE,
	planning_service_id INTEGER REFERENCES planning_service(planning_service_id) ON DELETE CASCADE,
	role_mission VARCHAR(64) NOT NULL,
	statut VARCHAR(16) NOT NULL DEFAULT 'planifiee' CHECK (statut IN ('planifiee', 'confirmee', 'remplacee', 'annulee', 'terminee')),
	commentaire VARCHAR(512),
	date_affectation TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS adhesion_commercant (
	adhesion_commercant_id SERIAL PRIMARY KEY,
	commercant_id INTEGER NOT NULL REFERENCES commercant(commercant_id) ON DELETE CASCADE,
	forfait_id INTEGER NOT NULL REFERENCES forfait(forfait_id),
	date_debut DATE NOT NULL DEFAULT CURRENT_DATE,
	date_fin DATE,
	renouvellement_automatique CHAR(1) NOT NULL DEFAULT '1' CHECK (renouvellement_automatique IN ('1', '0')),
	statut VARCHAR(16) NOT NULL DEFAULT 'active' CHECK (statut IN ('active', 'expiree', 'suspendue', 'resiliee')),
	created_by INTEGER REFERENCES compte(compte_id),
	UNIQUE (commercant_id, forfait_id, date_debut)
);

CREATE TABLE IF NOT EXISTS document_genere (
	document_genere_id SERIAL PRIMARY KEY,
	collecte_id INTEGER REFERENCES collecte(collecte_id) ON DELETE CASCADE,
	tournee_id INTEGER REFERENCES tournee(tournee_id) ON DELETE CASCADE,
	planning_service_id INTEGER REFERENCES planning_service(planning_service_id) ON DELETE CASCADE,
	type_document VARCHAR(16) NOT NULL CHECK (type_document IN ('pdf', 'xlsx', 'csv', 'ods', 'autre')),
	nom_fichier VARCHAR(128) NOT NULL,
	chemin_fichier VARCHAR(256) NOT NULL,
	date_generation TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	genere_par INTEGER REFERENCES compte(compte_id)
);

-- Un benevole peut s'inscrire directement sur une collecte a venir
-- ("voir/participer a la prochaine collecte" dans le cas d'utilisation),
-- independamment de son affectation eventuelle a une tournee.
CREATE TABLE IF NOT EXISTS participation_collecte (
	collecte_id INTEGER NOT NULL REFERENCES collecte(collecte_id) ON DELETE CASCADE,
	benevole_id INTEGER NOT NULL REFERENCES benevole(benevole_id) ON DELETE CASCADE,
	role_mission VARCHAR(64),
	commentaire VARCHAR(512),
	date_participation TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (collecte_id, benevole_id)
);

-- Catalogue de produits par type (nom + unite), utilise pour les quantites de
-- denrees suivies sur une collecte ou une distribution - independant de
-- produit/stock qui suivent des articles individuels code-barres.
CREATE TABLE IF NOT EXISTS stock_produit (
	stock_produit_id SERIAL PRIMARY KEY,
	nom VARCHAR(128) NOT NULL UNIQUE,
	unite VARCHAR(16) NOT NULL DEFAULT 'unites',
	quantite_disponible NUMERIC(12,3) NOT NULL DEFAULT 0 CHECK (quantite_disponible >= 0)
);

CREATE TABLE IF NOT EXISTS collecte_denree (
	collecte_denree_id SERIAL PRIMARY KEY,
	collecte_id INTEGER NOT NULL REFERENCES collecte(collecte_id) ON DELETE CASCADE,
	stock_produit_id INTEGER NOT NULL REFERENCES stock_produit(stock_produit_id),
	quantite NUMERIC(12,3) NOT NULL CHECK (quantite > 0),
	non_perissable CHAR(1) NOT NULL DEFAULT '0' CHECK (non_perissable IN ('1', '0')),
	dlc DATE,
	propose_par INTEGER NOT NULL REFERENCES compte(compte_id) ON DELETE CASCADE,
	propose_par_type VARCHAR(16) NOT NULL CHECK (propose_par_type IN ('staff', 'commercant')),
	confirmee CHAR(1) NOT NULL DEFAULT '0' CHECK (confirmee IN ('1', '0')),
	date_ajout TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Affectation ferme d'un benevole a une collecte, decidee par le staff -
-- distincte de participation_collecte qui n'est qu'une demande.
CREATE TABLE IF NOT EXISTS collecte_benevole_affecte (
	collecte_id INTEGER NOT NULL REFERENCES collecte(collecte_id) ON DELETE CASCADE,
	benevole_id INTEGER NOT NULL REFERENCES benevole(benevole_id) ON DELETE CASCADE,
	role_mission VARCHAR(64) NOT NULL,
	PRIMARY KEY (collecte_id, benevole_id)
);

-- Distribution aux adherents : entite autonome (lieu/date/quota), distincte
-- de tournee/tournee_etape qui modelisent une tournee logistique multi-arret.
CREATE TABLE IF NOT EXISTS distribution (
	distribution_id SERIAL PRIMARY KEY,
	lieu VARCHAR(128) NOT NULL,
	date_distribution DATE NOT NULL,
	heure_distribution TIME,
	statut VARCHAR(16) NOT NULL DEFAULT 'planifiee' CHECK (statut IN ('planifiee', 'confirmee', 'en_cours', 'terminee', 'annulee')),
	quota_par_adherent INTEGER NOT NULL DEFAULT 1 CHECK (quota_par_adherent >= 0),
	created_by INTEGER REFERENCES compte(compte_id),
	date_creation TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS distribution_denree (
	distribution_id INTEGER NOT NULL REFERENCES distribution(distribution_id) ON DELETE CASCADE,
	stock_produit_id INTEGER NOT NULL REFERENCES stock_produit(stock_produit_id),
	quantite NUMERIC(12,3) NOT NULL DEFAULT 0 CHECK (quantite >= 0),
	PRIMARY KEY (distribution_id, stock_produit_id)
);

CREATE TABLE IF NOT EXISTS distribution_benevole_affecte (
	distribution_id INTEGER NOT NULL REFERENCES distribution(distribution_id) ON DELETE CASCADE,
	benevole_id INTEGER NOT NULL REFERENCES benevole(benevole_id) ON DELETE CASCADE,
	role_mission VARCHAR(64) NOT NULL,
	PRIMARY KEY (distribution_id, benevole_id)
);

CREATE TABLE IF NOT EXISTS distribution_benevole_participant (
	distribution_id INTEGER NOT NULL REFERENCES distribution(distribution_id) ON DELETE CASCADE,
	benevole_id INTEGER NOT NULL REFERENCES benevole(benevole_id) ON DELETE CASCADE,
	date_participation TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (distribution_id, benevole_id)
);

CREATE TABLE IF NOT EXISTS reservation (
	reservation_id SERIAL PRIMARY KEY,
	distribution_id INTEGER NOT NULL REFERENCES distribution(distribution_id) ON DELETE CASCADE,
	adherent_id INTEGER NOT NULL REFERENCES adherent(adherent_id) ON DELETE CASCADE,
	stock_produit_id INTEGER NOT NULL REFERENCES stock_produit(stock_produit_id),
	quantite NUMERIC(12,3) NOT NULL CHECK (quantite > 0),
	date_reservation TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Forum de conseils anti-gaspi ("acceder/creer un forum de conseil").
CREATE TABLE IF NOT EXISTS forum_thread (
	forum_thread_id SERIAL PRIMARY KEY,
	compte_id INTEGER NOT NULL REFERENCES compte(compte_id) ON DELETE CASCADE,
	titre VARCHAR(128) NOT NULL,
	message VARCHAR(2000) NOT NULL,
	vues INTEGER NOT NULL DEFAULT 0,
	date_creation TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS forum_message (
	forum_message_id SERIAL PRIMARY KEY,
	forum_thread_id INTEGER NOT NULL REFERENCES forum_thread(forum_thread_id) ON DELETE CASCADE,
	compte_id INTEGER NOT NULL REFERENCES compte(compte_id) ON DELETE CASCADE,
	message VARCHAR(2000) NOT NULL,
	parent_id INTEGER REFERENCES forum_message(forum_message_id) ON DELETE SET NULL,
	date_envoi TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Recettes / cours de cuisine consultables ("acceder aux recettes/cours de cuisine").
CREATE TABLE IF NOT EXISTS ressource_cuisine (
	ressource_cuisine_id SERIAL PRIMARY KEY,
	titre VARCHAR(128) NOT NULL,
	ingredients TEXT[] NOT NULL DEFAULT '{}',
	outils TEXT[] NOT NULL DEFAULT '{}',
	contenu VARCHAR(4000) NOT NULL,
	video VARCHAR(256),
	created_by INTEGER REFERENCES compte(compte_id),
	date_creation TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Petites annonces d'echange entre particuliers ("annonces PAP") : covoiturage,
-- reparation, gardiennage, echange de services... cf. services proposes du CDC.
CREATE TABLE IF NOT EXISTS annonce_echange (
	annonce_echange_id SERIAL PRIMARY KEY,
	compte_id INTEGER NOT NULL REFERENCES compte(compte_id) ON DELETE CASCADE,
	categorie VARCHAR(16) NOT NULL CHECK (categorie IN ('covoiturage', 'reparation', 'gardiennage', 'location', 'vente', 'don')),
	titre VARCHAR(128) NOT NULL,
	description VARCHAR(1000) NOT NULL,
	prix NUMERIC(10,2) CHECK (prix >= 0),
	statut VARCHAR(16) NOT NULL DEFAULT 'ouverte' CHECK (statut IN ('ouverte', 'en_cours', 'cloturee')),
	date_publication TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS notification (
	notification_id SERIAL PRIMARY KEY,
	compte_id INTEGER NOT NULL REFERENCES compte(compte_id) ON DELETE CASCADE,
	type_notification VARCHAR(32) NOT NULL CHECK (type_notification IN ('signalement_traite', 'ticket_traite')),
	message VARCHAR(1000) NOT NULL,
	lien VARCHAR(256),
	date_notification TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	lu CHAR(1) NOT NULL DEFAULT '0' CHECK (lu IN ('1', '0'))
);

CREATE TABLE IF NOT EXISTS ticket (
	ticket_id SERIAL PRIMARY KEY,
	auteur_id INTEGER REFERENCES compte(compte_id) ON DELETE SET NULL,
	contact_nom VARCHAR(128),
	contact_email VARCHAR(128),
	sujet VARCHAR(128) NOT NULL,
	message VARCHAR(2000) NOT NULL,
	statut VARCHAR(16) NOT NULL DEFAULT 'ouvert' CHECK (statut IN ('ouvert', 'traite')),
	reponse VARCHAR(2000),
	traite_par INTEGER REFERENCES compte(compte_id) ON DELETE SET NULL,
	date_traitement TIMESTAMP,
	date_creation TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS message_annonce_echange (
	message_annonce_echange_id SERIAL PRIMARY KEY,
	annonce_echange_id INTEGER NOT NULL REFERENCES annonce_echange(annonce_echange_id) ON DELETE CASCADE,
	expediteur_id INTEGER NOT NULL REFERENCES compte(compte_id) ON DELETE CASCADE,
	message VARCHAR(1000) NOT NULL,
	prix_propose NUMERIC(10,2) CHECK (prix_propose >= 0),
	date_envoi TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS signalement (
	signalement_id SERIAL PRIMARY KEY,
	type_signalement VARCHAR(16) NOT NULL CHECK (type_signalement IN ('forum', 'forum_message', 'annonce_message')),
	forum_thread_id INTEGER REFERENCES forum_thread(forum_thread_id) ON DELETE SET NULL,
	forum_message_id INTEGER REFERENCES forum_message(forum_message_id) ON DELETE SET NULL,
	annonce_message_id INTEGER REFERENCES message_annonce_echange(message_annonce_echange_id) ON DELETE SET NULL,
	signale_par INTEGER NOT NULL REFERENCES compte(compte_id) ON DELETE CASCADE,
	motif VARCHAR(1000) NOT NULL,
	date_signalement TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	statut VARCHAR(16) NOT NULL DEFAULT 'ouvert' CHECK (statut IN ('ouvert', 'traite')),
	commentaire VARCHAR(1000),
	traite_par INTEGER REFERENCES compte(compte_id) ON DELETE SET NULL,
	date_traitement TIMESTAMP,
	-- le contenu signale peut etre supprime apres coup (le signalement et, une fois
	-- traite, son archive dans message_archive doivent survivre a cette suppression) :
	-- la coherence ci-dessous n'est donc garantie qu'a la creation, pas dans la duree.
	CHECK (
		(type_signalement = 'forum' AND forum_message_id IS NULL AND annonce_message_id IS NULL)
		OR (type_signalement = 'forum_message' AND annonce_message_id IS NULL)
		OR (type_signalement = 'annonce_message' AND forum_thread_id IS NULL AND forum_message_id IS NULL)
	)
);

CREATE TABLE IF NOT EXISTS message_archive (
	message_archive_id SERIAL PRIMARY KEY,
	signalement_id INTEGER NOT NULL UNIQUE REFERENCES signalement(signalement_id) ON DELETE CASCADE,
	auteur_id INTEGER REFERENCES compte(compte_id) ON DELETE SET NULL,
	message VARCHAR(2000) NOT NULL,
	date_message TIMESTAMP NOT NULL,
	date_archivage TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_personne_email ON personne(email);
CREATE INDEX IF NOT EXISTS idx_token_value ON token(token_value);
CREATE INDEX IF NOT EXISTS idx_forum_message_thread ON forum_message(forum_thread_id);
CREATE INDEX IF NOT EXISTS idx_annonce_echange_categorie ON annonce_echange(categorie);
CREATE INDEX IF NOT EXISTS idx_collecte_date_lieu ON collecte(date_collecte, lieu);
CREATE INDEX IF NOT EXISTS idx_collecte_partenaire ON collecte(partenaire_id);
CREATE INDEX IF NOT EXISTS idx_produit_barcode ON produit(barcode);
CREATE INDEX IF NOT EXISTS idx_produit_collecte ON produit(collecte_id);
CREATE INDEX IF NOT EXISTS idx_tournee_date ON tournee(date_tournee);
CREATE INDEX IF NOT EXISTS idx_planning_service_date ON planning_service(date_service);
CREATE INDEX IF NOT EXISTS idx_adhesion_association_fin ON adhesion_association(date_fin);
CREATE INDEX IF NOT EXISTS idx_adhesion_commercant_fin ON adhesion_commercant(date_fin);
CREATE INDEX IF NOT EXISTS idx_collecte_denree_collecte ON collecte_denree(collecte_id);
CREATE INDEX IF NOT EXISTS idx_distribution_denree_distribution ON distribution_denree(distribution_id);
CREATE INDEX IF NOT EXISTS idx_reservation_distribution ON reservation(distribution_id);
CREATE INDEX IF NOT EXISTS idx_notification_compte_lu ON notification(compte_id, lu);
CREATE INDEX IF NOT EXISTS idx_signalement_statut ON signalement(statut);
CREATE INDEX IF NOT EXISTS idx_ticket_statut ON ticket(statut);

CREATE OR REPLACE VIEW v_collectes_recherche AS
SELECT
	c.collecte_id AS id,
	c.lieu,
	c.date_collecte AS date,
	COALESCE(p.nom, '') AS partenaire,
	c.statut,
	c.description
FROM collecte c
LEFT JOIN partenaire p ON p.partenaire_id = c.partenaire_id;

CREATE OR REPLACE VIEW v_collectes_detail AS
SELECT
	c.collecte_id,
	c.lieu,
	c.date_collecte,
	c.heure_collecte,
	c.statut,
	c.description,
	s.nom AS site,
	co.raison_sociale AS commercant,
	p.nom AS partenaire
FROM collecte c
LEFT JOIN site s ON s.site_id = c.site_id
LEFT JOIN commercant co ON co.commercant_id = c.commercant_id
LEFT JOIN partenaire p ON p.partenaire_id = c.partenaire_id;

CREATE OR REPLACE VIEW v_planning_benevoles AS
SELECT
	ps.planning_service_id,
	ps.date_service,
	ps.heure_debut,
	ps.heure_fin,
	svc.libelle AS service,
	s.nom AS site,
	pe.nom,
	pe.prenom,
	af.role_mission,
	af.statut AS statut_affectation
FROM planning_service ps
INNER JOIN service svc ON svc.service_id = ps.service_id
LEFT JOIN site s ON s.site_id = ps.site_id
LEFT JOIN affectation_benevole af ON af.planning_service_id = ps.planning_service_id
LEFT JOIN benevole b ON b.benevole_id = af.benevole_id
LEFT JOIN personne pe ON pe.personne_id = b.personne_id;

CREATE OR REPLACE VIEW v_stock_barcode AS
SELECT
	pr.produit_id,
	pr.barcode,
	pr.nom,
	pr.quantite,
	pr.etat,
	pr.date_reception,
	st.nom AS stock,
	sit.nom AS site
FROM produit pr
LEFT JOIN stock st ON st.stock_id = pr.stock_id
LEFT JOIN site sit ON sit.site_id = st.site_id;

-- Donnees de depart minimales pour pouvoir demarrer l'application.
INSERT INTO forfait (libelle, type_forfait, periodicite, prix) VALUES
	('Adhesion standard', 'adhesion', 'annuel', 20.00),
	('Adhesion commercant', 'commerce', 'annuel', 50.00)
ON CONFLICT (libelle) DO NOTHING;

INSERT INTO site (nom, type_site, adresse, code_postal, ville, pays) VALUES
	('Siege Paris', 'siege', '1 rue de la Solidarite', '75001', 'Paris', 'France')
ON CONFLICT (nom) DO NOTHING;

-- Catalogue de competences de base pour les benevoles (chauffeurs,
-- cuisiniers, plombiers, ... - cf. cahier des charges).
INSERT INTO competence (libelle, description) VALUES
	('Chauffeur', 'Permis B, conduite des camionnettes de collecte/distribution'),
	('Cuisinier', 'Animation des cours de cuisine et ateliers anti-gaspi'),
	('Plombier', 'Interventions de plomberie dans le cadre du service reparation'),
	('Electricien', 'Interventions electriques dans le cadre du service reparation'),
	('Manutention', 'Chargement/dechargement, tri et stockage des denrees'),
	('Accueil', 'Accueil des beneficiaires lors des distributions')
ON CONFLICT (libelle) DO NOTHING;

-- Compte administrateur de demarrage (bootstrap) : comme il n'y a plus de
-- table de roles a peupler manuellement, il faut au moins un compte
-- administrateur pour pouvoir ensuite promouvoir tout autre compte via
-- PATCH /api/v1/admin/comptes. Identifiants : admin@nomorewaste.local /
-- ChangeMe123! - A CHANGER IMMEDIATEMENT APRES LE PREMIER DEPLOIEMENT
-- (le hash ci-dessous est un bcrypt de "ChangeMe123!").
WITH admin_personne AS (
	INSERT INTO personne (nom, prenom, email, telephone, adresse, code_postal, ville, pays)
	VALUES ('Admin', 'NoMoreWaste', 'admin@nomorewaste.local', '+33600000000', '1 rue de la Solidarite', '75001', 'Paris', 'France')
	ON CONFLICT (email) DO NOTHING
	RETURNING personne_id
)
INSERT INTO compte (personne_id, mot_de_passe, type_utilisateur)
SELECT personne_id, '$2a$10$dqeWXN4UP/YgRghRmZPavOm13RlgTrGTNJ5Xhixp7Y05U2ASt3uje', 'administrateur'
FROM admin_personne
ON CONFLICT (personne_id) DO NOTHING;

COMMIT;
