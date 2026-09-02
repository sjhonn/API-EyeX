#!/usr/bin/env python3
"""HTTP parity checks for the four EyeX backends.

Uses only the Python standard library so it can run in GitHub Actions without
extra dependencies. The simulation checks compare both cross-backend responses
and fixed reference vectors to avoid accepting four identical regressions.
"""
from __future__ import annotations

import argparse
import json
import sys
import time
from dataclasses import dataclass
from typing import Any
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen


@dataclass(frozen=True)
class Backend:
    name: str
    url: str


@dataclass(frozen=True)
class Case:
    name: str
    method: str
    path: str
    body: Any | None = None
    expected_status: int = 200
    expected_json: Any | None = None


def request(backend: Backend, case: Case) -> tuple[int, Any]:
    payload = None if case.body is None else json.dumps(case.body, separators=(",", ":")).encode()
    headers = {"Accept": "application/json"}
    if payload is not None:
        headers["Content-Type"] = "application/json"
    req = Request(backend.url.rstrip("/") + case.path, data=payload, headers=headers, method=case.method)
    try:
        with urlopen(req, timeout=8) as response:
            raw = response.read().decode("utf-8")
            return response.status, json.loads(raw)
    except HTTPError as exc:
        raw = exc.read().decode("utf-8")
        try:
            body = json.loads(raw)
        except json.JSONDecodeError:
            body = {"__raw__": raw}
        return exc.code, body


def wait_ready(backend: Backend, deadline: float) -> None:
    last: Exception | None = None
    while time.time() < deadline:
        try:
            status, _ = request(backend, Case("health", "GET", "/api/v1/theme/types"))
            if status == 200:
                return
        except (URLError, ConnectionError, TimeoutError, json.JSONDecodeError) as exc:
            last = exc
        time.sleep(0.35)
    raise RuntimeError(f"{backend.name} no estuvo disponible en {backend.url}: {last}")


