<?php
require_once __DIR__ . '/../inc/config.php';
require_once __DIR__ . '/../inc/auth.php';
require_once __DIR__ . '/../inc/helpers.php';
require_once __DIR__ . '/../inc/changeLanguage.php';
$lang = $LANGUAGE_CONTENTS;
require_once __DIR__ . '/../inc/xlsx.php';
require_once __DIR__ . '/../inc/ics.php';

function collect_benevole_events($compteId, $token) {
    $events = [];

    $collectesRes = api_get('/api/v1/collectes', ['from' => 0, 'size' => 200], $token);
    foreach (($collectesRes['data']['collectes'] ?? []) as $c) {
        $detail = api_get('/api/v1/collectes', ['id' => $c['collecte_id']], $token)['data'] ?? null;
        foreach (($detail['benevoles_affectes'] ?? []) as $b) {
            if ((int) $b['benevole_id'] === (int) $compteId) {
                $heure = $c['heure_collecte'] ?: '09:00:00';
                $events[] = [
                    'type_label' => $lang[380], 'titre' => $c['lieu'], 'date' => $c['date_collecte'],
                    'heure' => $heure, 'heure_fin' => ics_add_hours($heure, 2), 'lieu' => $c['lieu'],
                    'role' => $b['role_mission'], 'href' => 'collecte.php?id=' . $c['collecte_id'],
                    'uid' => 'collecte-' . $c['collecte_id'],
                ];
            }
        }
    }

    $distributionsRes = api_get('/api/v1/distributions', ['from' => 0, 'size' => 200], $token);
    foreach (($distributionsRes['data']['distributions'] ?? []) as $d) {
        $detail = api_get('/api/v1/distributions', ['id' => $d['distribution_id']], $token)['data'] ?? null;
        foreach (($detail['benevoles_affectes'] ?? []) as $b) {
            if ((int) $b['benevole_id'] === (int) $compteId) {
                $heure = $d['heure_distribution'] ?: '09:00:00';
                $events[] = [
                    'type_label' => $lang[381], 'titre' => $d['lieu'], 'date' => $d['date_distribution'],
                    'heure' => $heure, 'heure_fin' => ics_add_hours($heure, 2), 'lieu' => $d['lieu'],
                    'role' => $b['role_mission'], 'href' => 'distribution.php?id=' . $d['distribution_id'],
                    'uid' => 'distribution-' . $d['distribution_id'],
                ];
            }
        }
    }

    usort($events, fn($a, $b) => strcmp($a['date'] . $a['heure'], $b['date'] . $b['heure']));
    return $events;
}

$user = current_user();
$token = session_token();
$action = query_param('action');

if ($user && $action === 'export_xlsx') {
    $rows = [];
    foreach (collect_benevole_events($user['compte_id'], $token) as $e) {
        $rows[] = [$e['type_label'], $e['titre'], format_date_fr($e['date']), substr($e['heure'], 0, 5), $e['lieu'], $e['role']];
    }
    send_xlsx('mon-planning.xlsx', [$lang[106], $lang[374], $lang[14], $lang[103], $lang[104], $lang[105]], $rows);
}

if ($user && $action === 'export_ics') {
    $icsEvents = [];
    foreach (collect_benevole_events($user['compte_id'], $token) as $e) {
        $icsEvents[] = [
            'uid' => $e['uid'], 'start' => ics_datetime($e['date'], $e['heure']), 'end' => ics_datetime($e['date'], $e['heure_fin']),
            'title' => $e['type_label'] . ' - ' . $e['titre'], 'location' => $e['lieu'], 'description' => 'Role : ' . $e['role'],
        ];
    }
    send_ics('mon-planning.ics', $icsEvents);
}

$pageTitle = $lang[41];
$active = 'planning';
include __DIR__ . '/../inc/layout_top.php';

if (guard_login($user)) {
    $events = collect_benevole_events($user['compte_id'], $token);
    ?>
    <section class="section">
      <div class="section-head">
        <h1><?= h($lang[41]) ?></h1>
        <div class="actions">
          <a class="btn-secondary btn-sm" href="?action=export_xlsx"><?= h($lang[97]) ?></a>
          <a class="btn-secondary btn-sm" href="?action=export_ics"><?= h($lang[98]) ?></a>
        </div>
      </div>
      <p class="muted"><?= h($lang[253]) ?></p>
      <div class="table-wrap"><table>
        <thead><tr><th><?= h($lang[106]) ?></th><th><?= h($lang[104]) ?></th><th><?= h($lang[14]) ?></th><th><?= h($lang[103]) ?></th><th><?= h($lang[105]) ?></th><th></th></tr></thead>
        <tbody>
        <?php foreach ($events as $e): ?>
        <tr>
          <td><span class="badge"><?= h($e['type_label']) ?></span></td>
          <td><a href="<?= h($e['href']) ?>"><?= h($e['titre']) ?></a></td>
          <td><?= h(format_date_fr($e['date'])) ?></td>
          <td><?= h(substr($e['heure'], 0, 5)) ?></td>
          <td><?= h($e['role']) ?></td>
          <td><a class="btn-sm" href="<?= h(google_calendar_link($e['type_label'] . ' - ' . $e['titre'], ics_datetime($e['date'], $e['heure']), ics_datetime($e['date'], $e['heure_fin']), $e['lieu'], 'Role : ' . $e['role'])) ?>" target="_blank" rel="noopener"><?= h($lang[99]) ?></a></td>
        </tr>
        <?php endforeach; ?>
        <?php if (empty($events)): ?><tr><td colspan="6" class="muted"><?= h($lang[219]) ?></td></tr><?php endif; ?>
        </tbody>
      </table></div>
    </section>
    <?php
}

include __DIR__ . '/../inc/layout_bottom.php';
