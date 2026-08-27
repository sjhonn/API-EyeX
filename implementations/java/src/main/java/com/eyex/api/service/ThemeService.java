package com.eyex.api.service;

import com.eyex.api.model.CustomThemeRequest;
import com.eyex.api.model.Palette;
import com.eyex.api.model.QuickTestAnswers;
import com.eyex.api.model.ThemeResponse;
import org.springframework.stereotype.Service;

import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.regex.Pattern;

@Service
public class ThemeService {
    private static final List<String> TYPES = List.of("normal", "protanopia", "deuteranopia", "tritanopia", "achromatopsia", "low_vision");
    private static final Pattern HEX = Pattern.compile("^#[0-9A-Fa-f]{6}$");

    private final Map<String, Palette> legacy = new LinkedHashMap<>();
    private final Map<String, Map<String, Palette>> safe = new LinkedHashMap<>();

    public ThemeService() {
        legacy.put("normal", p("#F4F5F7", "#FFFFFF", "#20252B", "#2E6DA4", "#6B7785", "#C94C4C", "#3C8D5A"));
        legacy.put("protanopia", p("#1E1E1E", "#2A2A2A", "#F5F5F5", "#3F8FD2", "#E3B341", "#D96C3F", "#4FB3A5"));
        legacy.put("deuteranopia", p("#1E1E1E", "#2A2A2A", "#F5F5F5", "#4A90D9", "#D9A24A", "#D94A4A", "#4AD98C"));
        legacy.put("tritanopia", p("#202124", "#2D2F33", "#F5F5F5", "#D65DB1", "#4CC9A7", "#E05A47", "#64A66F"));
        legacy.put("achromatopsia", p("#202020", "#303030", "#F2F2F2", "#D0D0D0", "#A8A8A8", "#E0E0E0", "#BEBEBE"));

        safe.put("normal", modes(
                legacy.get("normal"),
                p("#181A1D", "#24272B", "#F5F7FA", "#5CA9E6", "#AAB4BE", "#FF7B72", "#56D364")
        ));
        safe.put("protanopia", modes(
                p("#F7F8FA", "#FFFFFF", "#1D2329", "#256EA6", "#916B00", "#A84824", "#237A70"),
                legacy.get("protanopia")
        ));
        safe.put("deuteranopia", modes(
                p("#F7F8FA", "#FFFFFF", "#1D2329", "#236FAE", "#8A6200", "#A83D3D", "#187A55"),
                legacy.get("deuteranopia")
        ));
        safe.put("tritanopia", modes(
                p("#F7F7F8", "#FFFFFF", "#202124", "#9B3F80", "#167A65", "#AA4234", "#347A42"),
                legacy.get("tritanopia")
        ));
        safe.put("achromatopsia", modes(
                p("#FAFAFA", "#FFFFFF", "#181818", "#4A4A4A", "#666666", "#303030", "#555555"),
                legacy.get("achromatopsia")
        ));
        safe.put("low_vision", modes(
                p("#FFFFFF", "#F2F2F2", "#000000", "#005FCC", "#6D4C00", "#A80000", "#006B35"),
                p("#000000", "#121212", "#FFFFFF", "#66B2FF", "#FFD166", "#FF6B6B", "#65E6A3")
        ));
    }

    public List<String> types() { return TYPES; }
    public boolean supported(String type) { return TYPES.contains(type); }

    public Optional<ThemeResponse> get(String type, String severityRaw, String modeRaw, Boolean highContrastRaw, boolean explicit) {
        if (!supported(type)) return Optional.empty();
        validateOptions(severityRaw, modeRaw);
        if (!explicit && !type.equals("low_vision")) return Optional.of(response(type, legacy.get(type)));
        String mode = blank(modeRaw) ? (type.equals("normal") ? "light" : "dark") : modeRaw;
        String severity = blank(severityRaw) ? "moderate" : severityRaw;
        Palette palette = type.equals("normal") ? safe.get("normal").get(mode) : mixPalette(safe.get("normal").get(mode), safe.get(type).get(mode), severityFactor(severity));
        if (Boolean.TRUE.equals(highContrastRaw) || type.equals("low_vision")) palette = highContrast(palette, type, mode);
        return Optional.of(response(type, ensureTextContrast(palette)));
    }

