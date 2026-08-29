<?php
require_once __DIR__ . '/../inc/config.php';
require_once __DIR__ . '/../inc/auth.php';
require_once __DIR__ . '/../inc/helpers.php';
require_once __DIR__ . '/../inc/changeLanguage.php';
$lang = $LANGUAGE_CONTENTS;

$user = current_user();
$sent = false;
$error = null;

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $fields = [
        'sujet' => $_POST['sujet'] ?? '',
        'message' => $_POST['message'] ?? '',
    ];
    if (!$user) {
        $fields['contact_nom'] = $_POST['contact_nom'] ?? '';
        $fields['contact_email'] = $_POST['contact_email'] ?? '';
    }
    $res = api_put('/api/v1/tickets', $fields, session_token());
    if ($res['status'] === 201) {
        $sent = true;
    } else {
        $error = $res['data']['error_description'] ?? $lang[291];
    }
}

$pageTitle = $lang[275];
$active = '';
include __DIR__ . '/../inc/layout_top.php';
?>

<section class="section">
  <h1><?= h($lang[275]) ?></h1>
  <p class="muted"><?= h($lang[277]) ?></p>

  <?php if ($sent): ?>
    <div class="note-box"><?= h($lang[276]) ?></div>
  <?php else: ?>
    <?php if ($error): ?><div class="note-box"><?= h($error) ?></div><?php endif; ?>
    <form method="post" class="card">
      <?php if (!$user): ?>
      <div class="grid grid-2">
        <div class="field"><label for="contact-nom"><?= h($lang[144]) ?></label><input id="contact-nom" name="contact_nom" required></div>
        <div class="field"><label for="contact-email"><?= h($lang[145]) ?></label><input id="contact-email" name="contact_email" type="email" required></div>
      </div>
      <?php endif; ?>
      <div class="field"><label for="sujet"><?= h($lang[12]) ?></label><input id="sujet" name="sujet" required></div>
      <div class="field"><label for="message"><?= h($lang[35]) ?></label><textarea id="message" name="message" required></textarea></div>
      <div class="actions"><button type="submit"><?= h($lang[77]) ?></button></div>
    </form>
  <?php endif; ?>
</section>

<?php include __DIR__ . '/../inc/layout_bottom.php'; ?>
