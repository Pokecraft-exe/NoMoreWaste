<?php
require_once __DIR__ . '/../inc/config.php';
require_once __DIR__ . '/../inc/auth.php';
require_once __DIR__ . '/../inc/helpers.php';
require_once __DIR__ . '/../inc/changeLanguage.php';
$lang = $LANGUAGE_CONTENTS;

$user = current_user();
$error = null;

$categories = [
    'covoiturage' => $lang[118], 'reparation' => $lang[116], 'gardiennage' => $lang[117],
    'location' => $lang[115], 'vente' => $lang[113], 'don' => 'Don',
];

if ($_SERVER['REQUEST_METHOD'] === 'POST' && $user) {
    $prix = trim($_POST['prix'] ?? '');
    $res = api_put('/api/v1/annonces', [
        'categorie' => $_POST['categorie'] ?? '',
        'titre' => $_POST['titre'] ?? '',
        'description' => $_POST['description'] ?? '',
        'prix' => $prix,
    ], session_token());
    if ($res['status'] === 201) {
        header('Location: annonces.php');
        exit;
    }
    $error = $res['data']['error_description'] ?? $lang[292];
}

$categorie = query_param('categorie');
$q = query_param('q');
$res = api_get('/api/v1/annonces', api_params(['from' => 0, 'size' => 30, 'categorie' => $categorie, 'q' => $q]), session_token());
$annonces = $res['data']['annonces'] ?? [];

$pageTitle = $lang[17];
$active = 'annonces';
include __DIR__ . '/../inc/layout_top.php';
?>

<section class="section">
  <div class="section-head">
    <div><h1><?= h($lang[184]) ?></h1><p class="muted"><?= h($lang[118]) ?>, <?= h($lang[349]) ?></p></div>
    <?php if ($user): ?><a class="btn-sm" href="#deposer"><?= h($lang[183]) ?></a><?php endif; ?>
  </div>

  <div class="chip-row">
    <a class="chip <?= $categorie === '' ? 'active' : '' ?>" href="annonces.php"><?= h($lang[110]) ?></a>
    <?php foreach ($categories as $val => $label): ?>
    <a class="chip <?= $categorie === $val ? 'active' : '' ?>" href="annonces.php?categorie=<?= h($val) ?>"><?= h($label) ?></a>
    <?php endforeach; ?>
  </div>

  <form class="filters" method="get">
    <input type="hidden" name="categorie" value="<?= h($categorie) ?>">
    <div class="field"><label for="q"><?= h($lang[79]) ?></label><input id="q" name="q" value="<?= h($q) ?>"></div>
    <div class="actions"><button type="submit"><?= h($lang[79]) ?></button></div>
  </form>

  <div class="results grid grid-auto">
    <?php foreach ($annonces as $a): ?>
    <a class="card-link" href="annonce.php?id=<?= (int) $a['annonce_echange_id'] ?>">
      <div class="card-title"><?= h($a['titre']) ?></div>
      <p class="muted"><?= h(mb_strimwidth($a['description'], 0, 90, '...')) ?></p>
      <div class="card-meta">
        <span class="badge"><?= h($categories[$a['categorie']] ?? $a['categorie']) ?></span>
        <span><?= $a['prix'] !== null ? h($a['prix']) . ' €' : $lang[112] ?></span>
      </div>
    </a>
    <?php endforeach; ?>
    <?php if (empty($annonces)): ?><div class="empty"><?= h($lang[210]) ?></div><?php endif; ?>
  </div>
</section>

<?php if ($user): ?>
<section class="section card" id="deposer">
  <h2><?= h($lang[183]) ?></h2>
  <?php if ($error): ?><div class="note-box"><?= h($error) ?></div><?php endif; ?>
  <form method="post">
    <div class="field"><label for="categorie-new"><?= h($lang[107]) ?></label>
      <select id="categorie-new" name="categorie" required>
        <?php foreach ($categories as $val => $label): ?><option value="<?= h($val) ?>"><?= h($label) ?></option><?php endforeach; ?>
      </select>
    </div>
    <div class="field"><label for="titre-new"><?= h($lang[20]) ?></label><input id="titre-new" name="titre" required></div>
    <div class="field"><label for="desc-new"><?= h($lang[22]) ?></label><textarea id="desc-new" name="description" required></textarea></div>
    <div class="field"><label for="prix-new"><?= h($lang[21]) ?> <?= h($lang[350]) ?></label><input id="prix-new" name="prix" type="number" min="0"></div>
    <div class="actions"><button type="submit"><?= h($lang[78]) ?></button></div>
  </form>
</section>
<?php endif; ?>

<?php include __DIR__ . '/../inc/layout_bottom.php'; ?>
