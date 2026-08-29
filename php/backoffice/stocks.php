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
    if ($action === 'create_stock') {
        $res = api_put('/api/v1/stocks', [
            'site_id' => $_POST['site_id'] ?? '', 'nom' => $_POST['nom'] ?? '',
            'type_stock' => $_POST['type_stock'] ?? '', 'capacite_max' => $_POST['capacite_max'] ?? '',
        ], $token);
        if ($res['status'] !== 201) {
            $error = $res['data']['error_description'] ?? $lang[292];
        }
    } elseif ($action === 'referencer_produit') {
        $res = api_put('/api/v1/produits', [
            'barcode' => $_POST['barcode'] ?? '', 'nom' => $_POST['nom'] ?? '',
            'description' => $_POST['description'] ?? '', 'stock_id' => $_POST['stock_id'] ?? '',
            'quantite' => $_POST['quantite'] ?? '1', 'poids_kg' => $_POST['poids_kg'] ?? '',
            'date_peremption' => $_POST['date_peremption'] ?? '',
        ], $token);
        if ($res['status'] !== 201) {
            $error = $res['data']['error_description'] ?? $lang[292];
        }
    } elseif ($action === 'update_produit') {
        api_patch('/api/v1/produits?id=' . intval($_POST['produit_id'] ?? 0), [
            'etat' => $_POST['etat'] ?? '', 'stock_id' => $_POST['stock_id'] ?? '',
        ], $token);
    } elseif ($action === 'mouvement') {
        $res = api_put('/api/v1/produits/mouvements', [
            'produit_id' => $_POST['produit_id'] ?? '', 'stock_id' => $_POST['stock_id'] ?? '',
            'type_mouvement' => $_POST['type_mouvement'] ?? '', 'quantite' => $_POST['quantite'] ?? '1',
            'commentaire' => $_POST['commentaire'] ?? '',
        ], $token);
        if ($res['status'] !== 201) {
            $error = $res['data']['error_description'] ?? $lang[292];
        }
    }
    if (!$error) {
        header('Location: stocks.php');
        exit;
    }
}

$q = query_param('q');
$pageTitle = $lang[409];
$active = 'stocks';
$navMode = 'back';
include __DIR__ . '/../inc/layout_top.php';

$sitesRes = api_get('/api/v1/sites', [], $token);
$sites = $sitesRes['data']['sites'] ?? [];
$stocksRes = api_get('/api/v1/stocks', [], $token);
$stocks = $stocksRes['data']['stocks'] ?? [];
$produitsRes = api_get('/api/v1/produits', api_params(['from' => 0, 'size' => 100, 'q' => $q]), $token);
$produits = $produitsRes['data']['produits'] ?? [];
$typesStock = ['sec', 'frais', 'surgeles', 'materiel', 'autre'];
$etatsProduit = ['recu', 'controle', 'stocke', 'reserve', 'redistribue', 'jete'];
$typesMouvement = ['entree', 'sortie', 'transfert', 'ajustement'];
$preselect = intval($_GET['produit_id'] ?? 0);
?>
<section class="section">
  <h1><?= h($lang[409]) ?></h1>
</section>

<section class="section">
  <div class="section-head"><h2><?= h($lang[410]) ?></h2><a class="btn-sm" href="#creer-stock"><?= h($lang[413]) ?></a></div>
  <div class="table-wrap"><table>
    <thead><tr><th><?= h($lang[64]) ?></th><th><?= h($lang[444]) ?></th><th><?= h($lang[411]) ?></th><th><?= h($lang[25]) ?></th></tr></thead>
    <tbody>
    <?php foreach ($stocks as $s):
      $site = null;
      foreach ($sites as $st) { if ((int) $st['site_id'] === (int) $s['site_id']) { $site = $st; break; } }
      ?>
    <tr>
      <td><?= h($s['nom']) ?></td>
      <td><?= h($site['nom'] ?? ('#' . $s['site_id'])) ?></td>
      <td><span class="badge"><?= h($s['type_stock']) ?></span></td>
      <td><span class="badge <?= $s['actif'] ? 'badge-success' : 'badge-danger' ?>"><?= $s['actif'] ? $lang[368] : $lang[369] ?></span></td>
    </tr>
    <?php endforeach; ?>
    <?php if (empty($stocks)): ?><tr><td colspan="4" class="muted"><?= h($lang[422]) ?></td></tr><?php endif; ?>
    </tbody>
  </table></div>
