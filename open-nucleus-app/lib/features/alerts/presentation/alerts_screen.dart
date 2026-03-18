import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/extensions/date_extensions.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_spacing.dart';
import '../../../shared/models/alert_models.dart';
import '../../../shared/widgets/empty_state.dart';
import '../../../shared/widgets/error_state.dart';
import '../../../shared/widgets/loading_skeleton.dart';
import '../../../shared/widgets/pagination_controls.dart';
import '../../../shared/widgets/severity_badge.dart';
import 'alerts_providers.dart';

/// Full alerts screen with summary cards, filterable data table, and a detail
/// panel for acknowledge/dismiss actions.
class AlertsScreen extends ConsumerWidget {
  const AlertsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Padding(
      padding: AppSpacing.pagePadding,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          // ── Header ──────────────────────────────────────────────────
          Text(
            'Alerts',
            style: TextStyle(
              fontSize: 24,
              fontWeight: FontWeight.w700,
              color: Theme.of(context).colorScheme.onSurface,
            ),
          ),
          const SizedBox(height: AppSpacing.md),

          // ── Summary Cards ───────────────────────────────────────────
          const _SummaryCardsRow(),
          const SizedBox(height: AppSpacing.md),

          // ── Filters ─────────────────────────────────────────────────
          const _FilterRow(),
          const SizedBox(height: AppSpacing.sm),

          // ── Content ─────────────────────────────────────────────────
          const Expanded(child: _AlertsContent()),
        ],
      ),
    );
  }
}

// =============================================================================
// Summary Cards Row
// =============================================================================

class _SummaryCardsRow extends ConsumerWidget {
  const _SummaryCardsRow();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final summaryAsync = ref.watch(alertSummaryProvider);

    return summaryAsync.when(
      loading: () => Row(
        children: List.generate(
          4,
          (i) => Expanded(
            child: Padding(
              padding: EdgeInsets.only(
                  right: i < 3 ? AppSpacing.sm : 0),
              child: LoadingSkeleton.card(height: 80),
            ),
          ),
        ),
      ),
      error: (err, _) => const SizedBox.shrink(),
      data: (summary) {
        return Row(
          children: [
            _SummaryCard(
              label: 'Total',
              count: summary.total,
              color: Theme.of(context).colorScheme.primary,
              icon: Icons.notifications,
            ),
            const SizedBox(width: AppSpacing.sm),
            _SummaryCard(
              label: 'Critical',
              count: summary.critical,
              color: AppColors.severityCritical,
              icon: Icons.error,
            ),
            const SizedBox(width: AppSpacing.sm),
            _SummaryCard(
              label: 'Warning',
              count: summary.warning,
              color: const Color(0xFFF57F17),
              icon: Icons.warning,
            ),
            const SizedBox(width: AppSpacing.sm),
            _SummaryCard(
              label: 'Info',
              count: summary.info,
              color: AppColors.severityInfo,
              icon: Icons.info,
            ),
            const SizedBox(width: AppSpacing.sm),
            _SummaryCard(
              label: 'Unacknowledged',
              count: summary.unacknowledged,
              color: AppColors.statusPending,
              icon: Icons.visibility_off,
            ),
          ],
        );
      },
    );
  }
}

class _SummaryCard extends StatelessWidget {
  const _SummaryCard({
    required this.label,
    required this.count,
    required this.color,
    required this.icon,
  });

