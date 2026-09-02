package com.eyex.api.model;

import java.util.List;

public record SimulateBatchResponse(
        String type,
        double severity,
        String model,
        List<SimulatedColor> results
) {}
