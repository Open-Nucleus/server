import 'package:flutter/material.dart';

import '../../../core/theme/app_spacing.dart';

/// Placeholder patient detail screen.
class PatientDetailScreen extends StatelessWidget {
  const PatientDetailScreen({required this.patientId, super.key});

  final String patientId;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Padding(
      padding: AppSpacing.pagePadding,
      child: Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.person, size: 64, color: colorScheme.primary),
            const SizedBox(height: AppSpacing.md),
            Text(
              'Patient Details',
              style: TextStyle(
                fontSize: 24,
                fontWeight: FontWeight.w700,
                color: colorScheme.onSurface,
              ),
            ),
            const SizedBox(height: AppSpacing.sm),
            Text(
              'Patient ID: $patientId',
              style: TextStyle(
                fontFamily: 'monospace',
                color: colorScheme.onSurfaceVariant,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
