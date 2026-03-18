import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../shared/models/consent_models.dart';
import '../../../shared/providers/dio_provider.dart';
import '../data/consent_api.dart';

// ---------------------------------------------------------------------------
// Data layer provider
// ---------------------------------------------------------------------------

/// Provides the [ConsentApi] HTTP client.
final consentApiProvider = Provider<ConsentApi>((ref) {
  final dio = ref.watch(dioProvider);
  return ConsentApi(dio);
});

// ---------------------------------------------------------------------------
// Patient consents (by patient ID)
// ---------------------------------------------------------------------------

/// Fetches the consent list for a specific patient.
final patientConsentsProvider =
    FutureProvider.autoDispose.family<ConsentListResponse, String>(
  (ref, patientId) async {
    final api = ref.watch(consentApiProvider);
    final envelope = await api.listConsents(patientId);
    return envelope.data!;
  },
);
