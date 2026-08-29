<?php

function ics_escape($value) {
    return str_replace(["\\", ";", ",", "\n"], ["\\\\", "\\;", "\\,", "\\n"], (string) $value);
}

function ics_datetime($date, $time = null) {
    $ts = strtotime($date . ' ' . ($time ?: '00:00:00'));
    return $ts ? date('Ymd\THis', $ts) : '';
}

function ics_add_hours($time, $hours) {
    $ts = strtotime($time ?: '09:00:00');
    return $ts ? date('H:i:s', $ts + $hours * 3600) : '11:00:00';
}

function build_ics($events) {
    $lines = ['BEGIN:VCALENDAR', 'VERSION:2.0', 'PRODID:-//NO MORE WASTE//Planning//FR', 'CALSCALE:GREGORIAN'];
    foreach ($events as $e) {
        $lines[] = 'BEGIN:VEVENT';
        $lines[] = 'UID:' . $e['uid'] . '@nomorewaste.local';
        $lines[] = 'DTSTAMP:' . gmdate('Ymd\THis\Z');
        $lines[] = 'DTSTART:' . $e['start'];
        $lines[] = 'DTEND:' . $e['end'];
        $lines[] = 'SUMMARY:' . ics_escape($e['title']);
        if (!empty($e['location'])) {
            $lines[] = 'LOCATION:' . ics_escape($e['location']);
        }
        if (!empty($e['description'])) {
            $lines[] = 'DESCRIPTION:' . ics_escape($e['description']);
        }
        $lines[] = 'END:VEVENT';
    }
    $lines[] = 'END:VCALENDAR';
    return implode("\r\n", $lines) . "\r\n";
}

function send_ics($filename, $events) {
    $content = build_ics($events);
    header('Content-Type: text/calendar; charset=utf-8');
    header('Content-Disposition: attachment; filename="' . $filename . '"');
    header('Content-Length: ' . strlen($content));
    echo $content;
    exit;
}

function google_calendar_link($title, $start, $end, $location = '', $details = '') {
    $params = [
        'action' => 'TEMPLATE',
        'text' => $title,
        'dates' => $start . '/' . $end,
        'location' => $location,
        'details' => $details,
    ];
    return 'https://calendar.google.com/calendar/render?' . http_build_query($params);
}
