import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/extensions/date_extensions.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_spacing.dart';
import '../../../shared/models/conflict_models.dart';
import '../../../shared/models/sync_models.dart';
import '../../../shared/widgets/empty_state.dart';
import '../../../shared/widgets/error_state.dart';
import '../../../shared/widgets/json_viewer.dart';
import '../../../shared/widgets/loading_skeleton.dart';
import '../../../shared/widgets/severity_badge.dart';
import 'sync_providers.dart';

/// Full sync screen with two sections:
///
/// **Top** -- Sync status, peer list, action buttons, and sync history.
/// **Bottom** -- Merge conflict list with a master-detail layout.
class SyncScreen extends ConsumerStatefulWidget {
  const SyncScreen({super.key});

  @override
  ConsumerState<SyncScreen> createState() => _SyncScreenState();
}

class _SyncScreenState extends ConsumerState<SyncScreen> {
  bool _historyExpanded = false;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: AppSpacing.pagePadding,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          // ── Header ──────────────────────────────────────────────────
          Text(
            'Sync & Conflicts',
            style: TextStyle(
              fontSize: 24,
              fontWeight: FontWeight.w700,
              color: Theme.of(context).colorScheme.onSurface,
            ),
          ),
          const SizedBox(height: AppSpacing.md),

          // ── Content ─────────────────────────────────────────────────
          Expanded(
            child: SingleChildScrollView(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  // ── Top Section: Status & Peers ─────────────────────
                  _SyncStatusCard(
                    onTriggerSync: _showTriggerSyncDialog,
                    onExportBundle: _showExportBundleDialog,
                    onImportBundle: _showImportBundleDialog,
                  ),
                  const SizedBox(height: AppSpacing.md),
                  const _PeerListSection(),
                  const SizedBox(height: AppSpacing.md),

                  // ── Sync History (expandable) ───────────────────────
                  _SyncHistorySection(
                    expanded: _historyExpanded,
                    onToggle: () =>
                        setState(() => _historyExpanded = !_historyExpanded),
                  ),
                  const SizedBox(height: AppSpacing.lg),

                  // ── Bottom Section: Conflicts ───────────────────────
                  const _ConflictSection(),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  // ---------------------------------------------------------------------------
  // Trigger Sync dialog
  // ---------------------------------------------------------------------------

  Future<void> _showTriggerSyncDialog() async {
    final peersAsync = ref.read(syncPeersProvider);
    final peers = peersAsync.valueOrNull?.peers ?? [];

    if (peers.isEmpty) {
      if (!mounted) return;
      ScaffoldMessenger.of(context)
        ..hideCurrentSnackBar()
        ..showSnackBar(
          const SnackBar(
            content: Text('No peers discovered. Cannot trigger sync.'),
            behavior: SnackBarBehavior.floating,
          ),
        );
      return;
    }

    String? selectedNode = peers.first.nodeId;

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) {
        return StatefulBuilder(
          builder: (ctx, setDialogState) {
            return AlertDialog(
              shape: RoundedRectangleBorder(
                borderRadius:
                    BorderRadius.circular(AppSpacing.borderRadiusLg),
              ),
              title: const Text('Trigger Sync'),
              content: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('Select a target node to sync with:'),
                  const SizedBox(height: AppSpacing.md),
                  DropdownButtonFormField<String>(
                    value: selectedNode,
                    decoration: InputDecoration(
                      labelText: 'Target Node',
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(
                            AppSpacing.borderRadiusMd),
                      ),
                    ),
                    items: peers
                        .map((p) => DropdownMenuItem(
                              value: p.nodeId,
                              child: Text(
                                '${p.nodeId} (${p.siteId})',
                                style: const TextStyle(fontSize: 13),
                              ),
                            ))
                        .toList(),
                    onChanged: (v) =>
                        setDialogState(() => selectedNode = v),
                  ),
                ],
              ),
              actions: [
                TextButton(
                  onPressed: () => Navigator.of(ctx).pop(false),
                  child: const Text('Cancel'),
                ),
                FilledButton(
                  onPressed: () => Navigator.of(ctx).pop(true),
                  child: const Text('Sync'),
                ),
              ],
            );
          },
        );
      },
    );

    if (confirmed == true && selectedNode != null) {
      try {
        final api = ref.read(syncApiProvider);
        await api.triggerSync(selectedNode!);
        ref.invalidate(syncStatusProvider);
        ref.invalidate(syncHistoryProvider);
        if (!mounted) return;
        ScaffoldMessenger.of(context)
          ..hideCurrentSnackBar()
          ..showSnackBar(
            SnackBar(
              content: Text('Sync triggered with $selectedNode'),
              behavior: SnackBarBehavior.floating,
            ),
          );
      } catch (e) {
        if (!mounted) return;
        ScaffoldMessenger.of(context)
          ..hideCurrentSnackBar()
          ..showSnackBar(
            SnackBar(
              content: Text('Sync failed: $e'),
              behavior: SnackBarBehavior.floating,
            ),
          );
      }
    }
  }

  // ---------------------------------------------------------------------------
  // Export Bundle dialog
  // ---------------------------------------------------------------------------

  Future<void> _showExportBundleDialog() async {
    final resourceTypesController = TextEditingController(
      text: 'Patient,Encounter,Observation',
    );
    final sinceController = TextEditingController(
      text: DateTime.now()
          .subtract(const Duration(days: 7))
          .toIso8601String()
          .split('T')
          .first,
    );

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) {
        return AlertDialog(
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(AppSpacing.borderRadiusLg),
          ),
          title: const Text('Export Bundle'),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              TextField(
                controller: resourceTypesController,
                decoration: InputDecoration(
                  labelText: 'Resource Types (comma-separated)',
                  border: OutlineInputBorder(
                    borderRadius:
                        BorderRadius.circular(AppSpacing.borderRadiusMd),
                  ),
                ),
              ),
              const SizedBox(height: AppSpacing.md),
              TextField(
                controller: sinceController,
                decoration: InputDecoration(
                  labelText: 'Since (YYYY-MM-DD)',
                  border: OutlineInputBorder(
                    borderRadius:
                        BorderRadius.circular(AppSpacing.borderRadiusMd),
                  ),
                ),
              ),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(ctx).pop(false),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: () => Navigator.of(ctx).pop(true),
              child: const Text('Export'),
            ),
          ],
        );
      },
    );

    if (confirmed == true) {
      try {
        final api = ref.read(syncApiProvider);
        final types = resourceTypesController.text
            .split(',')
            .map((s) => s.trim())
            .where((s) => s.isNotEmpty)
            .toList();
        final result = await api.exportBundle(
          resourceTypes: types,
          since: sinceController.text.trim(),
        );
        if (!mounted) return;
        ScaffoldMessenger.of(context)
          ..hideCurrentSnackBar()
          ..showSnackBar(
            SnackBar(
              content: Text(
                'Exported ${result.data?.resourceCount ?? 0} resources',
              ),
              behavior: SnackBarBehavior.floating,
            ),
          );
      } catch (e) {
        if (!mounted) return;
        ScaffoldMessenger.of(context)
          ..hideCurrentSnackBar()
          ..showSnackBar(
            SnackBar(
              content: Text('Export failed: $e'),
              behavior: SnackBarBehavior.floating,
            ),
          );
      }
    }

    resourceTypesController.dispose();
    sinceController.dispose();
  }

  // ---------------------------------------------------------------------------
  // Import Bundle dialog
  // ---------------------------------------------------------------------------

  Future<void> _showImportBundleDialog() async {
    final bundleDataController = TextEditingController();
    final authorController = TextEditingController();

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) {
        return AlertDialog(
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(AppSpacing.borderRadiusLg),
          ),
          title: const Text('Import Bundle'),
          content: SizedBox(
            width: 500,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                TextField(
                  controller: authorController,
                  decoration: InputDecoration(
                    labelText: 'Author',
                    border: OutlineInputBorder(
                      borderRadius:
                          BorderRadius.circular(AppSpacing.borderRadiusMd),
                    ),
                  ),
                ),
                const SizedBox(height: AppSpacing.md),
                TextField(
                  controller: bundleDataController,
                  maxLines: 8,
                  decoration: InputDecoration(
                    labelText: 'Bundle JSON',
                    alignLabelWithHint: true,
                    border: OutlineInputBorder(
                      borderRadius:
                          BorderRadius.circular(AppSpacing.borderRadiusMd),
                    ),
                  ),
                  style: const TextStyle(
                    fontFamily: 'monospace',
                    fontSize: 12,
                  ),
                ),
              ],
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(ctx).pop(false),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: () => Navigator.of(ctx).pop(true),
              child: const Text('Import'),
            ),
          ],
        );
      },
    );

    if (confirmed == true) {
      try {
        final api = ref.read(syncApiProvider);
        final statusData = ref.read(syncStatusProvider).valueOrNull;
        final result = await api.importBundle(
          bundleData: bundleDataController.text,
          format: 'json',
          author: authorController.text.trim(),
          nodeId: statusData?.nodeId ?? '',
          siteId: statusData?.siteId ?? '',
        );
        ref.invalidate(syncStatusProvider);
        ref.invalidate(syncHistoryProvider);
        if (!mounted) return;
        final data = result.data;
        ScaffoldMessenger.of(context)
          ..hideCurrentSnackBar()
          ..showSnackBar(
            SnackBar(
              content: Text(
                'Imported ${data?.resourcesImported ?? 0} resources, '
                '${data?.resourcesSkipped ?? 0} skipped',
              ),
              behavior: SnackBarBehavior.floating,
            ),
          );
      } catch (e) {
        if (!mounted) return;
        ScaffoldMessenger.of(context)
          ..hideCurrentSnackBar()
          ..showSnackBar(
            SnackBar(
              content: Text('Import failed: $e'),
              behavior: SnackBarBehavior.floating,
            ),
          );
      }
    }

    bundleDataController.dispose();
    authorController.dispose();
  }
}

