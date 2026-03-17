import 'package:flutter/material.dart';

import '../../core/theme/app_colors.dart';
import '../../features/shell/providers/connection_provider.dart';

/// A small coloured dot with an optional text label indicating status.
///
/// Common statuses: connected (green), disconnected (red),
/// checking (amber), unknown (grey).
class StatusIndicator extends StatelessWidget {
  const StatusIndicator({
    required this.color,
    this.label,
    this.size = 10,
    super.key,
  });

  /// Creates a [StatusIndicator] from a [ConnectionStatus] enum value.
  factory StatusIndicator.fromConnectionStatus(ConnectionStatus status) {
    switch (status) {
      case ConnectionStatus.connected:
        return const StatusIndicator(
          color: AppColors.statusActive,
          label: 'Connected',
        );
      case ConnectionStatus.disconnected:
        return const StatusIndicator(
          color: AppColors.statusError,
          label: 'Disconnected',
        );
      case ConnectionStatus.checking:
        return const StatusIndicator(
          color: AppColors.statusPending,
          label: 'Checking',
        );
      case ConnectionStatus.unknown:
        return const StatusIndicator(
          color: AppColors.statusInactive,
          label: 'Unknown',
        );
    }
  }

  final Color color;
  final String? label;
  final double size;

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          width: size,
          height: size,
          decoration: BoxDecoration(
            color: color,
            shape: BoxShape.circle,
            boxShadow: [
              BoxShadow(
                color: color.withOpacity(0.4),
                blurRadius: 4,
                spreadRadius: 1,
              ),
            ],
          ),
        ),
        if (label != null) ...[
          const SizedBox(width: 6),
          Text(
            label!,
            style: TextStyle(
              fontSize: 12,
              color: Theme.of(context).colorScheme.onSurfaceVariant,
            ),
          ),
        ],
      ],
    );
  }
}
