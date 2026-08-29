<?php
require_once __DIR__ . '/../inc/config.php';
require_once __DIR__ . '/../inc/auth.php';
require_once __DIR__ . '/../inc/helpers.php';
require_once __DIR__ . '/../inc/changeLanguage.php';
$lang = $LANGUAGE_CONTENTS;

$user = current_user();
$token = session_token();
require_staff_or_404($user);

$pageTitle = $lang[43];
$active = 'dashboard';
$navMode = 'back';
include __DIR__ . '/../inc/layout_top.php';

$signalementsOuverts = api_get('/api/v1/signalements', ['from' => 0, 'size' => 5, 'statut' => 'ouvert'], $token);
$ticketsOuverts = api_get('/api/v1/tickets', ['from' => 0, 'size' => 5, 'statut' => 'ouvert'], $token);
$comptesRes = api_get('/api/v1/admin/comptes', ['from' => 0, 'size' => 500], $token);
$comptes = $comptesRes['data']['comptes'] ?? [];
$candRecues = api_get('/api/v1/benevoles/candidatures', ['from' => 0, 'size' => 100, 'statut' => 'recue'], $token);
$candEtude = api_get('/api/v1/benevoles/candidatures', ['from' => 0, 'size' => 100, 'statut' => 'en_etude'], $token);
$collectesRes = api_get('/api/v1/collectes', ['from' => 0, 'size' => 5], $token);
$distributionsRes = api_get('/api/v1/distributions', ['from' => 0, 'size' => 5], $token);
$adhesionsExpirantRes = api_get('/api/v1/adherents', ['from' => 0, 'size' => 5, 'expirant_sous' => 30], $token);

$signalementsList = $signalementsOuverts['data']['signalements'] ?? [];
$ticketsList = $ticketsOuverts['data']['tickets'] ?? [];
$signalementsCount = count($signalementsList);
$ticketsCount = count($ticketsList);
$actifs = count(array_filter($comptes, fn($c) => $c['actif']));
$typeCounts = [];
foreach ($comptes as $c) {
    $typeCounts[$c['type_utilisateur']] = ($typeCounts[$c['type_utilisateur']] ?? 0) + 1;
}
$candidaturesList = array_merge($candRecues['data']['candidatures'] ?? [], $candEtude['data']['candidatures'] ?? []);
$candidaturesEnAttente = count($candidaturesList);
$candidaturesList = array_slice($candidaturesList, 0, 5);
$adhesionsExpirant = $adhesionsExpirantRes['data']['adherents'] ?? [];
$signalementTypeLabels = ['forum' => $lang[375], 'forum_message' => $lang[376], 'annonce_message' => $lang[377]];
?>
<section class="section">
  <h1><?= h($lang[43]) ?></h1>
  <p class="muted"><?= h($lang[305]) ?></p>
</section>

<?php if ($signalementsCount > 0 || $ticketsCount > 0): ?>
<section class="section">
  <div class="grid grid-2">
    <?php if ($signalementsCount > 0): ?>
    <a class="stat-tile clickable stat-danger" href="signalements.php">
      <div class="stat-value"><?= $signalementsCount ?></div><div class="stat-label"><?= h($lang[309]) ?></div>
    </a>
    <?php endif; ?>
    <?php if ($ticketsCount > 0): ?>
    <a class="stat-tile clickable stat-warning" href="tickets.php">
      <div class="stat-value"><?= $ticketsCount ?></div><div class="stat-label"><?= h($lang[170]) ?></div>
    </a>
    <?php endif; ?>
  </div>
</section>
<?php endif; ?>

<section class="section">
  <div class="stat-grid">
    <div class="stat-tile"><div class="stat-value"><?= count($comptes) ?></div><div class="stat-label"><?= h($lang[307]) ?></div></div>
    <div class="stat-tile"><div class="stat-value"><?= $actifs ?></div><div class="stat-label"><?= h($lang[306]) ?></div></div>
    <div class="stat-tile"><div class="stat-value"><?= $candidaturesEnAttente ?></div><div class="stat-label"><?= h($lang[308]) ?></div></div>
  </div>
</section>

<section class="section grid grid-2">
  <div>
    <div class="section-head"><h2><?= h($lang[310]) ?></h2><a class="btn-secondary btn-sm" href="collectes.php"><?= h($lang[85]) ?></a></div>
    <div class="table-wrap"><table>
      <thead><tr><th><?= h($lang[104]) ?></th><th><?= h($lang[14]) ?></th><th><?= h($lang[25]) ?></th></tr></thead>
      <tbody>
      <?php foreach (($collectesRes['data']['collectes'] ?? []) as $c): ?>
      <tr class="clickable" onclick="location.href='collectes.php?id=<?= (int) $c['collecte_id'] ?>'"><td><?= h($c['lieu']) ?></td><td><?= h($c['date_collecte']) ?></td><td><span class="badge"><?= h($c['statut']) ?></span></td></tr>
      <?php endforeach; ?>
      <?php if (empty($collectesRes['data']['collectes'])): ?><tr><td colspan="3" class="muted"><?= h($lang[223]) ?></td></tr><?php endif; ?>
      </tbody>
    </table></div>
  </div>
  <div>
    <div class="section-head"><h2><?= h($lang[311]) ?></h2><a class="btn-secondary btn-sm" href="distributions.php"><?= h($lang[85]) ?></a></div>
    <div class="table-wrap"><table>
      <thead><tr><th><?= h($lang[104]) ?></th><th><?= h($lang[14]) ?></th><th><?= h($lang[25]) ?></th></tr></thead>
      <tbody>
      <?php foreach (($distributionsRes['data']['distributions'] ?? []) as $d): ?>
      <tr class="clickable" onclick="location.href='distributions.php?id=<?= (int) $d['distribution_id'] ?>'"><td><?= h($d['lieu']) ?></td><td><?= h($d['date_distribution']) ?></td><td><span class="badge"><?= h($d['statut']) ?></span></td></tr>
      <?php endforeach; ?>
      <?php if (empty($distributionsRes['data']['distributions'])): ?><tr><td colspan="3" class="muted"><?= h($lang[224]) ?></td></tr><?php endif; ?>
      </tbody>
    </table></div>
  </div>
