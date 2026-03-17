import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../providers/shell_providers.dart';
import 'sidebar_nav.dart';
import 'top_bar.dart';

/// Route-to-title mapping for the top bar.
const _routeTitles = <String, String>{
  '/dashboard': 'Dashboard',
  '/patients': 'Patients',
  '/patients/new': 'New Patient',
  '/formulary': 'Formulary',
  '/sync': 'Sync',
  '/alerts': 'Alerts',
  '/integrity': 'Integrity',
  '/settings': 'Settings',
};

/// The main application shell used as a [ShellRoute] builder.
///
/// Layout:
/// ```
/// ┌──────────┬──────────────────────────────────┐
/// │          │  TopBar                           │
/// │ Sidebar  ├──────────────────────────────────┤
/// │          │  Content (child from router)      │
/// │          │                                   │
/// └──────────┴──────────────────────────────────┘
/// ```
class AppScaffold extends ConsumerStatefulWidget {
  const AppScaffold({required this.child, super.key});

  /// The routed child widget rendered in the content area.
  final Widget child;

  @override
  ConsumerState<AppScaffold> createState() => _AppScaffoldState();
}

class _AppScaffoldState extends ConsumerState<AppScaffold> {
  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    _updatePageTitle();
  }

  void _updatePageTitle() {
    final location = GoRouterState.of(context).uri.toString();
    String title = 'Open Nucleus';

    // Try exact match first, then prefix match.
    if (_routeTitles.containsKey(location)) {
      title = _routeTitles[location]!;
    } else {
      // Handle parameterised routes like /patients/:id
      for (final entry in _routeTitles.entries) {
        if (location.startsWith(entry.key) && entry.key != '/') {
          title = entry.value;
          break;
        }
      }
      // Special case: /patients/<uuid> → "Patient Details"
      final patientDetailPattern = RegExp(r'^/patients/[a-zA-Z0-9\-]+$');
      if (patientDetailPattern.hasMatch(location) &&
          location != '/patients/new') {
        title = 'Patient Details';
      }
    }

    // Schedule the provider update for after the current build.
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) {
        ref.read(currentPageTitleProvider.notifier).state = title;
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Row(
        children: [
          // ── Sidebar ────────────────────────────────────────────────
          const SidebarNav(),

          // ── Main Content Area ──────────────────────────────────────
          Expanded(
            child: Column(
              children: [
                const TopBar(),
                Expanded(child: widget.child),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
