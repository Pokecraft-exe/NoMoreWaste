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
        $commercantId = '';
        if (!empty($_POST['commercant_email'])) {
            $found = api_get('/api/v1/admin/comptes', ['q' => $_POST['commercant_email'], 'type_utilisateur' => 'commercant', 'from' => 0, 'size' => 5], $token);
            $match = $found['data']['comptes'][0] ?? null;
            $commercantId = $match ? $match['compte_id'] : '';
        }
        $res = api_put('/api/v1/collectes', [
            'lieu' => $_POST['lieu'] ?? '', 'date_collecte' => $_POST['date_collecte'] ?? '',
            'heure_collecte' => $_POST['heure_collecte'] ?? '', 'partenaire' => $_POST['partenaire'] ?? '',
            'commercant_id' => $commercantId, 'statut' => $_POST['statut'] ?? '', 'description' => $_POST['description'] ?? '',
        ], $token);
        if ($res['status'] !== 201) {
            $error = $res['data']['error_description'] ?? $lang[292];
        }
    } elseif ($action === 'update') {
        api_patch('/api/v1/collectes?id=' . intval($_POST['id'] ?? 0), [
            'lieu' => $_POST['lieu'] ?? '', 'date_collecte' => $_POST['date_collecte'] ?? '',
            'heure_collecte' => $_POST['heure_collecte'] ?? '', 'partenaire' => $_POST['partenaire'] ?? '',
            'statut' => $_POST['statut'] ?? '', 'description' => $_POST['description'] ?? '',
        ], $token);
    } elseif ($action === 'cancel') {
        api_delete('/api/v1/collectes', ['id' => intval($_POST['id'] ?? 0)], $token);
    } elseif ($action === 'affecter') {
        api_put('/api/v1/collectes/benevoles', [
            'collecte_id' => $_POST['collecte_id'] ?? '', 'benevole_id' => $_POST['benevole_id'] ?? '',
            'role_mission' => $_POST['role_mission'] ?? '',
        ], $token);
    } elseif ($action === 'generer_pdf') {
        api_put('/api/v1/documents', ['collecte_id' => $_POST['collecte_id'] ?? ''], $token);
    }
    if (!$error) {
        header('Location: collectes.php');
        exit;
    }
}

$q = query_param('q');
$statutFilter = query_param('statut');
$pageTitle = $lang[45];
$active = 'collectes';
$navMode = 'back';
include __DIR__ . '/../inc/layout_top.php';

$res = api_get('/api/v1/collectes', api_params(['from' => 0, 'size' => 100, 'q' => $q, 'statut' => $statutFilter]), $token);
$collectes = $res['data']['collectes'] ?? [];
$statuts = ['planifiee', 'confirmee', 'en_cours', 'terminee', 'annulee'];
$preselect = intval($_GET['id'] ?? 0);
?>
<section class="section">
  <div class="section-head"><h1><?= h($lang[45]) ?></h1><a class="btn-sm" href="#creer"><?= h($lang[178]) ?></a></div>
  <form class="filters" method="get">
    <div class="field"><label><?= h($lang[321]) ?></label><input name="q" value="<?= h($q) ?>"></div>
    <div class="field"><label><?= h($lang[25]) ?></label><select name="statut"><option value=""><?= h($lang[111]) ?></option><?php foreach ($statuts as $s): ?><option value="<?= h($s) ?>" <?= $statutFilter === $s ? 'selected' : '' ?>><?= h($s) ?></option><?php endforeach; ?></select></div>
    <div class="actions"><button type="submit"><?= h($lang[79]) ?></button></div>
  </form>
  <div class="table-wrap"><table>
    <thead><tr><th><?= h($lang[104]) ?></th><th><?= h($lang[14]) ?></th><th><?= h($lang[341]) ?></th><th><?= h($lang[25]) ?></th><th></th></tr></thead>
    <tbody>
    <?php foreach ($collectes as $c): ?>
    <tr>
      <td><?= h($c['lieu']) ?></td><td><?= h($c['date_collecte']) ?></td><td><?= h($c['partenaire'] ?? '-') ?></td>
      <td><span class="badge"><?= h($c['statut']) ?></span></td>
      <td><a class="btn-sm" href="#detail-<?= (int) $c['collecte_id'] ?>"><?= h($lang[86]) ?></a></td>
    </tr>
    <?php endforeach; ?>
    <?php if (empty($collectes)): ?><tr><td colspan="5" class="muted"><?= h($lang[211]) ?></td></tr><?php endif; ?>
    </tbody>
  </table></div>
</section>