// =============================================================================
// Sync Status Card
// =============================================================================

class _SyncStatusCard extends ConsumerWidget {
  const _SyncStatusCard({
    required this.onTriggerSync,
    required this.onExportBundle,
    required this.onImportBundle,
  });

  final VoidCallback onTriggerSync;
  final VoidCallback onExportBundle;
  final VoidCallback onImportBundle;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final statusAsync = ref.watch(syncStatusProvider);
    final colorScheme = Theme.of(context).colorScheme;

    return Card(
      elevation: 1,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(AppSpacing.borderRadiusLg),
      ),
      child: Padding(
        padding: AppSpacing.cardPadding,
        child: statusAsync.when(
          loading: () => const SizedBox(
            height: 80,
            child: Center(child: CircularProgressIndicator()),
          ),
          error: (err, _) => ErrorState(
            message: 'Failed to load sync status',
            details: err.toString(),
            onRetry: () => ref.invalidate(syncStatusProvider),
          ),
          data: (status) => Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // ── State indicator + metadata ──────────────────────────
              Row(
                children: [
                  _SyncStateBadge(state: status.state),
                  const SizedBox(width: AppSpacing.md),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Last Sync: ${_formatTimestamp(status.lastSync)}',
                          style: TextStyle(
                            fontSize: 13,
                            color: colorScheme.onSurfaceVariant,
                          ),
                        ),
                        const SizedBox(height: 2),
                        Text(
                          'Pending Changes: ${status.pendingChanges}',
                          style: TextStyle(
                            fontSize: 13,
                            color: colorScheme.onSurfaceVariant,
                          ),
                        ),
                      ],
                    ),
                  ),
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.end,
                    children: [
                      Text(
                        'Node: ${status.nodeId}',
                        style: TextStyle(
                          fontSize: 12,
                          fontFamily: 'monospace',
                          color: colorScheme.onSurfaceVariant,
                        ),
                      ),
                      const SizedBox(height: 2),
                      Text(
                        'Site: ${status.siteId}',
                        style: TextStyle(
                          fontSize: 12,
                          fontFamily: 'monospace',
                          color: colorScheme.onSurfaceVariant,
                        ),
                      ),
                    ],
                  ),
                ],
              ),
              const SizedBox(height: AppSpacing.md),

              // ── Action buttons ─────────────────────────────────────
              Wrap(
                spacing: AppSpacing.sm,
                runSpacing: AppSpacing.sm,
                children: [
                  FilledButton.icon(
                    onPressed: onTriggerSync,
                    icon: const Icon(Icons.sync, size: 18),
                    label: const Text('Trigger Sync'),
                  ),
                  OutlinedButton.icon(
                    onPressed: onExportBundle,
                    icon: const Icon(Icons.upload, size: 18),
                    label: const Text('Export Bundle'),
                  ),
                  OutlinedButton.icon(
                    onPressed: onImportBundle,
                    icon: const Icon(Icons.download, size: 18),
                    label: const Text('Import Bundle'),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  static String _formatTimestamp(String timestamp) {
    try {
      return DateTime.parse(timestamp).timeAgo;
    } catch (_) {
      return timestamp.isEmpty ? 'Never' : timestamp;
    }
  }
}

// =============================================================================
// Sync State Badge
// =============================================================================

class _SyncStateBadge extends StatelessWidget {
  const _SyncStateBadge({required this.state});

  final String state;

  @override
  Widget build(BuildContext context) {
    final (Color color, IconData icon) = _resolve(state);

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: color.withOpacity(0.12),
        borderRadius: BorderRadius.circular(AppSpacing.borderRadiusMd),
        border: Border.all(color: color.withOpacity(0.4)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 20, color: color),
          const SizedBox(width: 6),
          Text(
            state[0].toUpperCase() + state.substring(1),
            style: TextStyle(
              fontSize: 14,
              fontWeight: FontWeight.w600,
              color: color,
            ),
          ),
        ],
      ),
    );
  }

  static (Color, IconData) _resolve(String state) {
    switch (state.toLowerCase()) {
      case 'idle':
        return (AppColors.syncIdle, Icons.pause_circle_outline);
      case 'syncing':
        return (AppColors.syncSyncing, Icons.sync);
      case 'complete':
        return (AppColors.syncComplete, Icons.check_circle_outline);
      case 'error':
        return (AppColors.syncError, Icons.error_outline);
      default:
        return (AppColors.syncIdle, Icons.help_outline);
    }
  }
}

