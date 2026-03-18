import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_spacing.dart';
import '../../../shared/models/smart_models.dart';
import '../../../shared/widgets/confirm_dialog.dart';
import '../../../shared/widgets/data_table_card.dart';
import '../../../shared/providers/dio_provider.dart';
import '../../../shared/widgets/error_state.dart';
import 'settings_providers.dart';

/// Settings screen with connection settings, appearance, SMART client
/// management, and about section.
class SettingsScreen extends ConsumerStatefulWidget {
  const SettingsScreen({super.key});

  @override
  ConsumerState<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends ConsumerState<SettingsScreen> {
  late TextEditingController _serverUrlController;
  bool _testingConnection = false;
  String? _connectionTestResult;

  @override
  void initState() {
    super.initState();
    _serverUrlController = TextEditingController();
  }

  @override
  void dispose() {
    _serverUrlController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final serverUrl = ref.watch(serverUrlProvider);
    if (_serverUrlController.text.isEmpty) {
      _serverUrlController.text = serverUrl;
    }

    return Padding(
      padding: AppSpacing.pagePadding,
      child: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            _buildConnectionSettings(context),
            const SizedBox(height: AppSpacing.lg),
            _buildAppearance(context),
            const SizedBox(height: AppSpacing.lg),
            _buildSmartClientSection(context),
            const SizedBox(height: AppSpacing.lg),
            _buildAboutSection(context),
          ],
        ),
      ),
    );
  }

  // ===========================================================================
  // Connection Settings
  // ===========================================================================

  Widget _buildConnectionSettings(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Card(
      elevation: 1,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(AppSpacing.borderRadiusLg),
      ),
      child: Padding(
        padding: AppSpacing.cardPadding,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Connection Settings',
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w600,
                color: colorScheme.onSurface,
              ),
            ),
            const SizedBox(height: AppSpacing.md),

            // ── Server URL ──────────────────────────────────────────────
            Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _serverUrlController,
                    decoration: const InputDecoration(
                      labelText: 'Server URL',
                      hintText: 'https://localhost:8080',
                      border: OutlineInputBorder(),
                      prefixIcon: Icon(Icons.link, size: 18),
                    ),
                    onSubmitted: (value) {
                      ref.read(serverUrlProvider.notifier).state =
                          value.trim();
                    },
                  ),
                ),
                const SizedBox(width: AppSpacing.sm),
                FilledButton.icon(
                  onPressed: _testingConnection ? null : _testConnection,
                  icon: _testingConnection
                      ? const SizedBox(
                          width: 18,
                          height: 18,
                          child:
                              CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Icon(Icons.wifi_find, size: 18),
                  label: const Text('Test'),
                ),
              ],
            ),

            if (_connectionTestResult != null) ...[
              const SizedBox(height: AppSpacing.sm),
              Text(
                _connectionTestResult!,
                style: TextStyle(
                  fontSize: 13,
                  color: _connectionTestResult!.startsWith('Connected')
                      ? AppColors.statusActive
                      : colorScheme.error,
                ),
              ),
            ],

            const SizedBox(height: AppSpacing.md),

            // ── Accept Self-Signed Certificates ─────────────────────────
            Row(
              children: [
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Accept self-signed certificates',
                        style: TextStyle(
                          fontSize: 14,
                          color: colorScheme.onSurface,
                        ),
                      ),
                      const SizedBox(height: 2),
                      Text(
                        'Enable for local development. Disable in production.',
                        style: TextStyle(
                          fontSize: 12,
                          color: colorScheme.onSurfaceVariant,
                        ),
                      ),
                    ],
                  ),
                ),
                Switch(
                  value: true, // Controlled by AppConfig in a real setup
                  onChanged: (value) {
                    // In a real implementation, this would update the
                    // AppConfig and recreate the Dio instance.
                    ScaffoldMessenger.of(context)
                      ..hideCurrentSnackBar()
                      ..showSnackBar(
                        const SnackBar(
                          content: Text(
                              'Certificate setting updated. Restart may be required.'),
                          behavior: SnackBarBehavior.floating,
                        ),
                      );
                  },
                ),
              ],
            ),

            const SizedBox(height: AppSpacing.sm),

            // ── Connection Status ───────────────────────────────────────
            Row(
              children: [
                Container(
                  width: 10,
                  height: 10,
                  decoration: const BoxDecoration(
                    color: AppColors.statusInactive,
                    shape: BoxShape.circle,
                  ),
                ),
                const SizedBox(width: 6),
                Text(
                  'Connection status updates automatically',
                  style: TextStyle(
                    fontSize: 12,
                    color: colorScheme.onSurfaceVariant,
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _testConnection() async {
    final url = _serverUrlController.text.trim();
    if (url.isEmpty) return;

    setState(() {
      _testingConnection = true;
      _connectionTestResult = null;
    });

    // Save the URL to the provider.
    ref.read(serverUrlProvider.notifier).state = url;

    try {
      final dio = ref.read(dioProvider);
      final response = await dio.get('/health');
      if (response.statusCode == 200) {
        setState(() => _connectionTestResult = 'Connected successfully');
      } else {
        setState(() =>
            _connectionTestResult = 'Unexpected status: ${response.statusCode}');
      }
    } catch (e) {
      setState(
          () => _connectionTestResult = 'Connection failed: $e');
    } finally {
      setState(() => _testingConnection = false);
    }
  }

  // ===========================================================================
  // Appearance
  // ===========================================================================

  Widget _buildAppearance(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final currentMode = ref.watch(themeModePr);

    return Card(
      elevation: 1,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(AppSpacing.borderRadiusLg),
      ),
      child: Padding(
        padding: AppSpacing.cardPadding,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Appearance',
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w600,
                color: colorScheme.onSurface,
              ),
            ),
            const SizedBox(height: AppSpacing.md),
            Row(
              children: [
                Text(
                  'Theme Mode',
                  style: TextStyle(
                    fontSize: 14,
                    color: colorScheme.onSurface,
                  ),
                ),
                const SizedBox(width: AppSpacing.md),
                SegmentedButton<ThemeMode>(
                  segments: const [
                    ButtonSegment(
                      value: ThemeMode.light,
                      label: Text('Light'),
                      icon: Icon(Icons.light_mode, size: 18),
                    ),
                    ButtonSegment(
                      value: ThemeMode.dark,
                      label: Text('Dark'),
                      icon: Icon(Icons.dark_mode, size: 18),
                    ),
                    ButtonSegment(
                      value: ThemeMode.system,
                      label: Text('System'),
                      icon: Icon(Icons.settings_brightness, size: 18),
                    ),
                  ],
                  selected: {currentMode},
                  onSelectionChanged: (modes) {
                    ref.read(themeModePr.notifier).state = modes.first;
                  },
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  // ===========================================================================
  // SMART Client Management
  // ===========================================================================

  Widget _buildSmartClientSection(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final clientsAsync = ref.watch(smartClientListProvider);

    return Card(
      elevation: 1,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(AppSpacing.borderRadiusLg),
      ),
      child: ExpansionTile(
        shape: const RoundedRectangleBorder(),
        title: Text(
          'SMART Client Management',
          style: TextStyle(
            fontSize: 16,
            fontWeight: FontWeight.w600,
            color: colorScheme.onSurface,
          ),
        ),
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(
              AppSpacing.md,
              0,
              AppSpacing.md,
              AppSpacing.md,
            ),
            child: clientsAsync.when(
              loading: () => const SizedBox(
                height: 100,
                child: Center(child: CircularProgressIndicator()),
              ),
              error: (err, _) => ErrorState(
                message: 'Failed to load SMART clients',
                details: err.toString(),
                onRetry: () => ref.invalidate(smartClientListProvider),
              ),
              data: (clientList) => Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  // ── Register Button ──────────────────────────────────
                  Align(
                    alignment: Alignment.topRight,
                    child: FilledButton.icon(
                      onPressed: () =>
                          _showRegisterClientDialog(context),
                      icon: const Icon(Icons.add, size: 18),
                      label: const Text('Register Client'),
                    ),
                  ),
                  const SizedBox(height: AppSpacing.sm),

                  // ── Client Table ─────────────────────────────────────
                  DataTableCard(
                    title: 'Registered Clients',
                    emptyTitle: 'No SMART clients registered',
                    emptySubtitle:
                        'Register a client to enable SMART on FHIR authorization.',
                    columns: const [
                      DataColumn(label: Text('Client Name')),
                      DataColumn(label: Text('Client ID')),
                      DataColumn(label: Text('Scope')),
                      DataColumn(label: Text('Status')),
                      DataColumn(label: Text('Registered At')),
                      DataColumn(label: Text('Actions')),
                    ],
                    rows: clientList.clients.map((c) {
                      return DataRow(cells: [
                        DataCell(Text(
                          c.clientName,
                          style: const TextStyle(
                              fontWeight: FontWeight.w600,
                              fontSize: 13),
                        )),
                        DataCell(Text(
                          c.clientId,
                          style: const TextStyle(
                              fontFamily: 'monospace', fontSize: 12),
                        )),
                        DataCell(Text(
                          c.scope,
                          style: const TextStyle(fontSize: 12),
                        )),
                        DataCell(_ClientStatusBadge(status: c.status)),
                        DataCell(Text(
                          c.registeredAt,
                          style: const TextStyle(fontSize: 12),
                        )),
                        DataCell(Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            IconButton(
                              icon: const Icon(Icons.edit, size: 18),
                              tooltip: 'Edit',
                              onPressed: () =>
                                  _showEditClientDialog(context, c),
                            ),
                            IconButton(
                              icon: Icon(
                                Icons.delete,
                                size: 18,
                                color: colorScheme.error,
                              ),
                              tooltip: 'Delete',
                              onPressed: () =>
                                  _deleteClient(context, c.clientId),
                            ),
                          ],
                        )),
                      ]);
                    }).toList(),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Future<void> _showRegisterClientDialog(BuildContext context) async {
    final nameController = TextEditingController();
    final redirectUrisController = TextEditingController();
    final scopeController = TextEditingController(text: 'openid fhirUser');
    String tokenAuthMethod = 'client_secret_basic';
    final grantTypes = <String>{'authorization_code'};
    final launchModes = <String>{'launch-ehr'};
    String? errorMsg;

    await showDialog(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setDialogState) {
          return AlertDialog(
            shape: RoundedRectangleBorder(
              borderRadius:
                  BorderRadius.circular(AppSpacing.borderRadiusLg),
            ),
            title: const Text('Register SMART Client'),
            content: SizedBox(
              width: 520,
              child: SingleChildScrollView(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    TextField(
                      controller: nameController,
                      decoration: const InputDecoration(
                        labelText: 'Client Name *',
                        border: OutlineInputBorder(),
                      ),
                    ),
                    const SizedBox(height: AppSpacing.sm),
                    TextField(
                      controller: redirectUrisController,
                      maxLines: 3,
                      decoration: const InputDecoration(
                        labelText: 'Redirect URIs (one per line) *',
                        hintText: 'https://app.example.com/callback',
                        border: OutlineInputBorder(),
                      ),
                    ),
                    const SizedBox(height: AppSpacing.sm),
                    TextField(
                      controller: scopeController,
                      decoration: const InputDecoration(
                        labelText: 'Scope *',
                        border: OutlineInputBorder(),
                      ),
                    ),
                    const SizedBox(height: AppSpacing.sm),

                    // Grant Types
                    const Text('Grant Types',
                        style: TextStyle(
                            fontSize: 14, fontWeight: FontWeight.w500)),
                    Wrap(
                      spacing: 8,
                      children: [
                        'authorization_code',
                        'client_credentials',
                        'refresh_token',
                      ].map((gt) {
                        return FilterChip(
                          label: Text(gt, style: const TextStyle(fontSize: 12)),
                          selected: grantTypes.contains(gt),
                          onSelected: (selected) {
                            setDialogState(() {
                              if (selected) {
                                grantTypes.add(gt);
                              } else {
                                grantTypes.remove(gt);
                              }
                            });
                          },
                        );
                      }).toList(),
                    ),
                    const SizedBox(height: AppSpacing.sm),

                    // Token Auth Method
                    DropdownButtonFormField<String>(
                      value: tokenAuthMethod,
                      decoration: const InputDecoration(
                        labelText: 'Token Auth Method',
                        border: OutlineInputBorder(),
                      ),
                      items: const [
                        DropdownMenuItem(
                          value: 'client_secret_basic',
                          child: Text('client_secret_basic'),
                        ),
                        DropdownMenuItem(
                          value: 'client_secret_post',
                          child: Text('client_secret_post'),
                        ),
                        DropdownMenuItem(
                          value: 'private_key_jwt',
                          child: Text('private_key_jwt'),
                        ),
                        DropdownMenuItem(
                          value: 'none',
                          child: Text('none (public client)'),
                        ),
                      ],
                      onChanged: (v) {
                        if (v != null) {
                          setDialogState(() => tokenAuthMethod = v);
                        }
                      },
                    ),
                    const SizedBox(height: AppSpacing.sm),

                    // Launch Modes
                    const Text('Launch Modes',
                        style: TextStyle(
                            fontSize: 14, fontWeight: FontWeight.w500)),
                    Wrap(
                      spacing: 8,
                      children: ['launch-ehr', 'launch-standalone']
                          .map((lm) {
                        return FilterChip(
                          label: Text(lm, style: const TextStyle(fontSize: 12)),
                          selected: launchModes.contains(lm),
                          onSelected: (selected) {
                            setDialogState(() {
                              if (selected) {
                                launchModes.add(lm);
                              } else {
                                launchModes.remove(lm);
                              }
                            });
                          },
                        );
                      }).toList(),
                    ),

                    if (errorMsg != null) ...[
                      const SizedBox(height: AppSpacing.md),
                      Text(
                        errorMsg!,
                        style: TextStyle(
                          color: Theme.of(ctx).colorScheme.error,
                          fontSize: 13,
                        ),
                      ),
                    ],
                  ],
                ),
              ),
            ),
            actions: [
              TextButton(
                onPressed: () => Navigator.of(ctx).pop(),
                child: const Text('Cancel'),
              ),
              FilledButton(
                onPressed: () async {
                  final name = nameController.text.trim();
                  final uris = redirectUrisController.text
                      .split('\n')
                      .map((u) => u.trim())
                      .where((u) => u.isNotEmpty)
                      .toList();
                  final scope = scopeController.text.trim();

                  if (name.isEmpty || uris.isEmpty || scope.isEmpty) {
                    setDialogState(() =>
                        errorMsg = 'Name, redirect URIs, and scope are required.');
                    return;
                  }

                  try {
                    final api = ref.read(smartApiProvider);
                    await api.registerClient(RegisterClientRequest(
                      clientName: name,
                      redirectUris: uris,
                      scope: scope,
                      grantTypes: grantTypes.toList(),
                      tokenEndpointAuthMethod: tokenAuthMethod,
                      launchModes: launchModes.toList(),
                    ));
                    ref.invalidate(smartClientListProvider);
                    if (ctx.mounted) Navigator.of(ctx).pop();
                  } catch (e) {
                    setDialogState(() => errorMsg = e.toString());
                  }
                },
                child: const Text('Register'),
              ),
            ],
          );
        },
      ),
    );
  }

  Future<void> _showEditClientDialog(
      BuildContext context, ClientResponse client) async {
    final scopeController = TextEditingController(text: client.scope);
    String status = client.status;
    String? errorMsg;

    await showDialog(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setDialogState) {
          return AlertDialog(
            shape: RoundedRectangleBorder(
              borderRadius:
                  BorderRadius.circular(AppSpacing.borderRadiusLg),
            ),
            title: Text('Edit ${client.clientName}'),
            content: SizedBox(
              width: 400,
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  DropdownButtonFormField<String>(
                    value: status,
                    decoration: const InputDecoration(
                      labelText: 'Status',
                      border: OutlineInputBorder(),
                    ),
                    items: const [
                      DropdownMenuItem(
                          value: 'active', child: Text('Active')),
                      DropdownMenuItem(
                          value: 'suspended', child: Text('Suspended')),
                      DropdownMenuItem(
                          value: 'revoked', child: Text('Revoked')),
                    ],
                    onChanged: (v) {
                      if (v != null) setDialogState(() => status = v);
                    },
                  ),
                  const SizedBox(height: AppSpacing.sm),
                  TextField(
                    controller: scopeController,
                    decoration: const InputDecoration(
                      labelText: 'Scope',
                      border: OutlineInputBorder(),
                    ),
                  ),
                  if (errorMsg != null) ...[
                    const SizedBox(height: AppSpacing.md),
                    Text(
                      errorMsg!,
                      style: TextStyle(
                        color: Theme.of(ctx).colorScheme.error,
                        fontSize: 13,
                      ),
                    ),
                  ],
                ],
              ),
            ),
            actions: [
              TextButton(
                onPressed: () => Navigator.of(ctx).pop(),
                child: const Text('Cancel'),
              ),
              FilledButton(
                onPressed: () async {
                  try {
                    final api = ref.read(smartApiProvider);
                    await api.updateClient(
                      client.clientId,
                      UpdateClientRequest(
                        status: status,
                        scope: scopeController.text.trim(),
                      ),
                    );
                    ref.invalidate(smartClientListProvider);
                    if (ctx.mounted) Navigator.of(ctx).pop();
                  } catch (e) {
                    setDialogState(() => errorMsg = e.toString());
                  }
                },
                child: const Text('Update'),
              ),
            ],
          );
        },
      ),
    );
  }

  Future<void> _deleteClient(BuildContext context, String clientId) async {
    final confirmed = await ConfirmDialog.show(
      context: context,
      title: 'Delete SMART Client',
      message:
          'This will permanently remove the client "$clientId". This action cannot be undone.',
      confirmLabel: 'Delete',
      destructive: true,
    );

    if (!confirmed) return;

    try {
      final api = ref.read(smartApiProvider);
      await api.deleteClient(clientId);
      ref.invalidate(smartClientListProvider);

      if (mounted) {
        ScaffoldMessenger.of(context)
          ..hideCurrentSnackBar()
          ..showSnackBar(
            const SnackBar(
              content: Text('SMART client deleted'),
              behavior: SnackBarBehavior.floating,
            ),
          );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context)
          ..hideCurrentSnackBar()
          ..showSnackBar(
            SnackBar(
              content: Text('Failed to delete client: $e'),
              behavior: SnackBarBehavior.floating,
            ),
          );
      }
    }
  }

  // ===========================================================================
  // About
  // ===========================================================================

  Widget _buildAboutSection(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Card(
      elevation: 1,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(AppSpacing.borderRadiusLg),
      ),
      child: Padding(
        padding: AppSpacing.cardPadding,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'About',
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w600,
                color: colorScheme.onSurface,
              ),
            ),
            const SizedBox(height: AppSpacing.md),
            _aboutRow('Application', 'Open Nucleus Desktop'),
            _aboutRow('Version', '1.0.0-alpha'),
            _aboutRow('License', 'AGPLv3'),
            const SizedBox(height: AppSpacing.sm),
            InkWell(
              onTap: () {
                // In a real implementation, launch the URL.
                ScaffoldMessenger.of(context)
                  ..hideCurrentSnackBar()
                  ..showSnackBar(
                    const SnackBar(
                      content: Text(
                          'https://github.com/open-nucleus/open-nucleus'),
                      behavior: SnackBarBehavior.floating,
                    ),
                  );
              },
              child: Row(
                children: [
                  Icon(Icons.open_in_new,
                      size: 16, color: colorScheme.primary),
                  const SizedBox(width: 8),
                  Text(
                    'View project on GitHub',
                    style: TextStyle(
                      fontSize: 13,
                      color: colorScheme.primary,
                      decoration: TextDecoration.underline,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _aboutRow(String label, String value) {
    final colorScheme = Theme.of(context).colorScheme;

    return Padding(
      padding: const EdgeInsets.only(bottom: AppSpacing.xs),
      child: Row(
        children: [
          SizedBox(
            width: 100,
            child: Text(
              label,
              style: TextStyle(
                fontSize: 13,
                fontWeight: FontWeight.w600,
                color: colorScheme.onSurfaceVariant,
              ),
            ),
          ),
          Text(
            value,
            style: TextStyle(
              fontSize: 13,
              color: colorScheme.onSurface,
            ),
          ),
        ],
      ),
    );
  }
}

// =============================================================================
// Shared Widgets
// =============================================================================

class _ClientStatusBadge extends StatelessWidget {
  final String status;
  const _ClientStatusBadge({required this.status});

  @override
  Widget build(BuildContext context) {
    final Color color;
    switch (status.toLowerCase()) {
      case 'active':
      case 'approved':
        color = AppColors.statusActive;
        break;
      case 'pending':
        color = AppColors.statusPending;
        break;
      case 'suspended':
      case 'revoked':
        color = AppColors.statusError;
        break;
      default:
        color = AppColors.statusInactive;
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: color.withOpacity(0.15),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withOpacity(0.4)),
      ),
      child: Text(
        status,
        style: TextStyle(
          fontSize: 11,
          fontWeight: FontWeight.w600,
          color: color,
        ),
      ),
    );
  }
}
