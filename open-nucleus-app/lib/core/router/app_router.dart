import 'package:go_router/go_router.dart';

import '../../features/auth/presentation/login_screen.dart';

/// Central router configuration for Open Nucleus.
///
/// Routes will be added as feature screens are implemented.
abstract final class AppRouter {
  static final GoRouter router = GoRouter(
    initialLocation: '/login',
    routes: [
      GoRoute(
        path: '/login',
        builder: (context, state) => const LoginScreen(),
      ),
    ],
  );
}
