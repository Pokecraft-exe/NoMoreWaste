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
    if ($action === 'create_service') {
        $res = api_put('/api/v1/services', [
            'libelle' => $_POST['libelle'] ?? '', 'categorie_service' => $_POST['categorie_service'] ?? '',
            'description' => $_POST['description'] ?? '', 'tarif' => $_POST['tarif'] ?? '0',
        ], $token);
        if ($res['status'] !== 201) {
            $error = $res['data']['error_description'] ?? $lang[292];
        }
    } elseif ($action === 'update_service') {
        api_patch('/api/v1/services?id=' . intval($_POST['service_id'] ?? 0), [
            'description' => $_POST['description'] ?? '', 'tarif' => $_POST['tarif'] ?? '',
            'actif' => isset($_POST['actif']) ? '1' : '0',
        ], $token);
    } elseif ($action === 'add_date') {
        $res = api_put('/api/v1/services/planning', [
            'service_id' => $_POST['service_id'] ?? '', 'date_service' => $_POST['date_service'] ?? '',
            'heure_debut' => $_POST['heure_debut'] ?? '', 'heure_fin' => $_POST['heure_fin'] ?? '',
            'capacite' => $_POST['capacite'] ?? '0',
        ], $token);
        if ($res['status'] !== 201) {
            $error = $res['data']['error_description'] ?? $lang[292];
        }
    } elseif ($action === 'cancel_date') {
        api_delete('/api/v1/services/planning', ['id' => intval($_POST['planning_service_id'] ?? 0)], $token);
    }
    if (!$error) {
        header('Location: services.php');
        exit;
    }
}

$pageTitle = $lang[423];
$active = 'services';
$navMode = 'back';
include __DIR__ . '/../inc/layout_top.php';

$categoriesService = ['conseil', 'cuisine', 'vehicule', 'echange', 'reparation', 'gardiennage', 'autre'];
$res = api_get('/api/v1/services', [], $token);
$services = $res['data']['services'] ?? [];
$planningRes = api_get('/api/v1/services/planning', [], $token);
$planning = $planningRes['data']['planning'] ?? [];
$preselect = intval($_GET['service_id'] ?? 0);
?>
<section class="section">
  <div class="section-head"><h1><?= h($lang[423]) ?></h1><a class="btn-sm" href="#creer-service"><?= h($lang[426]) ?></a></div>
  <div class="table-wrap"><table>
    <thead><tr><th><?= h($lang[64]) ?></th><th><?= h($lang[424]) ?></th><th><?= h($lang[425]) ?></th><th><?= h($lang[25]) ?></th><th></th></tr></thead>
    <tbody>
    <?php foreach ($services as $s): ?>
    <tr>
      <td><?= h($s['libelle']) ?></td>
      <td><span class="badge"><?= h($s['categorie_service']) ?></span></td>
      <td><?= h($s['tarif']) ?> €</td>
      <td><span class="badge <?= $s['actif'] ? 'badge-success' : 'badge-danger' ?>"><?= $s['actif'] ? $lang[368] : $lang[369] ?></span></td>
      <td><a class="btn-sm" href="#service-<?= (int) $s['service_id'] ?>"><?= h($lang[86]) ?></a></td>
    </tr>
    <?php endforeach; ?>
    <?php if (empty($services)): ?><tr><td colspan="5" class="muted"><?= h($lang[432]) ?></td></tr><?php endif; ?>
    </tbody>
  </table></div>
</section>

<?php foreach ($services as $s): if ($preselect && $preselect !== (int) $s['service_id']) continue;
  $slots = array_filter($planning, fn($p) => (int) $p['service_id'] === (int) $s['service_id']);
  ?>
