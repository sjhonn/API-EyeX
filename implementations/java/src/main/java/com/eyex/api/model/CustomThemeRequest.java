package com.eyex.api.model;

import com.fasterxml.jackson.annotation.JsonProperty;

public record CustomThemeRequest(
        String type,
        Palette palette,
        String severity,
        String mode,
        @JsonProperty("high_contrast") Boolean highContrast
) {}
