from __future__ import annotations

import os
from dataclasses import dataclass, field
from pathlib import Path

import yaml


@dataclass
class OllamaConfig:
    enabled: bool = False
    port: int = 11434
    model: str = "gemma2:2b"
    timeout: int = 120
    allow_patient_data: bool = False
    max_restarts: int = 5


@dataclass
class SentinelConfig:
    grpc_port: int = 50056
    http_port: int = 8090
    sync_grpc_address: str = "localhost:50052"
    patient_service_address: str = "localhost:50051"
    repo_path: str = "/var/lib/open-nucleus/data"
    sqlite_path: str = "/var/lib/open-nucleus/index.db"
    hardware_profile: str = "pi4_8gb"
    ollama: OllamaConfig = field(default_factory=OllamaConfig)
    skills: list[str] = field(default_factory=lambda: [
        "idsr_cholera",
        "idsr_measles",
        "stockout_prediction",
        "medication_interaction",
        "vital_sign_trend",
    ])


def load_config(path: str | None = None) -> SentinelConfig:
    if path is None:
        path = os.environ.get(
            "SENTINEL_CONFIG",
            str(Path(__file__).parent.parent.parent / "config.yaml"),
        )

    cfg = SentinelConfig()

    p = Path(path)
    if not p.exists():
        return cfg

    with open(p) as f:
        raw = yaml.safe_load(f) or {}

    for k in ("grpc_port", "http_port", "sync_grpc_address",
              "patient_service_address", "repo_path", "sqlite_path",
              "hardware_profile"):
        if k in raw:
            setattr(cfg, k, raw[k])

    if "skills" in raw and isinstance(raw["skills"], list):
        cfg.skills = raw["skills"]

    if "ollama" in raw and isinstance(raw["ollama"], dict):
        ollama_raw = raw["ollama"]
        for k in ("enabled", "port", "model", "timeout",
                   "allow_patient_data", "max_restarts"):
            if k in ollama_raw:
                setattr(cfg.ollama, k, ollama_raw[k])

    return cfg
