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
    $benevoleId = intval($_POST['benevole_id'] ?? 0);
    if ($action === 'update') {
        api_patch('/api/v1/benevoles?id=' . $benevoleId, [
            'statut' => $_POST['statut'] ?? '', 'disponibilite' => $_POST['disponibilite'] ?? '',
        ], $token);
    } elseif ($action === 'competence') {
        $res = api_put('/api/v1/benevoles/competences', [
            'benevole_id' => $benevoleId, 'competence_id' => $_POST['competence_id'] ?? '',
            'niveau' => $_POST['niveau'] ?? '1',
        ], $token);
        if ($res['status'] !== 201) {
            $error = $res['data']['error_description'] ?? $lang[292];
        }
    } elseif ($action === 'affecter') {
        $res = api_put('/api/v1/benevoles/affectations', [
            'benevole_id' => $benevoleId, 'planning_service_id' => $_POST['planning_service_id'] ?? '',
            'role_mission' => $_POST['role_mission'] ?? '',
        ], $token);
        if ($res['status'] !== 201) {
            $error = $res['data']['error_description'] ?? $lang[292];
        }
    }
    if (!$error) {
        header('Location: benevoles.php');
        exit;
    }
}

$q = query_param('q');
$pageTitle = $lang[437];
$active = 'benevoles';
$navMode = 'back';
include __DIR__ . '/../inc/layout_top.php';

$res = api_get('/api/v1/benevoles', api_params(['from' => 0, 'size' => 100, 'q' => $q]), $token);
$benevoles = $res['data']['benevoles'] ?? [];
$catalogueRes = api_get('/api/v1/benevoles/competences', [], $token);
$catalogue = $catalogueRes['data']['competences'] ?? [];
$planningRes = api_get('/api/v1/services/planning', ['upcoming' => '1'], $token);
$planning = $planningRes['data']['planning'] ?? [];
$preselect = intval($_GET['id'] ?? 0);
?>
<section class="section">
  <h1><?= h($lang[437]) ?></h1>
  <form class="filters" method="get">
    <div class="field"><label><?= h($lang[325]) ?></label><input name="q" value="<?= h($q) ?>"></div>
    <div class="actions"><button type="submit"><?= h($lang[79]) ?></button></div>
  </form>
  <div class="table-wrap"><table>
    <thead><tr><th><?= h($lang[64]) ?></th><th><?= h($lang[2]) ?></th><th><?= h($lang[142]) ?></th><th><?= h($lang[25]) ?></th><th></th></tr></thead>
    <tbody>
    <?php foreach ($benevoles as $b): ?>
    <tr>
      <td><?= h($b['prenom'] . ' ' . $b['nom']) ?></td>
      <td><?= h($b['email']) ?></td>
      <td><?= h($b['disponibilite'] ?? '-') ?></td>
      <td><span class="badge"><?= h($b['statut']) ?></span></td>
      <td><a class="btn-sm" href="#detail-<?= (int) $b['benevole_id'] ?>"><?= h($lang[86]) ?></a></td>
    </tr>
    <?php endforeach; ?>
    <?php if (empty($benevoles)): ?><tr><td colspan="5" class="muted"><?= h($lang[214]) ?></td></tr><?php endif; ?>
    </tbody>
  </table></div>
</section>

<?php foreach ($benevoles as $b): if ($preselect && $preselect !== (int) $b['benevole_id']) continue;
  $competencesRes = api_get('/api/v1/benevoles/competences', ['benevole_id' => $b['benevole_id']], $token);
  $competences = $competencesRes['data']['competences'] ?? [];
  $affectationsRes = api_get('/api/v1/benevoles/affectations', ['benevole_id' => $b['benevole_id']], $token);
  $affectations = $affectationsRes['data']['affectations'] ?? [];
  ?>