</section>

<section class="section">
  <h2><?= h($lang[414]) ?></h2>
  <form class="filters" method="get">
    <div class="field"><label><?= h($lang[415]) ?> / <?= h($lang[64]) ?></label><input name="q" value="<?= h($q) ?>"></div>
    <div class="actions"><button type="submit"><?= h($lang[79]) ?></button></div>
  </form>
  <div class="table-wrap"><table>
    <thead><tr><th><?= h($lang[415]) ?></th><th><?= h($lang[64]) ?></th><th><?= h($lang[120]) ?></th><th><?= h($lang[25]) ?></th><th><?= h($lang[410]) ?></th><th></th></tr></thead>
    <tbody>
    <?php foreach ($produits as $p): ?>
    <tr>
      <td><code><?= h($p['barcode']) ?></code></td>
      <td><?= h($p['nom']) ?></td>
      <td><?= (int) $p['quantite'] ?></td>
      <td><span class="badge"><?= h($p['etat']) ?></span></td>
      <td><?= h($p['stock'] ?? '-') ?></td>
      <td><a class="btn-sm" href="?produit_id=<?= (int) $p['produit_id'] ?>#produit-<?= (int) $p['produit_id'] ?>"><?= h($lang[86]) ?></a></td>
    </tr>
    <?php endforeach; ?>
    <?php if (empty($produits)): ?><tr><td colspan="6" class="muted"><?= h($lang[420]) ?></td></tr><?php endif; ?>
    </tbody>
  </table></div>
</section>

<?php foreach ($produits as $p): if ($preselect && $preselect !== (int) $p['produit_id']) continue;
  $mouvementsRes = api_get('/api/v1/produits/mouvements', ['produit_id' => $p['produit_id']], $token);
  $mouvements = $mouvementsRes['data']['mouvements'] ?? [];
  ?>
<section class="section card" id="produit-<?= (int) $p['produit_id'] ?>">
  <h3><?= h($p['nom']) ?> - <code><?= h($p['barcode']) ?></code></h3>
  <?php if ($error): ?><div class="note-box"><?= h($error) ?></div><?php endif; ?>
  <form method="post">
    <input type="hidden" name="action" value="update_produit">
    <input type="hidden" name="produit_id" value="<?= (int) $p['produit_id'] ?>">
    <div class="grid grid-2">
      <div class="field"><label><?= h($lang[25]) ?></label>
        <select name="etat"><?php foreach ($etatsProduit as $e): ?><option value="<?= h($e) ?>" <?= $p['etat'] === $e ? 'selected' : '' ?>><?= h($e) ?></option><?php endforeach; ?></select>
      </div>
      <div class="field"><label><?= h($lang[410]) ?></label>
        <select name="stock_id"><option value=""><?= h($lang[111]) ?></option><?php foreach ($stocks as $s): ?><option value="<?= (int) $s['stock_id'] ?>"><?= h($s['nom']) ?></option><?php endforeach; ?></select>
      </div>
    </div>
    <div class="actions"><button type="submit"><?= h($lang[76]) ?></button></div>
  </form>

  <h4><?= h($lang[417]) ?></h4>
  <div class="table-wrap"><table><thead><tr><th><?= h($lang[418]) ?></th><th><?= h($lang[120]) ?></th><th><?= h($lang[446]) ?></th><th><?= h($lang[154]) ?></th></tr></thead><tbody>
  <?php foreach ($mouvements as $m): ?>
  <tr><td><span class="badge"><?= h($m['type_mouvement']) ?></span></td><td><?= (int) $m['quantite'] ?></td><td><?= h($m['commentaire'] ?? '-') ?></td><td><?= h(time_ago($m['date_mouvement'])) ?></td></tr>
  <?php endforeach; ?>
  <?php if (empty($mouvements)): ?><tr><td colspan="4" class="muted"><?= h($lang[421]) ?></td></tr><?php endif; ?>
  </tbody></table></div>
  <form method="post">
    <input type="hidden" name="action" value="mouvement">
    <input type="hidden" name="produit_id" value="<?= (int) $p['produit_id'] ?>">
    <div class="grid grid-2">
      <div class="field"><label><?= h($lang[418]) ?></label>
        <select name="type_mouvement"><?php foreach ($typesMouvement as $tm): ?><option value="<?= h($tm) ?>"><?= h($tm) ?></option><?php endforeach; ?></select>
      </div>
      <div class="field"><label><?= h($lang[120]) ?></label><input name="quantite" type="number" min="1" value="1" required></div>
    </div>
    <div class="field"><label><?= h($lang[410]) ?></label>
      <select name="stock_id"><option value=""><?= h($lang[111]) ?></option><?php foreach ($stocks as $s): ?><option value="<?= (int) $s['stock_id'] ?>"><?= h($s['nom']) ?></option><?php endforeach; ?></select>
    </div>
    <div class="field"><label><?= h($lang[446]) ?></label><input name="commentaire"></div>
    <div class="actions"><button type="submit"><?= h($lang[419]) ?></button></div>
  </form>