// =============================================================================
// Peer List Section
// =============================================================================

class _PeerListSection extends ConsumerWidget {
  const _PeerListSection();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final peersAsync = ref.watch(syncPeersProvider);
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
            Row(
              children: [
                Icon(Icons.devices, size: 18, color: colorScheme.primary),
                const SizedBox(width: AppSpacing.sm),
                Text(
                  'Discovered Peers',
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.w600,
                    color: colorScheme.onSurface,
                  ),
                ),
                const Spacer(),
                IconButton(
                  icon: const Icon(Icons.refresh, size: 18),
                  tooltip: 'Refresh peers',
                  onPressed: () => ref.invalidate(syncPeersProvider),
                ),
              ],
            ),
            const SizedBox(height: AppSpacing.sm),
            peersAsync.when(
              loading: () => LoadingSkeleton.table(rows: 3, cols: 5),
              error: (err, _) => ErrorState(
                message: 'Failed to load peers',
                details: err.toString(),
                onRetry: () => ref.invalidate(syncPeersProvider),
              ),
              data: (response) {
                if (response.peers.isEmpty) {
                  return const Padding(
                    padding: EdgeInsets.symmetric(vertical: AppSpacing.lg),
                    child: EmptyState(
                      icon: Icons.wifi_off,
                      title: 'No peers discovered',
                      subtitle:
                          'Ensure other nodes are on the same network.',
                    ),
                  );
                }
                return SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  child: DataTable(
                    showCheckboxColumn: false,
                    headingRowColor: WidgetStateProperty.all(
                      colorScheme.surfaceContainerHighest.withOpacity(0.5),
                    ),
                    columns: const [
                      DataColumn(label: Text('Node ID')),
                      DataColumn(label: Text('Site ID')),
                      DataColumn(label: Text('Last Seen')),
                      DataColumn(label: Text('State')),
                      DataColumn(label: Text('Latency')),
                    ],
                    rows: response.peers
                        .map((peer) => _buildPeerRow(context, peer))
                        .toList(),
                  ),
                );
              },
            ),
          ],
        ),
      ),
    );
  }

  DataRow _buildPeerRow(BuildContext context, PeerInfo peer) {
    final colorScheme = Theme.of(context).colorScheme;

    return DataRow(
      cells: [
        DataCell(
          Text(
            peer.nodeId,
            style: const TextStyle(
              fontSize: 13,
              fontFamily: 'monospace',
            ),
          ),
        ),
        DataCell(
          Text(
            peer.siteId,
            style: TextStyle(
              fontSize: 13,
              fontFamily: 'monospace',
              color: colorScheme.onSurfaceVariant,
            ),
          ),
        ),
        DataCell(
          Text(
            _formatTimestamp(peer.lastSeen),
            style: TextStyle(
              fontSize: 13,
              color: colorScheme.onSurfaceVariant,
            ),
          ),
        ),
        DataCell(SeverityBadge(
          severity: _peerStateToSeverity(peer.state),
        )),
        DataCell(_LatencyBar(latencyMs: peer.latencyMs)),
      ],
    );
  }

  static String _formatTimestamp(String timestamp) {
    try {
      return DateTime.parse(timestamp).timeAgo;
    } catch (_) {
      return timestamp;
    }
  }

  static String _peerStateToSeverity(String state) {
    switch (state.toLowerCase()) {
      case 'connected':
      case 'online':
        return 'low'; // green
      case 'syncing':
        return 'info'; // blue
      case 'offline':
      case 'disconnected':
        return 'critical'; // red
      default:
        return 'warning'; // amber
    }
  }
}

