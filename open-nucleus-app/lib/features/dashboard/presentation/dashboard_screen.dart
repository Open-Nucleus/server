import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/extensions/date_extensions.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_spacing.dart';
import '../../../shared/widgets/error_state.dart';
import '../../../shared/widgets/loading_skeleton.dart';
import '../../../shared/models/alert_models.dart';
import '../../../shared/models/anchor_models.dart';
import '../../../shared/models/sync_models.dart';
import '../../../shared/widgets/status_indicator.dart';
import '../data/dashboard_models.dart';
import 'dashboard_providers.dart';

/// Full dashboard screen with responsive grid of summary cards.
///
/// Fetches health, patient count, alert summary, sync status, and anchor
/// status in parallel and displays them in a 2-3 column grid.
class DashboardScreen extends ConsumerWidget {
  const DashboardScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final asyncData = ref.watch(dashboardDataProvider);

    return asyncData.when(
      loading: () => const Padding(
        padding: AppSpacing.pagePadding,
        child: _DashboardSkeleton(),
      ),
      error: (error, _) => ErrorState(
        message: 'Failed to load dashboard',
        details: error.toString(),
        onRetry: () => ref.invalidate(dashboardDataProvider),
      ),
      data: (data) => _DashboardContent(data: data),
    );
  }
}

// ---------------------------------------------------------------------------
// Main content
// ---------------------------------------------------------------------------

class _DashboardContent extends StatelessWidget {
  const _DashboardContent({required this.data});

  final DashboardData data;

  @override
  Widget build(BuildContext context) {
    final width = MediaQuery.of(context).size.width;
    // Responsive columns: 3 columns above 1200px, 2 columns above 800px.
    final crossAxisCount = width > 1200 ? 3 : 2;

    return SingleChildScrollView(
      padding: AppSpacing.pagePadding,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // ── Title Row ──────────────────────────────────────────────
          Row(
            children: [
              Text(
                'Dashboard',
                style: TextStyle(
                  fontSize: 24,
                  fontWeight: FontWeight.w700,
                  color: Theme.of(context).colorScheme.onSurface,
                ),
              ),
              const SizedBox(width: AppSpacing.md),
              StatusIndicator(
                color: data.healthy
                    ? AppColors.statusActive
                    : AppColors.statusError,
                label: data.healthy ? 'System Healthy' : 'System Unhealthy',
              ),
            ],
          ),
          const SizedBox(height: AppSpacing.lg),

          // ── Card Grid ──────────────────────────────────────────────
          GridView.count(
            crossAxisCount: crossAxisCount,
            mainAxisSpacing: AppSpacing.md,
            crossAxisSpacing: AppSpacing.md,
            childAspectRatio: 1.6,
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            children: [
              _NodeIdentityCard(data: data),
              _PatientStatsCard(count: data.patientCount),
              _AlertSummaryCard(summary: data.alertSummary),
              _SyncStatusCard(sync: data.syncStatus),
              _AnchorStatusCard(anchor: data.anchorStatus),
              const _QuickActionsCard(),
            ],
          ),
          const SizedBox(height: AppSpacing.md),

          // ── Recent Activity ────────────────────────────────────────
          const _RecentActivityCard(),
        ],
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Node Identity Card
// ---------------------------------------------------------------------------

class _NodeIdentityCard extends StatelessWidget {
  const _NodeIdentityCard({required this.data});

  final DashboardData data;

  @override
  Widget build(BuildContext context) {
    return _DashboardCard(
      title: 'Node Identity',
      icon: Icons.hub_outlined,
      iconColor: AppColors.primary,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          _LabelValue(
            label: 'Node ID',
            value: data.nodeId ?? 'Unknown',
            mono: true,
          ),
          const SizedBox(height: AppSpacing.xs),
          _LabelValue(
            label: 'Site ID',
            value: data.siteId ?? 'Unknown',
            mono: true,
          ),
          const SizedBox(height: AppSpacing.xs),
          _LabelValue(
            label: 'Role',
            value: data.syncStatus?.state == 'syncing' ? 'Active' : 'Standalone',
          ),
          const SizedBox(height: AppSpacing.sm),
          StatusIndicator(
            color: data.healthy
                ? AppColors.statusActive
                : AppColors.statusError,
            label: data.healthy ? 'Online' : 'Offline',
          ),
        ],
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Patient Stats Card
// ---------------------------------------------------------------------------

class _PatientStatsCard extends StatelessWidget {
  const _PatientStatsCard({required this.count});

  final int count;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return _DashboardCard(
      title: 'Patients',
      icon: Icons.people_outlined,
      iconColor: AppColors.secondary,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            count.toString(),
            style: TextStyle(
              fontSize: 40,
              fontWeight: FontWeight.w800,
              color: colorScheme.onSurface,
              height: 1.1,
            ),
          ),
          const SizedBox(height: AppSpacing.xs),
          Text(
            'Total registered patients',
            style: TextStyle(
              fontSize: 12,
              color: colorScheme.onSurfaceVariant,
            ),
          ),
          const Spacer(),
          Align(
            alignment: Alignment.bottomRight,
            child: Builder(
              builder: (context) => TextButton(
                onPressed: () => GoRouter.of(context).go('/patients'),
                child: const Text('View All'),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Alert Summary Card
// ---------------------------------------------------------------------------

class _AlertSummaryCard extends StatelessWidget {
  const _AlertSummaryCard({this.summary});

  final AlertSummaryResponse? summary;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final s = summary;

    return _DashboardCard(
      title: 'Alerts',
      icon: Icons.warning_amber_rounded,
      iconColor: AppColors.warning,
      child: s == null
          ? Text(
              'Unable to load alert data',
              style: TextStyle(
                fontSize: 13,
                color: colorScheme.onSurfaceVariant,
              ),
            )
          : Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                Row(
                  children: [
                    _AlertCount(
                      label: 'Critical',
                      count: s.critical,
                      color: AppColors.severityCritical,
                    ),
                    const SizedBox(width: AppSpacing.md),
                    _AlertCount(
                      label: 'Warning',
                      count: s.warning,
                      color: AppColors.severityWarning,
                    ),
                    const SizedBox(width: AppSpacing.md),
                    _AlertCount(
                      label: 'Info',
                      count: s.info,
                      color: AppColors.severityInfo,
                    ),
                  ],
                ),
                const SizedBox(height: AppSpacing.sm),
                Text(
                  '${s.unacknowledged} unacknowledged of ${s.total} total',
                  style: TextStyle(
                    fontSize: 12,
                    color: colorScheme.onSurfaceVariant,
                  ),
                ),
              ],
            ),
    );
  }
}

class _AlertCount extends StatelessWidget {
  const _AlertCount({
    required this.label,
    required this.count,
    required this.color,
  });

  final String label;
  final int count;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 8,
              height: 8,
              decoration: BoxDecoration(
                color: color,
                shape: BoxShape.circle,
              ),
            ),
            const SizedBox(width: 4),
            Text(
              count.toString(),
              style: TextStyle(
                fontSize: 20,
                fontWeight: FontWeight.w700,
                color: color,
              ),
            ),
          ],
        ),
        Text(
          label,
          style: TextStyle(
            fontSize: 11,
            color: Theme.of(context).colorScheme.onSurfaceVariant,
          ),
        ),
      ],
    );
  }
}

