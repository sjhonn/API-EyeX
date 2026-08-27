package com.eyex.api.model;

import com.fasterxml.jackson.annotation.JsonProperty;

public record QuickTestResponse(
        @JsonProperty("suggested_type") String suggestedType,
        String disclaimer
) {}
