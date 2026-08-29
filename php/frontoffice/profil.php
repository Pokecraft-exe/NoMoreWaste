<?php
require_once __DIR__ . '/../inc/config.php';
require_once __DIR__ . '/../inc/auth.php';
require_once __DIR__ . '/../inc/helpers.php';
require_once __DIR__ . '/../inc/changeLanguage.php';
$lang = $LANGUAGE_CONTENTS;

$user = current_user();
$token = session_token();
$message = null;
$annonceError = null;

$categories = [
    'covoiturage' => $lang[118], 'reparation' => $lang[116], 'gardiennage' => $lang[117],
    'location' => $lang[115], 'vente' => $lang[113], 'don' => 'Don',
];
$etats = ['ouverte', 'en_cours', 'cloturee'];

if ($_SERVER['REQUEST_METHOD'] === 'POST' && $user) {
    $action = $_POST['action'] ?? '';
    if ($action === 'update') {
        $res = api_patch('/api/v1/profil', [
            'telephone' => $_POST['telephone'] ?? '',
            'adresse' => $_POST['adresse'] ?? '',
            'code_postal' => $_POST['code_postal'] ?? '',
            'ville' => $_POST['ville'] ?? '',
        ], $token);
        $message = $res['status'] === 200 ? $lang[258] : ($res['data']['error_description'] ?? $lang[293]);
        unset($_SESSION['profil']);
        $user = current_user();
    } elseif ($action === 'renew') {
        $res = api_put('/api/v1/adhesion/renouveler', [], $token);
        $message = $res['status'] === 201 ? $lang[384] . $res['data']['date_fin'] . '.' : ($res['data']['error_description'] ?? $lang[293]);
        unset($_SESSION['profil']);
        $user = current_user();
    } elseif ($action === 'update_annonce') {
        $res = api_patch('/api/v1/annonces?id=' . intval($_POST['annonce_echange_id'] ?? 0), [
            'titre' => $_POST['titre'] ?? '', 'description' => $_POST['description'] ?? '',
            'prix' => trim($_POST['prix'] ?? ''), 'statut' => $_POST['statut'] ?? '',
        ], $token);
        if ($res['status'] !== 200) {
            $annonceError = $res['data']['error_description'] ?? $lang[293];
        }
    } elseif ($action === 'delete_annonce') {
        api_delete('/api/v1/annonces', ['id' => intval($_POST['annonce_echange_id'] ?? 0)], $token);
    }
}

$pageTitle = $lang[40];
$active = 'profil';
include __DIR__ . '/../inc/layout_top.php';

