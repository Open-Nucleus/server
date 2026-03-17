import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/theme/app_spacing.dart';
import '../../../shared/widgets/status_indicator.dart';
import '../providers/shell_providers.dart';

/// The horizontal top bar displayed above the content area.
///
/// Shows (left to right): page title, search field, connection indicator,
/// node/site chips, role badge, and notification bell.
class TopBar extends ConsumerWidget {
  const TopBar({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final pageTitle = ref.watch(currentPageTitleProvider);
    final connectionState = ref.watch(connectionProvider);
    final alertCount = ref.watch(unacknowledgedAlertCountProvider);
    final colorScheme = Theme.of(context).colorScheme;

    return Container(
      height: AppSpacing.topBarHeight,
      padding: const EdgeInsets.symmetric(horizontal: AppSpacing.md),
      decoration: BoxDecoration(
        color: colorScheme.surface,
        border: Border(
          bottom: BorderSide(color: colorScheme.outlineVariant),
        ),
      ),
      child: Row(
        children: [
          // ── Page Title ───────────────────────────────────────────────
          Text(
            pageTitle,
            style: TextStyle(
              fontSize: 18,
              fontWeight: FontWeight.w600,
              color: colorScheme.onSurface,
            ),
          ),

          const Spacer(),

          // ── Search Field ─────────────────────────────────────────────
          SizedBox(
            width: 260,
            height: 36,
            child: TextField(
              decoration: InputDecoration(
                hintText: 'Search...  (Ctrl+K)',
                hintStyle: TextStyle(
                  fontSize: 13,
                  color: colorScheme.onSurfaceVariant,
                ),
                prefixIcon: Icon(
                  Icons.search,
                  size: 18,
                  color: colorScheme.onSurfaceVariant,
                ),
                isDense: true,
                contentPadding: const EdgeInsets.symmetric(
                  vertical: AppSpacing.xs,
                  horizontal: AppSpacing.sm,
                ),
                border: OutlineInputBorder(
                  borderRadius:
                      BorderRadius.circular(AppSpacing.borderRadiusMd),
                  borderSide: BorderSide(color: colorScheme.outline),
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius:
                      BorderRadius.circular(AppSpacing.borderRadiusMd),
                  borderSide: BorderSide(color: colorScheme.outline),
                ),
              ),
              readOnly: true, // Placeholder -- real search handled elsewhere
              onTap: () {
                // TODO: Open command-palette / global search overlay
              },
            ),
          ),

          const SizedBox(width: AppSpacing.lg),

          // ── Connection Indicator ─────────────────────────────────────
          StatusIndicator.fromConnectionStatus(connectionState.status),

          const SizedBox(width: AppSpacing.md),

          // ── Node ID Chip ─────────────────────────────────────────────
          if (connectionState.nodeId != null)
            _InfoChip(
              label: 'Node',
              value: _shorten(connectionState.nodeId!),
              colorScheme: colorScheme,
            ),

          if (connectionState.nodeId != null)
            const SizedBox(width: AppSpacing.sm),

          // ── Site ID Chip ─────────────────────────────────────────────
          if (connectionState.siteId != null)
            _InfoChip(
              label: 'Site',
              value: _shorten(connectionState.siteId!),
              colorScheme: colorScheme,
            ),

          if (connectionState.siteId != null)
            const SizedBox(width: AppSpacing.sm),

          // ── Role Badge ───────────────────────────────────────────────
          // TODO: Pull role from auth state once AuthProvider exists.

          // ── Notification Bell ────────────────────────────────────────
          const SizedBox(width: AppSpacing.sm),
          _NotificationBell(count: alertCount),
        ],
      ),
    );
  }

  /// Shortens a UUID / long ID to its first 8 characters for display.
  static String _shorten(String id) {
    if (id.length <= 8) return id;
    return '${id.substring(0, 8)}...';
  }
}

// ─────────────────────────────────────────────────────────────────────────────
// Private sub-widgets
// ─────────────────────────────────────────────────────────────────────────────

class _InfoChip extends StatelessWidget {
  const _InfoChip({
    required this.label,
    required this.value,
    required this.colorScheme,
  });

  final String label;
  final String value;
  final ColorScheme colorScheme;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(
        horizontal: AppSpacing.sm,
        vertical: AppSpacing.xs,
      ),
      decoration: BoxDecoration(
        color: colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(AppSpacing.borderRadiusSm),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            '$label: ',
            style: TextStyle(
              fontSize: 11,
              fontWeight: FontWeight.w600,
              color: colorScheme.onSurfaceVariant,
            ),
          ),
          Text(
            value,
            style: TextStyle(
              fontSize: 11,
              fontFamily: 'monospace',
              color: colorScheme.onSurface,
            ),
          ),
        ],
      ),
    );
  }
}

class _NotificationBell extends StatelessWidget {
  const _NotificationBell({required this.count});
  final int count;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return IconButton(
      icon: Stack(
        clipBehavior: Clip.none,
        children: [
          Icon(Icons.notifications_outlined, color: colorScheme.onSurface),
          if (count > 0)
            Positioned(
              right: -6,
              top: -4,
              child: Container(
                padding: const EdgeInsets.all(2),
                decoration: BoxDecoration(
                  color: colorScheme.error,
                  borderRadius: BorderRadius.circular(10),
                ),
                constraints: const BoxConstraints(
                  minWidth: 16,
                  minHeight: 16,
                ),
                child: Text(
                  count > 99 ? '99+' : '$count',
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 10,
                    fontWeight: FontWeight.w600,
                  ),
                  textAlign: TextAlign.center,
                ),
              ),
            ),
        ],
      ),
      onPressed: () {
        // TODO: Navigate to alerts or show alerts panel
      },
      tooltip: count > 0 ? '$count unacknowledged alerts' : 'Alerts',
    );
  }
}