  final String label;
  final int count;
  final Color color;
  final IconData icon;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Expanded(
      child: Card(
        elevation: 1,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(AppSpacing.borderRadiusLg),
        ),
        child: Padding(
          padding: const EdgeInsets.symmetric(
            horizontal: AppSpacing.md,
            vertical: AppSpacing.sm,
          ),
          child: Row(
            children: [
              Container(
                width: 40,
                height: 40,
                decoration: BoxDecoration(
                  color: color.withOpacity(0.12),
                  borderRadius:
                      BorderRadius.circular(AppSpacing.borderRadiusMd),
                ),
                child: Icon(icon, size: 20, color: color),
              ),
              const SizedBox(width: AppSpacing.sm),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      '$count',
                      style: TextStyle(
                        fontSize: 20,
                        fontWeight: FontWeight.w700,
                        color: colorScheme.onSurface,
                      ),
                    ),
                    Text(
                      label,
                      style: TextStyle(
                        fontSize: 12,
                        color: colorScheme.onSurfaceVariant,
                      ),
                      overflow: TextOverflow.ellipsis,
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

// =============================================================================
// Filter Row
// =============================================================================

class _FilterRow extends ConsumerWidget {
  const _FilterRow();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final listState = ref.watch(alertListProvider);

    return Row(
      children: [
        // Severity filter
        SizedBox(
          width: 160,
          child: DropdownButtonFormField<String?>(
            value: listState.severityFilter,
            decoration: InputDecoration(
              labelText: 'Severity',
              isDense: true,
              contentPadding: const EdgeInsets.symmetric(
                horizontal: AppSpacing.sm,
                vertical: AppSpacing.sm,
              ),
              border: OutlineInputBorder(
                borderRadius:
                    BorderRadius.circular(AppSpacing.borderRadiusMd),
              ),
            ),
            style: const TextStyle(fontSize: 13),
            items: const [
              DropdownMenuItem(value: null, child: Text('All')),
              DropdownMenuItem(value: 'critical', child: Text('Critical')),
              DropdownMenuItem(value: 'warning', child: Text('Warning')),
              DropdownMenuItem(value: 'info', child: Text('Info')),
            ],
            onChanged: (v) {
              ref.read(alertListProvider.notifier).setSeverityFilter(v);
            },
          ),
        ),
        const SizedBox(width: AppSpacing.sm),

        // Status filter
        SizedBox(
          width: 180,
          child: DropdownButtonFormField<String?>(
            value: listState.statusFilter,
            decoration: InputDecoration(
              labelText: 'Status',
              isDense: true,
              contentPadding: const EdgeInsets.symmetric(
                horizontal: AppSpacing.sm,
                vertical: AppSpacing.sm,
              ),
              border: OutlineInputBorder(
                borderRadius:
                    BorderRadius.circular(AppSpacing.borderRadiusMd),
              ),
            ),
            style: const TextStyle(fontSize: 13),
            items: const [
              DropdownMenuItem(value: null, child: Text('All')),
              DropdownMenuItem(value: 'active', child: Text('Active')),
              DropdownMenuItem(
                  value: 'acknowledged', child: Text('Acknowledged')),
              DropdownMenuItem(value: 'dismissed', child: Text('Dismissed')),
            ],
            onChanged: (v) {
              ref.read(alertListProvider.notifier).setStatusFilter(v);
            },
          ),
        ),
        const SizedBox(width: AppSpacing.sm),

        // Clear filters
        if (listState.severityFilter != null ||
            listState.statusFilter != null)
          TextButton.icon(
            onPressed: () {
              ref.read(alertListProvider.notifier).clearFilters();
            },
            icon: const Icon(Icons.clear, size: 16),
            label: const Text('Clear'),
          ),
      ],
    );
  }
}

// =============================================================================
// Alerts Content (table + detail)
// =============================================================================

class _AlertsContent extends ConsumerWidget {
  const _AlertsContent();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final listState = ref.watch(alertListProvider);
    final selectedId = ref.watch(selectedAlertProvider);

    if (listState.isLoading && listState.alerts.isEmpty) {
      return LoadingSkeleton.table(rows: 8, cols: 6);
    }

    if (listState.error != null && listState.alerts.isEmpty) {
      return ErrorState(
        message: 'Failed to load alerts',
        details: listState.error,
        onRetry: () => ref.read(alertListProvider.notifier).fetch(),
      );
    }

    if (listState.alerts.isEmpty) {
      return const EmptyState(
        icon: Icons.notifications_none,
        title: 'No alerts',
        subtitle: 'No Sentinel alerts match the current filters.',
      );
    }

    return Column(
      children: [
        Expanded(
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // ── Left: Alert table ──────────────────────────────────
              Expanded(
                flex: selectedId != null ? 3 : 1,
                child: _AlertDataTable(
                  alerts: listState.alerts,
                  selectedId: selectedId,
                  isLoading: listState.isLoading,
                ),
              ),

              // ── Right: Detail panel ────────────────────────────────
              if (selectedId != null) ...[
                const SizedBox(width: AppSpacing.md),
                Expanded(
                  flex: 2,
                  child: _AlertDetailPanel(alertId: selectedId),
                ),
              ],
            ],
          ),
        ),

        // ── Pagination ───────────────────────────────────────────────
        PaginationControls(
          currentPage: listState.page,
          totalPages: listState.totalPages,
          totalItems: listState.totalItems,
          rowsPerPage: listState.perPage,
          onPageChanged: (page) {
            ref.read(alertListProvider.notifier).goToPage(page);
          },
          onRowsPerPageChanged: (perPage) {
            ref.read(alertListProvider.notifier).setPerPage(perPage);
          },
        ),
      ],
    );
  }
}

// =============================================================================
// Alert Data Table
// =============================================================================

class _AlertDataTable extends ConsumerWidget {
  const _AlertDataTable({
    required this.alerts,
    required this.selectedId,
    required this.isLoading,
  });

