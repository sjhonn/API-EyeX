export type EyeXType = "normal" | "protanopia" | "deuteranopia" | "tritanopia" | "achromatopsia" | "low_vision";
export type EyeXSimulationType = "protanopia" | "deuteranopia" | "tritanopia";
export type EyeXSeverity = "mild" | "moderate" | "severe";
export type EyeXMode = "dark" | "light";
export interface EyeXPalette {
    background: string;
    surface: string;
    text: string;
    primary: string;
    secondary: string;
    error: string;
    success: string;
}
export interface EyeXTheme {
    type: EyeXType;
    palette: EyeXPalette;
    contrast_ok: boolean;
}
export interface EyeXOptions {
    severity?: EyeXSeverity;
    mode?: EyeXMode;
    highContrast?: boolean;
}
export interface EyeXCustomRequest extends EyeXOptions {
    type: EyeXType;
    palette: EyeXPalette;
}
export interface EyeXTestAnswers {
    reds_look_darker: boolean;
    green_brown_confusion: boolean;
    blue_yellow_confusion: boolean;
    colors_look_gray: boolean;
}
export interface EyeXTestResult {
    suggested_type: EyeXType;
    disclaimer: string;
}
export interface EyeXSimulation {
    original: string;
    simulated: string;
    type: EyeXSimulationType;
    severity: number;
    model: string;
}
export interface EyeXSimulatedColor {
    original: string;
    simulated: string;
}
export interface EyeXBatchSimulation {
    type: EyeXSimulationType;
    severity: number;
    model: string;
    results: EyeXSimulatedColor[];
}
export declare class EyeXClient {
    private readonly baseURL;
    constructor(baseURL?: string);
    types(): Promise<EyeXType[]>;
    theme(type: EyeXType, options?: EyeXOptions): Promise<EyeXTheme>;
    custom(request: EyeXCustomRequest): Promise<EyeXTheme>;
    suggest(answers: EyeXTestAnswers): Promise<EyeXTestResult>;
    simulate(hex: string, type: EyeXSimulationType, severity?: number): Promise<EyeXSimulation>;
    simulateBatch(colors: string[], type: EyeXSimulationType, severity?: number): Promise<EyeXBatchSimulation>;
    private get;
    private post;
    private read;
}
