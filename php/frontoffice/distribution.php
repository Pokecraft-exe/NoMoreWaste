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
        api_put('/api/v1/distributions/participation', ['distribution_id' => $id], $token);
    } elseif ($action === 'reserver') {
        $res = api_put('/api/v1/reservations', [
            'distribution_id' => $id,
            'stock_produit_id' => intval($_POST['stock_produit_id'] ?? 0),
            'quantite' => $_POST['quantite'] ?? '',
        ], $token);
        if ($res['status'] !== 201) {
            $error = $res['data']['error_description'] ?? $lang[252];
        }
    } elseif ($action === 'candidater_benevole') {
        api_put('/api/v1/benevoles/candidatures', [
            'motivation' => $_POST['motivation'] ?? '',
            'disponibilite' => $_POST['disponibilite'] ?? '',
        ], $token);
    }
    if (!$error) {
        header('Location: distribution.php?id=' . $id);
        exit;
    }
}

$res = api_get('/api/v1/distributions', ['id' => $id], $token);
$distribution = $res['data'] ?? null;

$pageTitle = $distribution ? $distribution['lieu'] : $lang[381];
$active = '';
include __DIR__ . '/../inc/layout_top.php';

if (!$distribution) {
    echo '<div class="empty">Cette distribution n\'existe pas.</div>';
    include __DIR__ . '/../inc/layout_bottom.php';
    exit;
}

$mapQuery = urlencode($distribution['lieu']);
?>

<p><a href="index.php">&larr; <?= h($lang[102]) ?></a></p>
<section class="section">
  <h1><?= h($distribution['lieu']) ?></h1>
  <div class="card-meta">
    <span class="badge"><?= h($distribution['statut']) ?></span>
    <span><?= h(format_date_fr($distribution['date_distribution'])) ?> <?= h($lang[300]) ?> <?= h($distribution['heure_distribution']) ?></span>
  </div>
  <div class="map-embed"><iframe src="https://www.google.com/maps?q=<?= $mapQuery ?>&output=embed" loading="lazy" title="<?= h($lang[248]) ?>" referrerpolicy="no-referrer-when-downgrade"></iframe></div>
</section>

<section class="section">
  <h2><?= h($lang[245]) ?></h2>
  <?php if (empty($distribution['denrees'])): ?>
    <div class="empty"><?= h($lang[249]) ?></div>
  <?php else: ?>
  <div class="table-wrap"><table>
    <thead><tr><th><?= h($lang[119]) ?></th><th><?= h($lang[121]) ?></th><th><?= h($lang[122]) ?></th></tr></thead>
    <tbody>
    <?php foreach ($distribution['denrees'] as $d): ?>
    <tr><td><?= h($d['nom']) ?></td><td><?= h($d['quantite']) ?> <?= h($d['unite']) ?></td><td><?= h($d['restant']) ?> <?= h($d['unite']) ?></td></tr>
    <?php endforeach; ?>
    </tbody>
  </table></div>
  <?php endif; ?>
</section>

<section class="section">
  <h2><?= h($lang[92]) ?></h2>
  <?php if (!$user): ?>
    <div class="note-box"><?= h($lang[251]) ?></div>
  <?php elseif (is_benevole($user) || is_staff($user)): ?>
    <p class="muted"><?= h($lang[228]) ?></p>
    <form method="post"><input type="hidden" name="action" value="participer"><button type="submit"><?= h($lang[230]) ?></button></form>
  <?php elseif (is_adherent($user)): ?>
    <?php if ($error): ?><div class="note-box"><?= h($error) ?></div><?php endif; ?>
    <p class="muted"><?= h($lang[247]) ?> <?= h($distribution['quota_par_adherent']) ?> <?= h($lang[302]) ?></p>
    <form method="post">
      <input type="hidden" name="action" value="reserver">
      <div class="grid grid-2">
        <div class="field"><label for="stock_produit_id"><?= h($lang[119]) ?></label>
          <select id="stock_produit_id" name="stock_produit_id" required>
            <?php foreach (($distribution['denrees'] ?? []) as $d): ?><option value="<?= (int) $d['stock_produit_id'] ?>"><?= h($d['nom']) ?></option><?php endforeach; ?>
          </select>
        </div>
        <div class="field"><label for="quantite"><?= h($lang[120]) ?></label><input id="quantite" name="quantite" type="number" min="1" value="1" required></div>
      </div>
      <div class="actions"><button type="submit"><?= h($lang[93]) ?></button></div>
    </form>
  <?php else: ?>
    <div class="note-box"><?= h($lang[250]) ?></div>
    <form method="post" style="margin-top:10px;">
      <input type="hidden" name="action" value="candidater_benevole">
      <div class="field"><label for="motivation"><?= h($lang[141]) ?></label><textarea id="motivation" name="motivation" required></textarea></div>
      <div class="field"><label for="disponibilite"><?= h($lang[142]) ?></label><input id="disponibilite" name="disponibilite" required></div>
      <div class="actions"><button type="submit"><?= h($lang[244]) ?></button></div>
    </form>
  <?php endif; ?>
</section>

<?php include __DIR__ . '/../inc/layout_bottom.php'; ?>
