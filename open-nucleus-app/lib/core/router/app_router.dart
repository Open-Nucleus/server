import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../features/alerts/presentation/alerts_screen.dart';
import '../../features/anchor/presentation/anchor_screen.dart';
import '../../features/auth/presentation/auth_notifier.dart';
import '../../features/auth/presentation/auth_providers.dart';
import '../../features/auth/presentation/login_screen.dart';
import '../../features/dashboard/presentation/dashboard_screen.dart';
import '../../features/formulary/presentation/formulary_screen.dart';
import '../../features/patients/presentation/patient_detail_screen.dart';
import '../../features/patients/presentation/patient_form_screen.dart';
import '../../features/patients/presentation/patient_list_screen.dart';
import '../../features/settings/presentation/settings_screen.dart';
import '../../features/shell/presentation/app_scaffold.dart';
import '../../features/sync/presentation/sync_screen.dart';

/// Central router configuration for Open Nucleus.
///
/// Uses [go_router] with a [ShellRoute] wrapping the [AppScaffold] for all
/// authenticated pages. The `/login` route is outside the shell.
abstract final class AppRouter {
  /// Creates a [GoRouter] that observes [authNotifierProvider] for redirects.
  static GoRouter router(WidgetRef ref) {
    return GoRouter(
      initialLocation: '/dashboard',
      refreshListenable: _AuthRefreshListenable(ref),
      redirect: (context, state) {
        final authState = ref.read(authNotifierProvider);
        final isAuthenticated = authState is Authenticated;
        final isOnLogin = state.uri.toString() == '/login';

        // Not authenticated and not already on login -> redirect to login.
        if (!isAuthenticated && !isOnLogin) return '/login';

        // Authenticated but still on login -> redirect to dashboard.
        if (isAuthenticated && isOnLogin) return '/dashboard';

        return null; // No redirect needed.
      },
      routes: [
        // ── Login (no shell) ─────────────────────────────────────────
        GoRoute(
          path: '/login',
          builder: (context, state) => const LoginScreen(),
        ),

        // ── Shell (authenticated) ────────────────────────────────────
        ShellRoute(
          builder: (context, state, child) => AppScaffold(child: child),
          routes: [
            GoRoute(
              path: '/dashboard',
              pageBuilder: (context, state) => const NoTransitionPage(
                child: DashboardScreen(),
              ),
            ),
            GoRoute(
              path: '/patients',
              pageBuilder: (context, state) => const NoTransitionPage(
                child: PatientListScreen(),
              ),
              routes: [
                GoRoute(
                  path: 'new',
                  pageBuilder: (context, state) => const NoTransitionPage(
                    child: PatientFormScreen(),
                  ),
                ),
                GoRoute(
                  path: ':id',
                  pageBuilder: (context, state) {
                    final id = state.pathParameters['id']!;
                    return NoTransitionPage(
                      child: PatientDetailScreen(patientId: id),
                    );
                  },
                ),
              ],
            ),
            GoRoute(
              path: '/formulary',
              pageBuilder: (context, state) => const NoTransitionPage(
                child: FormularyScreen(),
              ),
            ),
            GoRoute(
              path: '/sync',
              pageBuilder: (context, state) => const NoTransitionPage(
                child: SyncScreen(),
              ),
            ),
            GoRoute(
              path: '/alerts',
              pageBuilder: (context, state) => const NoTransitionPage(
                child: AlertsScreen(),
              ),
            ),
            GoRoute(
              path: '/integrity',
              pageBuilder: (context, state) => const NoTransitionPage(
                child: AnchorScreen(),
              ),
            ),
            GoRoute(
              path: '/settings',
              pageBuilder: (context, state) => const NoTransitionPage(
                child: SettingsScreen(),
              ),
            ),
          ],
        ),
      ],
    );
  }
}

// ─────────────────────────────────────────────────────────────────────────────
// Auth-aware refresh listenable for GoRouter redirects
// ─────────────────────────────────────────────────────────────────────────────

/// A [ChangeNotifier] that listens to [authNotifierProvider] and notifies
/// [GoRouter] when the auth state changes so redirects are re-evaluated.
class _AuthRefreshListenable extends ChangeNotifier {
  _AuthRefreshListenable(this._ref) {
    _ref.listen(authNotifierProvider, (_, __) {
      notifyListeners();
    });
  }

  final WidgetRef _ref;
}
