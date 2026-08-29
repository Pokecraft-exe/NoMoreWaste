-- Migration a appliquer sur une base DEJA initialisee.
--
-- sql.sql n'est rejoue par l'image postgres qu'au tout premier demarrage
-- (volume de donnees vide) : une instance deja deployee n'a donc pas le
-- catalogue de competences ajoute pour la page back-office "Benevoles",
-- et le menu deroulant des competences y apparaitrait vide.
--
--   docker compose exec -T db psql -U postgres -d nomorewaste < db/migration-01-competences.sql
--
-- Idempotent : peut etre rejoue sans risque.

INSERT INTO competence (libelle, description) VALUES
	('Chauffeur', 'Permis B, conduite des camionnettes de collecte/distribution'),
	('Cuisinier', 'Animation des cours de cuisine et ateliers anti-gaspi'),
	('Plombier', 'Interventions de plomberie dans le cadre du service reparation'),
	('Electricien', 'Interventions electriques dans le cadre du service reparation'),
	('Manutention', 'Chargement/dechargement, tri et stockage des denrees'),
	('Accueil', 'Accueil des beneficiaires lors des distributions')
ON CONFLICT (libelle) DO NOTHING;
