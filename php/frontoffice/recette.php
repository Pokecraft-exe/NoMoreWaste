<?php
require_once __DIR__ . '/../inc/config.php';
require_once __DIR__ . '/../inc/auth.php';
require_once __DIR__ . '/../inc/helpers.php';
require_once __DIR__ . '/../inc/changeLanguage.php';
$lang = $LANGUAGE_CONTENTS;

$user = current_user();
$id = intval($_GET['id'] ?? 0);
$res = api_get('/api/v1/ressources-cuisine', ['id' => $id], session_token());
$recette = $res['data'] ?? null;

$pageTitle = $recette ? $recette['titre'] : $lang[131];
$active = 'cuisine';
include __DIR__ . '/../inc/layout_top.php';

if (!$recette) {
    echo '<div class="empty">Cette recette n\'existe pas.</div>';
} else {
    ?>
    <p><a href="cuisine.php">&larr; <?= h($lang[101]) ?></a></p>
    <section class="section">
      <h1><?= h($recette['titre']) ?></h1>
      <div class="grid grid-2">
        <div class="card">
          <h3><?= h($lang[127]) ?></h3>
          <ul><?php foreach ($recette['ingredients'] as $i): ?><li><?= h($i) ?></li><?php endforeach; ?></ul>
        </div>
        <div class="card">
          <h3><?= h($lang[129]) ?></h3>
          <ul><?php foreach ($recette['outils'] as $o): ?><li><?= h($o) ?></li><?php endforeach; ?></ul>
        </div>
      </div>
      <div class="card"><h3><?= h($lang[130]) ?></h3><p><?= nl2br(h($recette['contenu'])) ?></p></div>
      <?php if (!empty($recette['video'])): ?><p><a href="<?= h($recette['video']) ?>" target="_blank" rel="noopener"><?= h($lang[347]) ?> &rarr;</a></p><?php endif; ?>
    </section>
    <?php
}

include __DIR__ . '/../inc/layout_bottom.php';
