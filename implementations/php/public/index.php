<?php
declare(strict_types=1);

require_once dirname(__DIR__) . DIRECTORY_SEPARATOR . 'simulation.php';

const SUPPORTED_TYPES = ['normal', 'protanopia', 'deuteranopia', 'tritanopia', 'achromatopsia', 'low_vision'];

const LEGACY_PALETTES = [
    'normal' => ['background'=>'#F4F5F7','surface'=>'#FFFFFF','text'=>'#20252B','primary'=>'#2E6DA4','secondary'=>'#6B7785','error'=>'#C94C4C','success'=>'#3C8D5A'],
    'protanopia' => ['background'=>'#1E1E1E','surface'=>'#2A2A2A','text'=>'#F5F5F5','primary'=>'#3F8FD2','secondary'=>'#E3B341','error'=>'#D96C3F','success'=>'#4FB3A5'],
    'deuteranopia' => ['background'=>'#1E1E1E','surface'=>'#2A2A2A','text'=>'#F5F5F5','primary'=>'#4A90D9','secondary'=>'#D9A24A','error'=>'#D94A4A','success'=>'#4AD98C'],
    'tritanopia' => ['background'=>'#202124','surface'=>'#2D2F33','text'=>'#F5F5F5','primary'=>'#D65DB1','secondary'=>'#4CC9A7','error'=>'#E05A47','success'=>'#64A66F'],
    'achromatopsia' => ['background'=>'#202020','surface'=>'#303030','text'=>'#F2F2F2','primary'=>'#D0D0D0','secondary'=>'#A8A8A8','error'=>'#E0E0E0','success'=>'#BEBEBE'],
];

const SAFE_PALETTES = [
    'normal' => [
        'light' => LEGACY_PALETTES['normal'],
        'dark' => ['background'=>'#181A1D','surface'=>'#24272B','text'=>'#F5F7FA','primary'=>'#5CA9E6','secondary'=>'#AAB4BE','error'=>'#FF7B72','success'=>'#56D364'],
    ],
    'protanopia' => [
        'light' => ['background'=>'#F7F8FA','surface'=>'#FFFFFF','text'=>'#1D2329','primary'=>'#256EA6','secondary'=>'#916B00','error'=>'#A84824','success'=>'#237A70'],
        'dark' => LEGACY_PALETTES['protanopia'],
    ],
    'deuteranopia' => [
        'light' => ['background'=>'#F7F8FA','surface'=>'#FFFFFF','text'=>'#1D2329','primary'=>'#236FAE','secondary'=>'#8A6200','error'=>'#A83D3D','success'=>'#187A55'],
        'dark' => LEGACY_PALETTES['deuteranopia'],
    ],
    'tritanopia' => [
        'light' => ['background'=>'#F7F7F8','surface'=>'#FFFFFF','text'=>'#202124','primary'=>'#9B3F80','secondary'=>'#167A65','error'=>'#AA4234','success'=>'#347A42'],
        'dark' => LEGACY_PALETTES['tritanopia'],
    ],
    'achromatopsia' => [
        'light' => ['background'=>'#FAFAFA','surface'=>'#FFFFFF','text'=>'#181818','primary'=>'#4A4A4A','secondary'=>'#666666','error'=>'#303030','success'=>'#555555'],
        'dark' => LEGACY_PALETTES['achromatopsia'],
    ],
    'low_vision' => [
        'light' => ['background'=>'#FFFFFF','surface'=>'#F2F2F2','text'=>'#000000','primary'=>'#005FCC','secondary'=>'#6D4C00','error'=>'#A80000','success'=>'#006B35'],
        'dark' => ['background'=>'#000000','surface'=>'#121212','text'=>'#FFFFFF','primary'=>'#66B2FF','secondary'=>'#FFD166','error'=>'#FF6B6B','success'=>'#65E6A3'],
    ],
];

