<?php
require_once __DIR__ . '/../inc/config.php';
require_once __DIR__ . '/../inc/auth.php';
require_once __DIR__ . '/../inc/helpers.php';
require_once __DIR__ . '/../inc/changeLanguage.php';

$user = current_user();
$lang = $LANGUAGE_CONTENTS;

// The page we should send the visitor back to once they've picked a
// language. Only a bare frontoffice filename, or a backoffice one prefixed
// with ../backoffice/ (see layout_top.php's $langRedirect), is accepted —
// anything else is rejected so this can never be turned into an open redirect.
function safe_lang_redirect($raw) {
    $raw = (string) ($raw ?? '');
    if (preg_match('#^\.\./backoffice/([A-Za-z0-9_-]+\.php)$#', $raw, $m)) {
        return '../backoffice/' . $m[1];
    }
    $file = basename($raw);
    return $file !== '' ? $file : 'index.php';
}
$redirect = safe_lang_redirect($_GET['redirect'] ?? 'index.php');

$pageTitle = $lang[$LANGUAGES_TITLE] ?? $lang[5];
$active = '';
include __DIR__ . '/../inc/layout_top.php';
?>

<section class="section">
  <h1><?= h($lang[$LANGUAGES_HEADER] ?? $lang[6]) ?></h1>
  <p class="muted"><?= h($lang[272]) ?></p>

  <div class="table-wrap">
    <table>
      <tbody>
      <?php foreach ($AVAILABLE_LANGUAGES as $code): ?>
      <tr class="clickable<?= $code === $LOADED_LANGUAGE ? ' active-lang' : '' ?>">
        <td>
          <a class="lang-row-link" href="<?= h($redirect) ?>?lang=<?= h($code) ?>">
            <img src="../lang/<?= h($code) ?>.svg" alt="<?= h($code) ?> flag" height="32" width="44">
            <span><?= h($code) ?></span>
            <?php if ($code === $LOADED_LANGUAGE): ?><span class="badge badge-success"><?= h($lang[273]) ?></span><?php endif; ?>
          </a>
        </td>
      </tr>
      <?php endforeach; ?>
      <?php if (empty($AVAILABLE_LANGUAGES)): ?>
      <tr><td class="muted"><?= h($lang[274]) ?></td></tr>
      <?php endif; ?>
      </tbody>
    </table>
  </div>
</section>

<?php include __DIR__ . '/../inc/layout_bottom.php'; ?>
