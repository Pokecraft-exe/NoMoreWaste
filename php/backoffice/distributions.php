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
        $res = api_put('/api/v1/distributions', [
            'lieu' => $_POST['lieu'] ?? '', 'date_distribution' => $_POST['date_distribution'] ?? '',
            'heure_distribution' => $_POST['heure_distribution'] ?? '', 'statut' => $_POST['statut'] ?? '',
            'quota_par_adherent' => $_POST['quota_par_adherent'] ?? '0',
        ], $token);
        if ($res['status'] !== 201) {
            $error = $res['data']['error_description'] ?? $lang[292];
        }
    } elseif ($action === 'update') {
        api_patch('/api/v1/distributions?id=' . intval($_POST['id'] ?? 0), [
            'lieu' => $_POST['lieu'] ?? '', 'date_distribution' => $_POST['date_distribution'] ?? '',
            'heure_distribution' => $_POST['heure_distribution'] ?? '', 'statut' => $_POST['statut'] ?? '',
            'quota_par_adherent' => $_POST['quota_par_adherent'] ?? '0',
        ], $token);
    } elseif ($action === 'cancel') {
        api_delete('/api/v1/distributions', ['id' => intval($_POST['id'] ?? 0)], $token);
    } elseif ($action === 'allouer') {
        api_put('/api/v1/distributions/denrees', [
            'distribution_id' => $_POST['distribution_id'] ?? '', 'stock_produit_id' => $_POST['stock_produit_id'] ?? '',
            'quantite' => $_POST['quantite'] ?? '0',
        ], $token);
    } elseif ($action === 'affecter') {
        api_put('/api/v1/distributions/benevoles', [
            'distribution_id' => $_POST['distribution_id'] ?? '', 'benevole_id' => $_POST['benevole_id'] ?? '',
            'role_mission' => $_POST['role_mission'] ?? '',
        ], $token);
    }
    if (!$error) {
        header('Location: distributions.php');
        exit;
    }
}

$q = query_param('q');
$statutFilter = query_param('statut');
$pageTitle = $lang[46];
$active = 'distributions';
$navMode = 'back';
include __DIR__ . '/../inc/layout_top.php';

$res = api_get('/api/v1/distributions', api_params(['from' => 0, 'size' => 100, 'q' => $q, 'statut' => $statutFilter]), $token);
$distributions = $res['data']['distributions'] ?? [];
$statuts = ['planifiee', 'confirmee', 'en_cours', 'terminee', 'annulee'];
$preselect = intval($_GET['id'] ?? 0);
$produitsRes = api_get('/api/v1/stock-produits', ['q' => ''], $token);
$produits = $produitsRes['data']['produits'] ?? [];
?>
<section class="section">
  <div class="section-head"><h1><?= h($lang[46]) ?></h1><a class="btn-sm" href="#creer"><?= h($lang[179]) ?></a></div>
  <form class="filters" method="get">
    <div class="field"><label><?= h($lang[104]) ?></label><input name="q" value="<?= h($q) ?>"></div>
    <div class="field"><label><?= h($lang[25]) ?></label><select name="statut"><option value=""><?= h($lang[111]) ?></option><?php foreach ($statuts as $s): ?><option value="<?= h($s) ?>" <?= $statutFilter === $s ? 'selected' : '' ?>><?= h($s) ?></option><?php endforeach; ?></select></div>
    <div class="actions"><button type="submit"><?= h($lang[79]) ?></button></div>
  </form>
  <div class="table-wrap"><table>
    <thead><tr><th><?= h($lang[104]) ?></th><th><?= h($lang[14]) ?></th><th><?= h($lang[25]) ?></th><th></th></tr></thead>
    <tbody>
    <?php foreach ($distributions as $d): ?>
    <tr>
      <td><?= h($d['lieu']) ?></td><td><?= h($d['date_distribution']) ?></td>
      <td><span class="badge"><?= h($d['statut']) ?></span></td>
      <td><a class="btn-sm" href="#detail-<?= (int) $d['distribution_id'] ?>"><?= h($lang[86]) ?></a></td>
    </tr>
    <?php endforeach; ?>
    <?php if (empty($distributions)): ?><tr><td colspan="4" class="muted"><?= h($lang[212]) ?></td></tr><?php endif; ?>
    </tbody>
  </table></div>
</section>

<?php foreach ($distributions as $d): if ($preselect && $preselect !== (int) $d['distribution_id']) continue;
  $detail = api_get('/api/v1/distributions', ['id' => $d['distribution_id']], $token)['data'] ?? $d;
  ?>