function loadRootEnv(): void {
    $path = dirname(__DIR__, 3) . DIRECTORY_SEPARATOR . '.env';
    if (!is_file($path)) return;
    foreach (file($path, FILE_IGNORE_NEW_LINES | FILE_SKIP_EMPTY_LINES) ?: [] as $line) {
        $line = trim($line);
        if ($line === '' || str_starts_with($line, '#') || !str_contains($line, '=')) continue;
        [$key, $value] = array_map('trim', explode('=', $line, 2));
        if ($key !== '' && getenv($key) === false) putenv($key . '=' . trim($value, "\"'"));
    }
}
function jsonResponse(int $status, array $payload): never {
    http_response_code($status); header('Content-Type: application/json; charset=utf-8');
    echo json_encode($payload, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES), "\n"; exit;
}
function validType(string $type): bool { return in_array($type, SUPPORTED_TYPES, true); }
function validSeverity(string $value): bool { return in_array($value, ['mild','moderate','severe'], true); }
function validMode(string $value): bool { return in_array($value, ['dark','light'], true); }
function onlyKeys(array $value, array $allowed): bool { foreach(array_keys($value) as $key) if(!in_array($key,$allowed,true)) return false; return true; }
function parseHex(string $hex): array { $h=ltrim($hex,'#'); return [hexdec(substr($h,0,2)),hexdec(substr($h,2,2)),hexdec(substr($h,4,2))]; }
function formatHex(float $r,float $g,float $b): string { return sprintf('#%02X%02X%02X', max(0,min(255,(int)round($r))), max(0,min(255,(int)round($g))), max(0,min(255,(int)round($b)))); }
function mixHex(string $a,string $b,float $f): string { [$ar,$ag,$ab]=parseHex($a);[$br,$bg,$bb]=parseHex($b);return formatHex($ar*(1-$f)+$br*$f,$ag*(1-$f)+$bg*$f,$ab*(1-$f)+$bb*$f); }
function luminance(string $hex): float { [$r,$g,$b]=parseHex($hex);$lin=static function(int $v):float{$x=$v/255;return $x<=0.04045?$x/12.92:(($x+0.055)/1.055)**2.4;};return 0.2126*$lin($r)+0.7152*$lin($g)+0.0722*$lin($b); }
function contrastRatio(string $a,string $b): float { $la=luminance($a);$lb=luminance($b);return (max($la,$lb)+0.05)/(min($la,$lb)+0.05); }
function contrastOK(array $p): bool { return contrastRatio($p['text'],$p['background'])>=4.5 && contrastRatio($p['text'],$p['surface'])>=4.5; }
function ensureTextContrast(array $p): array { if(contrastOK($p))return $p;$w=min(contrastRatio('#FFFFFF',$p['background']),contrastRatio('#FFFFFF',$p['surface']));$b=min(contrastRatio('#000000',$p['background']),contrastRatio('#000000',$p['surface']));$p['text']=$w>=$b?'#FFFFFF':'#000000';return $p; }
function severityFactor(string $s): float { return $s==='mild'?0.35:($s==='severe'?1.0:0.70); }
function mixPalette(array $a,array $b,float $f): array { $out=[];foreach(['background','surface','text','primary','secondary','error','success'] as $k)$out[$k]=mixHex($a[$k],$b[$k],$f);return $out; }
function grayscaleHex(string $hex): string { [$r,$g,$b]=parseHex($hex);$v=0.2126*$r+0.7152*$g+0.0722*$b;return formatHex($v,$v,$v); }
function adaptMode(array $p,string $mode): array { if($mode==='dark'){$p['background']=mixHex($p['background'],'#181A1D',0.72);$p['surface']=mixHex($p['surface'],'#24272B',0.72);if(contrastRatio($p['text'],$p['background'])<4.5||contrastRatio($p['text'],$p['surface'])<4.5)$p['text']='#F5F7FA';}else{$p['background']=mixHex($p['background'],'#F4F6F8',0.72);$p['surface']=mixHex($p['surface'],'#FFFFFF',0.82);if(contrastRatio($p['text'],$p['background'])<4.5||contrastRatio($p['text'],$p['surface'])<4.5)$p['text']='#1A1D21';}return $p; }
function highContrast(array $p,string $type,string $mode): array { $anchor=SAFE_PALETTES[$type==='normal'?'low_vision':$type][$mode];if($mode==='dark'){$p['background']='#000000';$p['surface']='#121212';$p['text']='#FFFFFF';}else{$p['background']='#FFFFFF';$p['surface']='#F2F2F2';$p['text']='#000000';}foreach(['primary','secondary','error','success'] as $k)$p[$k]=mixHex($p[$k],$anchor[$k],0.45);return $p; }
function themeResponse(string $type,array $p): array { return ['type'=>$type,'palette'=>$p,'contrast_ok'=>contrastOK($p)]; }
function getTheme(string $type,string $severity='',string $mode='',bool $hc=false,bool $explicit=false): array {
    if($severity!==''&&!validSeverity($severity))return ['error'=>'invalid_parameter','message'=>'severity debe ser mild, moderate o severe'];
    if($mode!==''&&!validMode($mode))return ['error'=>'invalid_parameter','message'=>'mode debe ser dark o light'];
    if(!$explicit&&$type!=='low_vision')return themeResponse($type,LEGACY_PALETTES[$type]);
    $mode=$mode!==''?$mode:($type==='normal'?'light':'dark');$severity=$severity!==''?$severity:'moderate';
    $p=$type==='normal'?SAFE_PALETTES['normal'][$mode]:mixPalette(SAFE_PALETTES['normal'][$mode],SAFE_PALETTES[$type][$mode],severityFactor($severity));
    if($hc||$type==='low_vision')$p=highContrast($p,$type,$mode);$p=ensureTextContrast($p);return themeResponse($type,$p);
}
function validatePalette(mixed $p): ?string { if(!is_array($p))return 'palette es requerido';foreach(['background','surface','text','primary','secondary','error','success'] as $k){if(!isset($p[$k])||!is_string($p[$k])||preg_match('/^#[0-9A-Fa-f]{6}$/',$p[$k])!==1)return "$k debe usar formato #RRGGBB";}return null; }
function customTheme(array $body): array {
    $type=(string)($body['type']??'');if(!validType($type))return ['error'=>'invalid_type','message'=>'Tipo de daltonismo no soportado'];
    $err=validatePalette($body['palette']??null);if($err!==null)return ['error'=>'invalid_palette','message'=>$err];
    $severity=(string)($body['severity']??'');$mode=(string)($body['mode']??'');if($severity!==''&&!validSeverity($severity))return ['error'=>'invalid_palette','message'=>'severity debe ser mild, moderate o severe'];if($mode!==''&&!validMode($mode))return ['error'=>'invalid_palette','message'=>'mode debe ser dark o light'];
    $p=$body['palette'];$mode=$mode!==''?$mode:(luminance($p['background'])<0.35?'dark':'light');$severity=$severity!==''?$severity:'moderate';$f=severityFactor($severity);
    if($type==='achromatopsia'){foreach(['background','surface','text','primary','secondary','error','success'] as $k)$p[$k]=mixHex($p[$k],grayscaleHex($p[$k]),$f);}elseif($type!=='normal'){$anchor=SAFE_PALETTES[$type][$mode];foreach(['primary','secondary','error','success'] as $k)$p[$k]=mixHex($p[$k],$anchor[$k],$f);}
    $p=adaptMode($p,$mode);if(($body['high_contrast']??false)===true||$type==='low_vision')$p=highContrast($p,$type,$mode);$p=ensureTextContrast($p);return themeResponse($type,$p);
}

