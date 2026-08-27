package com.eyex.api.model;

import com.fasterxml.jackson.annotation.JsonProperty;

public record ThemeResponse(
        String type,
        Palette palette,
        @JsonProperty("contrast_ok") boolean contrastOk
) {}
