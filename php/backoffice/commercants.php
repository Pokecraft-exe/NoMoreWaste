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
    $commercantId = intval($_POST['commercant_id'] ?? 0);
    if ($action === 'update') {
        api_patch('/api/v1/commercants?id=' . $commercantId, [
            'email' => $_POST['email'] ?? '', 'telephone' => $_POST['telephone'] ?? '',
            'adresse' => $_POST['adresse'] ?? '', 'code_postal' => $_POST['code_postal'] ?? '',
            'ville' => $_POST['ville'] ?? '', 'actif' => isset($_POST['actif']) ? '1' : '0',
        ], $token);
    } elseif ($action === 'renouveler') {
        $res = api_put('/api/v1/commercants/adhesion', [
            'commercant_id' => $commercantId, 'forfait' => 'Adhesion commercant',
        ], $token);
        if ($res['status'] !== 201) {
            $error = $res['data']['error_description'] ?? $lang[292];
        }
    } elseif ($action === 'rappel') {
        $res = api_put('/api/v1/commercants/rappel', [
            'commercant_id' => $commercantId, 'message' => $_POST['message'] ?? '',
        ], $token);
        if ($res['status'] !== 201) {
            $error = $res['data']['error_description'] ?? $lang[292];
        }
    }
    if (!$error) {
        header('Location: commercants.php');
        exit;
    }
}

$q = query_param('q');
$pageTitle = $lang[396];
$active = 'commercants';
$navMode = 'back';
include __DIR__ . '/../inc/layout_top.php';

$res = api_get('/api/v1/commercants', api_params(['from' => 0, 'size' => 100, 'q' => $q]), $token);
$commercants = $res['data']['commercants'] ?? [];
$preselect = intval($_GET['id'] ?? 0);
?>
<section class="section">
  <h1><?= h($lang[396]) ?></h1>
  <form class="filters" method="get">
    <div class="field"><label><?= h($lang[397]) ?></label><input name="q" value="<?= h($q) ?>"></div>
    <div class="actions"><button type="submit"><?= h($lang[79]) ?></button></div>
  </form>
  <div class="table-wrap"><table>
    <thead><tr><th><?= h($lang[397]) ?></th><th><?= h($lang[2]) ?></th><th><?= h($lang[69]) ?></th><th><?= h($lang[25]) ?></th><th></th></tr></thead>
    <tbody>
    <?php foreach ($commercants as $c): ?>
    <tr>
      <td><?= h($c['raison_sociale']) ?></td>
      <td><?= h($c['email']) ?></td>
      <td><?= h($c['ville']) ?></td>
      <td><span class="badge <?= $c['actif'] ? 'badge-success' : 'badge-danger' ?>"><?= $c['actif'] ? $lang[368] : $lang[369] ?></span></td>
      <td><a class="btn-sm" href="#detail-<?= (int) $c['commercant_id'] ?>"><?= h($lang[86]) ?></a></td>
    </tr>
    <?php endforeach; ?>
    <?php if (empty($commercants)): ?><tr><td colspan="5" class="muted"><?= h($lang[407]) ?></td></tr><?php endif; ?>
    </tbody>
  </table></div>
</section>

<?php foreach ($commercants as $c): if ($preselect && $preselect !== (int) $c['commercant_id']) continue;
  $adhesionsRes = api_get('/api/v1/commercants/adhesion', ['commercant_id' => $c['commercant_id']], $token);
  $adhesions = $adhesionsRes['data']['adhesions'] ?? [];
  ?>
<section class="section card" id="detail-<?= (int) $c['commercant_id'] ?>">
  <h3><?= h($c['raison_sociale']) ?></h3>
  <?php if ($error): ?><div class="note-box"><?= h($error) ?></div><?php endif; ?>
  <form method="post">
    <input type="hidden" name="action" value="update">
    <input type="hidden" name="commercant_id" value="<?= (int) $c['commercant_id'] ?>">
    <div class="field"><label><?= h($lang[398]) ?></label><input value="<?= h($c['identifiant_legal'] ?? '-') ?>" disabled></div>
    <div class="field"><label><?= h($lang[2]) ?></label><input name="email" type="email" value="<?= h($c['email']) ?>"></div>
    <div class="field"><label><?= h($lang[66]) ?></label><input name="telephone" value="<?= h($c['telephone']) ?>"></div>
    <div class="field"><label><?= h($lang[67]) ?></label><input name="adresse" value="<?= h($c['adresse']) ?>"></div>
    <div class="grid grid-2">
      <div class="field"><label><?= h($lang[68]) ?></label><input name="code_postal" value="<?= h($c['code_postal']) ?>"></div>
      <div class="field"><label><?= h($lang[69]) ?></label><input name="ville" value="<?= h($c['ville']) ?>"></div>
    </div>
    <div class="field"><label><input type="checkbox" name="actif" <?= $c['actif'] ? 'checked' : '' ?>> <?= h($lang[368]) ?></label></div>
    <div class="actions"><button type="submit"><?= h($lang[76]) ?></button></div>
  </form>

  <h4><?= h($lang[402]) ?></h4>
  <div class="table-wrap"><table><thead><tr><th><?= h($lang[403]) ?></th><th><?= h($lang[395]) ?></th><th><?= h($lang[25]) ?></th></tr></thead><tbody>
  <?php foreach ($adhesions as $a): ?>
  <tr><td><?= h($a['date_debut']) ?></td><td><?= h($a['date_fin'] ?? '-') ?></td><td><span class="badge"><?= h($a['statut']) ?></span></td></tr>
  <?php endforeach; ?>
  <?php if (empty($adhesions)): ?><tr><td colspan="3" class="muted"><?= h($lang[408]) ?></td></tr><?php endif; ?>
  </tbody></table></div>
  <form method="post" style="margin-top:8px;">
    <input type="hidden" name="action" value="renouveler">
    <input type="hidden" name="commercant_id" value="<?= (int) $c['commercant_id'] ?>">
    <div class="actions"><button type="submit"><?= h($lang[257]) ?></button></div>
  </form>

  <h4><?= h($lang[404]) ?></h4>
  <form method="post">
    <input type="hidden" name="action" value="rappel">
    <input type="hidden" name="commercant_id" value="<?= (int) $c['commercant_id'] ?>">
    <div class="field"><label><?= h($lang[405]) ?></label><textarea name="message" required></textarea></div>
    <div class="actions"><button type="submit"><?= h($lang[78]) ?></button></div>
  </form>
</section>
<?php endforeach; ?>

<?php include __DIR__ . '/../inc/layout_bottom.php'; ?>
