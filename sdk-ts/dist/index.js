export class EyeXClient {
    baseURL;
    constructor(baseURL = "http://localhost:8080") {
        this.baseURL = baseURL;
    }
    async types() {
        const data = await this.get("/api/v1/theme/types");
        return data.types;
    }
    theme(type, options = {}) {
        const query = new URLSearchParams();
        if (options.severity)
            query.set("severity", options.severity);
        if (options.mode)
            query.set("mode", options.mode);
        if (typeof options.highContrast === "boolean")
            query.set("high_contrast", String(options.highContrast));
        const suffix = query.size ? `?${query}` : "";
        return this.get(`/api/v1/theme/${encodeURIComponent(type)}${suffix}`);
    }
    custom(request) {
        return this.post("/api/v1/theme/custom", {
            type: request.type, palette: request.palette, severity: request.severity,
            mode: request.mode, high_contrast: request.highContrast,
        });
    }
    suggest(answers) {
        return this.post("/api/v1/test/suggest", { answers });
    }
    simulate(hex, type, severity = 1) {
        return this.post("/api/v1/simulate", { hex, type, severity });
    }
    simulateBatch(colors, type, severity = 1) {
        return this.post("/api/v1/simulate/batch", { colors, type, severity });
    }
    async get(path) {
        const response = await fetch(`${this.baseURL.replace(/\/$/, "")}${path}`, { headers: { Accept: "application/json" } });
        return this.read(response);
    }
    async post(path, body) {
        const response = await fetch(`${this.baseURL.replace(/\/$/, "")}${path}`, {
            method: "POST",
            headers: { Accept: "application/json", "Content-Type": "application/json" },
            body: JSON.stringify(body),
        });
        return this.read(response);
    }
    async read(response) {
        const payload = await response.json();
        if (!response.ok)
            throw new Error(payload.message || payload.error || `EyeX HTTP ${response.status}`);
        return payload;
    }
}
