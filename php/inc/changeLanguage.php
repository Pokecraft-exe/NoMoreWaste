<?php

if (session_status() === PHP_SESSION_NONE) {
    session_start();
}

$DEFAULT_LANG = "fr-fr";
$LOADED_LANGUAGE = $_SESSION['lang'] ?? $DEFAULT_LANG;

if (isset($_GET['lang']) && file_exists(__DIR__ . "/../lang/" . $_GET['lang'] . ".txt")) {
    $LOADED_LANGUAGE = $_GET['lang'];
}

if (!file_exists(__DIR__ . "/../lang/" . $LOADED_LANGUAGE . ".txt")) {
    $LOADED_LANGUAGE = $DEFAULT_LANG;
}

// Persist the choice so it survives navigation to pages that don't pass
// ?lang= explicitly.
$_SESSION['lang'] = $LOADED_LANGUAGE;

function load_lang($ldl) {
    $language_contents = array();
    $handle = fopen(__DIR__ . "/../lang/" . $ldl . ".txt", "r");
    if ($handle) {
        while (($line = fgets($handle)) !== false) {
            $language_contents[] = rtrim($line, "\r\n");
        }
        fclose($handle);
    }
    return $language_contents;
}

$LANGUAGE_CONTENTS = load_lang($LOADED_LANGUAGE);

$AVAILABLE_LANGUAGES = scandir(__DIR__ . "/../lang/");
for ($i = 0; $i < count($AVAILABLE_LANGUAGES); $i++) {
    $AVAILABLE_LANGUAGES[$i] = substr($AVAILABLE_LANGUAGES[$i], 0, -4);
}

$AVAILABLE_LANGUAGES = array_values(array_unique($AVAILABLE_LANGUAGES));
array_shift($AVAILABLE_LANGUAGES);

$CONNECTION_TITLE = 0;
$INSCRIPTION_TITLE = 1;
$EMAIL_LABEL = 2;
$SECRET_LABEL = 3;
$CONFIRM_LABEL = 4;
$LANGUAGES_TITLE = 5;
$LANGUAGES_HEADER = 6;
$ROLE_SELECT_TITLE = 7;
$ROLE_SELECT_SUBTITLE = 8;

// Forums page constants
$FORUMS_TITLE = 9;
$FORUMS_OFFSET_LABEL = 10;
$FORUMS_RESOLVED_LABEL = 11;
$FORUMS_NAME_LABEL = 12;
$FORUMS_PREVIEW_LABEL = 13;
$FORUMS_DATE_LABEL = 14;
$FORUMS_RESOLVED_YES = 15;
$FORUMS_RESOLVED_NO = 16;

// Annonces page constants
$ANNONCES_TITLE = 17;
$ANNONCES_OFFSET_LABEL = 18;
$ANNONCES_CATEGORY_LABEL = 19;
$ANNONCES_NAME_LABEL = 20;
$ANNONCES_PRICE_LABEL = 21;
$ANNONCES_DESCRIPTION_LABEL = 22;
$ANNONCES_DATE_LABEL = 23;
$ANNONCES_IMAGE_LABEL = 24;
$ANNONCES_STATE_LABEL = 25;

// Forum detail constants
$FORUMS_REPLY_LABEL = 28;
$FORUMS_LIKE_LABEL = 29;
$FORUMS_DISLIKE_LABEL = 30;

// Annonce detail constants
$ANNONCE_SELLER_LABEL = 31;
$ANNONCE_BUYER_LABEL = 32;
$ANNONCE_MESSAGES_LABEL = 33;
$ANNONCE_BACK_LABEL = 34;
$ANNONCE_MESSAGE_LABEL = 35;
$ANNONCE_AUTHOR_LABEL = 36;
$ANNONCE_DATE_LABEL = 37;

// Common constants
$PREVIOUS_LABEL = 26;
$NEXT_LABEL = 27;