<?php foreach ($collectes as $c): if ($preselect && $preselect !== (int) $c['collecte_id']) continue;
  $detail = api_get('/api/v1/collectes', ['id' => $c['collecte_id']], $token)['data'] ?? $c;
  ?>
<section class="section card" id="detail-<?= (int) $c['collecte_id'] ?>">
  <h3><?= h($lang[315]) ?> <?= h($c['lieu']) ?></h3>
  <form method="post">
    <input type="hidden" name="action" value="update">
    <input type="hidden" name="id" value="<?= (int) $c['collecte_id'] ?>">
    <div class="field"><label><?= h($lang[104]) ?></label><input name="lieu" value="<?= h($c['lieu']) ?>" required></div>
    <div class="grid grid-2">
      <div class="field"><label><?= h($lang[14]) ?></label><input name="date_collecte" type="date" value="<?= h($c['date_collecte']) ?>" required></div>
      <div class="field"><label><?= h($lang[103]) ?></label><input name="heure_collecte" type="time" value="<?= h($c['heure_collecte']) ?>"></div>
    </div>
    <div class="field"><label><?= h($lang[341]) ?></label><input name="partenaire" value="<?= h($c['partenaire'] ?? '') ?>"></div>
    <div class="field"><label><?= h($lang[25]) ?></label><select name="statut"><?php foreach ($statuts as $s): ?><option value="<?= h($s) ?>" <?= $c['statut'] === $s ? 'selected' : '' ?>><?= h($s) ?></option><?php endforeach; ?></select></div>
    <div class="field"><label><?= h($lang[22]) ?></label><textarea name="description"><?= h($c['description'] ?? '') ?></textarea></div>
    <div class="actions"><button type="submit"><?= h($lang[76]) ?></button></div>
  </form>
  <form method="post" onsubmit="return confirm('<?= h($lang[288]) ?>');">
    <input type="hidden" name="action" value="cancel"><input type="hidden" name="id" value="<?= (int) $c['collecte_id'] ?>">
    <button type="submit" class="btn-danger"><?= h($lang[89]) ?></button>
  </form>
  <p><a href="../frontoffice/collecte.php?id=<?= (int) $c['collecte_id'] ?>" target="_blank"><?= h($lang[85]) ?> <?= h($lang[367]) ?> &rarr;</a></p>

  <h4><?= h($lang[436]) ?></h4>
  <?php $documents = api_get('/api/v1/documents', ['collecte_id' => $c['collecte_id']], $token)['data']['documents'] ?? []; ?>
  <?php if (!empty($documents)): ?>
  <ul>
    <?php foreach ($documents as $d): ?>
    <li><a href="document.php?id=<?= (int) $d['document_genere_id'] ?>"><?= h($lang[435]) ?> (<?= h(time_ago($d['date_generation'])) ?>)</a></li>
    <?php endforeach; ?>
  </ul>
  <?php endif; ?>
  <form method="post">
    <input type="hidden" name="action" value="generer_pdf">
    <input type="hidden" name="collecte_id" value="<?= (int) $c['collecte_id'] ?>">
    <div class="actions"><button type="submit" class="btn-secondary"><?= h($lang[436]) ?></button></div>
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
    <input type="hidden" name="collecte_id" value="<?= (int) $c['collecte_id'] ?>">
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
  <h2><?= h($lang[178]) ?></h2>
  <?php if ($error): ?><div class="note-box"><?= h($error) ?></div><?php endif; ?>
  <form method="post">
    <input type="hidden" name="action" value="create">
    <div class="field"><label><?= h($lang[104]) ?></label><input name="lieu" required></div>
    <div class="grid grid-2">
      <div class="field"><label><?= h($lang[14]) ?></label><input name="date_collecte" type="date" required></div>
      <div class="field"><label><?= h($lang[103]) ?></label><input name="heure_collecte" type="time"></div>
    </div>
    <div class="field"><label><?= h($lang[320]) ?></label><input name="partenaire"></div>
    <div class="field"><label><?= h($lang[319]) ?></label><input name="commercant_email" type="email"></div>
    <div class="field"><label><?= h($lang[25]) ?></label><select name="statut"><?php foreach ($statuts as $s): ?><option value="<?= h($s) ?>"><?= h($s) ?></option><?php endforeach; ?></select></div>
    <div class="field"><label><?= h($lang[22]) ?></label><textarea name="description"></textarea></div>
    <div class="actions"><button type="submit"><?= h($lang[83]) ?></button></div>
  </form>
</section>

<?php include __DIR__ . '/../inc/layout_bottom.php'; ?>
