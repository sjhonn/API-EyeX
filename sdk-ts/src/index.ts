export type EyeXType =
  | "normal"
  | "protanopia"
  | "deuteranopia"
  | "tritanopia"
  | "achromatopsia"
  | "low_vision";
export type EyeXSimulationType = "protanopia" | "deuteranopia" | "tritanopia";
export type EyeXSeverity = "mild" | "moderate" | "severe";
export type EyeXMode = "dark" | "light";

export interface EyeXPalette {
  background: string; surface: string; text: string; primary: string;
  secondary: string; error: string; success: string;
}
export interface EyeXTheme { type: EyeXType; palette: EyeXPalette; contrast_ok: boolean; }
export interface EyeXOptions { severity?: EyeXSeverity; mode?: EyeXMode; highContrast?: boolean; }
export interface EyeXCustomRequest extends EyeXOptions { type: EyeXType; palette: EyeXPalette; }
export interface EyeXTestAnswers {
  reds_look_darker: boolean; green_brown_confusion: boolean;
  blue_yellow_confusion: boolean; colors_look_gray: boolean;
}
export interface EyeXTestResult { suggested_type: EyeXType; disclaimer: string; }
export interface EyeXSimulation {
  original: string; simulated: string; type: EyeXSimulationType; severity: number; model: string;
}
export interface EyeXSimulatedColor { original: string; simulated: string; }
export interface EyeXBatchSimulation {
  type: EyeXSimulationType; severity: number; model: string; results: EyeXSimulatedColor[];
}

export class EyeXClient {
  constructor(private readonly baseURL = "http://localhost:8080") {}

  async types(): Promise<EyeXType[]> {
    const data = await this.get<{ types: EyeXType[] }>("/api/v1/theme/types");
    return data.types;
  }

  theme(type: EyeXType, options: EyeXOptions = {}): Promise<EyeXTheme> {
    const query = new URLSearchParams();
    if (options.severity) query.set("severity", options.severity);
    if (options.mode) query.set("mode", options.mode);
    if (typeof options.highContrast === "boolean") query.set("high_contrast", String(options.highContrast));
    const suffix = query.size ? `?${query}` : "";
    return this.get<EyeXTheme>(`/api/v1/theme/${encodeURIComponent(type)}${suffix}`);
  }

  custom(request: EyeXCustomRequest): Promise<EyeXTheme> {
    return this.post<EyeXTheme>("/api/v1/theme/custom", {
      type: request.type, palette: request.palette, severity: request.severity,
      mode: request.mode, high_contrast: request.highContrast,
    });
  }

  suggest(answers: EyeXTestAnswers): Promise<EyeXTestResult> {
    return this.post<EyeXTestResult>("/api/v1/test/suggest", { answers });
  }

  simulate(hex: string, type: EyeXSimulationType, severity = 1): Promise<EyeXSimulation> {
    return this.post<EyeXSimulation>("/api/v1/simulate", { hex, type, severity });
  }

  simulateBatch(colors: string[], type: EyeXSimulationType, severity = 1): Promise<EyeXBatchSimulation> {
    return this.post<EyeXBatchSimulation>("/api/v1/simulate/batch", { colors, type, severity });
  }

  private async get<T>(path: string): Promise<T> {
    const response = await fetch(`${this.baseURL.replace(/\/$/, "")}${path}`, { headers: { Accept: "application/json" } });
    return this.read<T>(response);
  }

  private async post<T>(path: string, body: unknown): Promise<T> {
    const response = await fetch(`${this.baseURL.replace(/\/$/, "")}${path}`, {
      method: "POST",
      headers: { Accept: "application/json", "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    return this.read<T>(response);
  }

  private async read<T>(response: Response): Promise<T> {
    const payload = await response.json() as { message?: string; error?: string } & T;
    if (!response.ok) throw new Error(payload.message || payload.error || `EyeX HTTP ${response.status}`);
    return payload as T;
  }
}
