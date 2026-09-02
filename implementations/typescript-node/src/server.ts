import { existsSync, readFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import Fastify from 'fastify';
import cors from '@fastify/cors';
import { MACHADO_MODEL, isSimulationType, normalizeSimulationHex, normalizeSimulationSeverity, simulateMachadoHex } from './simulation.js';

type ThemeType = 'normal' | 'protanopia' | 'deuteranopia' | 'tritanopia' | 'achromatopsia' | 'low_vision';
type Severity = 'mild' | 'moderate' | 'severe';
type Mode = 'dark' | 'light';

interface Palette {
  background: string;
  surface: string;
  text: string;
  primary: string;
  secondary: string;
  error: string;
  success: string;
}

interface ThemeResponse {
  type: ThemeType;
  palette: Palette;
  contrast_ok: boolean;
}

interface TypesResponse { types: ThemeType[]; }
interface ErrorResponse { error: string; message: string; }

interface CustomThemeRequest {
  type: string;
  palette: Palette;
  severity?: string;
  mode?: string;
  high_contrast?: boolean;
}

interface QuickTestRequest {
  answers: {
    reds_look_darker: boolean;
    green_brown_confusion: boolean;
    blue_yellow_confusion: boolean;
    colors_look_gray: boolean;
  };
}

function loadRootEnv(): void {
  const here = dirname(fileURLToPath(import.meta.url));
  const path = resolve(here, '../../../.env');
  if (!existsSync(path)) return;
  for (const rawLine of readFileSync(path, 'utf8').split(/\r?\n/)) {
    const line = rawLine.trim();
    if (!line || line.startsWith('#')) continue;
    const separator = line.indexOf('=');
    if (separator < 1) continue;
    const key = line.slice(0, separator).trim();
    const value = line.slice(separator + 1).trim().replace(/^["']|["']$/g, '');
    if (process.env[key] === undefined) process.env[key] = value;
  }
}

loadRootEnv();

const supportedTypes: ThemeType[] = ['normal', 'protanopia', 'deuteranopia', 'tritanopia', 'achromatopsia', 'low_vision'];

const legacyPalettes: Record<Exclude<ThemeType, 'low_vision'>, Palette> = {
  normal: { background: '#F4F5F7', surface: '#FFFFFF', text: '#20252B', primary: '#2E6DA4', secondary: '#6B7785', error: '#C94C4C', success: '#3C8D5A' },
  protanopia: { background: '#1E1E1E', surface: '#2A2A2A', text: '#F5F5F5', primary: '#3F8FD2', secondary: '#E3B341', error: '#D96C3F', success: '#4FB3A5' },
  deuteranopia: { background: '#1E1E1E', surface: '#2A2A2A', text: '#F5F5F5', primary: '#4A90D9', secondary: '#D9A24A', error: '#D94A4A', success: '#4AD98C' },
  tritanopia: { background: '#202124', surface: '#2D2F33', text: '#F5F5F5', primary: '#D65DB1', secondary: '#4CC9A7', error: '#E05A47', success: '#64A66F' },
  achromatopsia: { background: '#202020', surface: '#303030', text: '#F2F2F2', primary: '#D0D0D0', secondary: '#A8A8A8', error: '#E0E0E0', success: '#BEBEBE' },
};

const safePalettes: Record<ThemeType, Record<Mode, Palette>> = {
  normal: {
    light: legacyPalettes.normal,
    dark: { background: '#181A1D', surface: '#24272B', text: '#F5F7FA', primary: '#5CA9E6', secondary: '#AAB4BE', error: '#FF7B72', success: '#56D364' },
  },
  protanopia: {
    light: { background: '#F7F8FA', surface: '#FFFFFF', text: '#1D2329', primary: '#256EA6', secondary: '#916B00', error: '#A84824', success: '#237A70' },
    dark: legacyPalettes.protanopia,
  },
  deuteranopia: {
    light: { background: '#F7F8FA', surface: '#FFFFFF', text: '#1D2329', primary: '#236FAE', secondary: '#8A6200', error: '#A83D3D', success: '#187A55' },
    dark: legacyPalettes.deuteranopia,
  },
  tritanopia: {
    light: { background: '#F7F7F8', surface: '#FFFFFF', text: '#202124', primary: '#9B3F80', secondary: '#167A65', error: '#AA4234', success: '#347A42' },
    dark: legacyPalettes.tritanopia,
  },
  achromatopsia: {
    light: { background: '#FAFAFA', surface: '#FFFFFF', text: '#181818', primary: '#4A4A4A', secondary: '#666666', error: '#303030', success: '#555555' },
    dark: legacyPalettes.achromatopsia,
  },
  low_vision: {
    light: { background: '#FFFFFF', surface: '#F2F2F2', text: '#000000', primary: '#005FCC', secondary: '#6D4C00', error: '#A80000', success: '#006B35' },
    dark: { background: '#000000', surface: '#121212', text: '#FFFFFF', primary: '#66B2FF', secondary: '#FFD166', error: '#FF6B6B', success: '#65E6A3' },
  },
};

function isThemeType(value: string): value is ThemeType { return supportedTypes.includes(value as ThemeType); }
function isSeverity(value: string): value is Severity { return ['mild', 'moderate', 'severe'].includes(value); }
function isMode(value: string): value is Mode { return value === 'dark' || value === 'light'; }
function validHex(value: string): boolean { return /^#[0-9A-Fa-f]{6}$/.test(value); }
function hasOnlyKeys(value: unknown, allowed: readonly string[]): boolean {
  if (value === null || typeof value !== 'object' || Array.isArray(value)) return false;
  return Object.keys(value).every((key) => allowed.includes(key));
}

function parseHex(value: string): [number, number, number] {
  const raw = value.replace('#', '');
  return [Number.parseInt(raw.slice(0, 2), 16), Number.parseInt(raw.slice(2, 4), 16), Number.parseInt(raw.slice(4, 6), 16)];
}
function formatHex(r: number, g: number, b: number): string {
  const one = (v: number) => Math.max(0, Math.min(255, Math.round(v))).toString(16).padStart(2, '0').toUpperCase();
  return `#${one(r)}${one(g)}${one(b)}`;
}
function mixHex(a: string, b: string, factor: number): string {
  const av = parseHex(a); const bv = parseHex(b);
  return formatHex(av[0] * (1 - factor) + bv[0] * factor, av[1] * (1 - factor) + bv[1] * factor, av[2] * (1 - factor) + bv[2] * factor);
}
function relativeLuminance(value: string): number {
  const channels = parseHex(value).map((v) => {
    const x = v / 255;
    return x <= 0.04045 ? x / 12.92 : ((x + 0.055) / 1.055) ** 2.4;
  });
  return 0.2126 * channels[0] + 0.7152 * channels[1] + 0.0722 * channels[2];
}
function contrastRatio(a: string, b: string): number {
  const la = relativeLuminance(a); const lb = relativeLuminance(b);
  return (Math.max(la, lb) + 0.05) / (Math.min(la, lb) + 0.05);
}
function contrastOK(p: Palette): boolean { return contrastRatio(p.text, p.background) >= 4.5 && contrastRatio(p.text, p.surface) >= 4.5; }
function ensureTextContrast(p: Palette): Palette {
  if (contrastOK(p)) return p;
  const white = Math.min(contrastRatio('#FFFFFF', p.background), contrastRatio('#FFFFFF', p.surface));
  const black = Math.min(contrastRatio('#000000', p.background), contrastRatio('#000000', p.surface));
  return { ...p, text: white >= black ? '#FFFFFF' : '#000000' };
}
function severityFactor(severity: Severity): number { return severity === 'mild' ? 0.35 : severity === 'severe' ? 1 : 0.70; }
function mixPalette(a: Palette, b: Palette, factor: number): Palette {
  return {
    background: mixHex(a.background, b.background, factor), surface: mixHex(a.surface, b.surface, factor),
    text: mixHex(a.text, b.text, factor), primary: mixHex(a.primary, b.primary, factor),
    secondary: mixHex(a.secondary, b.secondary, factor), error: mixHex(a.error, b.error, factor), success: mixHex(a.success, b.success, factor),
  };
}
function grayscaleHex(value: string): string {
  const [r, g, b] = parseHex(value); const gray = 0.2126 * r + 0.7152 * g + 0.0722 * b;
  return formatHex(gray, gray, gray);
}
function adaptMode(p: Palette, mode: Mode): Palette {
  const out = { ...p };
  if (mode === 'dark') {
    out.background = mixHex(out.background, '#181A1D', 0.72); out.surface = mixHex(out.surface, '#24272B', 0.72);
    if (contrastRatio(out.text, out.background) < 4.5 || contrastRatio(out.text, out.surface) < 4.5) out.text = '#F5F7FA';
  } else {
    out.background = mixHex(out.background, '#F4F6F8', 0.72); out.surface = mixHex(out.surface, '#FFFFFF', 0.82);
    if (contrastRatio(out.text, out.background) < 4.5 || contrastRatio(out.text, out.surface) < 4.5) out.text = '#1A1D21';
  }
  return out;
}
function applyHighContrast(p: Palette, type: ThemeType, mode: Mode): Palette {
  const anchor = safePalettes[type === 'normal' ? 'low_vision' : type][mode];
  const out = { ...p };
  if (mode === 'dark') { out.background = '#000000'; out.surface = '#121212'; out.text = '#FFFFFF'; }
  else { out.background = '#FFFFFF'; out.surface = '#F2F2F2'; out.text = '#000000'; }
  out.primary = mixHex(out.primary, anchor.primary, 0.45); out.secondary = mixHex(out.secondary, anchor.secondary, 0.45);
  out.error = mixHex(out.error, anchor.error, 0.45); out.success = mixHex(out.success, anchor.success, 0.45);
  return out;
}
function response(type: ThemeType, palette: Palette): ThemeResponse { return { type, palette, contrast_ok: contrastOK(palette) }; }

function getTheme(type: ThemeType, severityRaw?: string, modeRaw?: string, highContrast = false, explicit = false): ThemeResponse | ErrorResponse {
  if (severityRaw && !isSeverity(severityRaw)) return { error: 'invalid_parameter', message: 'severity debe ser mild, moderate o severe' };
  if (modeRaw && !isMode(modeRaw)) return { error: 'invalid_parameter', message: 'mode debe ser dark o light' };
  if (!explicit && type !== 'low_vision') return response(type, legacyPalettes[type as Exclude<ThemeType, 'low_vision'>]);
  const mode: Mode = modeRaw && isMode(modeRaw) ? modeRaw : type === 'normal' ? 'light' : 'dark';
  const severity: Severity = severityRaw && isSeverity(severityRaw) ? severityRaw : 'moderate';
  let palette = type === 'normal' ? safePalettes.normal[mode] : mixPalette(safePalettes.normal[mode], safePalettes[type][mode], severityFactor(severity));
  if (highContrast || type === 'low_vision') palette = applyHighContrast(palette, type, mode);
  return response(type, ensureTextContrast(palette));
}

function validatePalette(palette: Palette): string | null {
  for (const [name, value] of Object.entries(palette || {})) if (!validHex(value)) return `${name} debe usar formato #RRGGBB`;
  for (const name of ['background', 'surface', 'text', 'primary', 'secondary', 'error', 'success'] as const) if (!palette?.[name]) return `${name} debe usar formato #RRGGBB`;
  return null;
}

function customTheme(body: CustomThemeRequest): ThemeResponse | ErrorResponse {
  if (!isThemeType(body.type)) return { error: 'invalid_type', message: 'Tipo de daltonismo no soportado' };
  const paletteError = validatePalette(body.palette);
  if (paletteError) return { error: 'invalid_palette', message: paletteError };
  if (body.severity && !isSeverity(body.severity)) return { error: 'invalid_palette', message: 'severity debe ser mild, moderate o severe' };
  if (body.mode && !isMode(body.mode)) return { error: 'invalid_palette', message: 'mode debe ser dark o light' };
  const mode: Mode = body.mode && isMode(body.mode) ? body.mode : relativeLuminance(body.palette.background) < 0.35 ? 'dark' : 'light';
  const severity: Severity = body.severity && isSeverity(body.severity) ? body.severity : 'moderate';
  let palette = { ...body.palette };
  const factor = severityFactor(severity);
  if (body.type === 'achromatopsia') {
    palette = {
      background: mixHex(palette.background, grayscaleHex(palette.background), factor),
      surface: mixHex(palette.surface, grayscaleHex(palette.surface), factor),
      text: mixHex(palette.text, grayscaleHex(palette.text), factor),
      primary: mixHex(palette.primary, grayscaleHex(palette.primary), factor),
      secondary: mixHex(palette.secondary, grayscaleHex(palette.secondary), factor),
      error: mixHex(palette.error, grayscaleHex(palette.error), factor),
      success: mixHex(palette.success, grayscaleHex(palette.success), factor),
    };
  } else if (body.type !== 'normal') {
    const anchor = safePalettes[body.type][mode];
    palette.primary = mixHex(palette.primary, anchor.primary, factor); palette.secondary = mixHex(palette.secondary, anchor.secondary, factor);
    palette.error = mixHex(palette.error, anchor.error, factor); palette.success = mixHex(palette.success, anchor.success, factor);
  }
  palette = adaptMode(palette, mode);
  if (body.high_contrast || body.type === 'low_vision') palette = applyHighContrast(palette, body.type, mode);
  return response(body.type, ensureTextContrast(palette));
}

const app = Fastify({ logger: true });
await app.register(cors, { origin: process.env.EYEX_ALLOWED_ORIGIN || '*', methods: ['GET', 'POST', 'OPTIONS'], allowedHeaders: ['Content-Type', 'Accept', 'Accept-Language', 'If-None-Match', 'X-API-Key'] });

app.setErrorHandler((error, _request, reply) => {
  const statusCode =
    typeof error === 'object' &&
    error !== null &&
    'statusCode' in error &&
    typeof error.statusCode === 'number'
      ? error.statusCode
      : undefined;

  if (statusCode === 400) {
    return reply.code(400).send({
      error: 'invalid_request',
      message: 'JSON de entrada inválido',
    });
  }

  app.log.error(error);
  return reply.code(500).send({
    error: 'internal_server_error',
    message: 'Error interno del servidor',
  });
});
app.setNotFoundHandler((_request, reply) => reply.code(404).send({ error: 'not_found', message: 'Recurso no encontrado' }));

app.get<{ Reply: TypesResponse }>('/api/v1/theme/types', async () => ({ types: [...supportedTypes] }));

app.get<{ Params: { type: string }; Querystring: { severity?: string; mode?: string; high_contrast?: string }; Reply: ThemeResponse | ErrorResponse }>(
  '/api/v1/theme/:type', async (request, reply) => {
    const { type } = request.params;
    if (!isThemeType(type)) return reply.code(400).send({ error: 'invalid_type', message: 'Tipo de daltonismo no soportado' });
    const { severity, mode, high_contrast } = request.query;
    if (high_contrast && !['true', 'false', '1', '0'].includes(high_contrast)) return reply.code(400).send({ error: 'invalid_parameter', message: 'high_contrast debe ser true o false' });
    const result = getTheme(type, severity, mode, high_contrast === 'true' || high_contrast === '1', Boolean(severity || mode || high_contrast));
    if ('error' in result) return reply.code(400).send(result);
    return result;
  },
);

app.post<{ Body: { hex?: unknown; type?: unknown; severity?: unknown } }>('/api/v1/simulate', async (request, reply) => {
  const body = request.body || {};
  if (!hasOnlyKeys(body, ['hex', 'type', 'severity'])) return reply.code(400).send({ error: 'invalid_request', message: 'JSON de entrada inválido' });
  if (!isSimulationType(body.type)) return reply.code(400).send({ error: 'invalid_type', message: 'Tipo de daltonismo no soportado' });
  const severity = normalizeSimulationSeverity(body.severity, Object.prototype.hasOwnProperty.call(body, 'severity'));
  if (severity === null) return reply.code(400).send({ error: 'invalid_parameter', message: 'severity debe estar entre 0 y 1' });
  const original = normalizeSimulationHex(body.hex);
  if (original === null) return reply.code(400).send({ error: 'invalid_color', message: 'hex debe usar formato #RRGGBB' });
  return { original, simulated: simulateMachadoHex(original, body.type, severity), type: body.type, severity, model: MACHADO_MODEL };
});

app.post<{ Body: { colors?: unknown; type?: unknown; severity?: unknown } }>('/api/v1/simulate/batch', async (request, reply) => {
  const body = request.body || {};
  if (!hasOnlyKeys(body, ['colors', 'type', 'severity'])) return reply.code(400).send({ error: 'invalid_request', message: 'JSON de entrada inválido' });
  if (!Array.isArray(body.colors)) return reply.code(400).send({ error: 'invalid_request', message: 'JSON de entrada inválido' });
  if (body.colors.length < 1 || body.colors.length > 256) return reply.code(400).send({ error: 'invalid_request', message: 'colors debe contener entre 1 y 256 colores' });
  if (!isSimulationType(body.type)) return reply.code(400).send({ error: 'invalid_type', message: 'Tipo de daltonismo no soportado' });
  const severity = normalizeSimulationSeverity(body.severity, Object.prototype.hasOwnProperty.call(body, 'severity'));
  if (severity === null) return reply.code(400).send({ error: 'invalid_parameter', message: 'severity debe estar entre 0 y 1' });
  const results: Array<{ original: string; simulated: string }> = [];
  for (const color of body.colors) {
    const original = normalizeSimulationHex(color);
    if (original === null) return reply.code(400).send({ error: 'invalid_color', message: 'cada color debe usar formato #RRGGBB' });
    results.push({ original, simulated: simulateMachadoHex(original, body.type, severity) });
  }
  return { type: body.type, severity, model: MACHADO_MODEL, results };
});

app.post<{ Body: CustomThemeRequest; Reply: ThemeResponse | ErrorResponse }>('/api/v1/theme/custom', async (request, reply) => {
  const result = customTheme(request.body);
  if ('error' in result) return reply.code(400).send(result);
  return result;
});

app.post<{ Body: QuickTestRequest }>('/api/v1/test/suggest', async (request) => {
  const a = request.body.answers;
  let suggested_type: ThemeType = 'normal';
  if (a.colors_look_gray) suggested_type = 'achromatopsia';
  else if (a.blue_yellow_confusion) suggested_type = 'tritanopia';
  else if (a.reds_look_darker && a.green_brown_confusion) suggested_type = 'protanopia';
  else if (a.green_brown_confusion) suggested_type = 'deuteranopia';
  else if (a.reds_look_darker) suggested_type = 'protanopia';
  return { suggested_type, disclaimer: 'Resultado orientativo. No es un diagnóstico médico.' };
});

const port = Number.parseInt(process.env.EYEX_PORT || '8080', 10);
if (!Number.isInteger(port) || port < 1 || port > 65535) throw new Error('EYEX_PORT must be an integer between 1 and 65535');
await app.listen({ port, host: '0.0.0.0' });
