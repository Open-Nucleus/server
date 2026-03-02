from __future__ import annotations

import asyncio
import logging
import signal

import grpc

from sentinel.config import load_config
from sentinel.gen.sentinel.v1 import sentinel_pb2_grpc
from sentinel.http.health_server import HealthServer
from sentinel.ollama.sidecar import OllamaSidecar
from sentinel.server.servicer import SentinelServiceServicer
from sentinel.store.seed import seed_stores
from sentinel.sync_subscriber import SyncEventSubscriber
from sentinel.agent.stub import StubAgent

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
)
logger = logging.getLogger("sentinel")


async def main() -> None:
    config = load_config()
    logger.info("Sentinel Agent starting (grpc=%d, http=%d)", config.grpc_port, config.http_port)

    # 1. Create and seed stores
    alert_store, inventory_store = seed_stores()
    logger.info("Stores seeded: %d alerts, inventory ready", alert_store.summary().total)

    # 2. Start Ollama sidecar (if enabled)
    ollama = OllamaSidecar(config.ollama)
    await ollama.start()

    # 3. Create stub agent
    agent = StubAgent(alert_store)

    # 4. Start gRPC server on :50056
    grpc_server = grpc.aio.server()
    servicer = SentinelServiceServicer(alert_store, inventory_store)
    sentinel_pb2_grpc.add_SentinelServiceServicer_to_server(servicer, grpc_server)
    grpc_server.add_insecure_port(f"[::]:{config.grpc_port}")
    await grpc_server.start()
    logger.info("gRPC server listening on :%d", config.grpc_port)

    # 5. Start HTTP management server on :8090
    http_server = HealthServer(
        alert_store=alert_store,
        inventory_store=inventory_store,
        ollama_sidecar=ollama,
        port=config.http_port,
        skills=config.skills,
    )
    await http_server.serve()
    logger.info("HTTP management server listening on :%d", config.http_port)

    # 6. Start background tasks
    sync_sub = SyncEventSubscriber(config.sync_grpc_address)
    tasks = [
        asyncio.create_task(ollama.watchdog_loop()),
        asyncio.create_task(sync_sub.subscribe_loop()),
        asyncio.create_task(agent.run()),
    ]

    # 7. Graceful shutdown
    stop = asyncio.Event()

    def _signal_handler():
        logger.info("Shutdown signal received")
        stop.set()

    loop = asyncio.get_running_loop()
    for sig in (signal.SIGINT, signal.SIGTERM):
        loop.add_signal_handler(sig, _signal_handler)

    logger.info("Sentinel Agent ready")
    await stop.wait()

    # Cleanup
    logger.info("Shutting down...")
    for t in tasks:
        t.cancel()
    await grpc_server.stop(grace=5)
    await ollama.stop()
    logger.info("Sentinel Agent stopped")


if __name__ == "__main__":
    asyncio.run(main())
