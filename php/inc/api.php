<?php

require_once __DIR__ . '/config.php';

function api_request($method, $path, $params = [], $token = null) {
    global $API_BASE_URL;

    $url = $API_BASE_URL . $path;
    $isGet = $method === 'GET' || $method === 'DELETE';
    if ($isGet && $params) {
        $url .= (strpos($path, '?') === false ? '?' : '&') . http_build_query($params);
    }

    $ch = curl_init($url);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_CUSTOMREQUEST, $method);
    curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, false);
    curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, false);
    if (!$isGet) {
        curl_setopt($ch, CURLOPT_POSTFIELDS, http_build_query($params));
    }

    $headers = [];
    if ($token) {
        $headers[] = 'Authorization: Bearer ' . $token;
    }
    if ($headers) {
        curl_setopt($ch, CURLOPT_HTTPHEADER, $headers);
    }

    $body = curl_exec($ch);
    $status = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    $error = curl_error($ch);
    curl_close($ch);

    if ($error) {
        return ['status' => 0, 'data' => null, 'error' => $error];
    }

    return ['status' => $status, 'data' => json_decode($body, true), 'error' => null];
}

function api_basic($method, $path, $email, $password) {
    global $API_BASE_URL;

    $ch = curl_init($API_BASE_URL . $path);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_CUSTOMREQUEST, $method);
    curl_setopt($ch, CURLOPT_USERPWD, $email . ':' . $password);
    curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, false);
    curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, false);

    $body = curl_exec($ch);
    $status = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    curl_close($ch);

    return ['status' => $status, 'data' => json_decode($body, true)];
}

function api_get($path, $params = [], $token = null) {
    return api_request('GET', $path, $params, $token);
}

function api_put($path, $params = [], $token = null) {
    return api_request('PUT', $path, $params, $token);
}

function api_patch($path, $params = [], $token = null) {
    return api_request('PATCH', $path, $params, $token);
}

function api_delete($path, $params = [], $token = null) {
    return api_request('DELETE', $path, $params, $token);
}