</section>
<?php endforeach; ?>

<section class="section card" id="creer-stock">
  <h2><?= h($lang[413]) ?></h2>
  <form method="post">
    <input type="hidden" name="action" value="create_stock">
    <div class="field"><label><?= h($lang[444]) ?></label>
      <select name="site_id"><?php foreach ($sites as $s): ?><option value="<?= (int) $s['site_id'] ?>"><?= h($s['nom']) ?> (<?= h($s['ville']) ?>)</option><?php endforeach; ?></select>
    </div>
    <div class="field"><label><?= h($lang[64]) ?></label><input name="nom" required></div>
    <div class="field"><label><?= h($lang[411]) ?></label>
      <select name="type_stock"><?php foreach ($typesStock as $t): ?><option value="<?= h($t) ?>"><?= h($t) ?></option><?php endforeach; ?></select>
    </div>
    <div class="field"><label><?= h($lang[412]) ?></label><input name="capacite_max" type="number" min="0" step="0.001"></div>
    <div class="actions"><button type="submit"><?= h($lang[83]) ?></button></div>
  </form>
</section>

<section class="section card" id="referencer">
  <h2><?= h($lang[416]) ?></h2>
  <form method="post">
    <input type="hidden" name="action" value="referencer_produit">
    <div class="field"><label><?= h($lang[415]) ?></label><input name="barcode" required></div>
    <div class="field"><label><?= h($lang[64]) ?></label><input name="nom" required></div>
    <div class="field"><label><?= h($lang[22]) ?></label><textarea name="description"></textarea></div>
    <div class="grid grid-2">
      <div class="field"><label><?= h($lang[120]) ?></label><input name="quantite" type="number" min="1" value="1"></div>
      <div class="field"><label><?= h($lang[410]) ?></label>
        <select name="stock_id"><option value=""><?= h($lang[111]) ?></option><?php foreach ($stocks as $s): ?><option value="<?= (int) $s['stock_id'] ?>"><?= h($s['nom']) ?></option><?php endforeach; ?></select>
      </div>
    </div>
    <div class="grid grid-2">
      <div class="field"><label><?= h($lang[445]) ?></label><input name="poids_kg" type="number" min="0" step="0.001"></div>
      <div class="field"><label><?= h($lang[126]) ?></label><input name="date_peremption" type="date"></div>
    </div>
    <div class="actions"><button type="submit"><?= h($lang[416]) ?></button></div>
  </form>
</section>

<?php include __DIR__ . '/../inc/layout_bottom.php'; ?>