def cases() -> list[Case]:
    checks = [
        Case(
            "theme-types", "GET", "/api/v1/theme/types",
            expected_json={"types": ["normal", "protanopia", "deuteranopia", "tritanopia", "achromatopsia", "low_vision"]},
        ),
        Case(
            "theme-legacy-deuteranopia", "GET", "/api/v1/theme/deuteranopia",
            expected_json={
                "type": "deuteranopia",
                "palette": {
                    "background": "#1E1E1E", "surface": "#2A2A2A", "text": "#F5F5F5",
                    "primary": "#4A90D9", "secondary": "#D9A24A", "error": "#D94A4A", "success": "#4AD98C",
                },
                "contrast_ok": True,
            },
        ),
        Case(
            "theme-options", "GET", "/api/v1/theme/protanopia?severity=mild&mode=light&high_contrast=true",
        ),
        Case(
            "custom-theme", "POST", "/api/v1/theme/custom",
            {
                "type": "deuteranopia", "severity": "moderate", "mode": "dark", "high_contrast": True,
                "palette": {
                    "background": "#101820", "surface": "#182430", "text": "#F8F9FA",
                    "primary": "#E63946", "secondary": "#2A9D8F", "error": "#D62828", "success": "#2A9D8F",
                },
            },
        ),
        Case(
            "test-suggest", "POST", "/api/v1/test/suggest",
            {"answers": {"reds_look_darker": False, "green_brown_confusion": False, "blue_yellow_confusion": True, "colors_look_gray": False}},
            expected_json={"suggested_type": "tritanopia", "disclaimer": "Resultado orientativo. No es un diagnóstico médico."},
        ),
        Case(
            "theme-invalid-type", "GET", "/api/v1/theme/no-existe", expected_status=400,
            expected_json={"error": "invalid_type", "message": "Tipo de daltonismo no soportado"},
        ),
        Case(
            "theme-invalid-parameter", "GET", "/api/v1/theme/protanopia?severity=extreme", expected_status=400,
            expected_json={"error": "invalid_parameter", "message": "severity debe ser mild, moderate o severe"},
        ),
        Case(
            "protan-065", "POST", "/api/v1/simulate",
            {"hex": "#FF0000", "type": "protanopia", "severity": 0.65},
            expected_json={"original": "#FF0000", "simulated": "#A05A00", "type": "protanopia", "severity": 0.65, "model": "machado-2009"},
        ),
        Case(
            "deutan-050", "POST", "/api/v1/simulate",
            {"hex": "#FF0000", "type": "deuteranopia", "severity": 0.5},
            expected_json={"original": "#FF0000", "simulated": "#C37600", "type": "deuteranopia", "severity": 0.5, "model": "machado-2009"},
        ),
        Case(
            "tritan-025", "POST", "/api/v1/simulate",
            {"hex": "#FF0000", "type": "tritanopia", "severity": 0.25},
            expected_json={"original": "#FF0000", "simulated": "#F42F1E", "type": "tritanopia", "severity": 0.25, "model": "machado-2009"},
        ),
        Case(
            "severity-zero", "POST", "/api/v1/simulate",
            {"hex": "#12abef", "type": "protanopia", "severity": 0},
            expected_json={"original": "#12ABEF", "simulated": "#12ABEF", "type": "protanopia", "severity": 0, "model": "machado-2009"},
        ),
        Case(
            "severity-default", "POST", "/api/v1/simulate",
            {"hex": "#FF0000", "type": "protanopia"},
            expected_json={"original": "#FF0000", "simulated": "#6D5F00", "type": "protanopia", "severity": 1, "model": "machado-2009"},
        ),
        Case(
            "batch", "POST", "/api/v1/simulate/batch",
            {"colors": ["#FF0000", "#00FF00", "#0000FF"], "type": "deuteranopia", "severity": 0.5},
            expected_json={
                "type": "deuteranopia", "severity": 0.5, "model": "machado-2009",
                "results": [
                    {"original": "#FF0000", "simulated": "#C37600"},
                    {"original": "#00FF00", "simulated": "#CDE52E"},
                    {"original": "#0000FF", "simulated": "#0036FD"},
                ],
            },
        ),
        Case(
            "invalid-type", "POST", "/api/v1/simulate",
            {"hex": "#FF0000", "type": "achromatopsia", "severity": 0.5}, 400,
            {"error": "invalid_type", "message": "Tipo de daltonismo no soportado"},
        ),
        Case(
            "invalid-severity", "POST", "/api/v1/simulate",
            {"hex": "#FF0000", "type": "protanopia", "severity": 1.1}, 400,
            {"error": "invalid_parameter", "message": "severity debe estar entre 0 y 1"},
        ),
        Case(
            "invalid-color", "POST", "/api/v1/simulate",
            {"hex": "red", "type": "protanopia", "severity": 0.5}, 400,
            {"error": "invalid_color", "message": "hex debe usar formato #RRGGBB"},
        ),
        Case(
            "empty-batch", "POST", "/api/v1/simulate/batch",
            {"colors": [], "type": "protanopia", "severity": 0.5}, 400,
            {"error": "invalid_request", "message": "colors debe contener entre 1 y 256 colores"},
        ),
        Case(
            "batch-colors-must-be-array", "POST", "/api/v1/simulate/batch",
            {"colors": {"first": "#FF0000"}, "type": "protanopia", "severity": 0.5}, 400,
            {"error": "invalid_request", "message": "JSON de entrada inválido"},
        ),
        Case(
            "unknown-field", "POST", "/api/v1/simulate",
            {"hex": "#FF0000", "type": "protanopia", "severity": 0.5, "unexpected": True}, 400,
            {"error": "invalid_request", "message": "JSON de entrada inválido"},
        ),
    ]

    # Exercise every 0.1 matrix anchor in each Machado family. These cases use
    # Go as the HTTP reference while the fixed vectors above independently pin
    # representative expected results and interpolated severities.
    colors = ["#336699", "#D97A24", "#2EC4B6"]
    for type_name in ("protanopia", "deuteranopia", "tritanopia"):
        for step in range(11):
            checks.append(Case(
                f"matrix-{type_name}-{step:02d}",
                "POST",
                "/api/v1/simulate",
                {"hex": colors[step % len(colors)], "type": type_name, "severity": step / 10},
            ))
    return checks


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--go", default="http://127.0.0.1:18080")
    parser.add_argument("--php", default="http://127.0.0.1:18081")
    parser.add_argument("--typescript", default="http://127.0.0.1:18082")
    parser.add_argument("--java", default="http://127.0.0.1:18083")
    parser.add_argument("--wait", type=float, default=75.0)
    args = parser.parse_args()
    backends = [Backend("go", args.go), Backend("php", args.php), Backend("typescript", args.typescript), Backend("java", args.java)]

    deadline = time.time() + args.wait
    for backend in backends:
        wait_ready(backend, deadline)

    failures: list[str] = []
    for case in cases():
        values: dict[str, tuple[int, Any]] = {}
        for backend in backends:
            try:
                values[backend.name] = request(backend, case)
            except Exception as exc:  # makes the CI report name/backend instead of a raw stack only
                failures.append(f"{case.name}/{backend.name}: request failed: {exc}")
                continue
        if len(values) != len(backends):
            continue

        reference = values[backends[0].name]
        for backend in backends:
            status, payload = values[backend.name]
            if status != case.expected_status:
                failures.append(f"{case.name}/{backend.name}: status={status}, esperado={case.expected_status}; body={payload}")
            if (status, payload) != reference:
                failures.append(f"{case.name}/{backend.name}: difiere de go; got={(status, payload)!r}; go={reference!r}")
            if case.expected_json is not None and payload != case.expected_json:
                failures.append(f"{case.name}/{backend.name}: vector esperado distinto; got={payload!r}; expected={case.expected_json!r}")
        print(f"OK {case.name}: status={reference[0]}")

    if failures:
        print("\nPARITY FAILED", file=sys.stderr)
        for failure in failures:
            print("-", failure, file=sys.stderr)
        return 1
    print(f"\nParity OK: {len(cases())} casos x {len(backends)} backends")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
