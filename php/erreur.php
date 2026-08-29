<?php
// Page d'erreur commune, branchee via ErrorDocument (cf. docker/front-*.conf).
// Volontairement autonome : pas d'appel a l'API ni a layout_top.php, pour
// qu'une panne du back-end n'empeche pas d'afficher l'erreur elle-meme.
require_once __DIR__ . '/inc/helpers.php';
require_once __DIR__ . '/inc/changeLanguage.php';
$lang = $LANGUAGE_CONTENTS;

$status = intval($_SERVER['REDIRECT_STATUS'] ?? 0);
$titles = [404 => $lang[450], 403 => $lang[451], 500 => $lang[452]];
$details = [404 => $lang[454], 403 => $lang[455], 500 => $lang[456]];

$title = $titles[$status] ?? $lang[453];
$detail = $details[$status] ?? $lang[453];
if (!$status) {
    $status = 404;
    $title = $lang[450];
    $detail = $lang[454];
}
http_response_code($status);
?>
<!DOCTYPE html>
<html lang="<?= str_starts_with($LOADED_LANGUAGE, 'en') ? 'en' : 'fr' ?>">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>NO MORE WASTE - <?= h($title) ?></title>
<link rel="stylesheet" href="/generic.css">
</head>
<body>
<main class="wrap">
  <section class="section">
    <h1><?= (int) $status ?></h1>
    <h2><?= h($title) ?></h2>
    <p class="muted"><?= h($detail) ?></p>
    <div class="actions">
      <a class="btn" href="/frontoffice/index.php"><?= h($lang[102]) ?></a>
      <a class="btn-secondary" href="/frontoffice/aide.php"><?= h($lang[57]) ?></a>
    </div>
  </section>
</main>
</body>
</html>
