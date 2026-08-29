<?php
require_once __DIR__ . '/../inc/config.php';
require_once __DIR__ . '/../inc/auth.php';
require_once __DIR__ . '/../inc/helpers.php';
require_once __DIR__ . '/../inc/changeLanguage.php';
$lang = $LANGUAGE_CONTENTS;

$user = current_user();
$threadId = intval($_GET['id'] ?? 0);
$error = null;
$reportSent = null;

if ($_SERVER['REQUEST_METHOD'] === 'POST' && $user) {
    $action = $_POST['action'] ?? '';
    $token = session_token();

    if ($action === 'reply') {
        $res = api_put('/api/v1/forum/messages', ['forum_thread_id' => $threadId, 'message' => $_POST['message'] ?? ''], $token);
        if ($res['status'] !== 201) {
            $error = $res['data']['error_description'] ?? $lang[291];
        }
    } elseif ($action === 'edit') {
        api_patch('/api/v1/forum?id=' . $threadId, ['titre' => $_POST['titre'] ?? '', 'message' => $_POST['message'] ?? ''], $token);
    } elseif ($action === 'delete') {
        api_delete('/api/v1/forum', ['id' => $threadId], $token);
        header('Location: forums.php');
        exit;
    } elseif ($action === 'delete_message') {
        api_delete('/api/v1/forum/messages', ['id' => intval($_POST['message_id'] ?? 0)], $token);
    } elseif ($action === 'report_forum') {
        $res = api_put('/api/v1/signalements', [
            'type_signalement' => 'forum',
            'forum_thread_id' => $threadId,
            'motif' => $_POST['motif'] ?? '',
        ], $token);
        $reportSent = $res['status'] === 201;
    } elseif ($action === 'report_message') {
        $res = api_put('/api/v1/signalements', [
            'type_signalement' => 'forum_message',
            'forum_thread_id' => $threadId,
            'forum_message_id' => intval($_POST['message_id'] ?? 0),
            'motif' => $_POST['motif'] ?? '',
        ], $token);
        $reportSent = $res['status'] === 201;
    }

    header('Location: forum.php?id=' . $threadId);
    exit;
}

$threadRes = api_get('/api/v1/forum', ['id' => $threadId], session_token());
$thread = $threadRes['data'] ?? null;

$pageTitle = $thread ? $thread['titre'] : $lang[383];
$active = 'forums';
include __DIR__ . '/../inc/layout_top.php';

if (!$thread) {
    echo '<div class="empty">Ce forum n\'existe pas ou a ete supprime.</div><p><a href="forums.php">&larr; Retour aux forums</a></p>';
    include __DIR__ . '/../inc/layout_bottom.php';
    exit;
}

$messagesRes = api_get('/api/v1/forum/messages', ['forum_thread_id' => $threadId], session_token());
$messages = $messagesRes['data']['messages'] ?? [];

$canManage = $user && ($user['compte_id'] == $thread['auteur_id'] || is_staff($user));
$canReportForum = $user && $user['compte_id'] != $thread['auteur_id'];
?>

<p><a href="forums.php">&larr; <?= h($lang[100]) ?></a></p>

<section class="section">
  <div class="section-head">
    <div>
      <h1><?= h($thread['titre']) ?></h1>
      <div class="card-meta"><span><?= h($lang[297]) ?> <?= h($thread['auteur']) ?></span><span><?= h(time_ago($thread['date_creation'])) ?></span><span><?= (int) $thread['vues'] ?> <?= h($lang[298]) ?></span></div>
    </div>
    <div class="actions" style="margin-bottom:0;">
      <?php if ($canManage): ?><a class="btn-secondary btn-sm" href="#modifier-forum"><?= h($lang[279]) ?></a><?php endif; ?>
      <?php if ($canReportForum): ?><a class="report-link" href="#signaler-forum"><?= h($lang[188]) ?></a><?php endif; ?>
    </div>
  </div>
  <div class="card"><?= h($thread['message']) ?></div>
</section>

