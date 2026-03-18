import 'package:flutter/material.dart';

/// Spacing and layout constants for Open Nucleus.
///
/// All values are `double` so they can be used directly with
/// [EdgeInsets], [SizedBox], and layout widgets.
class AppSpacing {
  AppSpacing._();

  // ── Base Scale ───────────────────────────────────────────────────────

  static const double xs = 4;
  static const double sm = 8;
  static const double md = 16;
  static const double lg = 24;
  static const double xl = 32;
  static const double xxl = 48;

  // ── Layout Dimensions ────────────────────────────────────────────────

  /// Width of the navigation sidebar when fully expanded.
  static const double sidebarExpandedWidth = 240;

  /// Width of the navigation sidebar when collapsed to icons only.
  static const double sidebarCollapsedWidth = 72;

  /// Height of the top application bar.
  static const double topBarHeight = 56;

  // ── Common Paddings ──────────────────────────────────────────────────

  /// Padding inside cards.
  static const EdgeInsets cardPadding = EdgeInsets.all(md);

  /// Padding for list-item rows.
  static const EdgeInsets listItemPadding = EdgeInsets.symmetric(
    horizontal: md,
    vertical: sm,
  );

  /// Page-level padding (content area).
  static const EdgeInsets pagePadding = EdgeInsets.all(lg);

  // ── Border Radii ─────────────────────────────────────────────────────

  static const double borderRadiusSm = 4;
  static const double borderRadiusMd = 8;
  static const double borderRadiusLg = 12;
  static const double borderRadiusXl = 16;
}