  final List<AlertDetail> alerts;
  final String? selectedId;
  final bool isLoading;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final colorScheme = Theme.of(context).colorScheme;

    return Card(
      elevation: 1,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(AppSpacing.borderRadiusLg),
      ),
      child: Stack(
        children: [
          SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            child: SingleChildScrollView(
              child: DataTable(
                showCheckboxColumn: false,
                headingRowColor: WidgetStateProperty.all(
                  colorScheme.surfaceContainerHighest.withOpacity(0.5),
                ),
                columns: const [
                  DataColumn(label: Text('Severity')),
                  DataColumn(label: Text('Type')),
                  DataColumn(label: Text('Title')),
                  DataColumn(label: Text('Patient ID')),
                  DataColumn(label: Text('Status')),
                  DataColumn(label: Text('Created At')),
                ],
                rows: alerts
                    .map((alert) => _buildRow(context, ref, alert))
                    .toList(),
              ),
            ),
          ),

          // Loading overlay
          if (isLoading)
            Positioned.fill(
              child: Container(
                color: colorScheme.surface.withOpacity(0.5),
                child: const Center(child: CircularProgressIndicator()),
              ),
            ),
        ],
      ),
    );
  }

  DataRow _buildRow(
    BuildContext context,
    WidgetRef ref,
    AlertDetail alert,
  ) {
    final colorScheme = Theme.of(context).colorScheme;
    final isSelected = alert.id == selectedId;

    return DataRow(
      selected: isSelected,
      color: isSelected
          ? WidgetStateProperty.all(
              colorScheme.primaryContainer.withOpacity(0.3))
          : null,
      onSelectChanged: (_) {
        ref.read(selectedAlertProvider.notifier).state = alert.id;
      },
      cells: [
        // Severity
        DataCell(SeverityBadge(severity: alert.severity)),

        // Type
        DataCell(
          Text(
            alert.type,
            style: TextStyle(
              fontSize: 13,
              color: colorScheme.onSurfaceVariant,
            ),
          ),
        ),

        // Title
        DataCell(
          ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 250),
            child: Text(
              alert.title,
              style: const TextStyle(
                fontWeight: FontWeight.w500,
              ),
              overflow: TextOverflow.ellipsis,
            ),
          ),
        ),

        // Patient ID
        DataCell(
          Text(
            alert.patientId,
            style: TextStyle(
              fontSize: 13,
              fontFamily: 'monospace',
              color: colorScheme.onSurfaceVariant,
            ),
          ),
        ),

        // Status
        DataCell(_AlertStatusBadge(status: alert.status)),

        // Created At
        DataCell(
          Text(
            _formatTimestamp(alert.createdAt),
            style: TextStyle(
              fontSize: 13,
              color: colorScheme.onSurfaceVariant,
            ),
          ),
        ),
      ],
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
// Alert Status Badge
// =============================================================================

