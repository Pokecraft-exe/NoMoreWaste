<?php
require_once __DIR__ . '/../inc/config.php';
require_once __DIR__ . '/../inc/auth.php';
require_once __DIR__ . '/../inc/helpers.php';
require_once __DIR__ . '/../inc/changeLanguage.php';
$lang = $LANGUAGE_CONTENTS;

$user = current_user();

$pageTitle = $lang[38];
$active = 'accueil';
include __DIR__ . '/../inc/layout_top.php';

$forumsRes = api_get('/api/v1/forum', ['from' => 0, 'size' => 20], session_token());
$threads = $forumsRes['data']['threads'] ?? [];
$topThread = null;
foreach ($threads as $t) {
    if (!$topThread || $t['vues'] > $topThread['vues']) {
        $topThread = $t;
    }
}

$showCollecte = !is_adherent($user);
$nextCollecte = null;
if ($showCollecte) {
    $res = api_get('/api/v1/collectes', ['from' => 0, 'size' => 1, 'prochaine' => 1], session_token());
    $nextCollecte = $res['data']['collectes'][0] ?? null;
}

$distRes = api_get('/api/v1/distributions', ['from' => 0, 'size' => 1, 'prochaine' => 1], session_token());
$nextDistribution = $distRes['data']['distributions'][0] ?? null;
?>

<section class="section">
  <h1><?= h($lang[200]) ?></h1>
  <p class="muted"><?= h($lang[201]) ?></p>
</section>

<section class="hero-grid">
  <?php if ($topThread): ?>
  <a class="hero-card trending" href="forum.php?id=<?= (int) $topThread['forum_thread_id'] ?>">
    <div>
      <div class="hero-label"><?= h($lang[198]) ?></div>
      <h3><?= h($topThread['titre']) ?></h3>
      <p><?= h(mb_strimwidth($topThread['message'], 0, 120, '...')) ?></p>
    </div>
    <div class="hero-cta"><?= (int) $topThread['vues'] ?> <?= h($lang[298]) ?> &rarr;</div>
  </a>
  <?php else: ?>
  <div class="hero-card trending"><h3><?= h($lang[199]) ?></h3></div>
  <?php endif; ?>

  <?php if ($nextCollecte): ?>
  <a class="hero-card collecte" href="collecte.php?id=<?= (int) $nextCollecte['collecte_id'] ?>">
    <div>
      <div class="hero-label"><?= h($lang[202]) ?></div>
      <h3><?= h($nextCollecte['lieu']) ?></h3>
      <p><?= h($lang[344]) ?> <?= h(format_date_fr($nextCollecte['date_collecte'])) ?> <?= h($lang[300]) ?> <?= h($nextCollecte['heure_collecte']) ?></p>
    </div>
    <div class="hero-cta"><?= h($lang[345]) ?> &rarr;</div>
  </a>
  <?php endif; ?>

  <?php if ($nextDistribution): ?>
  <a class="hero-card distribution" href="distribution.php?id=<?= (int) $nextDistribution['distribution_id'] ?>">
    <div>
      <div class="hero-label"><?= h($lang[203]) ?></div>
      <h3><?= h($nextDistribution['lieu']) ?></h3>
      <p><?= h($lang[344]) ?> <?= h(format_date_fr($nextDistribution['date_distribution'])) ?> <?= h($lang[300]) ?> <?= h($nextDistribution['heure_distribution']) ?></p>
    </div>
    <div class="hero-cta"><?= h($lang[346]) ?> &rarr;</div>
  </a>
  <?php endif; ?>
</section>

<section class="section grid grid-2">
  <div class="card">
    <h3><?= h($lang[195]) ?></h3>
    <p class="muted"><?= h($lang[197]) ?></p>
    <a class="btn btn-secondary" href="forums.php"><?= h($lang[95]) ?></a>
  </div>
  <div class="card">
    <h3><?= h($lang[204]) ?></h3>
    <p class="muted"><?= h($lang[205]) ?></p>
    <a class="btn btn-secondary" href="cuisine.php"><?= h($lang[96]) ?></a>
  </div>
</section>

<?php include __DIR__ . '/../inc/layout_bottom.php'; ?>
