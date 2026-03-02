from __future__ import annotations

from abc import ABC, abstractmethod
from typing import Any


class SentinelSkill(ABC):
    """Abstract base class for sentinel surveillance skills."""

    @abstractmethod
    def name(self) -> str: ...

    @abstractmethod
    def category(self) -> str: ...

    @abstractmethod
    def priority(self) -> int: ...

    @abstractmethod
    async def run(self, data_adapter: DataAdapter) -> list[Any]: ...


class DataAdapter(ABC):
    """Abstract interface for reading clinical data."""

    @abstractmethod
    async def query(self, resource_type: str, filters: dict) -> list[dict]: ...

    @abstractmethod
    async def aggregate(
        self, resource_type: str, group_by: list[str], agg: str, filters: dict
    ) -> list[dict]: ...

    @abstractmethod
    async def count(self, resource_type: str, filters: dict) -> int: ...


class AlertOutput(ABC):
    """Abstract interface for emitting alerts."""

    @abstractmethod
    def name(self) -> str: ...

    @abstractmethod
    def accepts(self, alert: Any) -> bool: ...

    @abstractmethod
    async def emit(self, alert: Any) -> bool: ...


class MemoryStore(ABC):
    """Abstract interface for agent memory (episodic, semantic, procedural)."""

    @abstractmethod
    async def store_episode(self, episode: dict) -> None: ...

    @abstractmethod
    async def get_episodes(self, limit: int = 50) -> list[dict]: ...

    @abstractmethod
    async def get_baselines(self, skill: str = "", site: str = "") -> list[dict]: ...

    @abstractmethod
    async def update_baseline(self, key: str, value: dict) -> None: ...


class LLMEngine(ABC):
    """Abstract interface for LLM inference."""

    @abstractmethod
    async def generate(self, prompt: str, context: dict | None = None) -> str: ...

    @abstractmethod
    async def health(self) -> bool: ...