// =============================================================================
// Latency Bar
// =============================================================================

class _LatencyBar extends StatelessWidget {
  const _LatencyBar({required this.latencyMs});

  final int latencyMs;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    // Clamp latency to 0-500ms for visual bar width.
    final fraction = (latencyMs / 500).clamp(0.0, 1.0);
    final color = latencyMs < 100
        ? AppColors.statusActive
        : latencyMs < 300
            ? AppColors.statusPending
            : AppColors.statusError;

    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        SizedBox(
          width: 60,
          height: 8,
          child: ClipRRect(
            borderRadius: BorderRadius.circular(4),
            child: LinearProgressIndicator(
              value: fraction,
              backgroundColor: colorScheme.surfaceContainerHighest,
              valueColor: AlwaysStoppedAnimation(color),
            ),
          ),
        ),
        const SizedBox(width: 6),
        Text(
          '${latencyMs}ms',
          style: TextStyle(
            fontSize: 11,
            fontFamily: 'monospace',
            color: colorScheme.onSurfaceVariant,
          ),
        ),
      ],
    );
  }
}

// =============================================================================
// Sync History Section (expandable)
// =============================================================================

class _SyncHistorySection extends ConsumerWidget {
  const _SyncHistorySection({
    required this.expanded,
    required this.onToggle,
  });

