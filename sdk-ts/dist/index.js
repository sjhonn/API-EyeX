export class EyeXClient {
    baseURL;
    constructor(baseURL = "http://localhost:8080") {
        this.baseURL = baseURL;
    }
    simulate(hex, type) {
        return this.post("/api/v1/simulate", { hex, type });
    }
    correct(hex, type) {
        return this.post("/api/v1/correct", { hex, type });
    }
    palette(baseHex, type) {
        return this.post("/api/v1/palette", { base_hex: baseHex, type });
    }
    async types() {
        const response = await fetch(`${this.baseURL}/api/v1/types`);
        return this.read(response);
    }
    async post(path, body) {
        const response = await fetch(`${this.baseURL}${path}`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(body),
        });
        return this.read(response);
    }
    async read(response) {
        const payload = await response.json();
        if (!response.ok) {
            const message = payload?.details || payload?.error || `EyeX HTTP ${response.status}`;
            throw new Error(message);
        }
        return payload;
    }
}