    public ThemeResponse custom(CustomThemeRequest req) {
        if (req == null || !supported(req.type())) throw new IllegalArgumentException("invalid_type");
        validatePalette(req.palette());
        validateOptions(req.severity(), req.mode());
        String mode = blank(req.mode()) ? (luminance(req.palette().background()) < 0.35 ? "dark" : "light") : req.mode();
        String severity = blank(req.severity()) ? "moderate" : req.severity();
        Palette palette = req.palette();
        double f = severityFactor(severity);
        if (req.type().equals("achromatopsia")) {
            palette = new Palette(
                    mixHex(palette.background(), grayscale(palette.background()), f),
                    mixHex(palette.surface(), grayscale(palette.surface()), f),
                    mixHex(palette.text(), grayscale(palette.text()), f),
                    mixHex(palette.primary(), grayscale(palette.primary()), f),
                    mixHex(palette.secondary(), grayscale(palette.secondary()), f),
                    mixHex(palette.error(), grayscale(palette.error()), f),
                    mixHex(palette.success(), grayscale(palette.success()), f)
            );
        } else if (!req.type().equals("normal")) {
            Palette anchor = safe.get(req.type()).get(mode);
            palette = new Palette(
                    palette.background(), palette.surface(), palette.text(),
                    mixHex(palette.primary(), anchor.primary(), f), mixHex(palette.secondary(), anchor.secondary(), f),
                    mixHex(palette.error(), anchor.error(), f), mixHex(palette.success(), anchor.success(), f)
            );
        }
        palette = adaptMode(palette, mode);
        if (Boolean.TRUE.equals(req.highContrast()) || req.type().equals("low_vision")) palette = highContrast(palette, req.type(), mode);
        return response(req.type(), ensureTextContrast(palette));
    }

    public String suggest(QuickTestAnswers a) {
        if (a == null) return "normal";
        if (a.colorsLookGray()) return "achromatopsia";
        if (a.blueYellowConfusion()) return "tritanopia";
        if (a.redsLookDarker() && a.greenBrownConfusion()) return "protanopia";
        if (a.greenBrownConfusion()) return "deuteranopia";
        if (a.redsLookDarker()) return "protanopia";
        return "normal";
    }

    private void validateOptions(String severity, String mode) {
        if (!blank(severity) && !List.of("mild", "moderate", "severe").contains(severity)) throw new IllegalArgumentException("severity debe ser mild, moderate o severe");
        if (!blank(mode) && !List.of("dark", "light").contains(mode)) throw new IllegalArgumentException("mode debe ser dark o light");
    }

    private void validatePalette(Palette p) {
        if (p == null) throw new IllegalArgumentException("palette es requerido");
        Map<String, String> values = Map.of("background", p.background(), "surface", p.surface(), "text", p.text(), "primary", p.primary(), "secondary", p.secondary(), "error", p.error(), "success", p.success());
        for (var entry : values.entrySet()) if (entry.getValue() == null || !HEX.matcher(entry.getValue()).matches()) throw new IllegalArgumentException(entry.getKey() + " debe usar formato #RRGGBB");
    }

    private ThemeResponse response(String type, Palette palette) { return new ThemeResponse(type, palette, contrastOK(palette)); }
    private boolean contrastOK(Palette p) { return contrast(p.text(), p.background()) >= 4.5 && contrast(p.text(), p.surface()) >= 4.5; }
    private Palette ensureTextContrast(Palette p) {
        if (contrastOK(p)) return p;
        double white = Math.min(contrast("#FFFFFF", p.background()), contrast("#FFFFFF", p.surface()));
        double black = Math.min(contrast("#000000", p.background()), contrast("#000000", p.surface()));
        return new Palette(p.background(), p.surface(), white >= black ? "#FFFFFF" : "#000000", p.primary(), p.secondary(), p.error(), p.success());
    }

