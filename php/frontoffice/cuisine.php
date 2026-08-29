<?php
require_once __DIR__ . '/../inc/config.php';
require_once __DIR__ . '/../inc/auth.php';
require_once __DIR__ . '/../inc/helpers.php';
require_once __DIR__ . '/../inc/changeLanguage.php';
$lang = $LANGUAGE_CONTENTS;

$user = current_user();
$ingredient = query_param('ingredient');
$outil = query_param('outil');
$nom = query_param('nom');
$res = api_get('/api/v1/ressources-cuisine', api_params(['from' => 0, 'size' => 30, 'ingredient' => $ingredient, 'outil' => $outil, 'nom' => $nom]), session_token());
$recettes = $res['data']['recettes'] ?? [];

$pageTitle = $lang[39];
$active = 'cuisine';
include __DIR__ . '/../inc/layout_top.php';
?>

<section class="section">
  <h1><?= h($lang[204]) ?></h1>
  <p class="muted"><?= h($lang[205]) ?></p>
  <form class="filters" method="get">
    <div class="field"><label for="ingredient"><?= h($lang[389]) ?></label><input id="ingredient" name="ingredient" value="<?= h($ingredient) ?>" placeholder="Ex: pain, carotte..."></div>
    <div class="field"><label for="outil"><?= h($lang[390]) ?></label><input id="outil" name="outil" value="<?= h($outil) ?>" placeholder="Ex: mixeur, four..."></div>
    <div class="field"><label for="nom"><?= h($lang[391]) ?></label><input id="nom" name="nom" value="<?= h($nom) ?>" placeholder="Ex: soupe, pesto..."></div>
    <div class="actions">
      <button type="submit"><?= h($lang[79]) ?></button>
      <?php if ($ingredient || $outil || $nom): ?><a class="btn-secondary" href="cuisine.php"><?= h($lang[392]) ?></a><?php endif; ?>
    </div>
  </form>

  <div class="results grid grid-auto">
    <?php foreach ($recettes as $r): ?>
    <a class="card-link" href="recette.php?id=<?= (int) $r['ressource_cuisine_id'] ?>">
      <div class="card-title"><?= h($r['titre']) ?></div>
      <p class="muted"><?= h(implode(', ', array_slice($r['ingredients'] ?: [], 0, 4))) ?></p>
      <p class="muted"><?= h(implode(', ', $r['outils'] ?: [])) ?></p>
    </a>
    <?php endforeach; ?>
    <?php if (empty($recettes)): ?><div class="empty"><?= h($lang[206]) ?></div><?php endif; ?>
  </div>
</section>

<?php include __DIR__ . '/../inc/layout_bottom.php'; ?>
