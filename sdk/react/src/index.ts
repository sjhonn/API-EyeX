import { useCallback, useEffect, useState } from 'react';

export type EyeXType = 'normal' | 'protanopia' | 'deuteranopia' | 'tritanopia' | 'achromatopsia' | 'low_vision';
export type EyeXSeverity = 'mild' | 'moderate' | 'severe';
export type EyeXMode = 'dark' | 'light';
export interface EyeXPalette { background:string; surface:string; text:string; primary:string; secondary:string; error:string; success:string; }
export interface EyeXTheme { type:EyeXType; palette:EyeXPalette; contrast_ok:boolean; }
export interface EyeXOptions { severity?:EyeXSeverity; mode?:EyeXMode; highContrast?:boolean; }
export interface EyeXCustomRequest extends EyeXOptions { type:EyeXType; palette:EyeXPalette; }
export interface EyeXTestAnswers { reds_look_darker:boolean; green_brown_confusion:boolean; blue_yellow_confusion:boolean; colors_look_gray:boolean; }
export interface EyeXTestResult { suggested_type:EyeXType; disclaimer:string; }

export class EyeXClient {
  constructor(public readonly baseUrl: string) {}
  private url(path:string):string { return `${this.baseUrl.replace(/\/$/, '')}${path}`; }
  private async json<T>(path:string, init?:RequestInit):Promise<T>{
    const response=await fetch(this.url(path),{...init,headers:{Accept:'application/json',...(init?.body?{'Content-Type':'application/json'}:{}),...(init?.headers||{})}});
    const data=await response.json();
    if(!response.ok)throw new Error(data.message||data.error||`HTTP ${response.status}`);
    return data as T;
  }
  async types():Promise<EyeXType[]>{ const data=await this.json<{types:EyeXType[]}>('/api/v1/theme/types'); return data.types; }
  async theme(type: EyeXType, options: EyeXOptions = {}): Promise<EyeXTheme> {
    const q = new URLSearchParams();
    if (options.severity) q.set('severity', options.severity);
    if (options.mode) q.set('mode', options.mode);
    if (typeof options.highContrast === 'boolean') q.set('high_contrast', String(options.highContrast));
    return this.json<EyeXTheme>(`/api/v1/theme/${encodeURIComponent(type)}${q.size ? `?${q}` : ''}`);
  }
  async custom(request:EyeXCustomRequest):Promise<EyeXTheme>{
    return this.json<EyeXTheme>('/api/v1/theme/custom',{method:'POST',body:JSON.stringify({type:request.type,palette:request.palette,severity:request.severity,mode:request.mode,high_contrast:request.highContrast})});
  }
  async suggest(answers:EyeXTestAnswers):Promise<EyeXTestResult>{
    return this.json<EyeXTestResult>('/api/v1/test/suggest',{method:'POST',body:JSON.stringify({answers})});
  }
}

export function applyEyeXPalette(palette: EyeXPalette, root: HTMLElement = document.documentElement): void {
  Object.entries(palette).forEach(([key, value]) => root.style.setProperty(`--eyex-${key}`, value));
}

export function useEyeXTheme(client: EyeXClient, initialType: EyeXType = 'normal', initialOptions: EyeXOptions = {}) {
  const [type, setType] = useState<EyeXType>(initialType);
  const [options, setOptions] = useState<EyeXOptions>(initialOptions);
  const [theme, setTheme] = useState<EyeXTheme | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const [loading, setLoading] = useState(false);
  const refresh = useCallback(async () => {
    setLoading(true); setError(null);
    try { const next = await client.theme(type, options); setTheme(next); applyEyeXPalette(next.palette); }
    catch (cause) { setError(cause instanceof Error ? cause : new Error('EyeX request failed')); }
    finally { setLoading(false); }
  }, [client, type, options]);
  useEffect(() => { void refresh(); }, [refresh]);
  return { type, setType, options, setOptions, theme, loading, error, refresh };
}