loadRootEnv();
header('Access-Control-Allow-Origin: '.(getenv('EYEX_ALLOWED_ORIGIN')?:'*'));header('Access-Control-Allow-Headers: Accept, Accept-Language, Content-Type, If-None-Match, X-API-Key');header('Access-Control-Allow-Methods: GET, POST, OPTIONS');
if(($_SERVER['REQUEST_METHOD']??'GET')==='OPTIONS'){http_response_code(204);exit;}
$method=$_SERVER['REQUEST_METHOD']??'GET';$path=parse_url($_SERVER['REQUEST_URI']??'/',PHP_URL_PATH)?:'/';
if($method==='GET'&&$path==='/api/v1/theme/types')jsonResponse(200,['types'=>SUPPORTED_TYPES]);
if($method==='GET'&&preg_match('#^/api/v1/theme/([^/]+)$#',$path,$m)===1){$type=rawurldecode($m[1]);if(!validType($type))jsonResponse(400,['error'=>'invalid_type','message'=>'Tipo de daltonismo no soportado']);$q=$_GET;$severity=(string)($q['severity']??'');$mode=(string)($q['mode']??'');$hcRaw=(string)($q['high_contrast']??'');if($hcRaw!==''&&!in_array($hcRaw,['true','false','1','0'],true))jsonResponse(400,['error'=>'invalid_parameter','message'=>'high_contrast debe ser true o false']);$result=getTheme($type,$severity,$mode,in_array($hcRaw,['true','1'],true),$severity!==''||$mode!==''||$hcRaw!=='');jsonResponse(isset($result['error'])?400:200,$result);}
if($method==='POST'&&$path==='/api/v1/simulate'){
    $body=json_decode(file_get_contents('php://input')?:'',true);
    if(!is_array($body)||!onlyKeys($body,['hex','type','severity']))jsonResponse(400,['error'=>'invalid_request','message'=>'JSON de entrada inválido']);
    $type=is_string($body['type']??null)?$body['type']:'';
    if(!validSimulationType($type))jsonResponse(400,['error'=>'invalid_type','message'=>'Tipo de daltonismo no soportado']);
    $provided=array_key_exists('severity',$body);$severity=simulationSeverity($body['severity']??null,$provided);
    if($severity===null)jsonResponse(400,['error'=>'invalid_parameter','message'=>'severity debe estar entre 0 y 1']);
    $original=normalizeSimulationHex($body['hex']??null);
    if($original===null)jsonResponse(400,['error'=>'invalid_color','message'=>'hex debe usar formato #RRGGBB']);
    jsonResponse(200,['original'=>$original,'simulated'=>simulateMachadoHex($original,$type,$severity),'type'=>$type,'severity'=>$severity,'model'=>MACHADO_MODEL]);
}
if($method==='POST'&&$path==='/api/v1/simulate/batch'){
    $body=json_decode(file_get_contents('php://input')?:'',true);
    if(!is_array($body)||!onlyKeys($body,['colors','type','severity']))jsonResponse(400,['error'=>'invalid_request','message'=>'JSON de entrada inválido']);
    $colors=$body['colors']??null;
    if(!is_array($colors)||!array_is_list($colors))jsonResponse(400,['error'=>'invalid_request','message'=>'JSON de entrada inválido']);
    if(count($colors)<1||count($colors)>256)jsonResponse(400,['error'=>'invalid_request','message'=>'colors debe contener entre 1 y 256 colores']);
    $type=is_string($body['type']??null)?$body['type']:'';
    if(!validSimulationType($type))jsonResponse(400,['error'=>'invalid_type','message'=>'Tipo de daltonismo no soportado']);
    $provided=array_key_exists('severity',$body);$severity=simulationSeverity($body['severity']??null,$provided);
    if($severity===null)jsonResponse(400,['error'=>'invalid_parameter','message'=>'severity debe estar entre 0 y 1']);
    $results=[];
    foreach($colors as $color){$original=normalizeSimulationHex($color);if($original===null)jsonResponse(400,['error'=>'invalid_color','message'=>'cada color debe usar formato #RRGGBB']);$results[]=['original'=>$original,'simulated'=>simulateMachadoHex($original,$type,$severity)];}
    jsonResponse(200,['type'=>$type,'severity'=>$severity,'model'=>MACHADO_MODEL,'results'=>$results]);
}
if($method==='POST'&&$path==='/api/v1/theme/custom'){$body=json_decode(file_get_contents('php://input')?:'',true);if(!is_array($body))jsonResponse(400,['error'=>'invalid_request','message'=>'JSON de entrada inválido']);$result=customTheme($body);jsonResponse(isset($result['error'])?400:200,$result);}
if($method==='POST'&&$path==='/api/v1/test/suggest'){$body=json_decode(file_get_contents('php://input')?:'',true);if(!is_array($body)||!is_array($body['answers']??null))jsonResponse(400,['error'=>'invalid_request','message'=>'JSON de entrada inválido']);$a=$body['answers'];$suggested='normal';if(($a['colors_look_gray']??false)===true)$suggested='achromatopsia';elseif(($a['blue_yellow_confusion']??false)===true)$suggested='tritanopia';elseif(($a['reds_look_darker']??false)===true&&($a['green_brown_confusion']??false)===true)$suggested='protanopia';elseif(($a['green_brown_confusion']??false)===true)$suggested='deuteranopia';elseif(($a['reds_look_darker']??false)===true)$suggested='protanopia';jsonResponse(200,['suggested_type'=>$suggested,'disclaimer'=>'Resultado orientativo. No es un diagnóstico médico.']);}
jsonResponse(404,['error'=>'not_found']);
