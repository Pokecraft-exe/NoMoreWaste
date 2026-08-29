<?php

require_once __DIR__ . '/api.php';

function session_token() {
    if (!empty($_SESSION['access_token']) && !empty($_SESSION['token_expires']) && $_SESSION['token_expires'] > time()) {
        return $_SESSION['access_token'];
    }
    return null;
}

function current_user() {
    $token = session_token();
    if (!$token) {
        return null;
    }
    if (isset($_SESSION['profil'])) {
        return $_SESSION['profil'];
    }
    $res = api_get('/api/v1/profil', [], $token);
    if ($res['status'] !== 200) {
        do_logout();
        return null;
    }
    $_SESSION['profil'] = $res['data'];
    return $res['data'];
}

function is_staff($user) {
    return $user && ($user['type_utilisateur'] === 'administrateur' || $user['type_utilisateur'] === 'responsable');
}

function is_adherent($user) {
    return $user && $user['type_utilisateur'] === 'adherent';
}

function is_benevole($user) {
    return $user && $user['type_utilisateur'] === 'benevole';
}

function is_commercant($user) {
    return $user && $user['type_utilisateur'] === 'commercant';
}

function do_login($email, $password) {
    $res = api_basic('POST', '/oauth/v3/token', $email, $password);
    if ($res['status'] !== 200 || empty($res['data']['access_token'])) {
        return $res['data']['error_description'] ?? 'Identifiants invalides.';
    }
    $_SESSION['access_token'] = $res['data']['access_token'];
    $_SESSION['token_expires'] = time() + intval($res['data']['expires_in']);
    unset($_SESSION['profil']);
    return null;
}

function do_logout() {
    unset($_SESSION['access_token'], $_SESSION['token_expires'], $_SESSION['profil']);
}

function do_register($fields) {
    $res = api_request('POST', '/oauth/v3/inscription', $fields);
    if ($res['status'] !== 201) {
        return $res['data']['error_description'] ?? 'Inscription impossible.';
    }
    return null;
}

function guard_login($user) {
    if ($user) {
        return true;
    }
    echo '<div class="access-denied"><h2>Connexion requise</h2><p class="muted">Cette page est reservee aux comptes identifies. ';
    echo '<a href="connexion.php">Connectez-vous / inscrivez-vous</a>.</p></div>';
    return false;
}

function require_staff_or_404($user) {
    if (is_staff($user)) {
        return;
    }
    http_response_code(404);
    echo '<!DOCTYPE html><html lang="fr"><head><meta charset="UTF-8">';
    echo '<meta name="viewport" content="width=device-width, initial-scale=1.0">';
    echo '<title>NO MORE WASTE - Page introuvable</title>';
    echo '<link rel="stylesheet" href="../../generic.css"></head><body><main class="wrap">';
    echo '<section class="section"><h1>404</h1><p class="muted">Cette page n\'existe pas.</p>';
    echo '<a class="btn" href="../frontoffice/index.php">Retour a l\'accueil</a></section></main></body></html>';
    exit;
}
