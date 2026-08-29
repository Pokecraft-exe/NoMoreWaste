<?php
require_once __DIR__ . '/../inc/config.php';
require_once __DIR__ . '/../inc/auth.php';
require_once __DIR__ . '/../inc/helpers.php';
require_once __DIR__ . '/../inc/changeLanguage.php';
$lang = $LANGUAGE_CONTENTS;

$user = current_user();
$token = session_token();
$id = intval($_GET['id'] ?? 0);
$error = null;

if ($_SERVER['REQUEST_METHOD'] === 'POST' && $user) {
    $action = $_POST['action'] ?? '';
    if ($action === 'participer') {
        api_put('/api/v1/collectes/participation', ['collecte_id' => $id, 'role_mission' => $_POST['role_mission'] ?? ''], $token);
    } elseif ($action === 'ajouter_denree') {
        $res = api_put('/api/v1/collectes/denrees', [
            'collecte_id' => $id,
            'stock_produit_id' => intval($_POST['stock_produit_id'] ?? 0),
            'quantite' => $_POST['quantite'] ?? '',
            'non_perissable' => isset($_POST['non_perissable']) ? '1' : '0',
            'dlc' => $_POST['dlc'] ?? '',
        ], $token);
        if ($res['status'] !== 201) {
            $error = $res['data']['error_description'] ?? $lang[378];
        }
    } elseif ($action === 'confirmer_denree') {
        api_patch('/api/v1/collectes/denrees?id=' . intval($_POST['denree_id'] ?? 0), ['confirmee' => '1'], $token);
    } elseif ($action === 'retirer_denree') {
        api_delete('/api/v1/collectes/denrees', ['id' => intval($_POST['denree_id'] ?? 0)], $token);
    } elseif ($action === 'confirmer_collecte') {
        api_put('/api/v1/collectes/confirmer', ['collecte_id' => $id], $token);
    } elseif ($action === 'nouveau_produit') {
        $res = api_put('/api/v1/stock-produits', ['nom' => $_POST['nom_produit'] ?? '', 'unite' => $_POST['unite'] ?? 'unites'], $token);
        header('Location: collecte.php?id=' . $id);
        exit;
    }
    header('Location: collecte.php?id=' . $id);
    exit;
}

$res = api_get('/api/v1/collectes', ['id' => $id], $token);
$collecte = $res['data'] ?? null;

$pageTitle = $collecte ? $collecte['lieu'] : $lang[380];
$active = '';
include __DIR__ . '/../inc/layout_top.php';

if (!$collecte) {
    echo '<div class="empty">Cette collecte n\'existe pas.</div>';
    include __DIR__ . '/../inc/layout_bottom.php';
    exit;
}

$canConfirmCollecte = $user && is_staff($user);
$canManage = $canConfirmCollecte;
if ($user && !$canManage) {
    foreach (($collecte['benevoles_affectes'] ?? []) as $b) {
        if ($b['benevole_id'] == $user['compte_id']) {
            $canManage = true;
        }
    }
}
$canPropose = $user && is_commercant($user) && $collecte['commercant_id'] == $user['compte_id'];

$produitsRes = api_get('/api/v1/stock-produits', ['q' => ''], $token);
$produits = $produitsRes['data']['produits'] ?? [];
?>

<p><a href="index.php">&larr; <?= h($lang[102]) ?></a></p>
<section class="section">
  <h1><?= h($collecte['lieu']) ?></h1>
  <div class="card-meta">
    <span class="badge"><?= h($collecte['statut']) ?></span>
    <span><?= h(format_date_fr($collecte['date_collecte'])) ?> <?= h($lang[300]) ?> <?= h($collecte['heure_collecte']) ?></span>
    <?php if (!empty($collecte['partenaire'])): ?><span><?= h($lang[301]) ?> <?= h($collecte['partenaire']) ?></span><?php endif; ?>
  </div>
  <?php if (!empty($collecte['description'])): ?><p><?= h($collecte['description']) ?></p><?php endif; ?>
</section>