  final bool expanded;
  final VoidCallback onToggle;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final colorScheme = Theme.of(context).colorScheme;
    final historyAsync = ref.watch(syncHistoryProvider);

    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(AppSpacing.borderRadiusMd),
        side: BorderSide(color: colorScheme.outline.withOpacity(0.3)),
      ),
      child: Column(
        children: [
          // ── Toggle Row ─────────────────────────────────────────────
          InkWell(
            onTap: onToggle,
            borderRadius: BorderRadius.circular(AppSpacing.borderRadiusMd),
            child: Padding(
              padding: const EdgeInsets.symmetric(
                horizontal: AppSpacing.md,
                vertical: AppSpacing.sm,
              ),
              child: Row(
                children: [
                  Icon(Icons.history, size: 18, color: colorScheme.primary),
                  const SizedBox(width: AppSpacing.sm),
                  Text(
                    'Sync History',
                    style: TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.w500,
                      color: colorScheme.onSurface,
                    ),
                  ),
                  const Spacer(),
                  Icon(
                    expanded ? Icons.expand_less : Icons.expand_more,
                    size: 20,
                    color: colorScheme.onSurfaceVariant,
                  ),
                ],
              ),
            ),
          ),

          // ── Timeline ───────────────────────────────────────────────
          if (expanded)
            Padding(
              padding: const EdgeInsets.fromLTRB(
                AppSpacing.md,
                0,
                AppSpacing.md,
                AppSpacing.md,
              ),
              child: Column(
                children: [
                  const Divider(height: 1),
                  const SizedBox(height: AppSpacing.sm),
                  historyAsync.when(
                    loading: () =>
                        LoadingSkeleton.table(rows: 4, cols: 5),
                    error: (err, _) => ErrorState(
                      message: 'Failed to load sync history',
                      details: err.toString(),
                      onRetry: () =>
                          ref.invalidate(syncHistoryProvider),
                    ),
                    data: (response) {
                      if (response.events.isEmpty) {
                        return const Padding(
                          padding: EdgeInsets.symmetric(
                              vertical: AppSpacing.lg),
                          child: EmptyState(
                            icon: Icons.history,
                            title: 'No sync events yet',
                            subtitle:
                                'Sync history will appear here after the first sync.',
                          ),
                        );
                      }
                      return Column(
                        children: response.events
                            .map((event) =>
                                _SyncEventTile(event: event))
                            .toList(),
                      );
                    },
                  ),
                ],
              ),
            ),
        ],
      ),
    );
  }
}

// =============================================================================
// Sync Event Tile
// =============================================================================

class _SyncEventTile extends StatelessWidget {
  const _SyncEventTile({required this.event});

  final SyncEvent event;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final isInbound = event.direction.toLowerCase() == 'in' ||
        event.direction.toLowerCase() == 'inbound';

