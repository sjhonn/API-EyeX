package com.eyex.api.model;

import java.util.List;

public record SimulateBatchRequest(List<String> colors, String type, Double severity) {}
