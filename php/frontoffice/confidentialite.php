<?php
require_once __DIR__ . '/../inc/config.php';
require_once __DIR__ . '/../inc/auth.php';
require_once __DIR__ . '/../inc/helpers.php';
require_once __DIR__ . '/../inc/changeLanguage.php';
$lang = $LANGUAGE_CONTENTS;

$user = current_user();
$pageTitle = $lang[271];
$active = '';
include __DIR__ . '/../inc/layout_top.php';
?>

<section class="section">
  <h1><?= h($lang[271]) ?></h1>

  <div class="card">
    <h3><?= h($lang[268]) ?></h3>
    <p><?= h($lang[359]) ?></p>
  </div>

  <div class="card" style="margin-top:12px;">
    <h3><?= h($lang[269]) ?></h3>
    <p><?= h($lang[360]) ?></p>
  </div>

  <div class="card" style="margin-top:12px;">
    <h3><?= h($lang[270]) ?></h3>
    <p><?= h($lang[361]) ?></p>
  </div>

  <div class="card" style="margin-top:12px;">
    <h3><?= h($lang[266]) ?></h3>
    <p><?= h($lang[366]) ?> <a href="aide.php"><?= h($lang[56]) ?></a>.</p>
  </div>

  <div class="card" style="margin-top:12px;">
    <h3><?= h($lang[267]) ?></h3>
    <p><?= h($lang[362]) ?></p>
  </div>
</section>

<?php include __DIR__ . '/../inc/layout_bottom.php'; ?>
