-- ===================================================================
-- NO MORE WASTE - donnees de demonstration pour sql.sql (schema reel).
-- Le compte administrateur est deja cree par le bootstrap de sql.sql
-- (admin@nomorewaste.local / ChangeMe123!) ; ce script ajoute des comptes
-- de chaque type et des donnees pour exercer chaque fonctionnalite.
-- Mots de passe en clair dans comptes.txt (hashes ici via pgcrypto).
-- Prerequis : sql.sql deja execute sur une base vide.
-- ===================================================================

BEGIN;

DO $$
DECLARE
    v_personne_id INT;
    v_compte_id INT;
    v_adherent_id INT;
    v_benevole_id INT;
    v_commercant_id INT;
    v_forfait_standard INT;
    v_forfait_commerce INT;

    v_julie_compte INT; v_julie_adherent INT;
    v_sophie_compte INT; v_sophie_adherent INT;
    v_karim_compte INT; v_karim_benevole INT;
    v_lea_compte INT; v_lea_benevole INT;
    v_nadia_compte INT; v_nadia_benevole INT;
    v_boulangerie_compte INT; v_boulangerie_commercant INT;
    v_epicerie_compte INT; v_epicerie_commercant INT;
    v_camille_compte INT;

    v_thread1 INT; v_thread2 INT;
    v_annonce1 INT; v_annonce2 INT;
    v_collecte1 INT; v_collecte2 INT; v_collecte3 INT;
    v_distribution1 INT; v_distribution2 INT;
    v_produit_pain INT; v_produit_legumes INT; v_produit_conserves INT;
    v_signalement1 INT;