</section>

<section class="section grid grid-2">
  <div>
    <div class="section-head"><h2><?= h($lang[308]) ?></h2><a class="btn-secondary btn-sm" href="candidatures.php"><?= h($lang[85]) ?></a></div>
    <div class="table-wrap"><table>
      <thead><tr><th><?= h($lang[149]) ?></th><th><?= h($lang[25]) ?></th></tr></thead>
      <tbody>
      <?php foreach ($candidaturesList as $c): ?>
      <tr class="clickable" onclick="location.href='candidatures.php#decider-<?= (int) $c['candidature_benevole_id'] ?>'">
        <td><?= h($c['prenom'] . ' ' . $c['nom']) ?></td>
        <td><span class="badge badge-warning"><?= h($c['statut']) ?></span></td>
      </tr>
      <?php endforeach; ?>
      <?php if (empty($candidaturesList)): ?><tr><td colspan="2" class="muted"><?= h($lang[213]) ?></td></tr><?php endif; ?>
      </tbody>
    </table></div>
  </div>
  <div>
    <div class="section-head"><h2><?= h($lang[394]) ?></h2><a class="btn-secondary btn-sm" href="adherents.php"><?= h($lang[85]) ?></a></div>
    <div class="table-wrap"><table>
      <thead><tr><th><?= h($lang[370]) ?></th><th><?= h($lang[395]) ?></th></tr></thead>
      <tbody>
      <?php foreach ($adhesionsExpirant as $a): ?>
      <tr class="clickable" onclick="location.href='adherents.php?id=<?= (int) $a['adherent_id'] ?>#detail-<?= (int) $a['adherent_id'] ?>'"><td><?= h($a['prenom'] . ' ' . $a['nom']) ?></td><td><?= h($a['adhesion_date_fin'] ?? '-') ?></td></tr>
      <?php endforeach; ?>
      <?php if (empty($adhesionsExpirant)): ?><tr><td colspan="2" class="muted"><?= h($lang[214]) ?></td></tr><?php endif; ?>
      </tbody>
    </table></div>
  </div>
</section>

<section class="section grid grid-2">
  <div>
    <div class="section-head"><h2><?= h($lang[309]) ?></h2><a class="btn-secondary btn-sm" href="signalements.php"><?= h($lang[85]) ?></a></div>
    <div class="table-wrap"><table>
      <thead><tr><th><?= h($lang[106]) ?></th><th><?= h($lang[155]) ?></th><th><?= h($lang[154]) ?></th></tr></thead>
      <tbody>
      <?php foreach ($signalementsList as $s): ?>
      <tr class="clickable" onclick="location.href='signalements.php?id=<?= (int) $s['signalement_id'] ?>'">
        <td><span class="badge"><?= h($signalementTypeLabels[$s['type_signalement']] ?? $s['type_signalement']) ?></span></td>
        <td><?= h($s['signale_par_nom']) ?></td>
        <td><?= h(time_ago($s['date_signalement'])) ?></td>
      </tr>
      <?php endforeach; ?>
      <?php if (empty($signalementsList)): ?><tr><td colspan="3" class="muted"><?= h($lang[216]) ?></td></tr><?php endif; ?>
      </tbody>
    </table></div>
  </div>
  <div>
    <div class="section-head"><h2><?= h($lang[170]) ?></h2><a class="btn-secondary btn-sm" href="tickets.php"><?= h($lang[85]) ?></a></div>
    <div class="table-wrap"><table>
      <thead><tr><th><?= h($lang[12]) ?></th><th><?= h($lang[171]) ?></th><th><?= h($lang[154]) ?></th></tr></thead>
      <tbody>
      <?php foreach ($ticketsList as $t): ?>
      <tr class="clickable" onclick="location.href='tickets.php?id=<?= (int) $t['ticket_id'] ?>'">
        <td><?= h($t['sujet']) ?></td>
        <td><?= h($t['auteur_id'] ? 'compte #' . $t['auteur_id'] : ($t['contact_nom'] . ' (' . $t['contact_email'] . ')')) ?></td>
        <td><?= h(time_ago($t['date_creation'])) ?></td>
      </tr>
      <?php endforeach; ?>
      <?php if (empty($ticketsList)): ?><tr><td colspan="3" class="muted"><?= h($lang[217]) ?></td></tr><?php endif; ?>
      </tbody>
    </table></div>
  </div>
</section>

<section class="section">
  <div class="section-head"><h2><?= h($lang[312]) ?></h2><a class="btn-secondary btn-sm" href="comptes.php"><?= h($lang[85]) ?></a></div>
  <div class="table-wrap"><table>
    <thead><tr><th><?= h($lang[106]) ?></th><th><?= h($lang[313]) ?></th></tr></thead>
    <tbody>
    <?php foreach ($typeCounts as $type => $n): ?>
    <tr><td><?= h(user_type_label($type)) ?></td><td><?= $n ?></td></tr>
    <?php endforeach; ?>
    </tbody>
  </table></div>
</section>

<?php include __DIR__ . '/../inc/layout_bottom.php'; ?>
