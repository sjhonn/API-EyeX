package com.eyex.api.model;

public record Palette(
        String background,
        String surface,
        String text,
        String primary,
        String secondary,
        String error,
        String success
) {}
