package com.eyex.api.web;

import com.eyex.api.model.CustomThemeRequest;
import com.eyex.api.model.ErrorResponse;
import com.eyex.api.model.QuickTestRequest;
import com.eyex.api.model.QuickTestResponse;
import com.eyex.api.model.TypesResponse;
import com.eyex.api.service.ThemeService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/v1")
public class ThemeController {
    private final ThemeService themeService;

    public ThemeController(ThemeService themeService) { this.themeService = themeService; }

    @GetMapping("/theme/types")
    public TypesResponse types() { return new TypesResponse(themeService.types()); }

    @GetMapping("/theme/{type}")
    public ResponseEntity<?> theme(
            @PathVariable String type,
            @RequestParam(required = false) String severity,
            @RequestParam(required = false) String mode,
            @RequestParam(name = "high_contrast", required = false) String highContrast) {
        if (!themeService.supported(type)) return ResponseEntity.badRequest().body(new ErrorResponse("invalid_type", "Tipo de daltonismo no soportado"));
        Boolean hc = null;
        if (highContrast != null) {
            if (highContrast.equals("true") || highContrast.equals("1")) hc = true;
            else if (highContrast.equals("false") || highContrast.equals("0")) hc = false;
            else return ResponseEntity.badRequest().body(new ErrorResponse("invalid_parameter", "high_contrast debe ser true o false"));
        }
        try {
            boolean explicit = severity != null || mode != null || highContrast != null;
            return ResponseEntity.ok(themeService.get(type, severity, mode, hc, explicit).orElseThrow());
        } catch (IllegalArgumentException ex) {
            return ResponseEntity.badRequest().body(new ErrorResponse("invalid_parameter", ex.getMessage()));
        }
    }

    @PostMapping("/theme/custom")
    public ResponseEntity<?> custom(@RequestBody CustomThemeRequest request) {
        try {
            return ResponseEntity.ok(themeService.custom(request));
        } catch (IllegalArgumentException ex) {
            if ("invalid_type".equals(ex.getMessage())) return ResponseEntity.badRequest().body(new ErrorResponse("invalid_type", "Tipo de daltonismo no soportado"));
            return ResponseEntity.badRequest().body(new ErrorResponse("invalid_palette", ex.getMessage()));
        }
    }

    @PostMapping("/test/suggest")
    public QuickTestResponse suggest(@RequestBody QuickTestRequest request) {
        return new QuickTestResponse(themeService.suggest(request.answers()), "Resultado orientativo. No es un diagnóstico médico.");
    }
}