BEGIN
    SELECT forfait_id INTO v_forfait_standard FROM forfait WHERE libelle = 'Adhesion standard';
    SELECT forfait_id INTO v_forfait_commerce FROM forfait WHERE libelle = 'Adhesion commercant';

    -- Adherents --------------------------------------------------------
    INSERT INTO personne (nom, prenom, email, telephone, adresse, code_postal, ville)
        VALUES ('Martin', 'Julie', 'julie.martin@example.com', '+33611110001', '4 rue des Lilas', '75011', 'Paris')
        RETURNING personne_id INTO v_personne_id;
    INSERT INTO compte (personne_id, mot_de_passe, type_utilisateur) VALUES (v_personne_id, crypt('Julie2026!', gen_salt('bf')), 'adherent') RETURNING compte_id INTO v_julie_compte;
    INSERT INTO adherent (personne_id) VALUES (v_personne_id) RETURNING adherent_id INTO v_julie_adherent;
    INSERT INTO adhesion_association (adherent_id, forfait_id, date_debut, date_fin) VALUES (v_julie_adherent, v_forfait_standard, current_date, current_date + interval '1 year');

    INSERT INTO personne (nom, prenom, email, telephone, adresse, code_postal, ville)
        VALUES ('Bernard', 'Sophie', 'sophie.bernard@example.com', '+33611110002', '9 avenue Foch', '69003', 'Lyon')
        RETURNING personne_id INTO v_personne_id;
    INSERT INTO compte (personne_id, mot_de_passe, type_utilisateur) VALUES (v_personne_id, crypt('Sophie2026!', gen_salt('bf')), 'adherent') RETURNING compte_id INTO v_sophie_compte;
    INSERT INTO adherent (personne_id) VALUES (v_personne_id) RETURNING adherent_id INTO v_sophie_adherent;
    INSERT INTO adhesion_association (adherent_id, forfait_id, date_debut, date_fin) VALUES (v_sophie_adherent, v_forfait_standard, current_date, current_date + interval '1 year');

    -- Benevoles ----------------------------------------------------------
    INSERT INTO personne (nom, prenom, email, telephone, adresse, code_postal, ville)
        VALUES ('Haddad', 'Karim', 'karim.haddad@example.com', '+33611110003', '2 place Bellecour', '69002', 'Lyon')
        RETURNING personne_id INTO v_personne_id;
    INSERT INTO compte (personne_id, mot_de_passe, type_utilisateur) VALUES (v_personne_id, crypt('Karim2026!', gen_salt('bf')), 'benevole') RETURNING compte_id INTO v_karim_compte;
    INSERT INTO benevole (personne_id, statut, disponibilite) VALUES (v_personne_id, 'actif', 'Week-ends') RETURNING benevole_id INTO v_karim_benevole;

    INSERT INTO personne (nom, prenom, email, telephone, adresse, code_postal, ville)
        VALUES ('Dubois', 'Lea', 'lea.dubois@example.com', '+33611110004', '18 rue de Belleville', '75020', 'Paris')
        RETURNING personne_id INTO v_personne_id;
    INSERT INTO compte (personne_id, mot_de_passe, type_utilisateur) VALUES (v_personne_id, crypt('Lea2026!', gen_salt('bf')), 'benevole') RETURNING compte_id INTO v_lea_compte;
    INSERT INTO benevole (personne_id, statut, disponibilite) VALUES (v_personne_id, 'actif', 'Soirees') RETURNING benevole_id INTO v_lea_benevole;

    INSERT INTO personne (nom, prenom, email, telephone, adresse, code_postal, ville)
        VALUES ('Cherif', 'Nadia', 'nadia.cherif@example.com', '+33611110005', '5 rue Garibaldi', '69007', 'Lyon')
        RETURNING personne_id INTO v_personne_id;
    INSERT INTO compte (personne_id, mot_de_passe, type_utilisateur) VALUES (v_personne_id, crypt('Nadia2026!', gen_salt('bf')), 'benevole') RETURNING compte_id INTO v_nadia_compte;
    INSERT INTO benevole (personne_id, statut, disponibilite) VALUES (v_personne_id, 'actif', 'Matinees') RETURNING benevole_id INTO v_nadia_benevole;

    -- Commercants ----------------------------------------------------------
    INSERT INTO personne (nom, prenom, email, telephone, adresse, code_postal, ville)
        VALUES ('Boulangerie du Coin', 'Contact', 'contact@boulangerieducoin.example.com', '+33611110006', '3 rue du Pain', '69001', 'Lyon')
        RETURNING personne_id INTO v_personne_id;
    INSERT INTO compte (personne_id, mot_de_passe, type_utilisateur) VALUES (v_personne_id, crypt('Boulangerie2026!', gen_salt('bf')), 'commercant') RETURNING compte_id INTO v_boulangerie_compte;
    INSERT INTO commercant (personne_id, raison_sociale, adresse, code_postal, ville, pays)
        VALUES (v_personne_id, 'Boulangerie du Coin', '3 rue du Pain', '69001', 'Lyon', 'France') RETURNING commercant_id INTO v_boulangerie_commercant;
    INSERT INTO adhesion_commercant (commercant_id, forfait_id, date_debut, date_fin) VALUES (v_boulangerie_commercant, v_forfait_commerce, current_date, current_date + interval '1 year');

    INSERT INTO personne (nom, prenom, email, telephone, adresse, code_postal, ville)
        VALUES ('Epicerie Verte', 'Contact', 'contact@epicerieverte.example.com', '+33611110007', '12 rue Verte', '75011', 'Paris')
        RETURNING personne_id INTO v_personne_id;
    INSERT INTO compte (personne_id, mot_de_passe, type_utilisateur) VALUES (v_personne_id, crypt('EpicerieVerte2026!', gen_salt('bf')), 'commercant') RETURNING compte_id INTO v_epicerie_compte;
    INSERT INTO commercant (personne_id, raison_sociale, adresse, code_postal, ville, pays)
        VALUES (v_personne_id, 'Epicerie Verte', '12 rue Verte', '75011', 'Paris', 'France') RETURNING commercant_id INTO v_epicerie_commercant;
    INSERT INTO adhesion_commercant (commercant_id, forfait_id, date_debut, date_fin) VALUES (v_epicerie_commercant, v_forfait_commerce, current_date, current_date + interval '1 year');

    -- Visiteur ----------------------------------------------------------
    INSERT INTO personne (nom, prenom, email, telephone)
        VALUES ('Petit', 'Camille', 'camille.petit@example.com', '+33611110008')
        RETURNING personne_id INTO v_personne_id;
    INSERT INTO compte (personne_id, mot_de_passe, type_utilisateur) VALUES (v_personne_id, crypt('Camille2026!', gen_salt('bf')), 'visiteur') RETURNING compte_id INTO v_camille_compte;

    -- Candidatures benevole (une en attente, une validee = Karim ci-dessus deja actif) --
    INSERT INTO candidature_benevole (personne_id, statut, motivation, disponibilite)
        SELECT personne_id, 'recue', 'Je veux aider a lutter contre le gaspillage.', 'Week-ends'
        FROM compte WHERE compte_id = v_camille_compte;

    -- Catalogue produits (denrees) ---------------------------------------
    INSERT INTO stock_produit (nom, unite) VALUES ('Pain', 'kg') RETURNING stock_produit_id INTO v_produit_pain;
    INSERT INTO stock_produit (nom, unite) VALUES ('Legumes', 'kg') RETURNING stock_produit_id INTO v_produit_legumes;
    INSERT INTO stock_produit (nom, unite) VALUES ('Conserves', 'unites') RETURNING stock_produit_id INTO v_produit_conserves;

    -- Collectes -----------------------------------------------------------
    INSERT INTO collecte (commercant_id, lieu, date_collecte, heure_collecte, statut, description)
        VALUES (v_boulangerie_commercant, 'Boulangerie du Coin, Lyon', current_date + 3, '18:00', 'planifiee', 'Recuperation des invendus de fin de journee.')
        RETURNING collecte_id INTO v_collecte1;
    INSERT INTO collecte_denree (collecte_id, stock_produit_id, quantite, non_perissable, propose_par, propose_par_type, confirmee)
        VALUES (v_collecte1, v_produit_pain, 12, '0', (SELECT compte_id FROM compte WHERE personne_id = (SELECT personne_id FROM commercant WHERE commercant_id = v_boulangerie_commercant)), 'commercant', '0');
    INSERT INTO collecte_benevole_affecte (collecte_id, benevole_id, role_mission) VALUES (v_collecte1, v_karim_benevole, 'Collecte et tri');
    INSERT INTO participation_collecte (collecte_id, benevole_id, role_mission) VALUES (v_collecte1, v_lea_benevole, 'chauffeur');

    INSERT INTO collecte (lieu, date_collecte, heure_collecte, statut, description)
        VALUES ('Marche de Belleville, Paris', current_date - 4, '09:00', 'terminee', 'Collecte hebdomadaire en marche.')
        RETURNING collecte_id INTO v_collecte2;
    INSERT INTO collecte_denree (collecte_id, stock_produit_id, quantite, non_perissable, propose_par, propose_par_type, confirmee, date_ajout)
        VALUES (v_collecte2, v_produit_legumes, 20, '0', (SELECT compte_id FROM compte LIMIT 1), 'staff', '1', now() - interval '4 days');
    UPDATE collecte SET stock_mis_a_jour = '1' WHERE collecte_id = v_collecte2;
    UPDATE stock_produit SET quantite_disponible = quantite_disponible + 20 WHERE stock_produit_id = v_produit_legumes;

    INSERT INTO collecte (lieu, date_collecte, statut, description)
        VALUES ('Marche des Lices, Rennes', current_date + 10, 'planifiee', 'Collecte mensuelle.')
        RETURNING collecte_id INTO v_collecte3;

    -- Distributions ---------------------------------------------------------
    INSERT INTO distribution (lieu, date_distribution, heure_distribution, statut, quota_par_adherent)
        VALUES ('Centre social de Belleville, Paris', current_date + 5, '14:00', 'planifiee', 2)
        RETURNING distribution_id INTO v_distribution1;
    INSERT INTO distribution_denree (distribution_id, stock_produit_id, quantite) VALUES (v_distribution1, v_produit_legumes, 15);
    INSERT INTO distribution_benevole_affecte (distribution_id, benevole_id, role_mission) VALUES (v_distribution1, v_nadia_benevole, 'Accueil des beneficiaires');
    INSERT INTO reservation (distribution_id, adherent_id, stock_produit_id, quantite) VALUES (v_distribution1, v_julie_adherent, v_produit_legumes, 2);

    INSERT INTO distribution (lieu, date_distribution, statut, quota_par_adherent)
        VALUES ('Foyer Saint-Martin, Marseille', current_date + 12, 'planifiee', 1)
        RETURNING distribution_id INTO v_distribution2;

    -- Forum -----------------------------------------------------------------
    INSERT INTO forum_thread (compte_id, titre, message, vues)
        VALUES (v_julie_compte, 'Astuces pour conserver les legumes plus longtemps', 'Quelqu''un a des astuces pour eviter le gaspillage de legumes frais ?', 12)
        RETURNING forum_thread_id INTO v_thread1;
    INSERT INTO forum_message (forum_thread_id, compte_id, message) VALUES (v_thread1, v_karim_compte, 'Je congele tout ce qui approche la date limite !');
    INSERT INTO forum_message (forum_thread_id, compte_id, message) VALUES (v_thread1, v_sophie_compte, 'Le pesto de fanes de radis marche tres bien aussi.');

    INSERT INTO forum_thread (compte_id, titre, message, vues)
        VALUES (v_lea_compte, 'Recette anti-gaspi de la semaine', 'Je partage ma recette de soupe de fanes.', 5)
        RETURNING forum_thread_id INTO v_thread2;

    -- Ressources cuisine ------------------------------------------------------
    INSERT INTO ressource_cuisine (titre, ingredients, outils, contenu, created_by)
        VALUES ('Soupe de fanes de radis', ARRAY['fanes de radis', 'oignon', 'bouillon'], ARRAY['mixeur'], 'Faire revenir l''oignon, ajouter les fanes et le bouillon, laisser mijoter 15 minutes puis mixer.', v_camille_compte);
    INSERT INTO ressource_cuisine (titre, ingredients, outils, contenu, created_by)
        VALUES ('Pain perdu aux fruits murs', ARRAY['pain rassis', 'oeufs', 'lait', 'fruits murs'], ARRAY['poele'], 'Tremper le pain dans le melange oeufs-lait, faire dorer a la poele, servir avec les fruits.', v_camille_compte);

    -- Annonces ----------------------------------------------------------------
    INSERT INTO annonce_echange (compte_id, categorie, titre, description, prix)
        VALUES (v_karim_compte, 'don', 'Cagette de legumes en trop', 'Je donne une cagette de legumes recuperes en collecte.', NULL)
        RETURNING annonce_echange_id INTO v_annonce1;
    INSERT INTO message_annonce_echange (annonce_echange_id, expediteur_id, message) VALUES (v_annonce1, v_sophie_compte, 'Bonjour, je suis interessee, comment on fait ?');

    INSERT INTO annonce_echange (compte_id, categorie, titre, description, prix)
        VALUES (v_boulangerie_compte, 'vente', 'Composteur de jardin', 'Composteur en bois, tres peu servi.', 25.00)
        RETURNING annonce_echange_id INTO v_annonce2;

    -- Tickets -------------------------------------------------------------------
    INSERT INTO ticket (auteur_id, sujet, message) VALUES (v_camille_compte, 'Comment devenir benevole ?', 'Bonjour, je souhaiterais savoir comment rejoindre l''equipe de benevoles.');
    INSERT INTO ticket (contact_nom, contact_email, sujet, message) VALUES ('Anonyme', 'anonyme@example.com', 'Question sur les horaires', 'Bonjour, quels sont vos horaires d''ouverture ?');

    -- Signalements (un traite avec archive, un ouvert) ---------------------------
    INSERT INTO signalement (type_signalement, forum_thread_id, forum_message_id, signale_par, motif, statut, commentaire, traite_par, date_traitement)
        SELECT 'forum_message', v_thread1, forum_message_id, v_sophie_compte, 'Message hors sujet.', 'traite', 'Verifie : le message reste dans le cadre du forum.', compte_id, now()
        FROM forum_message WHERE forum_thread_id = v_thread1 AND compte_id = v_karim_compte
        LIMIT 1
        RETURNING signalement_id INTO v_signalement1;
    INSERT INTO message_archive (signalement_id, auteur_id, message, date_message)
        SELECT v_signalement1, v_karim_compte, message, date_envoi FROM forum_message WHERE forum_thread_id = v_thread1 AND compte_id = v_karim_compte LIMIT 1;

    INSERT INTO signalement (type_signalement, forum_thread_id, signale_par, motif)
        VALUES ('forum', v_thread2, v_karim_compte, 'Titre trompeur par rapport au contenu.');
END $$;

COMMIT;