<section class="section">
  <h2><?= count($messages) ?> <?= h($lang[299]) ?></h2>
  <div>
    <?php foreach ($messages as $m): ?>
    <div class="thread-message">
      <span class="avatar"><?= h(mb_strtoupper(mb_substr($m['auteur'], 0, 1))) ?></span>
      <div class="bubble">
        <div class="msg-meta">
          <strong><?= h($m['auteur']) ?></strong>
          <span><?= h(time_ago($m['date_envoi'])) ?></span>
          <span style="margin-left:auto;display:flex;gap:6px;">
            <?php if ($user && $user['compte_id'] != $m['auteur_id']): ?>
            <a class="report-link" href="#signaler-message-<?= (int) $m['forum_message_id'] ?>"><?= h($lang[187]) ?></a>
            <?php endif; ?>
            <?php if ($user && ($user['compte_id'] == $m['auteur_id'] || is_staff($user))): ?>
            <form method="post" style="display:inline;" onsubmit="return confirm('<?= h($lang[281]) ?>');">
              <input type="hidden" name="action" value="delete_message">
              <input type="hidden" name="message_id" value="<?= (int) $m['forum_message_id'] ?>">
              <button type="submit" class="btn-sm btn-danger"><?= h($lang[81]) ?></button>
            </form>
            <?php endif; ?>
          </span>
        </div>
        <div><?= h($m['message']) ?></div>
      </div>
    </div>
    <?php if ($user && $user['compte_id'] != $m['auteur_id']): ?>
    <div class="card" id="signaler-message-<?= (int) $m['forum_message_id'] ?>" style="margin:8px 0 16px;">
      <form method="post">
        <input type="hidden" name="action" value="report_message">
        <input type="hidden" name="message_id" value="<?= (int) $m['forum_message_id'] ?>">
        <div class="field"><label><?= h($lang[189]) ?></label><textarea name="motif" required></textarea></div>
        <div class="actions"><button type="submit" class="btn-danger"><?= h($lang[190]) ?></button></div>
      </form>
    </div>
    <?php endif; ?>
    <?php endforeach; ?>
    <?php if (empty($messages)): ?><div class="empty"><?= h($lang[220]) ?></div><?php endif; ?>
  </div>

  <?php if ($user): ?>
  <form method="post" style="margin-top:16px;">
    <input type="hidden" name="action" value="reply">
    <?php if ($error): ?><div class="note-box"><?= h($error) ?></div><?php endif; ?>
    <div class="field"><label for="reply-text"><?= h($lang[329]) ?></label><textarea id="reply-text" name="message" required></textarea></div>
    <div class="actions"><button type="submit"><?= h($lang[77]) ?></button></div>
  </form>
  <?php else: ?>
  <div class="note-box" style="margin-top:16px;"><?= h($lang[356]) ?></div>
  <?php endif; ?>
</section>

<?php if ($canManage): ?>
<section class="section card" id="modifier-forum">
  <h2><?= h($lang[279]) ?></h2>
  <form method="post">
    <input type="hidden" name="action" value="edit">
    <div class="field"><label for="edit-titre"><?= h($lang[20]) ?></label><input id="edit-titre" name="titre" value="<?= h($thread['titre']) ?>" required></div>
    <div class="field"><label for="edit-message"><?= h($lang[278]) ?></label><textarea id="edit-message" name="message" required><?= h($thread['message']) ?></textarea></div>
    <div class="actions">
      <button type="submit"><?= h($lang[76]) ?></button>
    </div>
  </form>
  <form method="post" onsubmit="return confirm('Supprimer definitivement ce forum et toutes ses reponses ?');">
    <input type="hidden" name="action" value="delete">
    <button type="submit" class="btn-danger"><?= h($lang[175]) ?></button>
  </form>
</section>
<?php endif; ?>

<?php if ($canReportForum): ?>
<section class="section card" id="signaler-forum">
  <h2><?= h($lang[188]) ?></h2>
  <?php if ($reportSent !== null): ?><div class="note-box"><?= $reportSent ? $lang[290] : 'Erreur lors de l\'envoi.' ?></div><?php endif; ?>
  <form method="post">
    <input type="hidden" name="action" value="report_forum">
    <div class="field"><label for="motif-forum"><?= h($lang[156]) ?></label><textarea id="motif-forum" name="motif" required></textarea></div>
    <div class="actions"><button type="submit" class="btn-danger"><?= h($lang[190]) ?></button></div>
  </form>
</section>
<?php endif; ?>

<?php include __DIR__ . '/../inc/layout_bottom.php'; ?>
