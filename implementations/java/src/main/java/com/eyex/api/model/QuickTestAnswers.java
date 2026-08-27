package com.eyex.api.model;

import com.fasterxml.jackson.annotation.JsonProperty;

public record QuickTestAnswers(
        @JsonProperty("reds_look_darker") boolean redsLookDarker,
        @JsonProperty("green_brown_confusion") boolean greenBrownConfusion,
        @JsonProperty("blue_yellow_confusion") boolean blueYellowConfusion,
        @JsonProperty("colors_look_gray") boolean colorsLookGray
) {}
