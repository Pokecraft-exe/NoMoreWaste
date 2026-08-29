<?php
require_once __DIR__ . '/../inc/config.php';
require_once __DIR__ . '/../inc/auth.php';
require_once __DIR__ . '/../inc/helpers.php';
require_once __DIR__ . '/../inc/changeLanguage.php';
$lang = $LANGUAGE_CONTENTS;

$user = current_user();
$id = intval($_GET['id'] ?? 0);
$token = session_token();

$categories = [
    'covoiturage' => $lang[118], 'reparation' => $lang[116], 'gardiennage' => $lang[117],
    'location' => $lang[115], 'vente' => $lang[113], 'don' => 'Don',
];

if ($_SERVER['REQUEST_METHOD'] === 'POST' && $user) {
    $action = $_POST['action'] ?? '';
    if ($action === 'message') {
        api_put('/api/v1/annonces/messages', [
            'annonce_echange_id' => $id,
            'message' => $_POST['message'] ?? '',
            'prix_propose' => trim($_POST['prix_propose'] ?? ''),
        ], $token);
    } elseif ($action === 'report_message') {
        api_put('/api/v1/signalements', [
            'type_signalement' => 'annonce_message',
            'annonce_message_id' => intval($_POST['message_id'] ?? 0),
            'motif' => $_POST['motif'] ?? '',
        ], $token);
    }
    header('Location: annonce.php?id=' . $id);
    exit;
}

$res = api_get('/api/v1/annonces', ['id' => $id], $token);
$annonce = $res['data'] ?? null;

$pageTitle = $annonce ? $annonce['titre'] : $lang[382];
$active = 'annonces';
include __DIR__ . '/../inc/layout_top.php';

if (!$annonce) {
    echo '<div class="empty">Cette annonce n\'existe pas.</div>';
    include __DIR__ . '/../inc/layout_bottom.php';
    exit;
}

$isOwner = $user && $user['compte_id'] == $annonce['auteur_id'];
$msgRes = $user ? api_get('/api/v1/annonces/messages', ['annonce_echange_id' => $id], $token) : ['status' => 401];
$canSeeMessages = $msgRes['status'] === 200;
$messages = $msgRes['data']['messages'] ?? [];
?>

<section class="section">
  <h1><?= h($annonce['titre']) ?></h1>
  <div class="card-meta">
    <span class="badge"><?= h($categories[$annonce['categorie']] ?? $annonce['categorie']) ?></span>
    <span><?= $annonce['prix'] !== null ? h($annonce['prix']) . ' €' : $lang[112] ?></span>
    <span><?= h($lang[297]) ?> <?= h($annonce['auteur']) ?></span>
  </div>
  <p><?= nl2br(h($annonce['description'])) ?></p>
  <?php if ($isOwner): ?><p class="muted"><?= h($lang[194]) ?></p><?php endif; ?>

  <h4><?= h($lang[357]) ?></h4>
  <?php if (!$user): ?>
    <div class="note-box"><?= h($lang[191]) ?></div>
  <?php elseif (!$canSeeMessages): ?>
    <div class="note-box"><?= h($lang[192]) ?></div>
  <?php elseif (empty($messages)): ?>
    <div class="empty"><?= h($lang[193]) ?></div>
  <?php else: ?>
    <?php foreach ($messages as $m): ?>
    <div class="thread-message">
      <span class="avatar"><?= h(mb_strtoupper(mb_substr($m['expediteur'], 0, 1))) ?></span>
      <div class="bubble">
        <div class="msg-meta">
          <strong><?= h($m['expediteur']) ?></strong><span><?= h(time_ago($m['date_envoi'])) ?></span>
          <?php if ($user['compte_id'] != $m['expediteur_id']): ?>
          <a class="report-link" style="margin-left:auto;" href="#signaler-<?= (int) $m['message_annonce_echange_id'] ?>"><?= h($lang[187]) ?></a>
          <?php endif; ?>
        </div>
        <div><?= h($m['message']) ?></div>
        <?php if ($m['prix_propose'] !== null): ?><div class="badge badge-warning" style="margin-top:6px;"><?= h($lang[186]) ?> <?= h($m['prix_propose']) ?> €</div><?php endif; ?>
      </div>
    </div>
    <?php if ($user['compte_id'] != $m['expediteur_id']): ?>
    <div class="card" id="signaler-<?= (int) $m['message_annonce_echange_id'] ?>" style="margin:8px 0 16px;">
      <form method="post">
        <input type="hidden" name="action" value="report_message">
        <input type="hidden" name="message_id" value="<?= (int) $m['message_annonce_echange_id'] ?>">
        <div class="field"><label><?= h($lang[189]) ?></label><textarea name="motif" required></textarea></div>
        <div class="actions"><button type="submit" class="btn-danger"><?= h($lang[190]) ?></button></div>
      </form>
    </div>
    <?php endif; ?>
    <?php endforeach; ?>
  <?php endif; ?>

  <?php if ($user): ?>
  <form method="post">
    <input type="hidden" name="action" value="message">
    <div class="grid grid-2">
      <div class="field"><label for="msg"><?= h($lang[35]) ?></label><textarea id="msg" name="message" required></textarea></div>
      <div class="field"><label for="prix-p"><?= h($lang[185]) ?></label><input id="prix-p" name="prix_propose" type="number" min="0"></div>
    </div>
    <div class="actions"><button type="submit"><?= h($lang[77]) ?></button></div>
  </form>
  <?php endif; ?>
</section>

<?php include __DIR__ . '/../inc/layout_bottom.php'; ?>