<section class="section card" id="detail-<?= (int) $d['distribution_id'] ?>">
  <h3><?= h($lang[316]) ?> <?= h($d['lieu']) ?></h3>
  <form method="post">
    <input type="hidden" name="action" value="update">
    <input type="hidden" name="id" value="<?= (int) $d['distribution_id'] ?>">
    <div class="field"><label><?= h($lang[104]) ?></label><input name="lieu" value="<?= h($d['lieu']) ?>" required></div>
    <div class="grid grid-2">
      <div class="field"><label><?= h($lang[14]) ?></label><input name="date_distribution" type="date" value="<?= h($d['date_distribution']) ?>" required></div>
      <div class="field"><label><?= h($lang[103]) ?></label><input name="heure_distribution" type="time" value="<?= h($d['heure_distribution']) ?>"></div>
    </div>
    <div class="grid grid-2">
      <div class="field"><label><?= h($lang[25]) ?></label><select name="statut"><?php foreach ($statuts as $s): ?><option value="<?= h($s) ?>" <?= $d['statut'] === $s ? 'selected' : '' ?>><?= h($s) ?></option><?php endforeach; ?></select></div>
      <div class="field"><label><?= h($lang[246]) ?></label><input name="quota_par_adherent" type="number" min="0" value="<?= h($d['quota_par_adherent']) ?>"></div>
    </div>
    <div class="actions"><button type="submit"><?= h($lang[76]) ?></button></div>
  </form>
  <form method="post" onsubmit="return confirm('<?= h($lang[289]) ?>');">
    <input type="hidden" name="action" value="cancel"><input type="hidden" name="id" value="<?= (int) $d['distribution_id'] ?>">
    <button type="submit" class="btn-danger"><?= h($lang[90]) ?></button>
  </form>

  <h4><?= h($lang[124]) ?> <?= h($lang[348]) ?></h4>
  <div class="table-wrap"><table><thead><tr><th><?= h($lang[119]) ?></th><th><?= h($lang[120]) ?></th><th><?= h($lang[122]) ?></th></tr></thead><tbody>
  <?php foreach (($detail['denrees'] ?? []) as $den): ?>
  <tr><td><?= h($den['nom']) ?></td><td><?= h($den['quantite']) ?> <?= h($den['unite']) ?></td><td><?= h($den['restant']) ?> <?= h($den['unite']) ?></td></tr>
  <?php endforeach; ?>
  <?php if (empty($detail['denrees'])): ?><tr><td colspan="3" class="muted"><?= h($lang[227]) ?></td></tr><?php endif; ?>
  </tbody></table></div>
  <form method="post">
    <input type="hidden" name="action" value="allouer">
    <input type="hidden" name="distribution_id" value="<?= (int) $d['distribution_id'] ?>">
    <div class="grid grid-2">
      <div class="field"><label><?= h($lang[119]) ?></label><select name="stock_produit_id" required><?php foreach ($produits as $p): ?><option value="<?= (int) $p['stock_produit_id'] ?>"><?= h($p['nom']) ?></option><?php endforeach; ?></select></div>
      <div class="field"><label><?= h($lang[120]) ?></label><input name="quantite" type="number" min="0" required></div>
    </div>
    <div class="actions"><button type="submit"><?= h($lang[91]) ?></button></div>
  </form>

  <h4><?= h($lang[314]) ?></h4>
  <div class="table-wrap"><table><thead><tr><th><?= h($lang[64]) ?></th><th><?= h($lang[105]) ?></th></tr></thead><tbody>
  <?php foreach (($detail['benevoles_affectes'] ?? []) as $b): ?>
  <tr><td><?= h($b['prenom'] . ' ' . $b['nom']) ?></td><td><?= h($b['role_mission']) ?></td></tr>
  <?php endforeach; ?>
  <?php if (empty($detail['benevoles_affectes'])): ?><tr><td colspan="2" class="muted"><?= h($lang[222]) ?></td></tr><?php endif; ?>
  </tbody></table></div>
  <form method="post">
    <input type="hidden" name="action" value="affecter">
    <input type="hidden" name="distribution_id" value="<?= (int) $d['distribution_id'] ?>">
    <div class="note-box"><?= h($lang[318]) ?></div>
    <div class="grid grid-2">
      <div class="field"><label><?= h($lang[317]) ?></label><input name="benevole_id" type="number" required></div>
      <div class="field"><label><?= h($lang[105]) ?></label><input name="role_mission" required></div>
    </div>
    <div class="actions"><button type="submit"><?= h($lang[87]) ?></button></div>
  </form>
</section>
<?php endforeach; ?>

<section class="section card" id="creer">
  <h2><?= h($lang[179]) ?></h2>
  <?php if ($error): ?><div class="note-box"><?= h($error) ?></div><?php endif; ?>
  <form method="post">
    <input type="hidden" name="action" value="create">
    <div class="field"><label><?= h($lang[104]) ?></label><input name="lieu" required></div>
    <div class="grid grid-2">
      <div class="field"><label><?= h($lang[14]) ?></label><input name="date_distribution" type="date" required></div>
      <div class="field"><label><?= h($lang[103]) ?></label><input name="heure_distribution" type="time"></div>
    </div>
    <div class="grid grid-2">
      <div class="field"><label><?= h($lang[25]) ?></label><select name="statut"><?php foreach ($statuts as $s): ?><option value="<?= h($s) ?>"><?= h($s) ?></option><?php endforeach; ?></select></div>
      <div class="field"><label><?= h($lang[246]) ?></label><input name="quota_par_adherent" type="number" min="0" value="1"></div>
    </div>
    <div class="actions"><button type="submit"><?= h($lang[83]) ?></button></div>
  </form>
</section>

<?php include __DIR__ . '/../inc/layout_bottom.php'; ?>
