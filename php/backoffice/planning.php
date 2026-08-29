<?php
require_once __DIR__ . '/../inc/config.php';
require_once __DIR__ . '/../inc/auth.php';
require_once __DIR__ . '/../inc/helpers.php';
require_once __DIR__ . '/../inc/changeLanguage.php';
$lang = $LANGUAGE_CONTENTS;
require_once __DIR__ . '/../inc/xlsx.php';

function exclude_already_affected($participations, $affectes) {
    $affectedIds = array_map(fn($b) => (int) $b['benevole_id'], $affectes);
    return array_values(array_filter($participations, fn($p) => !in_array((int) $p['benevole_id'], $affectedIds, true)));
}

function collect_planning_events($token) {
    global $lang;
    $events = [];

    $collectesRes = api_get('/api/v1/collectes', ['from' => 0, 'size' => 200, 'prochaine' => 1], $token);
    foreach (($collectesRes['data']['collectes'] ?? []) as $c) {
        $detail = api_get('/api/v1/collectes', ['id' => $c['collecte_id']], $token)['data'] ?? [];
        $participationsRes = api_get('/api/v1/collectes/participation', ['collecte_id' => $c['collecte_id']], $token);
        $affectes = $detail['benevoles_affectes'] ?? [];
        $events[] = [
            'type' => 'collecte', 'type_label' => $lang[380], 'id' => $c['collecte_id'], 'lieu' => $c['lieu'],
            'date' => $c['date_collecte'], 'heure' => $c['heure_collecte'], 'statut' => $c['statut'],
            'affectes' => $affectes, 'participations' => exclude_already_affected($participationsRes['data']['benevoles'] ?? [], $affectes),
            'href' => 'collectes.php?id=' . $c['collecte_id'],
        ];
    }

    $distributionsRes = api_get('/api/v1/distributions', ['from' => 0, 'size' => 200, 'prochaine' => 1], $token);
    foreach (($distributionsRes['data']['distributions'] ?? []) as $d) {
        $detail = api_get('/api/v1/distributions', ['id' => $d['distribution_id']], $token)['data'] ?? [];
        $participationsRes = api_get('/api/v1/distributions/participation', ['distribution_id' => $d['distribution_id']], $token);
        $affectes = $detail['benevoles_affectes'] ?? [];
        $events[] = [
            'type' => 'distribution', 'type_label' => $lang[381], 'id' => $d['distribution_id'], 'lieu' => $d['lieu'],
            'date' => $d['date_distribution'], 'heure' => $d['heure_distribution'], 'statut' => $d['statut'],
            'affectes' => $affectes, 'participations' => exclude_already_affected($participationsRes['data']['benevoles'] ?? [], $affectes),
            'href' => 'distributions.php?id=' . $d['distribution_id'],
        ];
    }

    usort($events, fn($a, $b) => strcmp($a['date'], $b['date']));
    return $events;
}

$user = current_user();
$token = session_token();
require_staff_or_404($user);
$error = null;

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $action = $_POST['action'] ?? '';
    if ($action === 'create_collecte') {
        $res = api_put('/api/v1/collectes', [
            'lieu' => $_POST['lieu'] ?? '', 'date_collecte' => $_POST['date_collecte'] ?? '',
            'heure_collecte' => $_POST['heure_collecte'] ?? '', 'partenaire' => $_POST['partenaire'] ?? '',
            'commercant_id' => '', 'statut' => 'planifiee', 'description' => $_POST['description'] ?? '',
        ], $token);
        if ($res['status'] !== 201) {
            $error = $res['data']['error_description'] ?? $lang[292];
        }
    } elseif ($action === 'create_distribution') {
        $res = api_put('/api/v1/distributions', [
            'lieu' => $_POST['lieu'] ?? '', 'date_distribution' => $_POST['date_distribution'] ?? '',
            'heure_distribution' => $_POST['heure_distribution'] ?? '', 'statut' => 'planifiee',
            'quota_par_adherent' => $_POST['quota_par_adherent'] ?? '1',
        ], $token);
        if ($res['status'] !== 201) {
            $error = $res['data']['error_description'] ?? $lang[292];
        }
    } elseif ($action === 'accepter_collecte') {
        api_put('/api/v1/collectes/benevoles', [
            'collecte_id' => $_POST['collecte_id'] ?? '', 'benevole_id' => $_POST['benevole_id'] ?? '',
            'role_mission' => $_POST['role_mission'] ?: $lang[75],
        ], $token);
    } elseif ($action === 'accepter_distribution') {
        api_put('/api/v1/distributions/benevoles', [
            'distribution_id' => $_POST['distribution_id'] ?? '', 'benevole_id' => $_POST['benevole_id'] ?? '',
            'role_mission' => $_POST['role_mission'] ?: $lang[75],
        ], $token);
    }
    if (!$error) {
        header('Location: planning.php');
        exit;
    }
}

if (query_param('action') === 'export_xlsx') {
    $rows = [];
    foreach (collect_planning_events($token) as $e) {
        $benevoles = implode(', ', array_map(fn($b) => $b['prenom'] . ' ' . $b['nom'] . ' (' . $b['role_mission'] . ')', $e['affectes']));
        $rows[] = [$e['type_label'], $e['lieu'], format_date_fr($e['date']), substr($e['heure'] ?? '', 0, 5), $e['statut'], $benevoles];
    }
    send_xlsx('planning.xlsx', [$lang[106], $lang[104], $lang[14], $lang[103], $lang[25], $lang[373]], $rows);
}

