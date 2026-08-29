<?php
require_once __DIR__ . '/../inc/config.php';
require_once __DIR__ . '/../inc/auth.php';
require_once __DIR__ . '/../inc/helpers.php';
require_once __DIR__ . '/../inc/changeLanguage.php';
$lang = $LANGUAGE_CONTENTS;

$user = current_user();
$token = session_token();
require_staff_or_404($user);

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $action = $_POST['action'] ?? '';
    if ($action === 'resoudre') {
        api_patch('/api/v1/signalements?id=' . intval($_POST['signalement_id'] ?? 0), [
            'commentaire' => $_POST['commentaire'] ?? '',
        ], $token);
    } elseif ($action === 'supprimer_message') {
        $type = $_POST['type_signalement'] ?? '';
        if ($type === 'forum') {
            api_delete('/api/v1/forum', ['id' => intval($_POST['forum_thread_id'] ?? 0)], $token);
        } elseif ($type === 'forum_message') {
            api_delete('/api/v1/forum/messages', ['id' => intval($_POST['forum_message_id'] ?? 0)], $token);
        }
        api_patch('/api/v1/signalements?id=' . intval($_POST['signalement_id'] ?? 0), [
            'commentaire' => $_POST['commentaire'] ?? $lang[161],
        ], $token);
    }
    header('Location: signalements.php');
    exit;
}

$statutFilter = query_param('statut', 'ouvert');
$openId = intval($_GET['id'] ?? 0);
$pageTitle = $lang[50];
$active = 'signalements';
$navMode = 'back';
include __DIR__ . '/../inc/layout_top.php';

$res = api_get('/api/v1/signalements', api_params(['from' => 0, 'size' => 100, 'statut' => $statutFilter]), $token);
$signalements = $res['data']['signalements'] ?? [];
$typeLabels = ['forum' => $lang[375], 'forum_message' => $lang[376], 'annonce_message' => $lang[377]];
?>
<section class="section">
  <h1><?= h($lang[50]) ?></h1>
  <form class="filters" method="get">
    <div class="field"><label><?= h($lang[25]) ?></label>
      <select name="statut">
        <option value="ouvert" <?= $statutFilter === 'ouvert' ? 'selected' : '' ?>><?= h($lang[328]) ?></option>
        <option value="traite" <?= $statutFilter === 'traite' ? 'selected' : '' ?>><?= h($lang[327]) ?></option>
        <option value="" <?= $statutFilter === '' ? 'selected' : '' ?>><?= h($lang[111]) ?></option>
      </select>
    </div>
    <div class="actions"><button type="submit"><?= h($lang[80]) ?></button></div>
  </form>
  <div class="table-wrap"><table>
    <thead><tr><th><?= h($lang[106]) ?></th><th><?= h($lang[155]) ?></th><th><?= h($lang[156]) ?></th><th><?= h($lang[25]) ?></th><th><?= h($lang[154]) ?></th><th></th></tr></thead>
    <tbody>
    <?php foreach ($signalements as $s): ?>
    <tr>
      <td><?= h($typeLabels[$s['type_signalement']] ?? $s['type_signalement']) ?></td>
      <td><?= h($s['signale_par_nom']) ?></td>
      <td class="muted"><?= h(mb_strimwidth($s['motif'], 0, 60, '...')) ?></td>
      <td><span class="badge"><?= h($s['statut']) ?></span></td>
      <td><?= h(time_ago($s['date_signalement'])) ?></td>
      <td><a class="btn-sm" href="?statut=<?= h($statutFilter) ?>&id=<?= (int) $s['signalement_id'] ?>#detail-<?= (int) $s['signalement_id'] ?>"><?= h($lang[86]) ?></a></td>
    </tr>
    <?php endforeach; ?>
    <?php if (empty($signalements)): ?><tr><td colspan="6" class="muted"><?= h($lang[216]) ?></td></tr><?php endif; ?>
    </tbody>
  </table></div>
</section>

<?php if ($openId):
  $detailRes = api_get('/api/v1/signalements', ['id' => $openId], $token);
  $s = $detailRes['status'] === 200 ? $detailRes['data'] : null;
  if ($s): ?>
<section class="section card" id="detail-<?= $openId ?>">
  <h3><?= h($lang[158]) ?><?= $openId ?> - <?= h($typeLabels[$s['type_signalement']] ?? $s['type_signalement']) ?></h3>
  <p><strong><?= h($lang[157]) ?></strong> <?= nl2br(h($s['motif'])) ?></p>
  <p class="muted"><?= h(time_ago($s['date_signalement'])) ?></p>

  <?php if ($s['statut'] === 'traite'): ?>
    <h4><?= h($lang[165]) ?></h4>
    <?php if (!empty($s['message_signale'])): ?>
    <div class="thread-message">
      <div class="bubble" style="border-color:var(--danger);">
        <div class="msg-meta"><span><?= h(time_ago($s['message_signale']['date'])) ?></span></div>
        <div><?= nl2br(h($s['message_signale']['message'])) ?></div>
      </div>
    </div>
    <?php else: ?>
    <p class="muted"><?= h($lang[166]) ?></p>
    <?php endif; ?>
    <h4><?= h($lang[153]) ?></h4>
    <p><?= nl2br(h($s['commentaire'])) ?></p>
  <?php else: ?>
    <h4><?= h($lang[163]) ?></h4>
    <?php foreach (($s['discussion'] ?? []) as $m): ?>
    <div class="thread-message">
      <span class="avatar"><?= h(mb_strtoupper(mb_substr($m['auteur'], 0, 1))) ?></span>
      <div class="bubble" <?= $m['signale'] ? 'style="border-color:var(--danger);"' : '' ?>>
        <div class="msg-meta">
          <strong><?= h($m['auteur']) ?></strong>
          <span><?= h(time_ago($m['date'])) ?></span>
          <?php if ($m['signale']): ?><span class="badge badge-danger"><?= h($lang[159]) ?></span><?php endif; ?>
        </div>
        <div><?= nl2br(h($m['message'])) ?></div>
      </div>
    </div>
    <?php endforeach; ?>
    <?php if (empty($s['discussion'])): ?><p class="muted"><?= h($lang[164]) ?></p><?php endif; ?>

    <h4><?= h($lang[161]) ?></h4>
    <form method="post">
      <input type="hidden" name="signalement_id" value="<?= $openId ?>">
      <input type="hidden" name="type_signalement" value="<?= h($s['type_signalement']) ?>">
      <?php if (!empty($s['forum_thread_id'])): ?><input type="hidden" name="forum_thread_id" value="<?= (int) $s['forum_thread_id'] ?>"><?php endif; ?>
      <?php if (!empty($s['forum_message_id'])): ?><input type="hidden" name="forum_message_id" value="<?= (int) $s['forum_message_id'] ?>"><?php endif; ?>
      <div class="field"><label><?= h($lang[323]) ?></label><textarea name="commentaire" required></textarea></div>
      <div class="actions">
        <button type="submit" name="action" value="resoudre"><?= h($lang[162]) ?></button>
        <?php if (in_array($s['type_signalement'], ['forum', 'forum_message'], true)): ?>
        <button type="submit" name="action" value="supprimer_message" class="btn-danger" onclick="return confirm('<?= h($lang[281]) ?>');"><?= h($lang[81]) ?></button>
        <?php endif; ?>
      </div>
    </form>
  <?php endif; ?>
</section>
<?php endif; endif; ?>

<?php include __DIR__ . '/../inc/layout_bottom.php'; ?>
