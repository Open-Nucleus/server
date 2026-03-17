import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/config/app_config.dart';
import '../../../core/constants/api_paths.dart';

/// Connection states for the backend health check.
enum ConnectionStatus {
  unknown,
  checking,
  connected,
  disconnected,
}

/// State object for the connection provider.
class ConnectionState {
  final ConnectionStatus status;
  final String? nodeId;
  final String? siteId;

  const ConnectionState({
    this.status = ConnectionStatus.unknown,
    this.nodeId,
    this.siteId,
  });

  ConnectionState copyWith({
    ConnectionStatus? status,
    String? nodeId,
    String? siteId,
  }) {
    return ConnectionState(
      status: status ?? this.status,
      nodeId: nodeId ?? this.nodeId,
      siteId: siteId ?? this.siteId,
    );
  }
}

/// Notifier that polls the backend `/health` endpoint every 10 seconds
/// and exposes connection status, node ID, and site ID.
class ConnectionNotifier extends StateNotifier<ConnectionState> {
  ConnectionNotifier({AppConfig? config})
      : _config = config ?? const AppConfig(),
        super(const ConnectionState()) {
    _dio = Dio(BaseOptions(
      baseUrl: _config.serverUrl,
      connectTimeout: const Duration(seconds: 5),
      receiveTimeout: const Duration(seconds: 5),
    ));
    // Perform an immediate check, then start periodic polling.
    _check();
    _timer = Timer.periodic(_config.connectionPollInterval, (_) => _check());
  }

  final AppConfig _config;
  late final Dio _dio;
  Timer? _timer;

  Future<void> _check() async {
    state = state.copyWith(status: ConnectionStatus.checking);
    try {
      final response = await _dio.get(ApiPaths.health);
      if (response.statusCode == 200) {
        final data = response.data;
        String? nodeId;
        String? siteId;
        if (data is Map<String, dynamic>) {
          nodeId = data['node_id'] as String?;
          siteId = data['site_id'] as String?;
        }
        state = ConnectionState(
          status: ConnectionStatus.connected,
          nodeId: nodeId ?? state.nodeId,
          siteId: siteId ?? state.siteId,
        );
      } else {
        state = state.copyWith(status: ConnectionStatus.disconnected);
      }
    } catch (_) {
      state = state.copyWith(status: ConnectionStatus.disconnected);
    }
  }

  /// Force an immediate health check.
  Future<void> refresh() => _check();

  @override
  void dispose() {
    _timer?.cancel();
    _dio.close();
    super.dispose();
  }
}

/// Provider for connection state with automatic polling.
final connectionProvider =
    StateNotifierProvider<ConnectionNotifier, ConnectionState>(
  (ref) => ConnectionNotifier(),
);
