import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'connection_provider.dart';

// Re-export connection provider so consumers can import from one place.
export 'connection_provider.dart';

/// Whether the sidebar is expanded (true) or collapsed (false).
final sidebarExpandedProvider = StateProvider<bool>((ref) => true);

/// The current page title derived from the router location.
///
/// Updated by [AppScaffold] whenever the route changes.
final currentPageTitleProvider = StateProvider<String>((ref) => 'Dashboard');

/// The number of unacknowledged alerts shown on the notification bell.
///
/// This is a simple state provider that feature modules (e.g. alerts) can
/// update when they fetch new data from the backend.
final unacknowledgedAlertCountProvider = StateProvider<int>((ref) => 0);
