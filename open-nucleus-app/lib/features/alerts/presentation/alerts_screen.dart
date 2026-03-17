import 'package:flutter/material.dart';

import '../../../core/theme/app_spacing.dart';
import '../../../shared/widgets/empty_state.dart';

/// Placeholder alerts screen.
class AlertsScreen extends StatelessWidget {
  const AlertsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: AppSpacing.pagePadding,
      child: EmptyState(
        icon: Icons.notifications,
        title: 'Alerts',
        subtitle: 'Sentinel agent alerts and notifications will appear here.',
      ),
    );
  }
}
