import 'package:flutter/material.dart';

/// Convenience getters on [BuildContext] to reduce boilerplate when
/// accessing theme data, screen metrics, and showing snack bars.
extension ContextExtensions on BuildContext {
  // ── Theme Shortcuts ──────────────────────────────────────────────────

  /// Current [ThemeData].
  ThemeData get theme => Theme.of(this);

  /// Current [ColorScheme].
  ColorScheme get colorScheme => Theme.of(this).colorScheme;

  /// Current [TextTheme].
  TextTheme get textTheme => Theme.of(this).textTheme;

  // ── Screen Metrics ───────────────────────────────────────────────────

  /// Logical screen width in pixels.
  double get screenWidth => MediaQuery.sizeOf(this).width;

  /// Logical screen height in pixels.
  double get screenHeight => MediaQuery.sizeOf(this).height;

  /// `true` when the viewport is at least 1024 logical pixels wide,
  /// a reasonable breakpoint for desktop layouts.
  bool get isDesktop => screenWidth >= 1024;

  /// `true` when the viewport is at least 1440 logical pixels wide,
  /// allowing for wider side panels or split views.
  bool get isWideDesktop => screenWidth >= 1440;

  // ── Snack Bars ───────────────────────────────────────────────────────

  /// Show a plain [SnackBar] with [message].
  void showSnackBar(String message, {Duration? duration}) {
    ScaffoldMessenger.of(this)
      ..hideCurrentSnackBar()
      ..showSnackBar(
        SnackBar(
          content: Text(message),
          duration: duration ?? const Duration(seconds: 3),
          behavior: SnackBarBehavior.floating,
        ),
      );
  }

  /// Show an error [SnackBar] styled with the error colour.
  void showErrorSnackBar(String message, {Duration? duration}) {
    ScaffoldMessenger.of(this)
      ..hideCurrentSnackBar()
      ..showSnackBar(
        SnackBar(
          content: Text(message),
          backgroundColor: colorScheme.error,
          duration: duration ?? const Duration(seconds: 4),
          behavior: SnackBarBehavior.floating,
        ),
      );
  }

  /// Show a success [SnackBar] styled with green.
  void showSuccessSnackBar(String message, {Duration? duration}) {
    ScaffoldMessenger.of(this)
      ..hideCurrentSnackBar()
      ..showSnackBar(
        SnackBar(
          content: Text(message),
          backgroundColor: const Color(0xFF2E7D32),
          duration: duration ?? const Duration(seconds: 3),
          behavior: SnackBarBehavior.floating,
        ),
      );
  }
}