<section class="section">
  <h2><?= h($lang[124]) ?></h2>
  <?php if (empty($collecte['denrees'])): ?>
    <div class="empty"><?= h($lang[226]) ?></div>
  <?php else: ?>
  <div class="table-wrap"><table>
    <thead><tr><th><?= h($lang[119]) ?></th><th><?= h($lang[120]) ?></th><th><?= h($lang[126]) ?></th><th><?= h($lang[125]) ?></th><th><?= h($lang[25]) ?></th><?php if ($canManage): ?><th></th><?php endif; ?></tr></thead>
    <tbody>
    <?php foreach ($collecte['denrees'] as $d): ?>
    <tr>
      <td><?= h($d['nom']) ?></td>
      <td><?= h($d['quantite']) ?></td>
      <td><?= $d['non_perissable'] ? '<span class="badge badge-muted">Non perissable</span>' : h($d['dlc']) ?></td>
      <td><?= $d['propose_par_type'] === 'commercant' ? $lang[242] : $lang[379] ?></td>
      <td><?= $d['confirmee'] ? '<span class="badge badge-success">Collecte</span>' : '<span class="badge badge-warning">A confirmer</span>' ?></td>
      <?php if ($canManage): ?>
      <td>
        <?php if (!$d['confirmee']): ?>
        <form method="post" style="display:inline;"><input type="hidden" name="action" value="confirmer_denree"><input type="hidden" name="denree_id" value="<?= (int) $d['collecte_denree_id'] ?>"><button type="submit" class="btn-sm"><?= h($lang[241]) ?></button></form>
        <?php endif; ?>
        <form method="post" style="display:inline;" onsubmit="return confirm('<?= h($lang[286]) ?>');"><input type="hidden" name="action" value="retirer_denree"><input type="hidden" name="denree_id" value="<?= (int) $d['collecte_denree_id'] ?>"><button type="submit" class="btn-sm btn-danger"><?= h($lang[94]) ?></button></form>
      </td>
      <?php endif; ?>
    </tr>
    <?php endforeach; ?>
    </tbody>
  </table></div>
  <?php endif; ?>

  <?php if ($canConfirmCollecte): ?>
  <div class="actions" style="margin-top:14px;">
    <form method="post" onsubmit="return confirm('<?= h($lang[287]) ?>');">
      <input type="hidden" name="action" value="confirmer_collecte">
      <button type="submit" <?= $collecte['stock_mis_a_jour'] ? 'disabled' : '' ?>><?= $collecte['stock_mis_a_jour'] ? $lang[237] : $lang[236] ?></button>
    </form>
  </div>
  <?php endif; ?>

  <?php if ($user && (is_benevole($user) || is_staff($user)) && !$canManage): ?>
  <div class="card" style="margin-top:16px;">
    <h3><?= h($lang[92]) ?></h3>
    <p class="muted"><?= h($lang[228]) ?></p>
    <form method="post">
      <input type="hidden" name="action" value="participer">
      <div class="field"><label for="role_mission"><?= h($lang[140]) ?></label><input id="role_mission" name="role_mission"></div>
      <div class="actions"><button type="submit"><?= h($lang[229]) ?></button></div>
    </form>
  </div>
  <?php endif; ?>

  <?php if ($canManage || $canPropose): ?>
  <div class="card" style="margin-top:16px;">
    <h3><?= $canManage ? $lang[233] : $lang[231] ?></h3>
    <?php if ($error): ?><div class="note-box"><?= h($error) ?></div><?php endif; ?>
    <form method="post">
      <input type="hidden" name="action" value="ajouter_denree">
      <div class="field"><label for="stock_produit_id"><?= h($lang[119]) ?></label>
        <select id="stock_produit_id" name="stock_produit_id" required>
          <?php foreach ($produits as $p): ?><option value="<?= (int) $p['stock_produit_id'] ?>"><?= h($p['nom']) ?> (<?= h($p['unite']) ?>)</option><?php endforeach; ?>
        </select>
      </div>
      <div class="field"><label for="quantite"><?= h($lang[120]) ?></label><input id="quantite" name="quantite" type="number" min="1" value="1" required></div>
      <div class="field"><label><input type="checkbox" name="non_perissable" id="non_perissable"> <?= h($lang[240]) ?></label></div>
      <div class="field"><label for="dlc"><?= h($lang[126]) ?></label><input id="dlc" name="dlc" type="date"></div>
      <div class="actions"><button type="submit"><?= h($lang[235]) ?></button></div>
    </form>
  </div>
  <div class="card" style="margin-top:12px;">
    <h4><?= h($lang[232]) ?></h4>
    <form method="post">
      <input type="hidden" name="action" value="nouveau_produit">
      <div class="grid grid-2">
        <div class="field"><label for="nom_produit"><?= h($lang[64]) ?></label><input id="nom_produit" name="nom_produit" required></div>
        <div class="field"><label for="unite"><?= h($lang[123]) ?></label>
          <select id="unite" name="unite"><option><?= h($lang[352]) ?></option><option><?= h($lang[353]) ?></option><option><?= h($lang[355]) ?></option><option><?= h($lang[351]) ?></option><option><?= h($lang[354]) ?></option></select>
        </div>
      </div>
      <div class="actions"><button type="submit"><?= h($lang[234]) ?></button></div>
    </form>
  </div>
  <?php elseif (!$user): ?>
  <div class="note-box" style="margin-top:16px;"><?= h($lang[243]) ?></div>
  <?php endif; ?>
</section>

<?php include __DIR__ . '/../inc/layout_bottom.php'; ?>