    return Padding(
      padding: const EdgeInsets.symmetric(vertical: AppSpacing.xs),
      child: Row(
        children: [
          // Direction icon
          Icon(
            isInbound ? Icons.arrow_downward : Icons.arrow_upward,
            size: 16,
            color: isInbound ? AppColors.info : AppColors.secondary,
          ),
          const SizedBox(width: AppSpacing.sm),

          // Timestamp
          SizedBox(
            width: 140,
            child: Text(
              _formatTimestamp(event.timestamp),
              style: TextStyle(
                fontSize: 12,
                color: colorScheme.onSurfaceVariant,
              ),
            ),
          ),

          // Peer node
          SizedBox(
            width: 140,
            child: Text(
              event.peerNode,
              style: const TextStyle(
                fontSize: 12,
                fontFamily: 'monospace',
              ),
              overflow: TextOverflow.ellipsis,
            ),
          ),

          // State badge
          SizedBox(
            width: 90,
            child: SeverityBadge(
              severity: _stateToSeverity(event.state),
            ),
          ),

          const SizedBox(width: AppSpacing.sm),

          // Resources transferred
          Text(
            '${event.resourcesTransferred} resources',
            style: TextStyle(
              fontSize: 12,
              color: colorScheme.onSurfaceVariant,
            ),
          ),
        ],
      ),
    );
  }

  static String _formatTimestamp(String timestamp) {
    try {
      return DateTime.parse(timestamp).toDisplayDateTime;
    } catch (_) {
      return timestamp;
    }
  }

  static String _stateToSeverity(String state) {
    switch (state.toLowerCase()) {
      case 'complete':
      case 'success':
        return 'low';
      case 'in_progress':
      case 'syncing':
        return 'info';
      case 'error':
      case 'failed':
        return 'critical';
      default:
        return 'warning';
    }
  }
}

// =============================================================================
// Conflict Section
// =============================================================================

class _ConflictSection extends ConsumerWidget {
  const _ConflictSection();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final conflictsAsync = ref.watch(conflictListProvider);
    final colorScheme = Theme.of(context).colorScheme;
    final selectedId = ref.watch(selectedConflictProvider);

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
            // ── Section Header ────────────────────────────────────────
            Row(
              children: [
                Icon(Icons.merge_type, size: 18, color: colorScheme.primary),
                const SizedBox(width: AppSpacing.sm),
                Text(
                  'Conflicts',
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.w600,
                    color: colorScheme.onSurface,
                  ),
                ),
                const SizedBox(width: AppSpacing.sm),
                conflictsAsync.when(
                  loading: () => const SizedBox.shrink(),
                  error: (_, __) => const SizedBox.shrink(),
                  data: (response) => Container(
                    padding: const EdgeInsets.symmetric(
                        horizontal: 8, vertical: 2),
                    decoration: BoxDecoration(
                      color: response.total > 0
                          ? AppColors.severityWarning.withOpacity(0.15)
                          : colorScheme.surfaceContainerHighest,
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Text(
                      '${response.total}',
                      style: TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.w600,
                        color: response.total > 0
                            ? const Color(0xFFF57F17)
                            : colorScheme.onSurfaceVariant,
                      ),
                    ),
                  ),
                ),
                const Spacer(),
                IconButton(
                  icon: const Icon(Icons.refresh, size: 18),
                  tooltip: 'Refresh conflicts',
                  onPressed: () => ref.invalidate(conflictListProvider),
                ),
              ],
            ),
            const SizedBox(height: AppSpacing.sm),

            // ── Master-Detail ─────────────────────────────────────────
            conflictsAsync.when(
              loading: () => LoadingSkeleton.table(rows: 4, cols: 4),
              error: (err, _) => ErrorState(
                message: 'Failed to load conflicts',
                details: err.toString(),
                onRetry: () => ref.invalidate(conflictListProvider),
              ),
              data: (response) {
                if (response.conflicts.isEmpty) {
                  return const Padding(
                    padding: EdgeInsets.symmetric(vertical: AppSpacing.lg),
                    child: EmptyState(
                      icon: Icons.check_circle_outline,
                      title: 'No conflicts',
                      subtitle:
                          'All merge conflicts have been resolved.',
                    ),
                  );
                }

                return SizedBox(
                  height: 500,
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      // ── Left: Conflict list ────────────────────────
                      Expanded(
                        flex: 2,
                        child: SingleChildScrollView(
                          child: SingleChildScrollView(
                            scrollDirection: Axis.horizontal,
                            child: DataTable(
                              showCheckboxColumn: false,
                              headingRowColor: WidgetStateProperty.all(
                                colorScheme.surfaceContainerHighest
                                    .withOpacity(0.5),
                              ),
                              columns: const [
                                DataColumn(label: Text('Resource Type')),
                                DataColumn(label: Text('Resource ID')),
                                DataColumn(label: Text('Status')),
                                DataColumn(label: Text('Detected At')),
                              ],
                              rows: response.conflicts.map((conflict) {
                                final isSelected =
                                    conflict.id == selectedId;
                                return DataRow(
                                  selected: isSelected,
                                  color: isSelected
                                      ? WidgetStateProperty.all(
                                          colorScheme.primaryContainer
                                              .withOpacity(0.3))
                                      : null,
                                  onSelectChanged: (_) {
                                    ref
                                        .read(selectedConflictProvider
                                            .notifier)
                                        .state = conflict.id;
                                  },
                                  cells: [
                                    DataCell(Text(
                                      conflict.resourceType,
                                      style: const TextStyle(
                                        fontWeight: FontWeight.w500,
                                      ),
                                    )),
                                    DataCell(Text(
                                      conflict.resourceId,
                                      style: const TextStyle(
                                        fontSize: 13,
                                        fontFamily: 'monospace',
                                      ),
                                    )),
                                    DataCell(_ConflictStatusBadge(
                                      status: conflict.status,
                                    )),
                                    DataCell(Text(
                                      _formatTimestamp(
                                          conflict.detectedAt),
                                      style: TextStyle(
                                        fontSize: 13,
                                        color: colorScheme
                                            .onSurfaceVariant,
                                      ),
                                    )),
                                  ],
                                );
                              }).toList(),
                            ),
                          ),
                        ),
                      ),

                      const SizedBox(width: AppSpacing.md),

                      // ── Right: Conflict detail ─────────────────────
                      Expanded(
                        flex: 3,
                        child: selectedId == null
                            ? Center(
                                child: Text(
                                  'Select a conflict to view details',
                                  style: TextStyle(
                                    fontSize: 14,
                                    color: colorScheme.onSurfaceVariant,
                                  ),
                                ),
                              )
                            : _ConflictDetailPane(
                                conflictId: selectedId,
                              ),
                      ),
                    ],
                  ),
                );
              },
            ),
          ],
        ),
      ),
    );
  }

  static String _formatTimestamp(String timestamp) {
    try {
      return DateTime.parse(timestamp).toDisplayDateTime;
    } catch (_) {
      return timestamp;
    }
  }
}

