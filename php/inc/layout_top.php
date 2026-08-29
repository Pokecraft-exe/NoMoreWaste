<?php

require_once __DIR__ . '/changeLanguage.php';
$lang = $LANGUAGE_CONTENTS;

$navMode = $navMode ?? 'front';
$active = $active ?? '';
$pageTitle = $pageTitle ?? 'NO MORE WASTE';
$assetPrefix = $navMode === 'back' ? '../../' : '../../';

// The language switcher lives under frontoffice/, so a backoffice page has
// to point back at itself with the ../backoffice/ prefix or the "return to
// this page" link would 404.
$currentFile = basename($_SERVER['PHP_SELF']);
$langRedirect = $navMode === 'back' ? '../backoffice/' . $currentFile : $currentFile;
$langSwitchHref = ($navMode === 'back' ? '../frontoffice/' : '') . 'languages.php'
    . '?lang=' . h($LOADED_LANGUAGE) . '&redirect=' . urlencode($langRedirect);

$unreadCount = 0;
if ($user) {
    $notifRes = api_get('/api/v1/notifications', [], session_token());
    if ($notifRes['status'] === 200) {
        $unreadCount = intval($notifRes['data']['non_lues'] ?? 0);
    }
}

if ($navMode === 'back') {
    $links = [
        ['href' => 'index.php', 'label' => $lang[43], 'key' => 'dashboard'],
        ['href' => 'comptes.php', 'label' => $lang[44], 'key' => 'comptes'],
        ['href' => 'adherents.php', 'label' => $lang[442], 'key' => 'adherents'],
        ['href' => 'commercants.php', 'label' => $lang[396], 'key' => 'commercants'],
        ['href' => 'forums.php', 'label' => $lang[9], 'key' => 'forums'],
        ['href' => 'annonces.php', 'label' => $lang[17], 'key' => 'annonces'],
        ['href' => 'collectes.php', 'label' => $lang[45], 'key' => 'collectes'],
        ['href' => 'distributions.php', 'label' => $lang[46], 'key' => 'distributions'],
        ['href' => 'stocks.php', 'label' => $lang[409], 'key' => 'stocks'],
        ['href' => 'services.php', 'label' => $lang[423], 'key' => 'services'],
        ['href' => 'planning.php', 'label' => $lang[47], 'key' => 'planning'],
        ['href' => 'candidatures.php', 'label' => $lang[48], 'key' => 'candidatures'],
        ['href' => 'benevoles.php', 'label' => $lang[437], 'key' => 'benevoles'],
        ['href' => 'recettes.php', 'label' => $lang[49], 'key' => 'recettes'],
        ['href' => 'signalements.php', 'label' => $lang[50], 'key' => 'signalements'],
        ['href' => 'tickets.php', 'label' => $lang[51], 'key' => 'tickets'],
        ['href' => '../frontoffice/index.php', 'label' => $lang[52], 'key' => 'front'],
    ];
    $brandHref = 'index.php';
} else {
    $links = [
        ['href' => 'index.php', 'label' => $lang[38], 'key' => 'accueil'],
        ['href' => 'forums.php', 'label' => $lang[9], 'key' => 'forums'],
        ['href' => 'cuisine.php', 'label' => $lang[39], 'key' => 'cuisine'],
        ['href' => 'services.php', 'label' => $lang[423], 'key' => 'services'],
        ['href' => 'annonces.php', 'label' => $lang[17], 'key' => 'annonces'],
        ['href' => 'profil.php', 'label' => $lang[40], 'key' => 'profil'],
    ];
    if (is_benevole($user)) {
        $links[] = ['href' => 'planning.php', 'label' => $lang[41], 'key' => 'planning'];
    }
    if (is_staff($user)) {
        $links[] = ['href' => '../backoffice/index.php', 'label' => $lang[42], 'key' => 'backoffice'];
    }
    $brandHref = 'index.php';
}
?>
<!DOCTYPE html>
<html lang="<?= str_starts_with($LOADED_LANGUAGE, 'en') ? 'en' : 'fr' ?>">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>NO MORE WASTE - <?= h($pageTitle) ?></title>
<link rel="stylesheet" href="<?= h($assetPrefix) ?>generic.css">
</head>
<body>
<header class="topnav">
<a class="brand" href="<?= h($brandHref) ?>"><span class="brand-mark">N</span> NO MORE WASTE</a>
<nav>
<?php foreach ($links as $link): ?>
<a href="<?= h($link['href']) ?>" class="<?= $link['key'] === $active ? 'active' : '' ?>"><?= h($link['label']) ?></a>
<?php endforeach; ?>
</nav>
<div class="who">
<a class="lang-switch" href="<?= $langSwitchHref ?>" title="<?= h($lang[57]) ?>">
<img src="<?= h($assetPrefix) ?>lang/<?= h($LOADED_LANGUAGE) ?>.svg" alt="<?= h($LOADED_LANGUAGE) ?> language switch button" height="20" width="28">
</a>
<?php if ($user): ?>
<a class="notif-bell" href="<?= $navMode === 'back' ? '' : '' ?>notifications.php" title="<?= h($lang[58]) ?>" aria-label="<?= h($lang[58]) ?>">🔔<?php if ($unreadCount > 0): ?><span class="notif-count"><?= $unreadCount > 9 ? '9+' : $unreadCount ?></span><?php endif; ?></a>
<span class="badge"><?= h(full_name($user)) ?> &middot; <?= h(user_type_label($user['type_utilisateur'])) ?></span>
<a class="btn-secondary btn-sm" href="<?= $navMode === 'back' ? '../frontoffice/deconnexion.php' : 'deconnexion.php' ?>"><?= h($lang[53]) ?></a>
<?php else: ?>
<span class="badge"><?= h($lang[55]) ?></span>
<a class="btn-secondary btn-sm" href="<?= $navMode === 'back' ? '../frontoffice/connexion.php' : 'connexion.php' ?>"><?= h($lang[54]) ?></a>
<?php endif; ?>
</div>
</header>

<a class="help-fab" href="<?= $navMode === 'back' ? '../frontoffice/aide.php' : 'aide.php' ?>">? <?= h($lang[56]) ?></a>

<?php $legalPrefix = $navMode === 'back' ? '../frontoffice/' : ''; ?>
<main class="wrap">
