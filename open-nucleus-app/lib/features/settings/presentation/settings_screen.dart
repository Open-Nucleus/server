import 'package:flutter/material.dart';

import '../../../core/theme/app_spacing.dart';
import '../../../shared/widgets/empty_state.dart';

/// Placeholder settings screen.
class SettingsScreen extends StatelessWidget {
  const SettingsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: AppSpacing.pagePadding,
      child: EmptyState(
        icon: Icons.settings,
        title: 'Settings',
        subtitle: 'Server configuration, theme, and device management will appear here.',
      ),
    );
  }
}
