import 'package:flutter/material.dart';

/// Centralised colour palette for Open Nucleus.
///
/// All colours are static constants so they can be used in const contexts
/// and referenced from both [AppTheme] and individual widgets.
class AppColors {
  AppColors._();

  // ── Brand ────────────────────────────────────────────────────────────

  static const Color primary = Color(0xFF1A237E); // navy
  static const Color secondary = Color(0xFF00897B); // teal

  // ── Semantic ─────────────────────────────────────────────────────────

  static const Color success = Color(0xFF2E7D32);
  static const Color warning = Color(0xFFF57F17);
  static const Color error = Color(0xFFD32F2F);
  static const Color critical = Color(0xFFD32F2F);
  static const Color info = Color(0xFF1565C0);

  // ── Surface / Background — Light ─────────────────────────────────────

  static const Color surfaceLight = Color(0xFFFAFAFA);
  static const Color surfaceVariantLight = Color(0xFFF5F5F5);
  static const Color backgroundLight = Color(0xFFFFFFFF);
  static const Color onSurfaceLight = Color(0xFF1C1B1F);
  static const Color onSurfaceVariantLight = Color(0xFF49454F);

  // ── Surface / Background — Dark ──────────────────────────────────────

  static const Color surfaceDark = Color(0xFF1C1B1F);
  static const Color surfaceVariantDark = Color(0xFF2B2930);
  static const Color backgroundDark = Color(0xFF121212);
  static const Color onSurfaceDark = Color(0xFFE6E1E5);
  static const Color onSurfaceVariantDark = Color(0xFFCAC4D0);

  // ── Severity ─────────────────────────────────────────────────────────

  static const Color severityCritical = Color(0xFFD32F2F); // red
  static const Color severityHigh = Color(0xFFE64A19); // deep orange
  static const Color severityWarning = Color(0xFFFFC107); // amber
  static const Color severityInfo = Color(0xFF1565C0); // blue
  static const Color severityLow = Color(0xFF2E7D32); // green

  // ── Status ───────────────────────────────────────────────────────────

  static const Color statusActive = Color(0xFF2E7D32); // green
  static const Color statusInactive = Color(0xFF757575); // grey
  static const Color statusPending = Color(0xFFFFC107); // amber
  static const Color statusError = Color(0xFFD32F2F); // red

  // ── Sync State ───────────────────────────────────────────────────────

  static const Color syncIdle = Color(0xFF757575); // grey
  static const Color syncSyncing = Color(0xFF1565C0); // blue
  static const Color syncError = Color(0xFFD32F2F); // red
  static const Color syncComplete = Color(0xFF2E7D32); // green
}
