import 'package:flutter/material.dart';

import '../../../core/theme/app_spacing.dart';
import '../../../shared/widgets/empty_state.dart';

/// Placeholder sync screen.
class SyncScreen extends StatelessWidget {
  const SyncScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: AppSpacing.pagePadding,
      child: EmptyState(
        icon: Icons.sync,
        title: 'Sync',
        subtitle: 'Peer discovery, sync status, and conflict resolution will appear here.',
      ),
    );
  }
}
