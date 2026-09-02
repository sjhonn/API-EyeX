package com.eyex.api.web;

import com.eyex.api.model.CustomThemeRequest;
import com.eyex.api.model.ErrorResponse;
import com.eyex.api.model.QuickTestRequest;
import com.eyex.api.model.QuickTestResponse;
import com.eyex.api.model.SimulateBatchRequest;
import com.eyex.api.model.SimulateBatchResponse;
import com.eyex.api.model.SimulateRequest;
import com.eyex.api.model.SimulateResponse;
import com.eyex.api.model.SimulatedColor;
import com.eyex.api.model.TypesResponse;
import com.eyex.api.service.SimulationService;
import com.eyex.api.service.ThemeService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.ArrayList;
import java.util.List;

@RestController
@RequestMapping("/api/v1")
public class ThemeController {
    private final ThemeService themeService;
    private final SimulationService simulationService;

    public ThemeController(ThemeService themeService, SimulationService simulationService) {
        this.themeService = themeService;
        this.simulationService = simulationService;
    }

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

    @PostMapping("/simulate")
    public ResponseEntity<?> simulate(@RequestBody(required = false) SimulateRequest request) {
        if (request == null) {
            return ResponseEntity.badRequest().body(new ErrorResponse("invalid_request", "JSON de entrada inválido"));
        }
        if (!simulationService.supported(request.type())) {
            return ResponseEntity.badRequest().body(new ErrorResponse("invalid_type", "Tipo de daltonismo no soportado"));
        }
        final double severity;
        try {
            severity = simulationService.severityOrDefault(request.severity());
        } catch (IllegalArgumentException ex) {
            return ResponseEntity.badRequest().body(new ErrorResponse("invalid_parameter", ex.getMessage()));
        }
        String original = simulationService.normalizeHex(request.hex());
        if (original == null) {
            return ResponseEntity.badRequest().body(new ErrorResponse("invalid_color", "hex debe usar formato #RRGGBB"));
        }
        return ResponseEntity.ok(new SimulateResponse(
                original,
                simulationService.simulate(original, request.type(), severity),
                request.type(), severity, SimulationService.MODEL));
    }

    @PostMapping("/simulate/batch")
    public ResponseEntity<?> simulateBatch(@RequestBody(required = false) SimulateBatchRequest request) {
        if (request == null) {
            return ResponseEntity.badRequest().body(new ErrorResponse("invalid_request", "JSON de entrada inválido"));
        }
        if (request.colors() == null || request.colors().isEmpty() || request.colors().size() > 256) {
            return ResponseEntity.badRequest().body(new ErrorResponse("invalid_request", "colors debe contener entre 1 y 256 colores"));
        }
        if (!simulationService.supported(request.type())) {
            return ResponseEntity.badRequest().body(new ErrorResponse("invalid_type", "Tipo de daltonismo no soportado"));
        }
        final double severity;
        try {
            severity = simulationService.severityOrDefault(request.severity());
        } catch (IllegalArgumentException ex) {
            return ResponseEntity.badRequest().body(new ErrorResponse("invalid_parameter", ex.getMessage()));
        }
        List<SimulatedColor> results = new ArrayList<>(request.colors().size());
        for (String color : request.colors()) {
            String original = simulationService.normalizeHex(color);
            if (original == null) {
                return ResponseEntity.badRequest().body(new ErrorResponse("invalid_color", "cada color debe usar formato #RRGGBB"));
            }
            results.add(new SimulatedColor(original, simulationService.simulate(original, request.type(), severity)));
        }
        return ResponseEntity.ok(new SimulateBatchResponse(request.type(), severity, SimulationService.MODEL, results));
    }

    @PostMapping("/theme/custom")
    public ResponseEntity<?> custom(@RequestBody(required = false) CustomThemeRequest request) {
        if (request == null) return ResponseEntity.badRequest().body(new ErrorResponse("invalid_request", "JSON de entrada inválido"));
        try {
            return ResponseEntity.ok(themeService.custom(request));
        } catch (IllegalArgumentException ex) {
            if ("invalid_type".equals(ex.getMessage())) return ResponseEntity.badRequest().body(new ErrorResponse("invalid_type", "Tipo de daltonismo no soportado"));
            return ResponseEntity.badRequest().body(new ErrorResponse("invalid_palette", ex.getMessage()));
        }
    }

    @PostMapping("/test/suggest")
    public ResponseEntity<?> suggest(@RequestBody(required = false) QuickTestRequest request) {
        if (request == null || request.answers() == null) {
            return ResponseEntity.badRequest().body(new ErrorResponse("invalid_request", "JSON de entrada inválido"));
        }
        return ResponseEntity.ok(new QuickTestResponse(themeService.suggest(request.answers()), "Resultado orientativo. No es un diagnóstico médico."));
    }
}