// =============================================================================
// Conflict Status Badge
// =============================================================================

class _ConflictStatusBadge extends StatelessWidget {
  const _ConflictStatusBadge({required this.status});

  final String status;

  @override
  Widget build(BuildContext context) {
    final (Color color, String label) = _resolve(status);

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: color.withOpacity(0.15),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withOpacity(0.4)),
      ),
      child: Text(
        label,
        style: TextStyle(
          fontSize: 11,
          fontWeight: FontWeight.w600,
          color: color,
        ),
      ),
    );
  }

  static (Color, String) _resolve(String status) {
    switch (status.toLowerCase()) {
      case 'pending':
      case 'open':
        return (AppColors.severityWarning, 'Pending');
      case 'resolved':
        return (AppColors.statusActive, 'Resolved');
      case 'deferred':
        return (AppColors.severityInfo, 'Deferred');
      default:
        return (AppColors.statusInactive, status);
    }
  }
}

// =============================================================================
// Conflict Detail Pane
// =============================================================================

class _ConflictDetailPane extends ConsumerWidget {
  const _ConflictDetailPane({required this.conflictId});

  final String conflictId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final detailAsync = ref.watch(conflictDetailProvider(conflictId));
    final colorScheme = Theme.of(context).colorScheme;

    return detailAsync.when(
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (err, _) => ErrorState(
        message: 'Failed to load conflict',
        details: err.toString(),
        onRetry: () =>
            ref.invalidate(conflictDetailProvider(conflictId)),
      ),
      data: (detail) {
        return SingleChildScrollView(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              // ── Header ─────────────────────────────────────────────
              Text(
                '${detail.resourceType} / ${detail.resourceId}',
                style: TextStyle(
                  fontSize: 15,
                  fontWeight: FontWeight.w600,
                  color: colorScheme.onSurface,
                ),
              ),
              const SizedBox(height: AppSpacing.sm),

              // ── Side-by-Side JSON ──────────────────────────────────
              Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Local
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 8,
                            vertical: 4,
                          ),
                          decoration: BoxDecoration(
                            color: AppColors.info.withOpacity(0.1),
                            borderRadius: BorderRadius.circular(
                                AppSpacing.borderRadiusSm),
                          ),
                          child: Text(
                            'Local (${detail.localNode})',
                            style: const TextStyle(
                              fontSize: 12,
                              fontWeight: FontWeight.w600,
                              color: AppColors.info,
                            ),
                          ),
                        ),
                        const SizedBox(height: AppSpacing.xs),
                        if (detail.localVersion != null)
                          JsonViewer(
                            data: detail.localVersion!,
                            initiallyExpanded: false,
                          )
                        else
                          Text(
                            'No local version',
                            style: TextStyle(
                              fontSize: 13,
                              color: colorScheme.onSurfaceVariant,
                            ),
                          ),
                      ],
                    ),
                  ),
                  const SizedBox(width: AppSpacing.sm),

                  // Remote
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 8,
                            vertical: 4,
                          ),
                          decoration: BoxDecoration(
                            color: AppColors.secondary.withOpacity(0.1),
                            borderRadius: BorderRadius.circular(
                                AppSpacing.borderRadiusSm),
                          ),
                          child: Text(
                            'Remote (${detail.remoteNode})',
                            style: const TextStyle(
                              fontSize: 12,
                              fontWeight: FontWeight.w600,
                              color: AppColors.secondary,
                            ),
                          ),
                        ),
                        const SizedBox(height: AppSpacing.xs),
                        if (detail.remoteVersion != null)
                          JsonViewer(
                            data: detail.remoteVersion!,
                            initiallyExpanded: false,
                          )
                        else
                          Text(
                            'No remote version',
                            style: TextStyle(
                              fontSize: 13,
                              color: colorScheme.onSurfaceVariant,
                            ),
                          ),
                      ],
                    ),
                  ),
                ],
              ),
              const SizedBox(height: AppSpacing.md),

              // ── Resolution buttons ─────────────────────────────────
              Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: [
                  OutlinedButton(
                    onPressed: () =>
                        _showDeferDialog(context, ref, detail),
                    child: const Text('Defer'),
                  ),
                  const SizedBox(width: AppSpacing.sm),
                  FilledButton.tonal(
                    onPressed: () =>
                        _resolveConflict(context, ref, detail, 'accept_local'),
                    child: const Text('Accept Local'),
                  ),
                  const SizedBox(width: AppSpacing.sm),
                  FilledButton(
                    onPressed: () =>
                        _resolveConflict(context, ref, detail, 'accept_remote'),
                    child: const Text('Accept Remote'),
                  ),
                ],
              ),
            ],
          ),
        );
      },
    );
  }

  Future<void> _resolveConflict(
    BuildContext context,
    WidgetRef ref,
    ConflictDetail detail,
    String resolution,
  ) async {
    try {
      final api = ref.read(conflictApiProvider);
      await api.resolveConflict(ResolveConflictRequest(
        conflictId: detail.id,
        resolution: resolution,
        author: 'current_user',
      ));
      ref.invalidate(conflictListProvider);
      ref.invalidate(conflictDetailProvider(conflictId));
      ref.read(selectedConflictProvider.notifier).state = null;
      if (!context.mounted) return;
      ScaffoldMessenger.of(context)
        ..hideCurrentSnackBar()
        ..showSnackBar(
          SnackBar(
            content: Text('Conflict resolved: $resolution'),
            behavior: SnackBarBehavior.floating,
          ),
        );
    } catch (e) {
      if (!context.mounted) return;
      ScaffoldMessenger.of(context)
        ..hideCurrentSnackBar()
        ..showSnackBar(
          SnackBar(
            content: Text('Failed to resolve conflict: $e'),
            behavior: SnackBarBehavior.floating,
          ),
        );
    }
  }

  Future<void> _showDeferDialog(
    BuildContext context,
    WidgetRef ref,
    ConflictDetail detail,
  ) async {
    final reasonController = TextEditingController();

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) {
        return AlertDialog(
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(AppSpacing.borderRadiusLg),
          ),
          title: const Text('Defer Conflict'),
          content: TextField(
            controller: reasonController,
            maxLines: 3,
            decoration: InputDecoration(
              labelText: 'Reason for deferral',
              alignLabelWithHint: true,
              border: OutlineInputBorder(
                borderRadius:
                    BorderRadius.circular(AppSpacing.borderRadiusMd),
              ),
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(ctx).pop(false),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: () => Navigator.of(ctx).pop(true),
              child: const Text('Defer'),
            ),
          ],
        );
      },
    );

    if (confirmed == true && reasonController.text.trim().isNotEmpty) {
      try {
        final api = ref.read(conflictApiProvider);
        await api.deferConflict(DeferConflictRequest(
          conflictId: detail.id,
          reason: reasonController.text.trim(),
        ));
        ref.invalidate(conflictListProvider);
        ref.invalidate(conflictDetailProvider(conflictId));
        ref.read(selectedConflictProvider.notifier).state = null;
        if (!context.mounted) return;
        ScaffoldMessenger.of(context)
          ..hideCurrentSnackBar()
          ..showSnackBar(
            const SnackBar(
              content: Text('Conflict deferred'),
              behavior: SnackBarBehavior.floating,
            ),
          );
      } catch (e) {
        if (!context.mounted) return;
        ScaffoldMessenger.of(context)
          ..hideCurrentSnackBar()
          ..showSnackBar(
            SnackBar(
              content: Text('Failed to defer conflict: $e'),
              behavior: SnackBarBehavior.floating,
            ),
          );
      }
    }

    reasonController.dispose();
  }
}
