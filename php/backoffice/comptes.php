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
        $res = api_put('/api/v1/admin/comptes', [
            'email' => $_POST['email'] ?? '', 'mot_de_passe' => $_POST['mot_de_passe'] ?? '',
            'nom' => $_POST['nom'] ?? '', 'prenom' => $_POST['prenom'] ?? '',
            'type_utilisateur' => $_POST['type_utilisateur'] ?? 'visiteur',
            'telephone' => $_POST['telephone'] ?? '', 'adresse' => $_POST['adresse'] ?? '',
            'code_postal' => $_POST['code_postal'] ?? '', 'ville' => $_POST['ville'] ?? '',
        ], $token);
        if ($res['status'] !== 201) {
            $error = $res['data']['error_description'] ?? $lang[292];
        }
    } elseif ($action === 'update') {
        api_patch('/api/v1/admin/comptes', [
            'compte_id' => $_POST['compte_id'] ?? '',
            'type_utilisateur' => $_POST['type_utilisateur'] ?? '',
            'actif' => isset($_POST['actif']) ? '1' : '0',
        ], $token);
    }
    if (!$error) {
        header('Location: comptes.php');
        exit;
    }
}

$q = query_param('q');
$type = query_param('type');
$pageTitle = $lang[44];
$active = 'comptes';
$navMode = 'back';
include __DIR__ . '/../inc/layout_top.php';

$res = api_get('/api/v1/admin/comptes', api_params(['from' => 0, 'size' => 200, 'q' => $q, 'type_utilisateur' => $type]), $token);
$comptes = $res['data']['comptes'] ?? [];
// "responsable" fait partie des types que seul ce back-office peut
// attribuer (cf. adminComptesHandler) et compte comme staff : il doit donc
// etre proposable au filtre comme a l'edition.
$types = ['visiteur' => $lang[72], 'adherent' => $lang[370], 'benevole' => $lang[74], 'commercant' => $lang[371], 'responsable' => $lang[457], 'administrateur' => $lang[372]];
?>
<section class="section">
  <div class="section-head"><h1><?= h($lang[44]) ?></h1><a class="btn-sm" href="#creer-compte"><?= h($lang[176]) ?></a></div>
  <form class="filters" method="get">
    <div class="field"><label for="q"><?= h($lang[325]) ?></label><input id="q" name="q" value="<?= h($q) ?>"></div>
    <div class="field"><label for="type"><?= h($lang[106]) ?></label>
      <select id="type" name="type">
        <option value=""><?= h($lang[111]) ?></option>
        <?php foreach ($types as $val => $label): ?><option value="<?= h($val) ?>" <?= $type === $val ? 'selected' : '' ?>><?= h($label) ?></option><?php endforeach; ?>
      </select>
    </div>
    <div class="actions"><button type="submit"><?= h($lang[80]) ?></button></div>
  </form>
  <div class="table-wrap"><table>
    <thead><tr><th><?= h($lang[64]) ?></th><th><?= h($lang[2]) ?></th><th><?= h($lang[106]) ?></th><th><?= h($lang[25]) ?></th><th></th></tr></thead>
    <tbody>
    <?php foreach ($comptes as $c): ?>
    <tr>
      <td><?= h($c['prenom'] . ' ' . $c['nom']) ?></td>
      <td><?= h($c['email']) ?></td>
      <td><span class="badge"><?= h(user_type_label($c['type_utilisateur'])) ?></span></td>
      <td><span class="badge <?= $c['actif'] ? 'badge-success' : 'badge-danger' ?>"><?= $c['actif'] ? $lang[368] : $lang[369] ?></span></td>
      <td><a class="btn-sm" href="#modifier-<?= (int) $c['compte_id'] ?>"><?= h($lang[82]) ?></a></td>
    </tr>
    <?php endforeach; ?>
    <?php if (empty($comptes)): ?><tr><td colspan="5" class="muted"><?= h($lang[214]) ?></td></tr><?php endif; ?>
    </tbody>
  </table></div>
</section>

<?php foreach ($comptes as $c): ?>
<section class="section card" id="modifier-<?= (int) $c['compte_id'] ?>">
  <h3><?= h($c['prenom'] . ' ' . $c['nom']) ?></h3>
  <form method="post">
    <input type="hidden" name="action" value="update">
    <input type="hidden" name="compte_id" value="<?= (int) $c['compte_id'] ?>">
    <div class="field"><label><?= h($lang[324]) ?></label>
      <select name="type_utilisateur">
        <?php foreach ($types as $val => $label): ?><option value="<?= h($val) ?>" <?= $c['type_utilisateur'] === $val ? 'selected' : '' ?>><?= h($label) ?></option><?php endforeach; ?>
      </select>
    </div>
    <div class="field"><label><input type="checkbox" name="actif" <?= $c['actif'] ? 'checked' : '' ?>> <?= h($lang[326]) ?></label></div>
    <div class="actions"><button type="submit"><?= h($lang[76]) ?></button></div>
  </form>
</section>
<?php endforeach; ?>

<section class="section card" id="creer-compte">
  <h2><?= h($lang[176]) ?></h2>
  <?php if ($error): ?><div class="note-box"><?= h($error) ?></div><?php endif; ?>
  <form method="post">
    <input type="hidden" name="action" value="create">
    <div class="grid grid-2">
      <div class="field"><label><?= h($lang[65]) ?></label><input name="prenom" required></div>
      <div class="field"><label><?= h($lang[64]) ?></label><input name="nom" required></div>
    </div>
    <div class="field"><label><?= h($lang[2]) ?></label><input name="email" type="email" required></div>
    <div class="field"><label><?= h($lang[62]) ?></label><input name="mot_de_passe" type="password" required></div>
    <div class="field"><label><?= h($lang[106]) ?></label>
      <select name="type_utilisateur"><?php foreach ($types as $val => $label): ?><option value="<?= h($val) ?>"><?= h($label) ?></option><?php endforeach; ?></select>
    </div>
    <div class="field"><label><?= h($lang[66]) ?></label><input name="telephone"></div>
    <div class="field"><label><?= h($lang[67]) ?></label><input name="adresse"></div>
    <div class="grid grid-2">
      <div class="field"><label><?= h($lang[68]) ?></label><input name="code_postal"></div>
      <div class="field"><label><?= h($lang[69]) ?></label><input name="ville"></div>
    </div>
    <div class="actions"><button type="submit"><?= h($lang[83]) ?></button></div>
  </form>
</section>

<?php include __DIR__ . '/../inc/layout_bottom.php'; ?>
