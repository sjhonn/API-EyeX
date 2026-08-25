import { existsSync, readFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import Fastify from 'fastify';
import cors from '@fastify/cors';

type ThemeType = 'normal' | 'protanopia' | 'deuteranopia' | 'tritanopia' | 'achromatopsia';

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
}

interface TypesResponse {
  types: ThemeType[];
}

interface ErrorResponse {
  error: 'invalid_type';
  message: 'Tipo de daltonismo no soportado';
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

const supportedTypes: ThemeType[] = [
  'normal',
  'protanopia',
  'deuteranopia',
  'tritanopia',
  'achromatopsia',
];

const palettes: Record<ThemeType, Palette> = {
  normal: {
    background: '#F4F5F7', surface: '#FFFFFF', text: '#20252B', primary: '#2E6DA4',
    secondary: '#6B7785', error: '#C94C4C', success: '#3C8D5A',
  },
  protanopia: {
    background: '#1E1E1E', surface: '#2A2A2A', text: '#F5F5F5', primary: '#3F8FD2',
    secondary: '#E3B341', error: '#D96C3F', success: '#4FB3A5',
  },
  deuteranopia: {
    background: '#1E1E1E', surface: '#2A2A2A', text: '#F5F5F5', primary: '#4A90D9',
    secondary: '#D9A24A', error: '#D94A4A', success: '#4AD98C',
  },
  tritanopia: {
    background: '#202124', surface: '#2D2F33', text: '#F5F5F5', primary: '#D65DB1',
    secondary: '#4CC9A7', error: '#E05A47', success: '#64A66F',
  },
  achromatopsia: {
    background: '#202020', surface: '#303030', text: '#F2F2F2', primary: '#D0D0D0',
    secondary: '#A8A8A8', error: '#E0E0E0', success: '#BEBEBE',
  },
};

function isThemeType(value: string): value is ThemeType {
  return Object.prototype.hasOwnProperty.call(palettes, value);
}

const app = Fastify({ logger: true });
await app.register(cors, {
  origin: process.env.EYEX_ALLOWED_ORIGIN || '*',
  methods: ['GET', 'OPTIONS'],
});

app.get<{ Reply: TypesResponse }>('/api/v1/theme/types', async () => ({
  types: [...supportedTypes],
}));

app.get<{
  Params: { type: string };
  Reply: ThemeResponse | ErrorResponse;
}>('/api/v1/theme/:type', async (request, reply) => {
  const { type } = request.params;
  if (!isThemeType(type)) {
    return reply.code(400).send({
      error: 'invalid_type',
      message: 'Tipo de daltonismo no soportado',
    });
  }
  return {
    type,
    palette: palettes[type],
  };
});

const port = Number.parseInt(process.env.EYEX_PORT || '8080', 10);
if (!Number.isInteger(port) || port < 1 || port > 65535) {
  throw new Error('EYEX_PORT must be an integer between 1 and 65535');
}

await app.listen({ port, host: '0.0.0.0' });
