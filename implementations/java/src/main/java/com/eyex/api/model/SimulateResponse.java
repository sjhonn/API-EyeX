package com.eyex.api.model;

public record SimulateResponse(
        String original,
        String simulated,
        String type,
        double severity,
        String model
) {}
