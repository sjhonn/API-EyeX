export type EyeXType = "protanopia" | "protanomaly" | "deuteranopia" | "deuteranomaly" | "tritanopia" | "tritanomaly" | "acromatopsia";
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
export declare class EyeXClient {
    private readonly baseURL;
    constructor(baseURL?: string);
    simulate(hex: string, type: EyeXType): Promise<SimulateResponse>;
    correct(hex: string, type: EyeXType): Promise<CorrectResponse>;
    palette(baseHex: string, type: EyeXType): Promise<PaletteResponse>;
    types(): Promise<TypeInfo[]>;
    private post;
    private read;
}
