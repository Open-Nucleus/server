import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_spacing.dart';
import '../providers/shell_providers.dart';

/// A navigation item definition used by [SidebarNav].
class _NavItem {
  final String label;
  final IconData icon;
  final String route;

  const _NavItem({
    required this.label,
    required this.icon,
    required this.route,
  });
}

/// A section grouping within the sidebar.
class _NavSection {
  final String title;
  final List<_NavItem> items;

  const _NavSection({required this.title, required this.items});
}

/// The sidebar navigation panel.
///
/// Toggles between 240 px (expanded) and 72 px (collapsed) widths using an
/// [AnimatedContainer]. When collapsed, labels are hidden and tooltips are
/// shown on hover.
class SidebarNav extends ConsumerWidget {
  const SidebarNav({super.key});

  static const _sections = [
    _NavSection(
      title: 'Operations',
      items: [
        _NavItem(label: 'Dashboard', icon: Icons.dashboard, route: '/dashboard'),
        _NavItem(label: 'Patients', icon: Icons.people, route: '/patients'),
        _NavItem(label: 'Formulary', icon: Icons.medication, route: '/formulary'),
      ],
    ),
    _NavSection(
      title: 'System',
      items: [
        _NavItem(label: 'Sync', icon: Icons.sync, route: '/sync'),
        _NavItem(label: 'Alerts', icon: Icons.notifications, route: '/alerts'),
        _NavItem(label: 'Integrity', icon: Icons.verified, route: '/integrity'),
      ],
    ),
    _NavSection(
      title: 'Config',
      items: [
        _NavItem(label: 'Settings', icon: Icons.settings, route: '/settings'),
      ],
    ),
  ];

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final expanded = ref.watch(sidebarExpandedProvider);
    final currentPath = GoRouterState.of(context).uri.toString();
    final colorScheme = Theme.of(context).colorScheme;

    return AnimatedContainer(
      duration: const Duration(milliseconds: 200),
      curve: Curves.easeInOut,
      width: expanded
          ? AppSpacing.sidebarExpandedWidth
          : AppSpacing.sidebarCollapsedWidth,
      decoration: BoxDecoration(
        color: colorScheme.surface,
        border: Border(
          right: BorderSide(color: colorScheme.outlineVariant),
        ),
      ),
      child: Column(
        children: [
          // ── Logo / Brand ───────────────────────────────────────────────
          _LogoHeader(expanded: expanded),

          const SizedBox(height: AppSpacing.sm),

          // ── Nav Sections ───────────────────────────────────────────────
          Expanded(
            child: ListView(
              padding: const EdgeInsets.symmetric(horizontal: AppSpacing.sm),
              children: [
                for (final section in _sections) ...[
                  if (expanded)
                    _SectionHeader(title: section.title)
                  else
                    const Divider(height: AppSpacing.md),
                  for (final item in section.items)
                    _NavItemTile(
                      item: item,
                      expanded: expanded,
                      isActive: _isActive(currentPath, item.route),
                    ),
                  const SizedBox(height: AppSpacing.xs),
                ],
              ],
            ),
          ),

          // ── Collapse Toggle ────────────────────────────────────────────
          const Divider(height: 1),
          _CollapseButton(expanded: expanded),
        ],
      ),
    );
  }

  /// Determines if a nav item is active based on the current path.
  bool _isActive(String currentPath, String route) {
    if (route == '/dashboard') return currentPath == '/dashboard';
    return currentPath.startsWith(route);
  }
}

// ─────────────────────────────────────────────────────────────────────────────
// Private sub-widgets
// ─────────────────────────────────────────────────────────────────────────────

class _LogoHeader extends StatelessWidget {
  const _LogoHeader({required this.expanded});
  final bool expanded;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    return Container(
      height: AppSpacing.topBarHeight,
      padding: const EdgeInsets.symmetric(horizontal: AppSpacing.md),
      alignment: expanded ? Alignment.centerLeft : Alignment.center,
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.local_hospital, color: colorScheme.primary, size: 28),
          if (expanded) ...[
            const SizedBox(width: AppSpacing.sm),
            Text(
              'Open Nucleus',
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w700,
                color: colorScheme.onSurface,
              ),
            ),
          ],
        ],
      ),
    );
  }
}

class _SectionHeader extends StatelessWidget {
  const _SectionHeader({required this.title});
  final String title;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(
        left: AppSpacing.sm,
        top: AppSpacing.md,
        bottom: AppSpacing.xs,
      ),
      child: Text(
        title.toUpperCase(),
        style: TextStyle(
          fontSize: 11,
          fontWeight: FontWeight.w600,
          letterSpacing: 1.0,
          color: Theme.of(context).colorScheme.onSurfaceVariant,
        ),
      ),
    );
  }
}

class _NavItemTile extends ConsumerWidget {
  const _NavItemTile({
    required this.item,
    required this.expanded,
    required this.isActive,
  });

  final _NavItem item;
  final bool expanded;
  final bool isActive;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final colorScheme = Theme.of(context).colorScheme;

    final tile = Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(AppSpacing.borderRadiusMd),
      child: InkWell(
        borderRadius: BorderRadius.circular(AppSpacing.borderRadiusMd),
        onTap: () => context.go(item.route),
        hoverColor: colorScheme.primary.withOpacity(0.08),
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 150),
          padding: EdgeInsets.symmetric(
            horizontal: expanded ? AppSpacing.sm : 0,
            vertical: AppSpacing.sm,
          ),
          decoration: BoxDecoration(
            color: isActive
                ? colorScheme.primary.withOpacity(0.12)
                : Colors.transparent,
            borderRadius: BorderRadius.circular(AppSpacing.borderRadiusMd),
          ),
          child: Row(
            mainAxisAlignment:
                expanded ? MainAxisAlignment.start : MainAxisAlignment.center,
            children: [
              Icon(
                item.icon,
                size: 22,
                color: isActive
                    ? colorScheme.primary
                    : colorScheme.onSurfaceVariant,
              ),
              if (expanded) ...[
                const SizedBox(width: AppSpacing.sm),
                Expanded(
                  child: Text(
                    item.label,
                    style: TextStyle(
                      fontSize: 14,
                      fontWeight:
                          isActive ? FontWeight.w600 : FontWeight.w400,
                      color: isActive
                          ? colorScheme.primary
                          : colorScheme.onSurface,
                    ),
                  ),
                ),
              ],
            ],
          ),
        ),
      ),
    );

    if (expanded) return tile;

    // When collapsed, wrap with a tooltip.
    return Tooltip(
      message: item.label,
      preferBelow: false,
      child: tile,
    );
  }
}

class _CollapseButton extends ConsumerWidget {
  const _CollapseButton({required this.expanded});
  final bool expanded;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return InkWell(
      onTap: () {
        ref.read(sidebarExpandedProvider.notifier).state = !expanded;
      },
      child: Container(
        height: 48,
        alignment: Alignment.center,
        child: Icon(
          expanded ? Icons.chevron_left : Icons.chevron_right,
          color: Theme.of(context).colorScheme.onSurfaceVariant,
        ),
      ),
    );
  }
}
