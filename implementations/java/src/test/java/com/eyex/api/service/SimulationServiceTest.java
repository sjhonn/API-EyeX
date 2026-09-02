package com.eyex.api.service;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;

class SimulationServiceTest {
    private final SimulationService service = new SimulationService();

    @Test
    void interpolatesContinuousSeverity() {
        assertEquals("#A05A00", service.simulate("#FF0000", "protanopia", 0.65));
        assertEquals("#C37600", service.simulate("#FF0000", "deuteranopia", 0.50));
        assertEquals("#F42F1E", service.simulate("#FF0000", "tritanopia", 0.25));
    }

    @Test
    void zeroSeverityIsIdentity() {
        assertEquals("#12ABEF", service.simulate("#12ABEF", "protanopia", 0.0));
    }

    @Test
    void validatesSeverity() {
        assertThrows(IllegalArgumentException.class, () -> service.severityOrDefault(1.01));
        assertEquals(1.0, service.severityOrDefault(null));
    }
}
