<?php

$API_BASE_URL = getenv('API_BASE_URL') ?: 'https://localhost:8081';

if (session_status() === PHP_SESSION_NONE) {
    session_start();
}
