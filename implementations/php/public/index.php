<?php
declare(strict_types=1);

const SUPPORTED_TYPES = [
    'normal',
    'protanopia',
    'deuteranopia',
    'tritanopia',
    'achromatopsia',
];

const PALETTES = [
    'normal' => [
        'background' => '#F4F5F7',
        'surface' => '#FFFFFF',
        'text' => '#20252B',
        'primary' => '#2E6DA4',
        'secondary' => '#6B7785',
        'error' => '#C94C4C',
        'success' => '#3C8D5A',
    ],
    'protanopia' => [
        'background' => '#1E1E1E',
        'surface' => '#2A2A2A',
        'text' => '#F5F5F5',
        'primary' => '#3F8FD2',
        'secondary' => '#E3B341',
        'error' => '#D96C3F',
        'success' => '#4FB3A5',
    ],
    'deuteranopia' => [
        'background' => '#1E1E1E',
        'surface' => '#2A2A2A',
        'text' => '#F5F5F5',
        'primary' => '#4A90D9',
        'secondary' => '#D9A24A',
        'error' => '#D94A4A',
        'success' => '#4AD98C',
    ],
    'tritanopia' => [
        'background' => '#202124',
        'surface' => '#2D2F33',
        'text' => '#F5F5F5',
        'primary' => '#D65DB1',
        'secondary' => '#4CC9A7',
        'error' => '#E05A47',
        'success' => '#64A66F',
    ],
    'achromatopsia' => [
        'background' => '#202020',
        'surface' => '#303030',
        'text' => '#F2F2F2',
        'primary' => '#D0D0D0',
        'secondary' => '#A8A8A8',
        'error' => '#E0E0E0',
        'success' => '#BEBEBE',
    ],
];

function loadRootEnv(): void
{
    $path = dirname(__DIR__, 3) . DIRECTORY_SEPARATOR . '.env';
    if (!is_file($path)) {
        return;
    }
    foreach (file($path, FILE_IGNORE_NEW_LINES | FILE_SKIP_EMPTY_LINES) ?: [] as $line) {
        $line = trim($line);
        if ($line === '' || str_starts_with($line, '#') || !str_contains($line, '=')) {
            continue;
        }
        [$key, $value] = array_map('trim', explode('=', $line, 2));
        if ($key !== '' && getenv($key) === false) {
            putenv($key . '=' . trim($value, "\"'"));
        }
    }
}

function jsonResponse(int $status, array $payload): never
{
    http_response_code($status);
    header('Content-Type: application/json; charset=utf-8');
    echo json_encode($payload, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES), "\n";
    exit;
}

loadRootEnv();
header('Access-Control-Allow-Origin: ' . (getenv('EYEX_ALLOWED_ORIGIN') ?: '*'));
header('Access-Control-Allow-Headers: Content-Type');
header('Access-Control-Allow-Methods: GET, OPTIONS');

if ($_SERVER['REQUEST_METHOD'] === 'OPTIONS') {
    http_response_code(204);
    exit;
}

if ($_SERVER['REQUEST_METHOD'] !== 'GET') {
    jsonResponse(404, ['error' => 'not_found']);
}

$path = parse_url($_SERVER['REQUEST_URI'] ?? '/', PHP_URL_PATH) ?: '/';

if ($path === '/api/v1/theme/types') {
    jsonResponse(200, ['types' => SUPPORTED_TYPES]);
}

if (preg_match('#^/api/v1/theme/([^/]+)$#', $path, $matches) === 1) {
    $type = rawurldecode($matches[1]);
    if (!array_key_exists($type, PALETTES)) {
        jsonResponse(400, [
            'error' => 'invalid_type',
            'message' => 'Tipo de daltonismo no soportado',
        ]);
    }
    jsonResponse(200, [
        'type' => $type,
        'palette' => PALETTES[$type],
    ]);
}

jsonResponse(404, ['error' => 'not_found']);
