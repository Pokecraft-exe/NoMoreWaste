<?php

function h($value) {
    if ($value === null) {
        return '';
    }
    return htmlspecialchars((string) $value, ENT_QUOTES, 'UTF-8');
}

$USER_TYPE_LABELS = [
    'visiteur' => 'Visiteur(se)',
    'adherent' => 'Adhérent(e)',
    'benevole' => 'Bénévole',
    'commercant' => 'Commerçant(e)',
    'responsable' => 'Responsable',
    'administrateur' => 'Administrateur(rice)',
];

function user_type_label($type) {
    global $USER_TYPE_LABELS;
    return $USER_TYPE_LABELS[$type] ?? $type;
}

function time_ago($value) {
    if (!$value) {
        return '';
    }
    $ts = strtotime($value);
    if (!$ts) {
        return '';
    }
    $diffSeconds = time() - $ts;
    $hours = intdiv($diffSeconds, 3600);
    if ($hours < 1) {
        return "a l'instant";
    }
    if ($hours < 24) {
        return "il y a $hours h";
    }
    $days = intdiv($hours, 24);
    return "il y a $days j";
}

function format_date_fr($isoDate) {
    if (!$isoDate) {
        return '';
    }
    $ts = strtotime($isoDate);
    if (!$ts) {
        return $isoDate;
    }
    $jours = ['dimanche', 'lundi', 'mardi', 'mercredi', 'jeudi', 'vendredi', 'samedi'];
    $mois = ['janvier', 'fevrier', 'mars', 'avril', 'mai', 'juin', 'juillet', 'aout', 'septembre', 'octobre', 'novembre', 'decembre'];
    return $jours[date('w', $ts)] . ' ' . date('j', $ts) . ' ' . $mois[date('n', $ts) - 1];
}

function full_name($item) {
    return trim(($item['prenom'] ?? '') . ' ' . ($item['nom'] ?? ''));
}

function query_param($name, $default = '') {
    return isset($_GET[$name]) ? trim($_GET[$name]) : $default;
}

function api_params($params) {
    return array_filter($params, fn($v) => $v !== '' && $v !== null);
}
