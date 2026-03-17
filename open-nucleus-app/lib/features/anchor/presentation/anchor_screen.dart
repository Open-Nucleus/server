import 'package:flutter/material.dart';

import '../../../core/theme/app_spacing.dart';
import '../../../shared/widgets/empty_state.dart';

/// Placeholder integrity/anchor screen.
class AnchorScreen extends StatelessWidget {
  const AnchorScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: AppSpacing.pagePadding,
      child: EmptyState(
        icon: Icons.verified,
        title: 'Integrity',
        subtitle: 'IOTA Tangle anchoring and data integrity verification will appear here.',
      ),
    );
  }
}
