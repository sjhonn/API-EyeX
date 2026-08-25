package com.eyex.api.service;

import com.eyex.api.model.Palette;
import com.eyex.api.model.ThemeResponse;
import org.springframework.stereotype.Service;

import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;

@Service
public class ThemeService {
    private static final List<String> TYPES = List.of(
            "normal",
            "protanopia",
            "deuteranopia",
            "tritanopia",
            "achromatopsia"
    );

    private final Map<String, Palette> palettes = new LinkedHashMap<>();

    public ThemeService() {
        palettes.put("normal", new Palette("#F4F5F7", "#FFFFFF", "#20252B", "#2E6DA4", "#6B7785", "#C94C4C", "#3C8D5A"));
        palettes.put("protanopia", new Palette("#1E1E1E", "#2A2A2A", "#F5F5F5", "#3F8FD2", "#E3B341", "#D96C3F", "#4FB3A5"));
        palettes.put("deuteranopia", new Palette("#1E1E1E", "#2A2A2A", "#F5F5F5", "#4A90D9", "#D9A24A", "#D94A4A", "#4AD98C"));
        palettes.put("tritanopia", new Palette("#202124", "#2D2F33", "#F5F5F5", "#D65DB1", "#4CC9A7", "#E05A47", "#64A66F"));
        palettes.put("achromatopsia", new Palette("#202020", "#303030", "#F2F2F2", "#D0D0D0", "#A8A8A8", "#E0E0E0", "#BEBEBE"));
    }

    public List<String> types() {
        return TYPES;
    }

    public Optional<ThemeResponse> get(String type) {
        Palette palette = palettes.get(type);
        return palette == null ? Optional.empty() : Optional.of(new ThemeResponse(type, palette));
    }
}
