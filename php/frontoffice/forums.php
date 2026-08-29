<?php
require_once __DIR__ . '/../inc/config.php';
require_once __DIR__ . '/../inc/auth.php';
require_once __DIR__ . '/../inc/helpers.php';
require_once __DIR__ . '/../inc/changeLanguage.php';
$lang = $LANGUAGE_CONTENTS;

$user = current_user();
$error = null;

if ($_SERVER['REQUEST_METHOD'] === 'POST' && $user) {
    $res = api_put('/api/v1/forum', [
        'titre' => $_POST['titre'] ?? '',
        'message' => $_POST['message'] ?? '',
    ], session_token());
    if ($res['status'] === 201) {
        header('Location: forum.php?id=' . $res['data']['forum_thread_id']);
        exit;
    }
    $error = $res['data']['error_description'] ?? $lang[292];
}

$q = query_param('q');
$res = api_get('/api/v1/forum', api_params(['from' => 0, 'size' => 30, 'q' => $q]), session_token());
$threads = $res['data']['threads'] ?? [];

$pageTitle = $lang[9];
$active = 'forums';
include __DIR__ . '/../inc/layout_top.php';
?>

<section class="section">
  <div class="section-head">
    <div><h1><?= h($lang[195]) ?></h1><p class="muted"><?= h($lang[196]) ?></p></div>
    <?php if ($user): ?><a class="btn-sm" href="#nouveau-forum"><?= h($lang[182]) ?></a><?php endif; ?>
  </div>

  <form class="filters" method="get">
    <div class="field"><label for="q"><?= h($lang[79]) ?></label><input id="q" name="q" value="<?= h($q) ?>"></div>
    <div class="actions"><button type="submit"><?= h($lang[79]) ?></button></div>
  </form>

  <div class="results">
    <?php foreach ($threads as $t): ?>
    <a class="card-link" href="forum.php?id=<?= (int) $t['forum_thread_id'] ?>">
      <div class="card-title"><?= h($t['titre']) ?></div>
      <p class="muted"><?= h(mb_strimwidth($t['message'], 0, 120, '...')) ?></p>
      <div class="card-meta"><span><?= h($lang[297]) ?> <?= h($t['auteur']) ?></span><span><?= h(time_ago($t['date_creation'])) ?></span><span><?= (int) $t['vues'] ?> <?= h($lang[298]) ?></span></div>
    </a>
    <?php endforeach; ?>
    <?php if (empty($threads)): ?><div class="empty"><?= h($lang[207]) ?></div><?php endif; ?>
  </div>
</section>

<?php if ($user): ?>
<section class="section card" id="nouveau-forum">
  <h2><?= h($lang[182]) ?></h2>
  <?php if ($error): ?><div class="note-box"><?= h($error) ?></div><?php endif; ?>
  <form method="post">
    <div class="field"><label for="titre"><?= h($lang[20]) ?></label><input id="titre" name="titre" required></div>
    <div class="field"><label for="message"><?= h($lang[35]) ?></label><textarea id="message" name="message" required></textarea></div>
    <div class="actions"><button type="submit"><?= h($lang[78]) ?></button></div>
  </form>
</section>
<?php endif; ?>

<?php include __DIR__ . '/../inc/layout_bottom.php'; ?>
