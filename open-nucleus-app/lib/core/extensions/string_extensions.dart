/// Convenience helpers on [String] used throughout the Open Nucleus UI.
extension StringExtensions on String {
  /// Capitalise the first character, leaving the rest unchanged.
  ///
  /// ```dart
  /// 'hello'.capitalize // 'Hello'
  /// ```
  String get capitalize {
    if (isEmpty) return this;
    return '${this[0].toUpperCase()}${substring(1)}';
  }

  /// Convert to Title Case (capitalise the first letter of every word).
  ///
  /// ```dart
  /// 'open nucleus app'.titleCase // 'Open Nucleus App'
  /// ```
  String get titleCase {
    if (isEmpty) return this;
    return split(' ').map((word) => word.capitalize).join(' ');
  }

  /// Truncate to [maxLength] characters and append an ellipsis if needed.
  ///
  /// ```dart
  /// 'A very long string'.truncate(10) // 'A very lon...'
  /// ```
  String truncate(int maxLength, {String ellipsis = '...'}) {
    if (length <= maxLength) return this;
    return '${substring(0, maxLength)}$ellipsis';
  }

  /// Basic email validation using a simplified RFC 5322 pattern.
  bool get isValidEmail {
    if (isEmpty) return false;
    return RegExp(
      r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$',
    ).hasMatch(this);
  }

  /// Attempt to parse this string into a [DateTime].
  ///
  /// Returns `null` if parsing fails.
  DateTime? toDateTime() {
    return DateTime.tryParse(this);
  }
}
