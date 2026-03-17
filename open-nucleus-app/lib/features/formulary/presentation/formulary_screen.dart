import 'package:flutter/material.dart';

import '../../../core/theme/app_spacing.dart';
import '../../../shared/widgets/empty_state.dart';

/// Placeholder formulary screen.
class FormularyScreen extends StatelessWidget {
  const FormularyScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: AppSpacing.pagePadding,
      child: EmptyState(
        icon: Icons.medication,
        title: 'Formulary',
        subtitle: 'Medication catalog and stock management will appear here.',
      ),
    );
  }
}
