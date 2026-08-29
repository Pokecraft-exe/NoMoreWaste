<?php
require_once __DIR__ . '/../inc/config.php';
require_once __DIR__ . '/../inc/auth.php';
require_once __DIR__ . '/../inc/helpers.php';
require_once __DIR__ . '/../inc/changeLanguage.php';
$lang = $LANGUAGE_CONTENTS;

$user = current_user();
$token = session_token();
require_staff_or_404($user);
$error = null;

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $action = $_POST['action'] ?? '';
    if ($action === 'create') {
        $email = $_POST['email'] ?? '';
        $found = api_get('/api/v1/admin/comptes', ['q' => $email, 'from' => 0, 'size' => 5], $token);
        $match = $found['data']['comptes'][0] ?? null;
        if ($match) {
            $res = api_put('/api/v1/forum', [
                'compte_id' => $match['compte_id'],
                'titre' => $_POST['titre'] ?? '', 'message' => $_POST['message'] ?? '',
            ], $token);
            if ($res['status'] !== 201) {
                $error = $res['data']['error_description'] ?? $lang[292];
            }
        } else {
            $error = $lang[214];
        }
    } elseif ($action === 'update_thread') {
        api_patch('/api/v1/forum?id=' . intval($_POST['forum_thread_id'] ?? 0), [
            'titre' => $_POST['titre'] ?? '', 'message' => $_POST['message'] ?? '',
        ], $token);
    } elseif ($action === 'delete_thread') {
        api_delete('/api/v1/forum', ['id' => intval($_POST['forum_thread_id'] ?? 0)], $token);
    } elseif ($action === 'delete_message') {
        api_delete('/api/v1/forum/messages', ['id' => intval($_POST['forum_message_id'] ?? 0)], $token);
    }
    if (!$error) {
        header('Location: forums.php' . (!empty($_POST['forum_thread_id']) && $action !== 'delete_thread' ? '?id=' . intval($_POST['forum_thread_id']) : ''));
        exit;
    }
}

$q = query_param('q');
$openId = intval($_GET['id'] ?? 0);
$pageTitle = $lang[9];
$active = 'forums';
$navMode = 'back';
include __DIR__ . '/../inc/layout_top.php';

$res = api_get('/api/v1/forum', api_params(['from' => 0, 'size' => 100, 'q' => $q]), $token);
$threads = $res['data']['threads'] ?? [];
?>
<section class="section">
  <div class="section-head"><h1><?= h($lang[172]) ?></h1><a class="btn-sm" href="#creer"><?= h($lang[182]) ?></a></div>
  <form class="filters" method="get">
    <div class="field"><label><?= h($lang[20]) ?></label><input name="q" value="<?= h($q) ?>"></div>
    <div class="actions"><button type="submit"><?= h($lang[79]) ?></button></div>
  </form>
  <div class="table-wrap"><table>
    <thead><tr><th><?= h($lang[20]) ?></th><th><?= h($lang[148]) ?></th><th><?= h($lang[342]) ?></th><th><?= h($lang[343]) ?></th><th></th></tr></thead>
    <tbody>
    <?php foreach ($threads as $t): ?>
    <tr>
      <td><?= h($t['titre']) ?></td><td><?= h($t['auteur']) ?></td><td><?= (int) $t['vues'] ?></td>
      <td><?= h(time_ago($t['date_creation'])) ?></td>
      <td><a class="btn-sm" href="?id=<?= (int) $t['forum_thread_id'] ?>#detail-<?= (int) $t['forum_thread_id'] ?>"><?= h($lang[86]) ?></a></td>
    </tr>
    <?php endforeach; ?>
    <?php if (empty($threads)): ?><tr><td colspan="5" class="muted"><?= h($lang[208]) ?></td></tr><?php endif; ?>
    </tbody>
  </table></div>
</section>

<?php if ($openId):
  $detailRes = api_get('/api/v1/forum', ['id' => $openId], $token);
  $thread = $detailRes['status'] === 200 ? $detailRes['data'] : null;
  $messagesRes = $thread ? api_get('/api/v1/forum/messages', ['forum_thread_id' => $openId], $token) : null;
  $messages = $messagesRes['data']['messages'] ?? [];
  if ($thread): ?>
<section class="section card" id="detail-<?= $openId ?>">
  <h3><?= h($thread['titre']) ?></h3>
  <p class="muted"><?= h($lang[296]) ?> <?= h($thread['auteur']) ?> - <?= h(time_ago($thread['date_creation'])) ?></p>
  <form method="post">
    <input type="hidden" name="action" value="update_thread">
    <input type="hidden" name="forum_thread_id" value="<?= $openId ?>">
    <div class="field"><label><?= h($lang[20]) ?></label><input name="titre" value="<?= h($thread['titre']) ?>" required></div>
    <div class="field"><label><?= h($lang[35]) ?></label><textarea name="message" required><?= h($thread['message']) ?></textarea></div>
    <div class="actions"><button type="submit"><?= h($lang[76]) ?></button></div>
  </form>
  <form method="post" onsubmit="return confirm('<?= h($lang[282]) ?>');">
    <input type="hidden" name="action" value="delete_thread">
    <input type="hidden" name="forum_thread_id" value="<?= $openId ?>">
    <button type="submit" class="btn-danger"><?= h($lang[174]) ?></button>
  </form>

  <h4><?= h($lang[147]) ?></h4>
  <?php foreach ($messages as $m): ?>
  <div class="thread-message">
    <span class="avatar"><?= h(mb_strtoupper(mb_substr($m['auteur'], 0, 1))) ?></span>
    <div class="bubble">
      <div class="msg-meta">
        <strong><?= h($m['auteur']) ?></strong>
        <span><?= h(time_ago($m['date_envoi'])) ?></span>
        <span style="margin-left:auto;">
          <form method="post" style="display:inline;" onsubmit="return confirm('<?= h($lang[281]) ?>');">
            <input type="hidden" name="action" value="delete_message">
            <input type="hidden" name="forum_message_id" value="<?= (int) $m['forum_message_id'] ?>">
            <input type="hidden" name="forum_thread_id" value="<?= $openId ?>">
            <button type="submit" class="btn-sm btn-danger"><?= h($lang[81]) ?></button>
          </form>
        </span>
      </div>
      <div><?= h($m['message']) ?></div>
    </div>
  </div>
  <?php endforeach; ?>
  <?php if (empty($messages)): ?><p class="muted"><?= h($lang[221]) ?></p><?php endif; ?>
</section>
<?php endif; endif; ?>

<section class="section card" id="creer">
  <h2><?= h($lang[182]) ?></h2>
  <?php if ($error): ?><div class="note-box"><?= h($error) ?></div><?php endif; ?>
  <form method="post">
    <input type="hidden" name="action" value="create">
    <div class="field"><label><?= h($lang[393]) ?></label><input name="email" type="email" required></div>
    <div class="field"><label><?= h($lang[20]) ?></label><input name="titre" required></div>
    <div class="field"><label><?= h($lang[35]) ?></label><textarea name="message" required></textarea></div>
    <div class="actions"><button type="submit"><?= h($lang[78]) ?></button></div>
  </form>
</section>

<?php include __DIR__ . '/../inc/layout_bottom.php'; ?>
