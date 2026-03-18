/// Application-level configuration for the Open Nucleus desktop client.
///
/// Holds connection settings, polling intervals, and development flags.
/// Values can be overridden at startup via environment variables or a
/// local config file; the defaults here target a developer workstation
/// talking to a local backend over self-signed TLS.
class AppConfig {
  /// Base URL of the Open Nucleus API gateway.
  final String serverUrl;

  /// Whether to accept self-signed TLS certificates.
  ///
  /// Defaults to `true` during development so the app can talk to a local
  /// backend that uses the auto-generated TLS certs from `pkg/tls`.
  /// **Must** be `false` in production builds.
  final bool acceptSelfSignedCerts;

  /// How often the app checks whether the backend is reachable.
  final Duration connectionPollInterval;

  /// How often the app polls the sync status endpoint for new data.
  final Duration syncPollInterval;

  /// How often the app polls the alerts endpoint for Sentinel notifications.
  final Duration alertPollInterval;

  const AppConfig({
    this.serverUrl = 'https://localhost:8080',
    this.acceptSelfSignedCerts = true,
    this.connectionPollInterval = const Duration(seconds: 10),
    this.syncPollInterval = const Duration(seconds: 5),
    this.alertPollInterval = const Duration(seconds: 30),
  });

  /// Creates a copy with the given fields replaced.
  AppConfig copyWith({
    String? serverUrl,
    bool? acceptSelfSignedCerts,
    Duration? connectionPollInterval,
    Duration? syncPollInterval,
    Duration? alertPollInterval,
  }) {
    return AppConfig(
      serverUrl: serverUrl ?? this.serverUrl,
      acceptSelfSignedCerts:
          acceptSelfSignedCerts ?? this.acceptSelfSignedCerts,
      connectionPollInterval:
          connectionPollInterval ?? this.connectionPollInterval,
      syncPollInterval: syncPollInterval ?? this.syncPollInterval,
      alertPollInterval: alertPollInterval ?? this.alertPollInterval,
    );
  }
}