$pageTitle = $lang[47];
$active = 'planning';
$navMode = 'back';
include __DIR__ . '/../inc/layout_top.php';

$events = collect_planning_events($token);
?>
<section class="section">
  <div class="section-head">
    <h1><?= h($lang[47]) ?></h1>
    <div class="actions">
      <a class="btn-sm" href="#creer-collecte"><?= h($lang[178]) ?></a>
      <a class="btn-sm" href="#creer-distribution"><?= h($lang[179]) ?></a>
      <a class="btn-secondary btn-sm" href="?action=export_xlsx"><?= h($lang[97]) ?></a>
    </div>
  </div>
  <p class="muted"><?= h($lang[333]) ?></p>
</section>

<?php foreach ($events as $e): ?>
<section class="section card">
  <div class="section-head">
    <h3><span class="badge"><?= h($e['type_label']) ?></span> <?= h($e['lieu']) ?></h3>
    <a class="btn-secondary btn-sm" href="<?= h($e['href']) ?>"><?= h($lang[84]) ?></a>
  </div>
  <p class="muted"><?= h(format_date_fr($e['date'])) ?><?php if ($e['heure']): ?> - <?= h(substr($e['heure'], 0, 5)) ?><?php endif; ?> - <span class="badge"><?= h($e['statut']) ?></span></p>

  <h4><?= h($lang[314]) ?></h4>
  <div class="table-wrap"><table><thead><tr><th><?= h($lang[64]) ?></th><th><?= h($lang[105]) ?></th></tr></thead><tbody>
  <?php foreach ($e['affectes'] as $b): ?>
  <tr><td><?= h($b['prenom'] . ' ' . $b['nom']) ?></td><td><?= h($b['role_mission']) ?></td></tr>
  <?php endforeach; ?>
  <?php if (empty($e['affectes'])): ?><tr><td colspan="2" class="muted"><?= h($lang[222]) ?></td></tr><?php endif; ?>
  </tbody></table></div>

  <?php if (!empty($e['participations'])): ?>
  <h4><?= h($lang[332]) ?></h4>
  <div class="table-wrap"><table><thead><tr><th><?= h($lang[64]) ?></th><th></th></tr></thead><tbody>
  <?php foreach ($e['participations'] as $p): ?>
  <tr>
    <td><?= h($p['prenom'] . ' ' . $p['nom']) ?></td>
    <td>
      <form method="post" style="display:flex;gap:6px;align-items:center;">
        <input type="hidden" name="action" value="accepter_<?= h($e['type']) ?>">
        <input type="hidden" name="<?= h($e['type']) ?>_id" value="<?= (int) $e['id'] ?>">
        <input type="hidden" name="benevole_id" value="<?= (int) $p['benevole_id'] ?>">
        <input name="role_mission" placeholder="<?= h($lang[105]) ?>" value="<?= h($p['role_mission'] ?? '') ?>" style="max-width:160px;">
        <button type="submit" class="btn-sm"><?= h($lang[88]) ?></button>
      </form>
    </td>
  </tr>
  <?php endforeach; ?>
  </tbody></table></div>
  <?php endif; ?>
</section>
<?php endforeach; ?>
<?php if (empty($events)): ?>
<section class="section"><p class="muted"><?= h($lang[225]) ?></p></section>
<?php endif; ?>

<section class="section card" id="creer-collecte">
  <h2><?= h($lang[178]) ?></h2>
  <?php if ($error): ?><div class="note-box"><?= h($error) ?></div><?php endif; ?>
  <form method="post">
    <input type="hidden" name="action" value="create_collecte">
    <div class="field"><label><?= h($lang[104]) ?></label><input name="lieu" required></div>
    <div class="grid grid-2">
      <div class="field"><label><?= h($lang[14]) ?></label><input name="date_collecte" type="date" required></div>
      <div class="field"><label><?= h($lang[103]) ?></label><input name="heure_collecte" type="time"></div>
    </div>
    <div class="field"><label><?= h($lang[341]) ?></label><input name="partenaire"></div>
    <div class="field"><label><?= h($lang[22]) ?></label><textarea name="description"></textarea></div>
    <div class="actions"><button type="submit"><?= h($lang[83]) ?></button></div>
  </form>
</section>

<section class="section card" id="creer-distribution">
  <h2><?= h($lang[179]) ?></h2>
  <form method="post">
    <input type="hidden" name="action" value="create_distribution">
    <div class="field"><label><?= h($lang[104]) ?></label><input name="lieu" required></div>
    <div class="grid grid-2">
      <div class="field"><label><?= h($lang[14]) ?></label><input name="date_distribution" type="date" required></div>
      <div class="field"><label><?= h($lang[103]) ?></label><input name="heure_distribution" type="time"></div>
    </div>
    <div class="field"><label><?= h($lang[246]) ?></label><input name="quota_par_adherent" type="number" min="0" value="1"></div>
    <div class="actions"><button type="submit"><?= h($lang[83]) ?></button></div>
  </form>
</section>

<?php include __DIR__ . '/../inc/layout_bottom.php'; ?>
