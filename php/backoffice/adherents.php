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
    $adherentId = intval($_POST['adherent_id'] ?? 0);
    if ($action === 'update') {
        api_patch('/api/v1/adherents?id=' . $adherentId, ['statut' => $_POST['statut'] ?? ''], $token);
    } elseif ($action === 'rappel') {
        $res = api_put('/api/v1/adherents/rappel', [
            'adhesion_association_id' => intval($_POST['adhesion_association_id'] ?? 0),
            'canal' => $_POST['canal'] ?? 'email',
            'message' => $_POST['message'] ?? '',
        ], $token);
        if ($res['status'] !== 201) {
            $error = $res['data']['error_description'] ?? $lang[292];
        }
    }
    if (!$error) {
        header('Location: adherents.php');
        exit;
    }
}

$q = query_param('q');
$pageTitle = $lang[442];
$active = 'adherents';
$navMode = 'back';
include __DIR__ . '/../inc/layout_top.php';

$res = api_get('/api/v1/adherents', api_params(['from' => 0, 'size' => 100, 'q' => $q]), $token);
$adherents = $res['data']['adherents'] ?? [];
$statuts = ['actif', 'suspendu', 'radie', 'en_attente'];
$preselect = intval($_GET['id'] ?? 0);
?>
<section class="section">
  <h1><?= h($lang[442]) ?></h1>
  <form class="filters" method="get">
    <div class="field"><label><?= h($lang[325]) ?></label><input name="q" value="<?= h($q) ?>"></div>
    <div class="actions"><button type="submit"><?= h($lang[79]) ?></button></div>
  </form>
  <div class="table-wrap"><table>
    <thead><tr><th><?= h($lang[64]) ?></th><th><?= h($lang[2]) ?></th><th><?= h($lang[443]) ?></th><th><?= h($lang[25]) ?></th><th></th></tr></thead>
    <tbody>
    <?php foreach ($adherents as $a): ?>
    <tr>
      <td><?= h($a['prenom'] . ' ' . $a['nom']) ?></td>
      <td><?= h($a['email']) ?></td>
      <td><?= h($a['date_inscription']) ?></td>
      <td><span class="badge"><?= h($a['statut']) ?></span></td>
      <td><a class="btn-sm" href="#detail-<?= (int) $a['adherent_id'] ?>"><?= h($lang[86]) ?></a></td>
    </tr>
    <?php endforeach; ?>
    <?php if (empty($adherents)): ?><tr><td colspan="5" class="muted"><?= h($lang[215]) ?></td></tr><?php endif; ?>
    </tbody>
  </table></div>
</section>

<?php foreach ($adherents as $a): if ($preselect && $preselect !== (int) $a['adherent_id']) continue;
  $detail = api_get('/api/v1/adherents', ['id' => $a['adherent_id']], $token)['data'] ?? $a;
  ?>
<section class="section card" id="detail-<?= (int) $a['adherent_id'] ?>">
  <h3><?= h($a['prenom'] . ' ' . $a['nom']) ?></h3>
  <?php if ($error): ?><div class="note-box"><?= h($error) ?></div><?php endif; ?>
  <p class="muted"><?= h($a['email']) ?></p>
  <?php if (!empty($detail['adhesion_date_fin'])): ?><p><?= h($lang[256]) ?> <?= h($detail['adhesion_date_fin']) ?></p><?php endif; ?>

  <form method="post">
    <input type="hidden" name="action" value="update">
    <input type="hidden" name="adherent_id" value="<?= (int) $a['adherent_id'] ?>">
    <div class="field"><label><?= h($lang[25]) ?></label>
      <select name="statut"><?php foreach ($statuts as $s): ?><option value="<?= h($s) ?>" <?= $a['statut'] === $s ? 'selected' : '' ?>><?= h($s) ?></option><?php endforeach; ?></select>
    </div>
    <div class="actions"><button type="submit"><?= h($lang[76]) ?></button></div>
  </form>

  <?php if (!empty($detail['adhesion_association_id'])): ?>
  <h4><?= h($lang[404]) ?></h4>
  <form method="post">
    <input type="hidden" name="action" value="rappel">
    <input type="hidden" name="adherent_id" value="<?= (int) $a['adherent_id'] ?>">
    <input type="hidden" name="adhesion_association_id" value="<?= (int) $detail['adhesion_association_id'] ?>">
    <div class="field"><label><?= h($lang[405]) ?></label><textarea name="message" required></textarea></div>
    <input type="hidden" name="canal" value="email">
    <div class="actions"><button type="submit"><?= h($lang[78]) ?></button></div>
  </form>
  <?php endif; ?>
</section>
<?php endforeach; ?>

<?php include __DIR__ . '/../inc/layout_bottom.php'; ?>
