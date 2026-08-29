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
        $email = $_POST['email'] ?? '';
        $found = api_get('/api/v1/admin/comptes', ['q' => $email, 'from' => 0, 'size' => 5], $token);
        $match = $found['data']['comptes'][0] ?? null;
        if ($match) {
            $res = api_put('/api/v1/benevoles/candidatures', [
                'personne_id' => $match['compte_id'],
                'motivation' => $_POST['motivation'] ?? '',
                'disponibilite' => $_POST['disponibilite'] ?? '',
            ], $token);
            if ($res['status'] !== 201) {
                $error = $res['data']['error_description'] ?? $lang[292];
            }
        } else {
            $error = $lang[214];
        }
    } elseif ($action === 'decide') {
        api_patch('/api/v1/benevoles/candidatures?id=' . intval($_POST['id'] ?? 0), [
            'statut' => $_POST['statut'] ?? '',
            'commentaire' => $_POST['commentaire'] ?? '',
        ], $token);
    } elseif ($action === 'delete') {
        api_delete('/api/v1/benevoles/candidatures', ['id' => intval($_POST['id'] ?? 0)], $token);
    }
    if (!$error) {
        header('Location: candidatures.php');
        exit;
    }
}

$q = query_param('q');
$statutFilter = query_param('statut');
$pageTitle = $lang[48];
$active = 'candidatures';
$navMode = 'back';
include __DIR__ . '/../inc/layout_top.php';

$res = api_get('/api/v1/benevoles/candidatures', api_params(['from' => 0, 'size' => 100, 'q' => $q, 'statut' => $statutFilter]), $token);
$candidatures = $res['data']['candidatures'] ?? [];
$statuts = ['recue' => $lang[336], 'en_etude' => $lang[337], 'validee' => $lang[338], 'refusee' => $lang[339], 'archivee' => $lang[340]];
?>
<section class="section">
  <div class="section-head"><h1><?= h($lang[151]) ?></h1><a class="btn-sm" href="#creer"><?= h($lang[177]) ?></a></div>
  <form class="filters" method="get">
    <div class="field"><label><?= h($lang[149]) ?></label><input name="q" value="<?= h($q) ?>"></div>
    <div class="field"><label><?= h($lang[25]) ?></label>
      <select name="statut">
        <option value=""><?= h($lang[111]) ?></option>
        <?php foreach ($statuts as $val => $label): ?><option value="<?= h($val) ?>" <?= $statutFilter === $val ? 'selected' : '' ?>><?= h($label) ?></option><?php endforeach; ?>
      </select>
    </div>
    <div class="actions"><button type="submit"><?= h($lang[80]) ?></button></div>
  </form>
  <div class="table-wrap"><table>
    <thead><tr><th><?= h($lang[149]) ?></th><th><?= h($lang[336]) ?></th><th><?= h($lang[141]) ?></th><th><?= h($lang[25]) ?></th><th></th></tr></thead>
    <tbody>
    <?php foreach ($candidatures as $c): ?>
    <tr>
      <td><?= h($c['prenom'] . ' ' . $c['nom']) ?></td>
      <td><?= h(time_ago($c['date_candidature'])) ?></td>
      <td class="muted"><?= h(mb_strimwidth($c['motivation'] ?? '', 0, 60, '...')) ?></td>
      <td><span class="badge"><?= h($c['statut']) ?></span></td>
      <td><a class="btn-sm" href="#decider-<?= (int) $c['candidature_benevole_id'] ?>"><?= h($lang[160]) ?></a></td>
    </tr>
    <?php endforeach; ?>
    <?php if (empty($candidatures)): ?><tr><td colspan="5" class="muted"><?= h($lang[213]) ?></td></tr><?php endif; ?>
    </tbody>
  </table></div>
</section>

<?php foreach ($candidatures as $c): ?>
<section class="section card" id="decider-<?= (int) $c['candidature_benevole_id'] ?>">
  <h3><?= h($lang[150]) ?> <?= h($c['prenom'] . ' ' . $c['nom']) ?></h3>
  <p><strong><?= h($lang[143]) ?></strong> <?= h($c['disponibilite'] ?? '-') ?></p>
  <p><strong><?= h($lang[331]) ?></strong> <?= h($c['motivation'] ?? '-') ?></p>
  <form method="post">
    <input type="hidden" name="action" value="decide">
    <input type="hidden" name="id" value="<?= (int) $c['candidature_benevole_id'] ?>">
    <div class="field"><label><?= h($lang[25]) ?></label>
      <select name="statut"><?php foreach ($statuts as $val => $label): ?><option value="<?= h($val) ?>" <?= $c['statut'] === $val ? 'selected' : '' ?>><?= h($label) ?></option><?php endforeach; ?></select>
    </div>
    <div class="field"><label><?= h($lang[152]) ?></label><textarea name="commentaire"><?= h($c['commentaire'] ?? '') ?></textarea></div>
    <div class="note-box"><?= h($lang[322]) ?></div>
    <div class="actions">
      <button type="submit"><?= h($lang[76]) ?></button>
    </div>
  </form>
  <form method="post" onsubmit="return confirm('<?= h($lang[285]) ?>');">
    <input type="hidden" name="action" value="delete">
    <input type="hidden" name="id" value="<?= (int) $c['candidature_benevole_id'] ?>">
    <button type="submit" class="btn-danger"><?= h($lang[81]) ?></button>
  </form>
</section>
<?php endforeach; ?>

<section class="section card" id="creer">
  <h2><?= h($lang[177]) ?></h2>
  <?php if ($error): ?><div class="note-box"><?= h($error) ?></div><?php endif; ?>
  <form method="post">
    <input type="hidden" name="action" value="create">
    <div class="field"><label><?= h($lang[330]) ?></label><input name="email" type="email" required></div>
    <div class="field"><label><?= h($lang[142]) ?></label><input name="disponibilite"></div>
    <div class="field"><label><?= h($lang[141]) ?></label><textarea name="motivation"></textarea></div>
    <div class="actions"><button type="submit"><?= h($lang[83]) ?></button></div>
  </form>
</section>

<?php include __DIR__ . '/../inc/layout_bottom.php'; ?>
