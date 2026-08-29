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
        $res = api_put('/api/v1/ressources-cuisine', [
            'titre' => $_POST['titre'] ?? '', 'ingredients' => $_POST['ingredients'] ?? '',
            'outils' => $_POST['outils'] ?? '', 'contenu' => $_POST['contenu'] ?? '', 'video' => $_POST['video'] ?? '',
        ], $token);
        if ($res['status'] !== 201) {
            $error = $res['data']['error_description'] ?? $lang[292];
        }
    } elseif ($action === 'update') {
        api_patch('/api/v1/ressources-cuisine?id=' . intval($_POST['id'] ?? 0), [
            'titre' => $_POST['titre'] ?? '', 'ingredients' => $_POST['ingredients'] ?? '',
            'outils' => $_POST['outils'] ?? '', 'contenu' => $_POST['contenu'] ?? '', 'video' => $_POST['video'] ?? '',
        ], $token);
    } elseif ($action === 'delete') {
        api_delete('/api/v1/ressources-cuisine', ['id' => intval($_POST['id'] ?? 0)], $token);
    }
    if (!$error) {
        header('Location: recettes.php');
        exit;
    }
}

$q = query_param('q');
$pageTitle = $lang[133];
$active = 'recettes';
$navMode = 'back';
include __DIR__ . '/../inc/layout_top.php';

$res = api_get('/api/v1/ressources-cuisine', api_params(['from' => 0, 'size' => 100, 'q' => $q]), $token);
$recettes = $res['data']['recettes'] ?? [];
?>
<section class="section">
  <div class="section-head"><h1><?= h($lang[133]) ?></h1><a class="btn-sm" href="#creer"><?= h($lang[180]) ?></a></div>
  <form class="filters" method="get">
    <div class="field"><label><?= h($lang[137]) ?></label><input name="q" value="<?= h($q) ?>"></div>
    <div class="actions"><button type="submit"><?= h($lang[79]) ?></button></div>
  </form>
  <div class="table-wrap"><table>
    <thead><tr><th><?= h($lang[20]) ?></th><th><?= h($lang[128]) ?></th><th></th></tr></thead>
    <tbody>
    <?php foreach ($recettes as $r): ?>
    <tr>
      <td><?= h($r['titre']) ?></td>
      <td class="muted"><?= h(implode(', ', $r['ingredients'] ?? [])) ?></td>
      <td><a class="btn-sm" href="#detail-<?= (int) $r['ressource_cuisine_id'] ?>"><?= h($lang[82]) ?></a></td>
    </tr>
    <?php endforeach; ?>
    <?php if (empty($recettes)): ?><tr><td colspan="3" class="muted"><?= h($lang[215]) ?></td></tr><?php endif; ?>
    </tbody>
  </table></div>
</section>

<?php foreach ($recettes as $r):
  $detail = api_get('/api/v1/ressources-cuisine', ['id' => $r['ressource_cuisine_id']], $token)['data'] ?? $r;
  ?>
<section class="section card" id="detail-<?= (int) $r['ressource_cuisine_id'] ?>">
  <h3><?= h($r['titre']) ?></h3>
  <form method="post">
    <input type="hidden" name="action" value="update">
    <input type="hidden" name="id" value="<?= (int) $r['ressource_cuisine_id'] ?>">
    <div class="field"><label><?= h($lang[20]) ?></label><input name="titre" value="<?= h($detail['titre']) ?>" required></div>
    <div class="field"><label><?= h($lang[136]) ?></label><input name="ingredients" value="<?= h(implode(', ', $detail['ingredients'] ?? [])) ?>"></div>
    <div class="field"><label><?= h($lang[135]) ?></label><input name="outils" value="<?= h(implode(', ', $detail['outils'] ?? [])) ?>"></div>
    <div class="field"><label><?= h($lang[134]) ?></label><textarea name="contenu" required><?= h($detail['contenu'] ?? '') ?></textarea></div>
    <div class="field"><label><?= h($lang[132]) ?></label><input name="video" value="<?= h($detail['video'] ?? '') ?>"></div>
    <div class="actions"><button type="submit"><?= h($lang[76]) ?></button></div>
  </form>
  <form method="post" onsubmit="return confirm('<?= h($lang[284]) ?>');">
    <input type="hidden" name="action" value="delete"><input type="hidden" name="id" value="<?= (int) $r['ressource_cuisine_id'] ?>">
    <button type="submit" class="btn-danger"><?= h($lang[81]) ?></button>
  </form>
</section>
<?php endforeach; ?>

<section class="section card" id="creer">
  <h2><?= h($lang[181]) ?></h2>
  <?php if ($error): ?><div class="note-box"><?= h($error) ?></div><?php endif; ?>
  <form method="post">
    <input type="hidden" name="action" value="create">
    <div class="field"><label><?= h($lang[20]) ?></label><input name="titre" required></div>
    <div class="field"><label><?= h($lang[136]) ?></label><input name="ingredients"></div>
    <div class="field"><label><?= h($lang[135]) ?></label><input name="outils"></div>
    <div class="field"><label><?= h($lang[134]) ?></label><textarea name="contenu" required></textarea></div>
    <div class="field"><label><?= h($lang[132]) ?></label><input name="video"></div>
    <div class="actions"><button type="submit"><?= h($lang[83]) ?></button></div>
  </form>
</section>

<?php include __DIR__ . '/../inc/layout_bottom.php'; ?>
