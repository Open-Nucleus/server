import 'package:flutter/material.dart';

import '../../../core/theme/app_spacing.dart';
import '../../../shared/widgets/empty_state.dart';

/// Placeholder patient list screen.
class PatientListScreen extends StatelessWidget {
  const PatientListScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: AppSpacing.pagePadding,
      child: EmptyState(
        icon: Icons.people,
        title: 'No patients yet',
        subtitle: 'Create your first patient record to get started.',
        actionLabel: 'Create Patient',
        onAction: () {
          // TODO: Navigate to patient form
        },
      ),
    );
  }
}
