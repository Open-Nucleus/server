import os
import tempfile

from sentinel.config import SentinelConfig, load_config


def test_default_config():
    cfg = SentinelConfig()
    assert cfg.grpc_port == 50056
    assert cfg.http_port == 8090
    assert cfg.ollama.enabled is False
    assert cfg.hardware_profile == "pi4_8gb"


def test_load_config_missing_file():
    cfg = load_config("/nonexistent/path.yaml")
    assert cfg.grpc_port == 50056


def test_load_config_from_yaml():
    yaml_content = """
grpc_port: 60000
http_port: 9999
hardware_profile: hub_16gb
ollama:
  enabled: true
  model: llama3:8b
  port: 12345
skills:
  - custom_skill_1
"""
    with tempfile.NamedTemporaryFile(mode="w", suffix=".yaml", delete=False) as f:
        f.write(yaml_content)
        f.flush()
        cfg = load_config(f.name)

    assert cfg.grpc_port == 60000
    assert cfg.http_port == 9999
    assert cfg.hardware_profile == "hub_16gb"
    assert cfg.ollama.enabled is True
    assert cfg.ollama.model == "llama3:8b"
    assert cfg.ollama.port == 12345
    assert cfg.skills == ["custom_skill_1"]
    os.unlink(f.name)


def test_load_config_partial_yaml():
    yaml_content = """
grpc_port: 55555
"""
    with tempfile.NamedTemporaryFile(mode="w", suffix=".yaml", delete=False) as f:
        f.write(yaml_content)
        f.flush()
        cfg = load_config(f.name)

    assert cfg.grpc_port == 55555
    assert cfg.http_port == 8090  # default
    assert cfg.ollama.enabled is False  # default
    os.unlink(f.name)