// ---------------------------------------------------------------------------
// Sync Status Card
// ---------------------------------------------------------------------------

class _SyncStatusCard extends StatelessWidget {
  const _SyncStatusCard({this.sync});

  final SyncStatusResponse? sync;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final s = sync;

    return _DashboardCard(
      title: 'Sync',
      icon: Icons.sync_outlined,
      iconColor: AppColors.syncSyncing,
      child: s == null
          ? Text(
              'Unable to load sync data',
              style: TextStyle(
                fontSize: 13,
                color: colorScheme.onSurfaceVariant,
              ),
            )
          : Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                Row(
                  children: [
                    StatusIndicator(
                      color: _syncStateColor(s.state),
                      label: _syncStateLabel(s.state),
                    ),
                  ],
                ),
                const SizedBox(height: AppSpacing.sm),
                _LabelValue(
                  label: 'Last Sync',
                  value: _formatTimeAgo(s.lastSync),
                ),
                const SizedBox(height: AppSpacing.xs),
                _LabelValue(
                  label: 'Pending',
                  value: '${s.pendingChanges} changes',
                ),
              ],
            ),
    );
  }

  static Color _syncStateColor(String state) {
    switch (state.toLowerCase()) {
      case 'idle':
        return AppColors.syncIdle;
      case 'syncing':
        return AppColors.syncSyncing;
      case 'error':
        return AppColors.syncError;
      case 'complete':
        return AppColors.syncComplete;
      default:
        return AppColors.syncIdle;
    }
  }

  static String _syncStateLabel(String state) {
    switch (state.toLowerCase()) {
      case 'idle':
        return 'Idle';
      case 'syncing':
        return 'Syncing...';
      case 'error':
        return 'Error';
      case 'complete':
        return 'Complete';
      default:
        return state;
    }
  }
}

// ---------------------------------------------------------------------------
// Anchor Status Card
// ---------------------------------------------------------------------------

class _AnchorStatusCard extends StatelessWidget {
  const _AnchorStatusCard({this.anchor});

  final AnchorStatusResponse? anchor;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final a = anchor;

    return _DashboardCard(
      title: 'Integrity Anchoring',
      icon: Icons.verified_outlined,
      iconColor: AppColors.success,
      child: a == null
          ? Text(
              'Unable to load anchor data',
              style: TextStyle(
                fontSize: 13,
                color: colorScheme.onSurfaceVariant,
              ),
            )
          : Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                _LabelValue(
                  label: 'Merkle Root',
                  value: _truncateHash(a.merkleRoot),
                  mono: true,
                ),
                const SizedBox(height: AppSpacing.xs),
                _LabelValue(
                  label: 'Last Anchored',
                  value: _formatTimeAgo(a.lastAnchorTime),
                ),
                const SizedBox(height: AppSpacing.xs),
                _LabelValue(
                  label: 'Queue Depth',
                  value: a.queueDepth.toString(),
                ),
              ],
            ),
    );
  }

  static String _truncateHash(String hash) {
    if (hash.length <= 16) return hash;
    return '${hash.substring(0, 8)}...${hash.substring(hash.length - 8)}';
  }
}

