from __future__ import annotations

import asyncio
import logging

logger = logging.getLogger(__name__)


class SyncEventSubscriber:
    """Connects to Sync Service gRPC stream and yields sync events.

    This is a skeleton — real implementation requires SubscribeEvents RPC
    on the Sync Service which isn't wired yet. For now, the subscriber
    logs connection attempts and retries on disconnect.
    """

    def __init__(self, sync_grpc_address: str) -> None:
        self._address = sync_grpc_address
        self._connected = False

    async def subscribe_loop(self) -> None:
        while True:
            try:
                logger.info(
                    "Attempting to connect to Sync Service at %s...",
                    self._address,
                )
                # Stub: in production this would open a gRPC stream
                # channel = grpc.aio.insecure_channel(self._address)
                # stub = SyncServiceStub(channel)
                # stream = stub.SubscribeEvents(...)
                # async for event in stream: ...
                self._connected = False
                logger.info(
                    "Sync event subscriber: stub mode — "
                    "SubscribeEvents RPC not available yet"
                )
                await asyncio.sleep(30)
            except Exception as e:
                self._connected = False
                logger.warning(
                    "Sync subscriber disconnected: %s. Retrying in 30s...", e
                )
                await asyncio.sleep(30)

    @property
    def connected(self) -> bool:
        return self._connected
