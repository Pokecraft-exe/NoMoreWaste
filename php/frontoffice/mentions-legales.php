<?php
require_once __DIR__ . '/../inc/config.php';
require_once __DIR__ . '/../inc/auth.php';
require_once __DIR__ . '/../inc/helpers.php';
require_once __DIR__ . '/../inc/changeLanguage.php';
$lang = $LANGUAGE_CONTENTS;

$user = current_user();
$pageTitle = $lang[260];
$active = '';
include __DIR__ . '/../inc/layout_top.php';
?>

<section class="section">
  <h1><?= h($lang[260]) ?></h1>

  <div class="card">
    <h3><?= h($lang[261]) ?></h3>
    <p><?= h($lang[364]) ?></p>
    <p><?= h($lang[334]) ?><br>
      <?= h($lang[265]) ?> <a href="aide.php"><?= h($lang[56]) ?></a>.</p>
  </div>

  <div class="card" style="margin-top:12px;">
    <h3><?= h($lang[262]) ?></h3>
    <p><?= h($lang[335]) ?></p>
  </div>

  <div class="card" style="margin-top:12px;">
    <h3><?= h($lang[263]) ?></h3>
    <p><?= h($lang[363]) ?></p>
  </div>

  <div class="card" style="margin-top:12px;">
    <h3><?= h($lang[264]) ?></h3>
    <p><?= h($lang[365]) ?></p>
  </div>
</section>

<?php include __DIR__ . '/../inc/layout_bottom.php'; ?>
