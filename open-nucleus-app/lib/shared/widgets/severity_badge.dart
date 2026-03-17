import 'package:flutter/material.dart';

import '../../core/theme/app_colors.dart';

/// A small coloured chip that shows a severity level.
///
/// Supported severities: critical, high, warning, info, low.
class SeverityBadge extends StatelessWidget {
  const SeverityBadge({
    required this.severity,
    super.key,
  });

  /// The severity string (case-insensitive). Recognised values:
  /// `critical`, `high`, `warning`, `info`, `low`.
  final String severity;

  @override
  Widget build(BuildContext context) {
    final (Color bg, Color fg, String label) = _resolve(severity);

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: bg.withOpacity(0.15),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: bg.withOpacity(0.4)),
      ),
      child: Text(
        label,
        style: TextStyle(
          fontSize: 11,
          fontWeight: FontWeight.w600,
          color: fg,
        ),
      ),
    );
  }

  static (Color, Color, String) _resolve(String severity) {
    switch (severity.toLowerCase()) {
      case 'critical':
        return (AppColors.severityCritical, AppColors.severityCritical, 'Critical');
      case 'high':
        return (AppColors.severityHigh, AppColors.severityHigh, 'High');
      case 'warning':
        return (AppColors.severityWarning, const Color(0xFFF57F17), 'Warning');
      case 'info':
        return (AppColors.severityInfo, AppColors.severityInfo, 'Info');
      case 'low':
        return (AppColors.severityLow, AppColors.severityLow, 'Low');
      default:
        return (AppColors.statusInactive, AppColors.statusInactive, severity);
    }
  }
}