<section class="section card" id="service-<?= (int) $s['service_id'] ?>">
  <h3><?= h($s['libelle']) ?></h3>
  <?php if ($error): ?><div class="note-box"><?= h($error) ?></div><?php endif; ?>
  <p class="muted"><?= h($s['description']) ?></p>
  <form method="post">
    <input type="hidden" name="action" value="update_service">
    <input type="hidden" name="service_id" value="<?= (int) $s['service_id'] ?>">
    <div class="field"><label><?= h($lang[22]) ?></label><textarea name="description"><?= h($s['description']) ?></textarea></div>
    <div class="field"><label><?= h($lang[425]) ?></label><input name="tarif" type="number" min="0" step="0.01" value="<?= h($s['tarif']) ?>"></div>
    <div class="field"><label><input type="checkbox" name="actif" <?= $s['actif'] ? 'checked' : '' ?>> <?= h($lang[368]) ?></label></div>
    <div class="actions"><button type="submit"><?= h($lang[76]) ?></button></div>
  </form>

  <h4><?= h($lang[427]) ?></h4>
  <div class="table-wrap"><table><thead><tr><th><?= h($lang[14]) ?></th><th><?= h($lang[103]) ?></th><th><?= h($lang[428]) ?></th><th><?= h($lang[25]) ?></th><th></th></tr></thead><tbody>
  <?php foreach ($slots as $p): ?>
  <tr>
    <td><?= h(format_date_fr($p['date_service'])) ?></td>
    <td><?= h(substr($p['heure_debut'], 0, 5)) ?> - <?= h(substr($p['heure_fin'], 0, 5)) ?></td>
    <td><?= (int) $p['capacite'] ?></td>
    <td><span class="badge"><?= h($p['statut']) ?></span></td>
    <td>
      <form method="post" style="display:inline;" onsubmit="return confirm('<?= h($lang[448]) ?>');">
        <input type="hidden" name="action" value="cancel_date">
        <input type="hidden" name="planning_service_id" value="<?= (int) $p['planning_service_id'] ?>">
        <button type="submit" class="btn-sm btn-danger"><?= h($lang[447]) ?></button>
      </form>
    </td>
  </tr>
  <?php endforeach; ?>
  <?php if (empty($slots)): ?><tr><td colspan="5" class="muted"><?= h($lang[433]) ?></td></tr><?php endif; ?>
  </tbody></table></div>
  <form method="post">
    <input type="hidden" name="action" value="add_date">
    <input type="hidden" name="service_id" value="<?= (int) $s['service_id'] ?>">
    <div class="grid grid-2">
      <div class="field"><label><?= h($lang[14]) ?></label><input name="date_service" type="date" required></div>
      <div class="field"><label><?= h($lang[428]) ?></label><input name="capacite" type="number" min="0" value="5" required></div>
    </div>
    <div class="grid grid-2">
      <div class="field"><label><?= h($lang[103]) ?> (début)</label><input name="heure_debut" type="time" required></div>
      <div class="field"><label><?= h($lang[103]) ?> (fin)</label><input name="heure_fin" type="time" required></div>
    </div>
    <div class="actions"><button type="submit"><?= h($lang[83]) ?></button></div>
  </form>
</section>
<?php endforeach; ?>

<section class="section card" id="creer-service">
  <h2><?= h($lang[426]) ?></h2>
  <form method="post">
    <input type="hidden" name="action" value="create_service">
    <div class="field"><label><?= h($lang[64]) ?></label><input name="libelle" required></div>
    <div class="field"><label><?= h($lang[424]) ?></label>
      <select name="categorie_service"><?php foreach ($categoriesService as $val): ?><option value="<?= h($val) ?>"><?= h($val) ?></option><?php endforeach; ?></select>
    </div>
    <div class="field"><label><?= h($lang[22]) ?></label><textarea name="description" required></textarea></div>
    <div class="field"><label><?= h($lang[425]) ?></label><input name="tarif" type="number" min="0" step="0.01" value="0"></div>
    <div class="actions"><button type="submit"><?= h($lang[83]) ?></button></div>
  </form>
</section>

<?php include __DIR__ . '/../inc/layout_bottom.php'; ?>
