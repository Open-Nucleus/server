from __future__ import annotations

import grpc
import pytest

from sentinel.gen.sentinel.v1 import sentinel_pb2_grpc
from sentinel.server.servicer import SentinelServiceServicer
from sentinel.store.alert_store import AlertStore
from sentinel.store.inventory_store import InventoryStore
from sentinel.store.seed import seed_stores


@pytest.fixture
def alert_store() -> AlertStore:
    store, _ = seed_stores()
    return store


@pytest.fixture
def inventory_store() -> InventoryStore:
    _, store = seed_stores()
    return store


@pytest.fixture
def empty_alert_store() -> AlertStore:
    return AlertStore()


@pytest.fixture
def empty_inventory_store() -> InventoryStore:
    return InventoryStore()


@pytest.fixture
async def grpc_channel(alert_store, inventory_store):
    """Create an in-process gRPC server and channel for testing."""
    server = grpc.aio.server()
    servicer = SentinelServiceServicer(alert_store, inventory_store)
    sentinel_pb2_grpc.add_SentinelServiceServicer_to_server(servicer, server)
    port = server.add_insecure_port("[::]:0")
    await server.start()
    channel = grpc.aio.insecure_channel(f"localhost:{port}")
    yield channel
    await channel.close()
    await server.stop(grace=0)


@pytest.fixture
async def grpc_stub(grpc_channel):
    return sentinel_pb2_grpc.SentinelServiceStub(grpc_channel)
