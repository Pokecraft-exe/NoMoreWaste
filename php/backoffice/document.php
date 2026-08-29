<?php
// Streams a generated PDF (recapitulatif de collecte, tournee ou service)
// back to the browser as a download - proxies /api/v1/documents?id=X
// without going through api_get()'s JSON decoding, since the response body
// here is the raw PDF file.
require_once __DIR__ . '/../inc/config.php';
require_once __DIR__ . '/../inc/auth.php';

$user = current_user();
$token = session_token();
require_staff_or_404($user);

$id = intval($_GET['id'] ?? 0);
if (!$id) {
    http_response_code(400);
    exit('Missing id');
}

$ch = curl_init($API_BASE_URL . '/api/v1/documents?id=' . $id);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch, CURLOPT_HTTPHEADER, ['Authorization: Bearer ' . $token]);
curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, false);
curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, false);
curl_setopt($ch, CURLOPT_HEADER, true);
$response = curl_exec($ch);
$status = curl_getinfo($ch, CURLINFO_HTTP_CODE);
$headerSize = curl_getinfo($ch, CURLINFO_HEADER_SIZE);
curl_close($ch);

if ($status !== 200 || $response === false) {
    http_response_code($status ?: 502);
    exit('Document unavailable');
}

$rawHeaders = substr($response, 0, $headerSize);
$body = substr($response, $headerSize);

foreach (['Content-Type', 'Content-Disposition', 'Content-Length'] as $name) {
    if (preg_match('/^' . preg_quote($name, '/') . ':\s*(.+)$/mi', $rawHeaders, $m)) {
        header($name . ': ' . trim($m[1]));
    }
}

echo $body;