class _AlertStatusBadge extends StatelessWidget {
  const _AlertStatusBadge({required this.status});

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
      case 'active':
        return (AppColors.severityWarning, 'Active');
      case 'acknowledged':
        return (AppColors.severityInfo, 'Acknowledged');
      case 'dismissed':
        return (AppColors.statusInactive, 'Dismissed');
      default:
        return (AppColors.statusInactive, status);
    }
  }
}

// =============================================================================
// Alert Detail Panel
// =============================================================================

class _AlertDetailPanel extends ConsumerWidget {
  const _AlertDetailPanel({required this.alertId});

  final String alertId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final detailAsync = ref.watch(alertDetailProvider(alertId));
    final colorScheme = Theme.of(context).colorScheme;

    return Card(
      elevation: 1,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(AppSpacing.borderRadiusLg),
      ),
      child: Padding(
        padding: AppSpacing.cardPadding,
        child: detailAsync.when(
          loading: () => const Center(child: CircularProgressIndicator()),
          error: (err, _) => ErrorState(
            message: 'Failed to load alert',
            details: err.toString(),
            onRetry: () =>
                ref.invalidate(alertDetailProvider(alertId)),
          ),
          data: (alert) {
            return SingleChildScrollView(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // ── Header ─────────────────────────────────────────
                  Row(
                    children: [
                      SeverityBadge(severity: alert.severity),
                      const SizedBox(width: AppSpacing.sm),
                      Expanded(
                        child: Text(
                          alert.title,
                          style: TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.w600,
                            color: colorScheme.onSurface,
                          ),
                        ),
                      ),
                      IconButton(
                        icon: const Icon(Icons.close, size: 18),
                        tooltip: 'Close detail',
                        onPressed: () {
                          ref.read(selectedAlertProvider.notifier).state =
                              null;
                        },
                      ),
                    ],
                  ),
                  const SizedBox(height: AppSpacing.sm),

                  // ── Metadata ───────────────────────────────────────
                  _DetailField(
                    label: 'Type',
                    value: alert.type,
                  ),
                  _DetailField(
                    label: 'Patient ID',
                    value: alert.patientId,
                    monospace: true,
                  ),
                  _DetailField(
                    label: 'Status',
                    value: alert.status,
                  ),
                  _DetailField(
                    label: 'Created At',
                    value: _formatTimestamp(alert.createdAt),
                  ),
                  if (alert.acknowledgedAt != null)
                    _DetailField(
                      label: 'Acknowledged At',
                      value: _formatTimestamp(alert.acknowledgedAt!),
                    ),
                  if (alert.acknowledgedBy != null)
                    _DetailField(
                      label: 'Acknowledged By',
                      value: alert.acknowledgedBy!,
                    ),
                  const SizedBox(height: AppSpacing.sm),

                  // ── Description ────────────────────────────────────
                  Text(
                    'Description',
                    style: TextStyle(
                      fontSize: 12,
                      fontWeight: FontWeight.w600,
                      color: colorScheme.onSurfaceVariant,
                    ),
                  ),
                  const SizedBox(height: AppSpacing.xs),
                  Container(
                    width: double.infinity,
                    padding: const EdgeInsets.all(AppSpacing.sm),
                    decoration: BoxDecoration(
                      color: colorScheme.surfaceContainerLowest,
                      borderRadius: BorderRadius.circular(
                          AppSpacing.borderRadiusMd),
                      border: Border.all(
                          color: colorScheme.outlineVariant),
                    ),
                    child: Text(
                      alert.description,
                      style: TextStyle(
                        fontSize: 13,
                        color: colorScheme.onSurface,
                      ),
                    ),
                  ),
                  const SizedBox(height: AppSpacing.md),

                  // ── Action Buttons ─────────────────────────────────
                  Row(
                    mainAxisAlignment: MainAxisAlignment.end,
                    children: [
                      if (alert.status.toLowerCase() != 'dismissed')
                        OutlinedButton.icon(
                          onPressed: () =>
                              _showDismissDialog(context, ref, alert),
                          icon: const Icon(Icons.cancel_outlined,
                              size: 18),
                          label: const Text('Dismiss'),
                        ),
                      if (alert.status.toLowerCase() != 'dismissed')
                        const SizedBox(width: AppSpacing.sm),
                      if (alert.status.toLowerCase() == 'active')
                        FilledButton.icon(
                          onPressed: () =>
                              _acknowledgeAlert(context, ref, alert),
                          icon: const Icon(Icons.check_circle_outline,
                              size: 18),
                          label: const Text('Acknowledge'),
                        ),
                    ],
                  ),
                ],
              ),
            );
          },
        ),
      ),
    );
  }

  Future<void> _acknowledgeAlert(
    BuildContext context,
    WidgetRef ref,
    AlertDetail alert,
  ) async {
    try {
      final api = ref.read(alertApiProvider);
      await api.acknowledgeAlert(alert.id);
      ref.invalidate(alertDetailProvider(alertId));
      ref.read(alertListProvider.notifier).fetch();
      ref.invalidate(alertSummaryProvider);
      if (!context.mounted) return;
      ScaffoldMessenger.of(context)
        ..hideCurrentSnackBar()
        ..showSnackBar(
          const SnackBar(
            content: Text('Alert acknowledged'),
            behavior: SnackBarBehavior.floating,
          ),
        );
    } catch (e) {
      if (!context.mounted) return;
      ScaffoldMessenger.of(context)
        ..hideCurrentSnackBar()
        ..showSnackBar(
          SnackBar(
            content: Text('Failed to acknowledge alert: $e'),
            behavior: SnackBarBehavior.floating,
          ),
        );
    }
  }

  Future<void> _showDismissDialog(
    BuildContext context,
    WidgetRef ref,
    AlertDetail alert,
  ) async {
    final reasonController = TextEditingController();

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) {
        return AlertDialog(
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(AppSpacing.borderRadiusLg),
          ),
          title: const Text('Dismiss Alert'),
          content: TextField(
            controller: reasonController,
            maxLines: 3,
            decoration: InputDecoration(
              labelText: 'Reason for dismissal',
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
              child: const Text('Dismiss'),
            ),
          ],
        );
      },
    );

    if (confirmed == true && reasonController.text.trim().isNotEmpty) {
      try {
        final api = ref.read(alertApiProvider);
        await api.dismissAlert(alert.id, reasonController.text.trim());
        ref.invalidate(alertDetailProvider(alertId));
        ref.read(alertListProvider.notifier).fetch();
        ref.invalidate(alertSummaryProvider);
        if (!context.mounted) return;
        ScaffoldMessenger.of(context)
          ..hideCurrentSnackBar()
          ..showSnackBar(
            const SnackBar(
              content: Text('Alert dismissed'),
              behavior: SnackBarBehavior.floating,
            ),
          );
      } catch (e) {
        if (!context.mounted) return;
        ScaffoldMessenger.of(context)
          ..hideCurrentSnackBar()
          ..showSnackBar(
            SnackBar(
              content: Text('Failed to dismiss alert: $e'),
              behavior: SnackBarBehavior.floating,
            ),
          );
      }
    }

    reasonController.dispose();
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
// Detail Field Helper
// =============================================================================

class _DetailField extends StatelessWidget {
  const _DetailField({
    required this.label,
    required this.value,
    this.monospace = false,
  });

  final String label;
  final String value;
  final bool monospace;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Padding(
      padding: const EdgeInsets.only(bottom: AppSpacing.xs),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 120,
            child: Text(
              label,
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.w600,
                color: colorScheme.onSurfaceVariant,
              ),
            ),
          ),
          Expanded(
            child: Text(
              value,
              style: TextStyle(
                fontSize: 13,
                fontFamily: monospace ? 'monospace' : null,
                color: colorScheme.onSurface,
              ),
            ),
          ),
        ],
      ),
    );
  }
}