if (guard_login($user)) {
    ?>
    <section class="section">
      <h1><?= h($lang[40]) ?></h1>
      <?php if ($message): ?><div class="note-box"><?= h($message) ?></div><?php endif; ?>
      <div class="card">
        <p><strong><?= h(full_name($user)) ?></strong> &middot; <?= h(user_type_label($user['type_utilisateur'])) ?></p>
        <p class="muted"><?= h($user['email']) ?></p>
      </div>

      <form method="post" class="card">
        <input type="hidden" name="action" value="update">
        <div class="field"><label for="telephone"><?= h($lang[66]) ?></label><input id="telephone" name="telephone" value="<?= h($user['telephone']) ?>"></div>
        <div class="field"><label for="adresse"><?= h($lang[67]) ?></label><input id="adresse" name="adresse" value="<?= h($user['adresse']) ?>"></div>
        <div class="grid grid-2">
          <div class="field"><label for="code_postal"><?= h($lang[68]) ?></label><input id="code_postal" name="code_postal" value="<?= h($user['code_postal']) ?>"></div>
          <div class="field"><label for="ville"><?= h($lang[69]) ?></label><input id="ville" name="ville" value="<?= h($user['ville']) ?>"></div>
        </div>
        <div class="actions"><button type="submit"><?= h($lang[76]) ?></button></div>
      </form>

      <?php if (is_adherent($user)): ?>
      <div class="card">
        <h3><?= h($lang[254]) ?></h3>
        <p><?= h($lang[255]) ?> <span class="badge"><?= h($user['adhesion_statut'] ?? 'aucune') ?></span></p>
        <?php if (!empty($user['adhesion_date_fin'])): ?><p><?= h($lang[256]) ?> <?= h($user['adhesion_date_fin']) ?></p><?php endif; ?>
        <form method="post"><input type="hidden" name="action" value="renew"><button type="submit"><?= h($lang[257]) ?></button></form>
      </div>
      <?php else: ?>
      <p class="muted"><?= h($lang[259]) ?></p>
      <?php endif; ?>
    </section>

    <?php
    $mesForumsRes = api_get('/api/v1/forum', ['from' => 0, 'size' => 50, 'auteur_id' => $user['compte_id']], $token);
    $mesForums = $mesForumsRes['data']['threads'] ?? [];
    $mesAnnoncesRes = api_get('/api/v1/annonces', ['from' => 0, 'size' => 50, 'auteur_id' => $user['compte_id']], $token);
    $mesAnnonces = $mesAnnoncesRes['data']['annonces'] ?? [];
    ?>

    <section class="section">
      <div class="section-head"><h2><?= h($lang[385]) ?></h2><a class="btn-secondary btn-sm" href="forums.php"><?= h($lang[95]) ?></a></div>
      <div class="results grid grid-auto">
        <?php foreach ($mesForums as $f): ?>
        <a class="card-link" href="forum.php?id=<?= (int) $f['forum_thread_id'] ?>">
          <div class="card-title"><?= h($f['titre']) ?></div>
          <div class="card-meta"><span><?= (int) $f['vues'] ?> <?= h($lang[298]) ?></span><span><?= h(time_ago($f['date_creation'])) ?></span></div>
        </a>
        <?php endforeach; ?>
        <?php if (empty($mesForums)): ?><div class="empty"><?= h($lang[387]) ?></div><?php endif; ?>
      </div>
    </section>

    <section class="section">
      <div class="section-head"><h2><?= h($lang[386]) ?></h2><a class="btn-secondary btn-sm" href="annonces.php"><?= h($lang[85]) ?></a></div>
      <?php if ($annonceError): ?><div class="note-box"><?= h($annonceError) ?></div><?php endif; ?>
      <div class="table-wrap"><table>
        <thead><tr><th><?= h($lang[20]) ?></th><th><?= h($lang[19]) ?></th><th><?= h($lang[25]) ?></th><th><?= h($lang[21]) ?></th><th></th></tr></thead>
        <tbody>
        <?php foreach ($mesAnnonces as $a): ?>
        <tr>
          <td><?= h($a['titre']) ?></td>
          <td><span class="badge"><?= h($categories[$a['categorie']] ?? $a['categorie']) ?></span></td>
          <td><span class="badge"><?= h($a['statut']) ?></span></td>
          <td><?= $a['prix'] !== null ? h($a['prix']) . ' €' : $lang[112] ?></td>
          <td><a class="btn-sm" href="#annonce-<?= (int) $a['annonce_echange_id'] ?>"><?= h($lang[82]) ?></a></td>
        </tr>
        <?php endforeach; ?>
        <?php if (empty($mesAnnonces)): ?><tr><td colspan="5" class="muted"><?= h($lang[388]) ?></td></tr><?php endif; ?>
        </tbody>
      </table></div>

      <?php foreach ($mesAnnonces as $a): ?>
      <div class="card" id="annonce-<?= (int) $a['annonce_echange_id'] ?>" style="margin-top:12px;">
        <h3><?= h($a['titre']) ?></h3>
        <form method="post">
          <input type="hidden" name="action" value="update_annonce">
          <input type="hidden" name="annonce_echange_id" value="<?= (int) $a['annonce_echange_id'] ?>">
          <div class="field"><label><?= h($lang[20]) ?></label><input name="titre" value="<?= h($a['titre']) ?>" required></div>
          <div class="field"><label><?= h($lang[22]) ?></label><textarea name="description" required><?= h($a['description']) ?></textarea></div>
          <div class="grid grid-2">
            <div class="field"><label><?= h($lang[21]) ?></label><input name="prix" type="number" min="0" value="<?= h($a['prix']) ?>"></div>
            <div class="field"><label><?= h($lang[25]) ?></label>
              <select name="statut"><?php foreach ($etats as $e): ?><option value="<?= h($e) ?>" <?= $a['statut'] === $e ? 'selected' : '' ?>><?= h($e) ?></option><?php endforeach; ?></select>
            </div>
          </div>
          <div class="actions"><button type="submit"><?= h($lang[76]) ?></button></div>
        </form>
        <form method="post" onsubmit="return confirm('<?= h($lang[284]) ?>');" style="margin-top:8px;">
          <input type="hidden" name="action" value="delete_annonce">
          <input type="hidden" name="annonce_echange_id" value="<?= (int) $a['annonce_echange_id'] ?>">
          <button type="submit" class="btn-danger"><?= h($lang[81]) ?></button>
        </form>
      </div>
      <?php endforeach; ?>
    </section>
    <?php
}

include __DIR__ . '/../inc/layout_bottom.php';