<section class="section card" id="detail-<?= (int) $b['benevole_id'] ?>">
  <h3><?= h($b['prenom'] . ' ' . $b['nom']) ?></h3>
  <?php if ($error): ?><div class="note-box"><?= h($error) ?></div><?php endif; ?>
  <p class="muted"><?= h($b['email']) ?></p>
  <form method="post">
    <input type="hidden" name="action" value="update">
    <input type="hidden" name="benevole_id" value="<?= (int) $b['benevole_id'] ?>">
    <div class="field"><label><?= h($lang[142]) ?></label><input name="disponibilite" value="<?= h($b['disponibilite'] ?? '') ?>"></div>
    <div class="field"><label><?= h($lang[25]) ?></label>
      <select name="statut"><?php foreach (['actif', 'inactif', 'suspendu'] as $s): ?><option value="<?= h($s) ?>" <?= $b['statut'] === $s ? 'selected' : '' ?>><?= h($s) ?></option><?php endforeach; ?></select>
    </div>
    <div class="actions"><button type="submit"><?= h($lang[76]) ?></button></div>
  </form>

  <h4><?= h($lang[438]) ?></h4>
  <div class="table-wrap"><table><thead><tr><th><?= h($lang[64]) ?></th><th><?= h($lang[440]) ?></th></tr></thead><tbody>
  <?php foreach ($competences as $c): ?>
  <tr><td><?= h($c['libelle']) ?></td><td><?= (int) $c['niveau'] ?> / 5</td></tr>
  <?php endforeach; ?>
  <?php if (empty($competences)): ?><tr><td colspan="2" class="muted"><?= h($lang[214]) ?></td></tr><?php endif; ?>
  </tbody></table></div>
  <form method="post">
    <input type="hidden" name="action" value="competence">
    <input type="hidden" name="benevole_id" value="<?= (int) $b['benevole_id'] ?>">
    <div class="grid grid-2">
      <div class="field"><label><?= h($lang[438]) ?></label>
        <select name="competence_id"><?php foreach ($catalogue as $c): ?><option value="<?= (int) $c['competence_id'] ?>"><?= h($c['libelle']) ?></option><?php endforeach; ?></select>
      </div>
      <div class="field"><label><?= h($lang[440]) ?> (1-5)</label><input name="niveau" type="number" min="1" max="5" value="3"></div>
    </div>
    <div class="actions"><button type="submit"><?= h($lang[439]) ?></button></div>
  </form>

  <h4><?= h($lang[441]) ?></h4>
  <div class="table-wrap"><table><thead><tr><th><?= h($lang[423]) ?></th><th><?= h($lang[105]) ?></th><th><?= h($lang[25]) ?></th></tr></thead><tbody>
  <?php foreach ($affectations as $a): ?>
  <tr><td><?= h($a['planning_service_id'] ? '#' . $a['planning_service_id'] : '-') ?></td><td><?= h($a['role_mission']) ?></td><td><span class="badge"><?= h($a['statut']) ?></span></td></tr>
  <?php endforeach; ?>
  <?php if (empty($affectations)): ?><tr><td colspan="3" class="muted"><?= h($lang[214]) ?></td></tr><?php endif; ?>
  </tbody></table></div>
  <form method="post">
    <input type="hidden" name="action" value="affecter">
    <input type="hidden" name="benevole_id" value="<?= (int) $b['benevole_id'] ?>">
    <div class="field"><label><?= h($lang[427]) ?></label>
      <select name="planning_service_id"><?php foreach ($planning as $p): ?><option value="<?= (int) $p['planning_service_id'] ?>"><?= h($p['service']) ?> - <?= h(format_date_fr($p['date_service'])) ?></option><?php endforeach; ?></select>
    </div>
    <div class="field"><label><?= h($lang[105]) ?></label><input name="role_mission" required></div>
    <div class="actions"><button type="submit"><?= h($lang[88]) ?></button></div>
  </form>
</section>
<?php endforeach; ?>

<?php include __DIR__ . '/../inc/layout_bottom.php'; ?>
