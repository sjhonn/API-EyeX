package com.eyex.api.web;

import com.eyex.api.model.ErrorResponse;
import com.eyex.api.model.ThemeResponse;
import com.eyex.api.model.TypesResponse;
import com.eyex.api.service.ThemeService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api/v1/theme")
public class ThemeController {
    private final ThemeService themeService;

    public ThemeController(ThemeService themeService) {
        this.themeService = themeService;
    }

    @GetMapping("/types")
    public TypesResponse types() {
        return new TypesResponse(themeService.types());
    }

    @GetMapping("/{type}")
    public ResponseEntity<?> theme(@PathVariable String type) {
        return themeService.get(type)
                .<ResponseEntity<?>>map(ResponseEntity::ok)
                .orElseGet(() -> ResponseEntity.badRequest().body(
                        new ErrorResponse("invalid_type", "Tipo de daltonismo no soportado")
                ));
    }
}
