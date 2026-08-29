<?php
require_once __DIR__ . '/../inc/changeLanguage.php';
$lang = $LANGUAGE_CONTENTS;
require_once __DIR__ . '/../inc/config.php';
require_once __DIR__ . '/../inc/auth.php';

do_logout();
header('Location: index.php');
exit;
