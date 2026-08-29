<?php
require_once __DIR__ . '/../inc/config.php';
require_once __DIR__ . '/../inc/auth.php';
require_once __DIR__ . '/../inc/helpers.php';
require_once __DIR__ . '/../inc/changeLanguage.php';
$lang = $LANGUAGE_CONTENTS;

$user = current_user();
if ($user) {
    header('Location: index.php');
    exit;
}

$loginError = null;
$registerError = null;

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    if (($_POST['action'] ?? '') === 'login') {
        $loginError = do_login($_POST['email'] ?? '', $_POST['mot_de_passe'] ?? '');
        if (!$loginError) {
            header('Location: index.php');
            exit;
        }
    } elseif (($_POST['action'] ?? '') === 'register') {
        $fields = [
            'email' => $_POST['email'] ?? '',
            'mot_de_passe' => $_POST['mot_de_passe'] ?? '',
            'mot_de_passe_confirmation' => $_POST['mot_de_passe_confirmation'] ?? '',
            'nom' => $_POST['nom'] ?? '',
            'prenom' => $_POST['prenom'] ?? '',
            'telephone' => $_POST['telephone'] ?? '',
            'adresse' => $_POST['adresse'] ?? '',
            'code_postal' => $_POST['code_postal'] ?? '',
            'ville' => $_POST['ville'] ?? '',
            'type_utilisateur' => $_POST['type_utilisateur'] ?? 'visiteur',
        ];
        $registerError = do_register($fields);
        if (!$registerError) {
            $loginError = do_login($fields['email'], $fields['mot_de_passe']);
            if (!$loginError) {
                header('Location: index.php');
                exit;
            }
        }
    }
}

$pageTitle = $lang[61];
$active = '';
include __DIR__ . '/../inc/layout_top.php';
?>

<section class="section grid grid-2">
  <div class="card">
    <h2><?= h($lang[59]) ?></h2>
    <?php if ($loginError): ?><div class="note-box"><?= h($loginError) ?></div><?php endif; ?>
    <form method="post">
      <input type="hidden" name="action" value="login">
      <div class="field"><label for="login-email"><?= h($lang[2]) ?></label><input id="login-email" name="email" type="email" required></div>
      <div class="field"><label for="login-mdp"><?= h($lang[62]) ?></label><input id="login-mdp" name="mot_de_passe" type="password" required></div>
      <div class="actions"><button type="submit"><?= h($lang[54]) ?></button></div>
    </form>
  </div>

  <div class="card">
    <h2><?= h($lang[60]) ?></h2>
    <?php if ($registerError): ?><div class="note-box"><?= h($registerError) ?></div><?php endif; ?>
    <form method="post">
      <input type="hidden" name="action" value="register">
      <div class="grid grid-2">
        <div class="field"><label for="reg-prenom"><?= h($lang[65]) ?></label><input id="reg-prenom" name="prenom" required></div>
        <div class="field"><label for="reg-nom"><?= h($lang[64]) ?></label><input id="reg-nom" name="nom" required></div>
      </div>
      <div class="field"><label for="reg-email"><?= h($lang[2]) ?></label><input id="reg-email" name="email" type="email" required></div>
      <div class="grid grid-2">
        <div class="field"><label for="reg-mdp"><?= h($lang[62]) ?></label><input id="reg-mdp" name="mot_de_passe" type="password" required></div>
        <div class="field"><label for="reg-mdp2"><?= h($lang[63]) ?></label><input id="reg-mdp2" name="mot_de_passe_confirmation" type="password" required></div>
      </div>
      <div class="field"><label for="reg-type"><?= h($lang[70]) ?></label>
        <select id="reg-type" name="type_utilisateur">
          <option value="visiteur"><?= h($lang[72]) ?></option>
          <option value="adherent"><?= h($lang[71]) ?></option>
          <option value="commercant"><?= h($lang[73]) ?></option>
        </select>
      </div>
      <div class="field"><label for="reg-tel"><?= h($lang[66]) ?></label><input id="reg-tel" name="telephone" placeholder="+33612345678"></div>
      <div class="field"><label for="reg-adresse"><?= h($lang[67]) ?></label><input id="reg-adresse" name="adresse"></div>
      <div class="grid grid-2">
        <div class="field"><label for="reg-cp"><?= h($lang[68]) ?></label><input id="reg-cp" name="code_postal" placeholder="75001"></div>
        <div class="field"><label for="reg-ville"><?= h($lang[69]) ?></label><input id="reg-ville" name="ville"></div>
      </div>
      <div class="actions"><button type="submit"><?= h($lang[1]) ?></button></div>
    </form>
  </div>
</section>

<?php include __DIR__ . '/../inc/layout_bottom.php'; ?>
