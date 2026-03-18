import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import 'core/router/app_router.dart';
import 'core/theme/app_theme.dart';
import 'features/settings/presentation/settings_providers.dart';

class OpenNucleusApp extends ConsumerStatefulWidget {
  const OpenNucleusApp({super.key});

  @override
  ConsumerState<OpenNucleusApp> createState() => _OpenNucleusAppState();
}

class _OpenNucleusAppState extends ConsumerState<OpenNucleusApp> {
  late final GoRouter _router;

  @override
  void initState() {
    super.initState();
    // Build the router once with the WidgetRef so auth redirect works.
    _router = AppRouter.router(ref);
  }

  @override
  void dispose() {
    _router.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    // Watch themeMode from settings so the app re-renders on theme change.
    final themeMode = ref.watch(themeModePr);

    return MaterialApp.router(
      title: 'Open Nucleus',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.light,
      darkTheme: AppTheme.dark,
      themeMode: themeMode,
      routerConfig: _router,
    );
  }
}
