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
    if ($action === 'repondre') {
        api_patch('/api/v1/tickets?id=' . intval($_POST['ticket_id'] ?? 0), [
            'reponse' => $_POST['reponse'] ?? '',
        ], $token);
    }
    header('Location: tickets.php');
    exit;
}

$statutFilter = query_param('statut', 'ouvert');
$openId = intval($_GET['id'] ?? 0);
$pageTitle = $lang[51];
$active = 'tickets';
$navMode = 'back';
include __DIR__ . '/../inc/layout_top.php';

$res = api_get('/api/v1/tickets', api_params(['from' => 0, 'size' => 100, 'statut' => $statutFilter]), $token);
$tickets = $res['data']['tickets'] ?? [];
?>
<section class="section">
  <h1><?= h($lang[169]) ?></h1>
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
    <thead><tr><th><?= h($lang[12]) ?></th><th><?= h($lang[171]) ?></th><th><?= h($lang[154]) ?></th><th><?= h($lang[25]) ?></th><th></th></tr></thead>
    <tbody>
    <?php foreach ($tickets as $t): ?>
    <tr>
      <td><?= h($t['sujet']) ?></td>
      <td><?= h($t['auteur_id'] ? 'compte #' . $t['auteur_id'] : ($t['contact_nom'] . ' (' . $t['contact_email'] . ')')) ?></td>
      <td><?= h(time_ago($t['date_creation'])) ?></td>
      <td><span class="badge"><?= h($t['statut']) ?></span></td>
      <td><a class="btn-sm" href="?statut=<?= h($statutFilter) ?>&id=<?= (int) $t['ticket_id'] ?>#detail-<?= (int) $t['ticket_id'] ?>"><?= h($lang[86]) ?></a></td>
    </tr>
    <?php endforeach; ?>
    <?php if (empty($tickets)): ?><tr><td colspan="5" class="muted"><?= h($lang[217]) ?></td></tr><?php endif; ?>
    </tbody>
  </table></div>
</section>

<?php if ($openId):
  $t = null;
  foreach ($tickets as $row) { if ((int) $row['ticket_id'] === $openId) { $t = $row; break; } }
  if (!$t) {
      $allRes = api_get('/api/v1/tickets', ['from' => 0, 'size' => 500], $token);
      foreach (($allRes['data']['tickets'] ?? []) as $row) { if ((int) $row['ticket_id'] === $openId) { $t = $row; break; } }
  }
  if ($t): ?>
<section class="section card" id="detail-<?= $openId ?>">
  <h3><?= h($t['sujet']) ?></h3>
  <p class="muted"><?= h($lang[171]) ?> <?= h($t['auteur_id'] ? 'compte #' . $t['auteur_id'] : ($t['contact_nom'] . ' (' . $t['contact_email'] . ')')) ?> - <?= h(time_ago($t['date_creation'])) ?></p>
  <p><?= nl2br(h($t['message'])) ?></p>

  <?php if ($t['statut'] === 'traite'): ?>
    <h4><?= h($lang[167]) ?></h4>
    <p><?= nl2br(h($t['reponse'])) ?></p>
  <?php else: ?>
    <h4><?= h($lang[28]) ?></h4>
    <form method="post">
      <input type="hidden" name="action" value="repondre">
      <input type="hidden" name="ticket_id" value="<?= $openId ?>">
      <div class="field"><label><?= h($lang[146]) ?></label><textarea name="reponse" required></textarea></div>
      <div class="actions"><button type="submit"><?= h($lang[168]) ?></button></div>
    </form>
  <?php endif; ?>
</section>
<?php endif; endif; ?>

<?php include __DIR__ . '/../inc/layout_bottom.php'; ?>
