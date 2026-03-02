from __future__ import annotations

import asyncio
import logging
import os
from datetime import datetime, timezone

from sentinel.config import OllamaConfig

logger = logging.getLogger(__name__)


class OllamaSidecar:
    def __init__(self, config: OllamaConfig) -> None:
        self.config = config
        self.process: asyncio.subprocess.Process | None = None
        self._restart_count = 0
        self._last_restart: datetime | None = None

    async def start(self) -> None:
        if not self.config.enabled:
            return
        if await self._is_running():
            return

        self.process = await asyncio.create_subprocess_exec(
            "ollama", "serve",
            env={
                **os.environ,
                "OLLAMA_HOST": f"127.0.0.1:{self.config.port}",
                "OLLAMA_NUM_PARALLEL": "1",
                "OLLAMA_MAX_LOADED_MODELS": "1",
            },
            stdout=asyncio.subprocess.DEVNULL,
            stderr=asyncio.subprocess.DEVNULL,
        )
        logger.info("Ollama sidecar started on port %d", self.config.port)

    async def stop(self) -> None:
        if self.process is not None:
            self.process.terminate()
            try:
                await asyncio.wait_for(self.process.wait(), timeout=10)
            except asyncio.TimeoutError:
                self.process.kill()
            self.process = None
            logger.info("Ollama sidecar stopped")

    async def health(self) -> bool:
        return await self._is_running()

    async def _is_running(self) -> bool:
        if self.process is None:
            return False
        return self.process.returncode is None

    async def watchdog_loop(self) -> None:
        while True:
            await asyncio.sleep(30)

            if not self.config.enabled:
                continue

            if not await self._is_running():
                if self._restart_count >= self.config.max_restarts:
                    logger.error(
                        "Ollama exceeded max restarts (%d). "
                        "Agent running in degraded mode.",
                        self.config.max_restarts,
                    )
                    continue

                logger.warning("Ollama not running. Attempting restart...")
                self._restart_count += 1
                self._last_restart = datetime.now(timezone.utc)

                try:
                    await self.start()
                    logger.info("Ollama restarted successfully.")
                except Exception as e:
                    logger.error("Ollama restart failed: %s", e)

    def status(self) -> dict:
        return {
            "enabled": self.config.enabled,
            "running": self.process is not None and self.process.returncode is None,
            "model": self.config.model,
            "restart_count": self._restart_count,
            "last_restart": self._last_restart.isoformat() if self._last_restart else None,
        }
