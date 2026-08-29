<?php
require_once __DIR__ . '/../inc/config.php';
require_once __DIR__ . '/../inc/auth.php';
require_once __DIR__ . '/../inc/helpers.php';
require_once __DIR__ . '/../inc/changeLanguage.php';
$lang = $LANGUAGE_CONTENTS;

$user = current_user();

$pageTitle = $lang[58];
$active = '';
include __DIR__ . '/../inc/layout_top.php';

if (guard_login($user)) {
    $res = api_get('/api/v1/notifications', [], session_token());
    $notifications = $res['data']['notifications'] ?? [];
    api_patch('/api/v1/notifications', [], session_token());
    ?>
    <section class="section">
      <h1><?= h($lang[58]) ?></h1>
      <?php if (empty($notifications)): ?>
        <div class="empty"><?= h($lang[218]) ?></div>
      <?php else: ?>
        <?php foreach ($notifications as $n): ?>
        <div class="thread-message">
          <div class="bubble">
            <div class="msg-meta"><span><?= h(time_ago($n['date_notification'])) ?></span></div>
            <div><?= h($n['message']) ?></div>
          </div>
        </div>
        <?php endforeach; ?>
      <?php endif; ?>
    </section>
    <?php
}

include __DIR__ . '/../inc/layout_bottom.php';
