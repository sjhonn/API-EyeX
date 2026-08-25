export type EyeXType =
  | "protanopia"
  | "protanomaly"
  | "deuteranopia"
  | "deuteranomaly"
  | "tritanopia"
  | "tritanomaly"
  | "acromatopsia";

export interface SimulateResponse {
  original: string;
  simulated: string;
}

export interface CorrectResponse {
  original: string;
  corrected: string;
  contrast_ratio: number;
}

export interface PaletteResponse {
  background: string;
  suggested_variants: string[];
}

export interface TypeInfo {
  id: EyeXType;
  name: string;
  description: string;
}

export class EyeXClient {
  constructor(private readonly baseURL = "http://localhost:8080") {}

  simulate(hex: string, type: EyeXType): Promise<SimulateResponse> {
    return this.post<SimulateResponse>("/api/v1/simulate", { hex, type });
  }

  correct(hex: string, type: EyeXType): Promise<CorrectResponse> {
    return this.post<CorrectResponse>("/api/v1/correct", { hex, type });
  }

  palette(baseHex: string, type: EyeXType): Promise<PaletteResponse> {
    return this.post<PaletteResponse>("/api/v1/palette", { base_hex: baseHex, type });
  }

  async types(): Promise<TypeInfo[]> {
    const response = await fetch(`${this.baseURL}/api/v1/types`);
    return this.read<TypeInfo[]>(response);
  }

  private async post<T>(path: string, body: unknown): Promise<T> {
    const response = await fetch(`${this.baseURL}${path}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    return this.read<T>(response);
  }

  private async read<T>(response: Response): Promise<T> {
    const payload = await response.json();
    if (!response.ok) {
      const message = payload?.details || payload?.error || `EyeX HTTP ${response.status}`;
      throw new Error(message);
    }
    return payload as T;
  }
}
