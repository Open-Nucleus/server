/// Convenience date formatting helpers for the Open Nucleus UI.
///
/// These avoid pulling in a full `intl` dependency for simple cases.
extension DateExtensions on DateTime {
  static const _months = [
    'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun',
    'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec',
  ];

  /// Human-readable date, e.g. "Mar 15, 2026".
  String get toDisplayDate {
    return '${_months[month - 1]} $day, $year';
  }

  /// Human-readable date + time, e.g. "Mar 15, 2026 09:30".
  String get toDisplayDateTime {
    final h = hour.toString().padLeft(2, '0');
    final m = minute.toString().padLeft(2, '0');
    return '$toDisplayDate $h:$m';
  }

  /// ISO 8601 string (UTC).
  String get toIso8601 => toUtc().toIso8601String();

  /// FHIR-compliant date string, e.g. "2026-03-15".
  String get toFhirDate {
    final y = year.toString().padLeft(4, '0');
    final m = month.toString().padLeft(2, '0');
    final d = day.toString().padLeft(2, '0');
    return '$y-$m-$d';
  }

  /// Relative time description, e.g. "2 hours ago", "just now".
  String get timeAgo {
    final now = DateTime.now();
    final diff = now.difference(this);

    if (diff.isNegative) return 'in the future';
    if (diff.inSeconds < 60) return 'just now';
    if (diff.inMinutes < 60) {
      final mins = diff.inMinutes;
      return '$mins ${mins == 1 ? 'minute' : 'minutes'} ago';
    }
    if (diff.inHours < 24) {
      final hours = diff.inHours;
      return '$hours ${hours == 1 ? 'hour' : 'hours'} ago';
    }
    if (diff.inDays < 7) {
      final days = diff.inDays;
      return '$days ${days == 1 ? 'day' : 'days'} ago';
    }
    if (diff.inDays < 30) {
      final weeks = (diff.inDays / 7).floor();
      return '$weeks ${weeks == 1 ? 'week' : 'weeks'} ago';
    }
    if (diff.inDays < 365) {
      final months = (diff.inDays / 30).floor();
      return '$months ${months == 1 ? 'month' : 'months'} ago';
    }
    final years = (diff.inDays / 365).floor();
    return '$years ${years == 1 ? 'year' : 'years'} ago';
  }
}