    private Palette adaptMode(Palette p, String mode) {
        if (mode.equals("dark")) {
            String bg = mixHex(p.background(), "#181A1D", 0.72), surface = mixHex(p.surface(), "#24272B", 0.72), text = p.text();
            if (contrast(text, bg) < 4.5 || contrast(text, surface) < 4.5) text = "#F5F7FA";
            return new Palette(bg, surface, text, p.primary(), p.secondary(), p.error(), p.success());
        }
        String bg = mixHex(p.background(), "#F4F6F8", 0.72), surface = mixHex(p.surface(), "#FFFFFF", 0.82), text = p.text();
        if (contrast(text, bg) < 4.5 || contrast(text, surface) < 4.5) text = "#1A1D21";
        return new Palette(bg, surface, text, p.primary(), p.secondary(), p.error(), p.success());
    }

    private Palette highContrast(Palette p, String type, String mode) {
        String anchorType = type.equals("normal") ? "low_vision" : type;
        Palette a = safe.get(anchorType).get(mode);
        String bg = mode.equals("dark") ? "#000000" : "#FFFFFF";
        String surface = mode.equals("dark") ? "#121212" : "#F2F2F2";
        String text = mode.equals("dark") ? "#FFFFFF" : "#000000";
        return new Palette(bg, surface, text, mixHex(p.primary(), a.primary(), .45), mixHex(p.secondary(), a.secondary(), .45), mixHex(p.error(), a.error(), .45), mixHex(p.success(), a.success(), .45));
    }

    private Palette mixPalette(Palette a, Palette b, double f) {
        return new Palette(mixHex(a.background(), b.background(), f), mixHex(a.surface(), b.surface(), f), mixHex(a.text(), b.text(), f), mixHex(a.primary(), b.primary(), f), mixHex(a.secondary(), b.secondary(), f), mixHex(a.error(), b.error(), f), mixHex(a.success(), b.success(), f));
    }
    private double severityFactor(String s) { return s.equals("mild") ? .35 : s.equals("severe") ? 1 : .70; }
    private double contrast(String a, String b) { double la=luminance(a), lb=luminance(b); return (Math.max(la,lb)+.05)/(Math.min(la,lb)+.05); }
    private double luminance(String hex) { int[] rgb=parseHex(hex); return .2126*linear(rgb[0])+.7152*linear(rgb[1])+.0722*linear(rgb[2]); }
    private double linear(int value) { double x=value/255.0; return x<=.04045?x/12.92:Math.pow((x+.055)/1.055,2.4); }
    private int[] parseHex(String hex) { String h=hex.substring(1); return new int[]{Integer.parseInt(h.substring(0,2),16),Integer.parseInt(h.substring(2,4),16),Integer.parseInt(h.substring(4,6),16)}; }
    private String mixHex(String a, String b, double f) { int[] x=parseHex(a),y=parseHex(b); return String.format("#%02X%02X%02X", mix(x[0],y[0],f),mix(x[1],y[1],f),mix(x[2],y[2],f)); }
    private int mix(int a,int b,double f) { return Math.max(0,Math.min(255,(int)Math.round(a*(1-f)+b*f))); }
    private String grayscale(String hex) { int[] c=parseHex(hex); int g=(int)Math.round(.2126*c[0]+.7152*c[1]+.0722*c[2]); return String.format("#%02X%02X%02X",g,g,g); }
    private static Palette p(String bg,String surface,String text,String primary,String secondary,String error,String success) { return new Palette(bg,surface,text,primary,secondary,error,success); }
    private static Map<String, Palette> modes(Palette light, Palette dark) { return Map.of("light",light,"dark",dark); }
    private static boolean blank(String value) { return value == null || value.isBlank(); }
}
