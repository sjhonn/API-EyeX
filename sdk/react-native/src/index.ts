export type EyeXType = 'normal' | 'protanopia' | 'deuteranopia' | 'tritanopia' | 'achromatopsia' | 'low_vision';
export type EyeXSeverity = 'mild' | 'moderate' | 'severe';
export type EyeXMode = 'dark' | 'light';
export interface EyeXPalette { background:string; surface:string; text:string; primary:string; secondary:string; error:string; success:string; }
export interface EyeXTheme { type:EyeXType; palette:EyeXPalette; contrast_ok:boolean; }
export interface EyeXOptions { severity?:EyeXSeverity; mode?:EyeXMode; highContrast?:boolean; }
export interface EyeXCustomRequest extends EyeXOptions { type:EyeXType; palette:EyeXPalette; }
export interface EyeXTestAnswers { reds_look_darker:boolean; green_brown_confusion:boolean; blue_yellow_confusion:boolean; colors_look_gray:boolean; }
export interface EyeXTestResult { suggested_type:EyeXType; disclaimer:string; }
export type EyeXSimulationType = 'protanopia' | 'deuteranopia' | 'tritanopia';
export interface EyeXSimulation { original:string; simulated:string; type:EyeXSimulationType; severity:number; model:string; }
export interface EyeXSimulatedColor { original:string; simulated:string; }
export interface EyeXBatchSimulation { type:EyeXSimulationType; severity:number; model:string; results:EyeXSimulatedColor[]; }

export class EyeXClient {
  constructor(private readonly baseUrl: string) {}
  private async json<T>(path:string,init?:RequestInit):Promise<T>{const response=await fetch(`${this.baseUrl.replace(/\/$/,'')}${path}`,{...init,headers:{Accept:'application/json',...(init?.body?{'Content-Type':'application/json'}:{}),...(init?.headers||{})}});const data=await response.json();if(!response.ok)throw new Error(data.message||data.error||`HTTP ${response.status}`);return data as T;}
  async types():Promise<EyeXType[]>{const data=await this.json<{types:EyeXType[]}>('/api/v1/theme/types');return data.types;}
  async theme(type: EyeXType, options: EyeXOptions = {}): Promise<EyeXTheme> {
    const params: string[] = [];
    if (options.severity) params.push(`severity=${encodeURIComponent(options.severity)}`);
    if (options.mode) params.push(`mode=${encodeURIComponent(options.mode)}`);
    if (typeof options.highContrast === 'boolean') params.push(`high_contrast=${options.highContrast}`);
    return this.json<EyeXTheme>(`/api/v1/theme/${encodeURIComponent(type)}${params.length ? `?${params.join('&')}` : ''}`);
  }
  async custom(request:EyeXCustomRequest):Promise<EyeXTheme>{return this.json<EyeXTheme>('/api/v1/theme/custom',{method:'POST',body:JSON.stringify({type:request.type,palette:request.palette,severity:request.severity,mode:request.mode,high_contrast:request.highContrast})});}
  async simulate(hex: string, type: EyeXSimulationType, severity = 1): Promise<EyeXSimulation> {
    return this.json<EyeXSimulation>('/api/v1/simulate',{method:'POST',body:JSON.stringify({hex,type,severity})});
  }
  async simulateBatch(colors: string[], type: EyeXSimulationType, severity = 1): Promise<EyeXBatchSimulation> {
    return this.json<EyeXBatchSimulation>('/api/v1/simulate/batch',{method:'POST',body:JSON.stringify({colors,type,severity})});
  }
  async suggest(answers:EyeXTestAnswers):Promise<EyeXTestResult>{return this.json<EyeXTestResult>('/api/v1/test/suggest',{method:'POST',body:JSON.stringify({answers})});}
}
