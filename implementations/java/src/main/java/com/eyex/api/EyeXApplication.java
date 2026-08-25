package com.eyex.api;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.HashMap;
import java.util.Map;

@SpringBootApplication
public class EyeXApplication {
    public static void main(String[] args) {
        SpringApplication application = new SpringApplication(EyeXApplication.class);
        application.setDefaultProperties(loadDotEnv());
        application.run(args);
    }

    private static Map<String, Object> loadDotEnv() {
        Map<String, Object> values = new HashMap<>();
        Path current = Paths.get("").toAbsolutePath();

        for (int i = 0; i < 5 && current != null; i++, current = current.getParent()) {
            Path candidate = current.resolve(".env");
            if (!Files.isRegularFile(candidate)) {
                continue;
            }
            try {
                for (String rawLine : Files.readAllLines(candidate)) {
                    String line = rawLine.trim();
                    if (line.isEmpty() || line.startsWith("#") || !line.contains("=")) {
                        continue;
                    }
                    String[] parts = line.split("=", 2);
                    String key = parts[0].trim();
                    String value = parts[1].trim().replaceAll("^[\\\"']|[\\\"']$", "");
                    if (!key.isEmpty()) {
                        values.putIfAbsent(key, value);
                    }
                }
                break;
            } catch (IOException ignored) {
                break;
            }
        }
        return values;
    }
}
