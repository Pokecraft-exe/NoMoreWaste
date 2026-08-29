<?php
require_once __DIR__ . '/../inc/config.php';
require_once __DIR__ . '/../inc/auth.php';
require_once __DIR__ . '/../inc/helpers.php';
require_once __DIR__ . '/../inc/changeLanguage.php';
$lang = $LANGUAGE_CONTENTS;

$user = current_user();
$token = session_token();
$error = null;

if ($_SERVER['REQUEST_METHOD'] === 'POST' && $user) {
    $action = $_POST['action'] ?? '';
    if ($action === 'inscrire') {
        $res = api_put('/api/v1/services/inscriptions', ['planning_service_id' => $_POST['planning_service_id'] ?? ''], $token);
        if ($res['status'] !== 201) {
            $error = $res['data']['error_description'] ?? $lang[293];
        }
    } elseif ($action === 'desinscrire') {
        api_delete('/api/v1/services/inscriptions', ['planning_service_id' => $_POST['planning_service_id'] ?? ''], $token);
    }
    if (!$error) {
        header('Location: services.php');
        exit;
    }
}

$pageTitle = $lang[423];
$active = 'services';
include __DIR__ . '/../inc/layout_top.php';

$res = api_get('/api/v1/services', [], $token);
$services = $res['data']['services'] ?? [];
$planningRes = api_get('/api/v1/services/planning', ['upcoming' => '1'], $token);
$planning = $planningRes['data']['planning'] ?? [];

$mesInscriptionsList = [];
$mesInscriptions = [];
if ($user && is_adherent($user)) {
    $mesInscriptionsList = api_get('/api/v1/services/inscriptions', [], $token)['data']['inscriptions'] ?? [];
    foreach ($mesInscriptionsList as $i) {
        $mesInscriptions[(int) $i['planning_service_id']] = true;
    }
}
?>

<section class="section">
  <h1><?= h($lang[423]) ?></h1>
  <?php if ($error): ?><div class="note-box"><?= h($error) ?></div><?php endif; ?>
  <div class="grid grid-auto">
    <?php foreach ($services as $s): ?>
    <div class="card">
      <div class="card-title"><?= h($s['libelle']) ?> <span class="badge"><?= h($s['categorie_service']) ?></span></div>
      <p class="muted"><?= h($s['description']) ?></p>
      <p><?= $s['tarif'] > 0 ? h($s['tarif']) . ' €' : $lang[112] ?></p>

      <?php $slots = array_filter($planning, fn($p) => (int) $p['service_id'] === (int) $s['service_id']); ?>
      <?php if (empty($slots)): ?>
      <p class="muted"><?= h($lang[433]) ?></p>
      <?php else: ?>
      <div class="table-wrap"><table><thead><tr><th><?= h($lang[14]) ?></th><th><?= h($lang[103]) ?></th><th><?= h($lang[434]) ?></th><th></th></tr></thead><tbody>
      <?php foreach ($slots as $p): ?>
      <tr>
        <td><?= h(format_date_fr($p['date_service'])) ?></td>
        <td><?= h(substr($p['heure_debut'], 0, 5)) ?> - <?= h(substr($p['heure_fin'], 0, 5)) ?></td>
        <td><?= (int) $p['capacite'] ?></td>
        <td>
          <?php if ($user && is_adherent($user)): ?>
            <?php if (!empty($mesInscriptions[(int) $p['planning_service_id']])): ?>
            <form method="post"><input type="hidden" name="action" value="desinscrire"><input type="hidden" name="planning_service_id" value="<?= (int) $p['planning_service_id'] ?>"><button type="submit" class="btn-sm btn-danger"><?= h($lang[430]) ?></button></form>
            <?php else: ?>
            <form method="post"><input type="hidden" name="action" value="inscrire"><input type="hidden" name="planning_service_id" value="<?= (int) $p['planning_service_id'] ?>"><button type="submit" class="btn-sm"><?= h($lang[429]) ?></button></form>
            <?php endif; ?>
          <?php endif; ?>
        </td>
      </tr>
      <?php endforeach; ?>
      </tbody></table></div>
      <?php endif; ?>
    </div>
    <?php endforeach; ?>
    <?php if (empty($services)): ?><div class="empty"><?= h($lang[432]) ?></div><?php endif; ?>
  </div>
</section>

<?php if ($user && is_adherent($user)): ?>
<section class="section">
  <h2><?= h($lang[431]) ?></h2>
  <div class="table-wrap"><table><thead><tr><th><?= h($lang[64]) ?></th><th><?= h($lang[14]) ?></th><th><?= h($lang[103]) ?></th><th><?= h($lang[25]) ?></th></tr></thead><tbody>
  <?php foreach ($mesInscriptionsList as $i): ?>
  <tr>
    <td><?= h($i['service']) ?></td>
    <td><?= h(format_date_fr($i['date_service'])) ?></td>
    <td><?= h(substr($i['heure_debut'], 0, 5)) ?> - <?= h(substr($i['heure_fin'], 0, 5)) ?></td>
    <td><span class="badge"><?= h($i['statut']) ?></span></td>
  </tr>
  <?php endforeach; ?>
  <?php if (empty($mesInscriptionsList)): ?><tr><td colspan="4" class="muted"><?= h($lang[219]) ?></td></tr><?php endif; ?>
  </tbody></table></div>
</section>
<?php elseif (!$user): ?>
<section class="section"><div class="note-box"><?= h($lang[449]) ?></div></section>
<?php endif; ?>

<?php include __DIR__ . '/../inc/layout_bottom.php'; ?>
