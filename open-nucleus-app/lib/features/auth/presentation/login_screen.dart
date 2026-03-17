import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/constants/api_paths.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_spacing.dart';
import '../../../shared/models/auth_models.dart';
import '../../../shared/providers/dio_provider.dart';
import '../../../shared/utils/ed25519_utils.dart';
import 'auth_notifier.dart';
import 'auth_providers.dart';
import 'device_notifier.dart';

// ---------------------------------------------------------------------------
// Connection test state
// ---------------------------------------------------------------------------

enum _ConnectionStatus { untested, testing, connected, failed }

// ---------------------------------------------------------------------------
// LoginScreen
// ---------------------------------------------------------------------------

class LoginScreen extends ConsumerStatefulWidget {
  const LoginScreen({super.key});

  @override
  ConsumerState<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends ConsumerState<LoginScreen> {
  final _serverUrlController =
      TextEditingController(text: 'https://localhost:8080');
  final _practitionerIdController = TextEditingController();

  _ConnectionStatus _connectionStatus = _ConnectionStatus.untested;
  String? _errorMessage;

  @override
  void dispose() {
    _serverUrlController.dispose();
    _practitionerIdController.dispose();
    super.dispose();
  }

  // ── Connection test ────────────────────────────────────────────────

  Future<void> _testConnection() async {
    setState(() {
      _connectionStatus = _ConnectionStatus.testing;
      _errorMessage = null;
    });

    try {
      final dio = Dio(
        BaseOptions(
          baseUrl: _serverUrlController.text.trim(),
          connectTimeout: const Duration(seconds: 5),
          receiveTimeout: const Duration(seconds: 5),
        ),
      );

      final response = await dio.get(ApiPaths.health);

      if (response.statusCode == 200) {
        setState(() => _connectionStatus = _ConnectionStatus.connected);
      } else {
        setState(() => _connectionStatus = _ConnectionStatus.failed);
      }
    } catch (_) {
      setState(() => _connectionStatus = _ConnectionStatus.failed);
    }
  }

  // ── Login ──────────────────────────────────────────────────────────

  bool get _canLogin {
    final deviceState = ref.read(deviceNotifierProvider);
    return _connectionStatus == _ConnectionStatus.connected &&
        deviceState is DeviceReady &&
        _practitionerIdController.text.trim().isNotEmpty;
  }

  Future<void> _handleLogin() async {
    final deviceState = ref.read(deviceNotifierProvider);
    if (deviceState is! DeviceReady) return;

    setState(() => _errorMessage = null);

    final now = DateTime.now().toUtc().toIso8601String();
    final nonce = 'login:$now';

    final signature =
        await Ed25519Utils.sign(deviceState.keypair, nonce);

    final request = LoginRequest(
      deviceId: deviceState.fingerprint,
      publicKey: deviceState.publicKeyBase64,
      challengeResponse: ChallengeResponseDTO(
        nonce: nonce,
        signature: signature,
        timestamp: now,
      ),
      practitionerId: _practitionerIdController.text.trim(),
    );

    await ref.read(authNotifierProvider.notifier).login(
          request: request,
          keypairFingerprint: deviceState.fingerprint,
        );
  }

  // ── Build ──────────────────────────────────────────────────────────

  @override
  Widget build(BuildContext context) {
    final authState = ref.watch(authNotifierProvider);
    final deviceState = ref.watch(deviceNotifierProvider);

    // Show error from auth notifier.
    if (authState is AuthError && _errorMessage == null) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted) {
          setState(() => _errorMessage = authState.message);
        }
      });
    }

    final isLoading = authState is AuthLoading;

    return Scaffold(
      body: Container(
        decoration: const BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
            colors: [
              AppColors.primary, // navy
              AppColors.secondary, // teal
            ],
          ),
        ),
        child: Center(
          child: SizedBox(
            width: 400,
            height: 560,
            child: Card(
              elevation: 8,
              shape: RoundedRectangleBorder(
                borderRadius:
                    BorderRadius.circular(AppSpacing.borderRadiusXl),
              ),
              child: Padding(
                padding: const EdgeInsets.all(AppSpacing.lg),
                child: SingleChildScrollView(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      // ── Title ──────────────────────────────────────
                      const SizedBox(height: AppSpacing.sm),
                      Text(
                        'Open Nucleus',
                        style:
                            Theme.of(context).textTheme.headlineMedium?.copyWith(
                                  fontWeight: FontWeight.w700,
                                  color: AppColors.primary,
                                ),
                      ),
                      const SizedBox(height: AppSpacing.xs),
                      Text(
                        'Electronic Health Record',
                        style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                              color: Colors.grey[600],
                            ),
                      ),

                      const SizedBox(height: AppSpacing.lg),
                      const Divider(),
                      const SizedBox(height: AppSpacing.md),

                      // ── Server URL ─────────────────────────────────
                      TextField(
                        controller: _serverUrlController,
                        decoration: InputDecoration(
                          labelText: 'Server URL',
                          hintText: 'https://localhost:8080',
                          prefixIcon: const Icon(Icons.dns_outlined),
                          suffixIcon: SizedBox(
                            width: 110,
                            child: Padding(
                              padding:
                                  const EdgeInsets.only(right: AppSpacing.xs),
                              child: TextButton.icon(
                                onPressed: _connectionStatus ==
                                        _ConnectionStatus.testing
                                    ? null
                                    : _testConnection,
                                icon: _connectionStatus ==
                                        _ConnectionStatus.testing
                                    ? const SizedBox(
                                        width: 14,
                                        height: 14,
                                        child: CircularProgressIndicator(
                                            strokeWidth: 2),
                                      )
                                    : const Icon(Icons.wifi_find, size: 16),
                                label: const Text('Test', style: TextStyle(fontSize: 12)),
                              ),
                            ),
                          ),
                        ),
                      ),

                      const SizedBox(height: AppSpacing.sm),

                      // ── Connection status indicator ────────────────
                      _buildConnectionStatus(),

                      const SizedBox(height: AppSpacing.md),

                      // ── Keypair section ────────────────────────────
                      _buildKeypairSection(deviceState),

                      const SizedBox(height: AppSpacing.md),

                      // ── Practitioner ID ────────────────────────────
                      TextField(
                        controller: _practitionerIdController,
                        decoration: const InputDecoration(
                          labelText: 'Practitioner ID',
                          hintText: 'e.g. practitioner-001',
                          prefixIcon: Icon(Icons.badge_outlined),
                        ),
                        onChanged: (_) => setState(() {}),
                      ),

                      const SizedBox(height: AppSpacing.lg),

                      // ── Login button ───────────────────────────────
                      SizedBox(
                        width: double.infinity,
                        height: 48,
                        child: FilledButton(
                          onPressed:
                              _canLogin && !isLoading ? _handleLogin : null,
                          child: isLoading
                              ? const SizedBox(
                                  width: 20,
                                  height: 20,
                                  child: CircularProgressIndicator(
                                    strokeWidth: 2,
                                    color: Colors.white,
                                  ),
                                )
                              : const Text('Login'),
                        ),
                      ),

                      // ── Error display ──────────────────────────────
                      if (_errorMessage != null) ...[
                        const SizedBox(height: AppSpacing.md),
                        Container(
                          width: double.infinity,
                          padding: const EdgeInsets.all(AppSpacing.sm),
                          decoration: BoxDecoration(
                            color: AppColors.error.withValues(alpha: 0.1),
                            borderRadius: BorderRadius.circular(
                                AppSpacing.borderRadiusSm),
                            border: Border.all(
                              color: AppColors.error.withValues(alpha: 0.3),
                            ),
                          ),
                          child: Text(
                            _errorMessage!,
                            style: Theme.of(context)
                                .textTheme
                                .bodySmall
                                ?.copyWith(color: AppColors.error),
                          ),
                        ),
                      ],
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  // ── Connection status widget ──────────────────────────────────────

  Widget _buildConnectionStatus() {
    final IconData icon;
    final Color color;
    final String label;

    switch (_connectionStatus) {
      case _ConnectionStatus.untested:
        icon = Icons.circle_outlined;
        color = Colors.grey;
        label = 'Not tested';
      case _ConnectionStatus.testing:
        icon = Icons.sync;
        color = AppColors.info;
        label = 'Testing...';
      case _ConnectionStatus.connected:
        icon = Icons.check_circle;
        color = AppColors.success;
        label = 'Connected';
      case _ConnectionStatus.failed:
        icon = Icons.error;
        color = AppColors.error;
        label = 'Connection failed';
    }

    return Row(
      children: [
        Icon(icon, size: 16, color: color),
        const SizedBox(width: AppSpacing.xs),
        Text(label, style: TextStyle(fontSize: 12, color: color)),
      ],
    );
  }

  // ── Keypair section widget ────────────────────────────────────────

  Widget _buildKeypairSection(DeviceState deviceState) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(AppSpacing.sm),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surfaceContainerLowest,
        borderRadius: BorderRadius.circular(AppSpacing.borderRadiusMd),
        border: Border.all(
          color: Theme.of(context).colorScheme.outlineVariant,
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.key, size: 16, color: Colors.grey[600]),
              const SizedBox(width: AppSpacing.xs),
              Text(
                'Device Keypair',
                style: Theme.of(context).textTheme.labelMedium,
              ),
              const Spacer(),
              if (deviceState is DeviceReady)
                TextButton.icon(
                  onPressed: () => ref
                      .read(deviceNotifierProvider.notifier)
                      .generateNewKeypair(),
                  icon: const Icon(Icons.refresh, size: 14),
                  label: const Text('Generate New',
                      style: TextStyle(fontSize: 11)),
                ),
            ],
          ),
          const SizedBox(height: AppSpacing.xs),
          if (deviceState is DeviceLoading)
            const Padding(
              padding: EdgeInsets.symmetric(vertical: AppSpacing.xs),
              child: LinearProgressIndicator(),
            )
          else if (deviceState is DeviceReady)
            Row(
              children: [
                Text(
                  'Fingerprint: ',
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: Colors.grey[600],
                      ),
                ),
                SelectableText(
                  deviceState.fingerprint,
                  style: const TextStyle(
                    fontFamily: 'monospace',
                    fontSize: 13,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ],
            )
          else if (deviceState is DeviceError)
            Text(
              'Error: ${deviceState.message}',
              style: Theme.of(context)
                  .textTheme
                  .bodySmall
                  ?.copyWith(color: AppColors.error),
            ),
        ],
      ),
    );
  }
}
