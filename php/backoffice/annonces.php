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
    if ($action === 'delete') {
        api_delete('/api/v1/annonces', ['id' => intval($_POST['annonce_echange_id'] ?? 0)], $token);
    }
    header('Location: annonces.php');
    exit;
}

$q = query_param('q');
$categorie = query_param('categorie');
$statutFilter = query_param('statut');
$pageTitle = $lang[17];
$active = 'annonces';
$navMode = 'back';
include __DIR__ . '/../inc/layout_top.php';

$res = api_get('/api/v1/annonces', api_params(['from' => 0, 'size' => 100, 'q' => $q, 'categorie' => $categorie, 'statut' => $statutFilter]), $token);
$annonces = $res['data']['annonces'] ?? [];
$categories = [
    'covoiturage' => $lang[118], 'reparation' => $lang[116], 'gardiennage' => $lang[117],
    'location' => $lang[115], 'vente' => $lang[113], 'don' => 'Don',
];
$etats = ['ouverte', 'en_cours', 'cloturee'];
?>
<section class="section">
  <h1><?= h($lang[173]) ?></h1>
  <p class="muted"><?= h($lang[304]) ?></p>
  <form class="filters" method="get">
    <div class="field"><label><?= h($lang[139]) ?></label><input name="q" value="<?= h($q) ?>"></div>
    <div class="field"><label><?= h($lang[19]) ?></label><select name="categorie"><option value=""><?= h($lang[110]) ?></option><?php foreach ($categories as $val => $label): ?><option value="<?= h($val) ?>" <?= $categorie === $val ? 'selected' : '' ?>><?= h($label) ?></option><?php endforeach; ?></select></div>
    <div class="field"><label><?= h($lang[25]) ?></label><select name="statut"><option value=""><?= h($lang[111]) ?></option><?php foreach ($etats as $e): ?><option value="<?= h($e) ?>" <?= $statutFilter === $e ? 'selected' : '' ?>><?= h($e) ?></option><?php endforeach; ?></select></div>
    <div class="actions"><button type="submit"><?= h($lang[79]) ?></button></div>
  </form>
  <div class="table-wrap"><table>
    <thead><tr><th><?= h($lang[20]) ?></th><th><?= h($lang[19]) ?></th><th><?= h($lang[36]) ?></th><th><?= h($lang[108]) ?></th><th><?= h($lang[21]) ?></th><th><?= h($lang[109]) ?></th><th></th></tr></thead>
    <tbody>
    <?php foreach ($annonces as $a): ?>
    <tr>
      <td><a href="../frontoffice/annonce.php?id=<?= (int) $a['annonce_echange_id'] ?>" target="_blank"><?= h($a['titre']) ?></a></td>
      <td><span class="badge"><?= h($categories[$a['categorie']] ?? $a['categorie']) ?></span></td>
      <td><?= h($a['auteur'] ?? '-') ?></td>
      <td><span class="badge"><?= h($a['statut']) ?></span></td>
      <td><?= $a['prix'] !== null ? h($a['prix']) . ' €' : $lang[112] ?></td>
      <td><?= h(time_ago($a['date_publication'])) ?></td>
      <td>
        <form method="post" onsubmit="return confirm('<?= h($lang[283]) ?>');">
          <input type="hidden" name="action" value="delete">
          <input type="hidden" name="annonce_echange_id" value="<?= (int) $a['annonce_echange_id'] ?>">
          <button type="submit" class="btn-sm btn-danger"><?= h($lang[81]) ?></button>
        </form>
      </td>
    </tr>
    <?php endforeach; ?>
    <?php if (empty($annonces)): ?><tr><td colspan="7" class="muted"><?= h($lang[209]) ?></td></tr><?php endif; ?>
    </tbody>
  </table></div>
</section>

<?php include __DIR__ . '/../inc/layout_bottom.php'; ?>