// ---------------------------------------------------------------------------
// Quick Actions Card
// ---------------------------------------------------------------------------

class _QuickActionsCard extends StatelessWidget {
  const _QuickActionsCard();

  @override
  Widget build(BuildContext context) {
    return _DashboardCard(
      title: 'Quick Actions',
      icon: Icons.bolt_outlined,
      iconColor: AppColors.primary,
      child: Wrap(
        spacing: AppSpacing.sm,
        runSpacing: AppSpacing.sm,
        children: [
          FilledButton.icon(
            onPressed: () => GoRouter.of(context).go('/patients/new'),
            icon: const Icon(Icons.person_add_outlined, size: 18),
            label: const Text('New Patient'),
          ),
          OutlinedButton.icon(
            onPressed: () => GoRouter.of(context).go('/sync'),
            icon: const Icon(Icons.sync, size: 18),
            label: const Text('Trigger Sync'),
          ),
          OutlinedButton.icon(
            onPressed: () => GoRouter.of(context).go('/alerts'),
            icon: const Icon(Icons.notifications_outlined, size: 18),
            label: const Text('View Alerts'),
          ),
        ],
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Recent Activity Card
// ---------------------------------------------------------------------------

class _RecentActivityCard extends StatelessWidget {
  const _RecentActivityCard();

  @override
  Widget build(BuildContext context) {
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
                Icon(
                  Icons.history_outlined,
                  size: 20,
                  color: colorScheme.onSurfaceVariant,
                ),
                const SizedBox(width: AppSpacing.sm),
                Text(
                  'Recent Activity',
                  style: TextStyle(
                    fontSize: 15,
                    fontWeight: FontWeight.w600,
                    color: colorScheme.onSurface,
                  ),
                ),
              ],
            ),
            const SizedBox(height: AppSpacing.md),
            Center(
              child: Padding(
                padding: const EdgeInsets.symmetric(vertical: AppSpacing.lg),
                child: Text(
                  'No recent activity',
                  style: TextStyle(
                    fontSize: 13,
                    color: colorScheme.onSurfaceVariant,
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Skeleton loader for the dashboard
// ---------------------------------------------------------------------------

class _DashboardSkeleton extends StatelessWidget {
  const _DashboardSkeleton();

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const LoadingSkeleton(width: 200, height: 28),
        const SizedBox(height: AppSpacing.lg),
        GridView.count(
          crossAxisCount: 2,
          mainAxisSpacing: AppSpacing.md,
          crossAxisSpacing: AppSpacing.md,
          childAspectRatio: 1.6,
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          children: List.generate(
            6,
            (_) => LoadingSkeleton.card(height: 160),
          ),
        ),
        const SizedBox(height: AppSpacing.md),
        LoadingSkeleton.card(height: 120),
      ],
    );
  }
}

// ---------------------------------------------------------------------------
// Shared dashboard card wrapper
// ---------------------------------------------------------------------------

class _DashboardCard extends StatelessWidget {
  const _DashboardCard({
    required this.title,
    required this.icon,
    required this.iconColor,
    required this.child,
  });

  final String title;
  final IconData icon;
  final Color iconColor;
  final Widget child;

  @override
  Widget build(BuildContext context) {
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
            // ── Card Header ──────────────────────────────────────────
            Row(
              children: [
                Icon(icon, size: 20, color: iconColor),
                const SizedBox(width: AppSpacing.sm),
                Text(
                  title,
                  style: TextStyle(
                    fontSize: 15,
                    fontWeight: FontWeight.w600,
                    color: colorScheme.onSurface,
                  ),
                ),
              ],
            ),
            const SizedBox(height: AppSpacing.md),

            // ── Card Body ────────────────────────────────────────────
            Expanded(child: child),
          ],
        ),
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

/// A label: value row for dashboard cards.
class _LabelValue extends StatelessWidget {
  const _LabelValue({
    required this.label,
    required this.value,
    this.mono = false,
  });

  final String label;
  final String value;
  final bool mono;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(
          '$label: ',
          style: TextStyle(
            fontSize: 12,
            fontWeight: FontWeight.w500,
            color: colorScheme.onSurfaceVariant,
          ),
        ),
        Flexible(
          child: Text(
            value,
            style: TextStyle(
              fontSize: 12,
              fontWeight: FontWeight.w600,
              color: colorScheme.onSurface,
              fontFamily: mono ? 'monospace' : null,
            ),
            overflow: TextOverflow.ellipsis,
          ),
        ),
      ],
    );
  }
}

/// Parses an ISO 8601 timestamp and returns a human-readable time-ago string.
String _formatTimeAgo(String isoTimestamp) {
  try {
    final dt = DateTime.parse(isoTimestamp);
    return dt.timeAgo;
  } catch (_) {
    return isoTimestamp;
  }
}

