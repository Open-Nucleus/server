import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/constants/api_paths.dart';
import '../../../shared/models/api_envelope.dart';
import '../../../shared/models/clinical_models.dart';
import '../../../shared/models/consent_models.dart';
import '../../../shared/models/patient_models.dart';
import '../../../shared/providers/dio_provider.dart';

// ---------------------------------------------------------------------------
// Patient detail (full bundle)
// ---------------------------------------------------------------------------

/// Fetches the full [PatientBundle] for a given patient ID.
///
/// Returns the patient resource plus all clinical sub-resources (encounters,
/// observations, conditions, medication requests, allergy intolerances, flags).
final patientDetailProvider =
    FutureProvider.family<PatientBundle, String>((ref, patientId) async {
  final dio = ref.watch(dioProvider);
  final response = await dio.get(ApiPaths.patient(patientId));
  final envelope = ApiEnvelope.fromJson(
    response.data as Map<String, dynamic>,
    (data) => PatientBundle.fromJson(data as Map<String, dynamic>),
  );
  if (envelope.isError || envelope.data == null) {
    throw Exception(envelope.error?.message ?? 'Failed to load patient');
  }
  return envelope.data!;
});

// ---------------------------------------------------------------------------
// Individual clinical resource providers
// ---------------------------------------------------------------------------

/// Fetches paginated encounters for a patient.
final patientEncountersProvider =
    FutureProvider.family<ClinicalListResponse, String>((ref, patientId) async {
  final dio = ref.watch(dioProvider);
  final response = await dio.get(ApiPaths.encounters(patientId));
  return _parseClinicalList(response);
});

/// Fetches paginated observations (vitals) for a patient.
final patientObservationsProvider =
    FutureProvider.family<ClinicalListResponse, String>((ref, patientId) async {
  final dio = ref.watch(dioProvider);
  final response = await dio.get(ApiPaths.observations(patientId));
  return _parseClinicalList(response);
});

/// Fetches paginated conditions for a patient.
final patientConditionsProvider =
    FutureProvider.family<ClinicalListResponse, String>((ref, patientId) async {
  final dio = ref.watch(dioProvider);
  final response = await dio.get(ApiPaths.conditions(patientId));
  return _parseClinicalList(response);
});

/// Fetches paginated medication requests for a patient.
final patientMedicationsProvider =
    FutureProvider.family<ClinicalListResponse, String>((ref, patientId) async {
  final dio = ref.watch(dioProvider);
  final response = await dio.get(ApiPaths.medicationRequests(patientId));
  return _parseClinicalList(response);
});

/// Fetches paginated allergy intolerances for a patient.
final patientAllergiesProvider =
    FutureProvider.family<ClinicalListResponse, String>((ref, patientId) async {
  final dio = ref.watch(dioProvider);
  final response = await dio.get(ApiPaths.allergyIntolerances(patientId));
  return _parseClinicalList(response);
});

/// Fetches paginated immunizations for a patient.
final patientImmunizationsProvider =
    FutureProvider.family<ClinicalListResponse, String>((ref, patientId) async {
  final dio = ref.watch(dioProvider);
  final response = await dio.get(ApiPaths.immunizations(patientId));
  return _parseClinicalList(response);
});

/// Fetches paginated procedures for a patient.
final patientProceduresProvider =
    FutureProvider.family<ClinicalListResponse, String>((ref, patientId) async {
  final dio = ref.watch(dioProvider);
  final response = await dio.get(ApiPaths.procedures(patientId));
  return _parseClinicalList(response);
});

/// Fetches paginated consents for a patient.
final patientConsentsProvider =
    FutureProvider.family<ConsentListResponse, String>((ref, patientId) async {
  final dio = ref.watch(dioProvider);
  final response = await dio.get(ApiPaths.patientConsents(patientId));
  final envelope = ApiEnvelope.fromJson(
    response.data as Map<String, dynamic>,
    (data) => ConsentListResponse.fromJson(data as Map<String, dynamic>),
  );
  if (envelope.isError || envelope.data == null) {
    throw Exception(envelope.error?.message ?? 'Failed to load consents');
  }
  return envelope.data!;
});

/// Fetches the Git history (audit trail) for a patient.
final patientHistoryProvider =
    FutureProvider.family<PatientHistoryResponse, String>(
        (ref, patientId) async {
  final dio = ref.watch(dioProvider);
  final response = await dio.get(ApiPaths.patientHistory(patientId));
  final envelope = ApiEnvelope.fromJson(
    response.data as Map<String, dynamic>,
    (data) => PatientHistoryResponse.fromJson(data as Map<String, dynamic>),
  );
  if (envelope.isError || envelope.data == null) {
    throw Exception(envelope.error?.message ?? 'Failed to load history');
  }
  return envelope.data!;
});

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

ClinicalListResponse _parseClinicalList(Response<dynamic> response) {
  final envelope = ApiEnvelope.fromJson(
    response.data as Map<String, dynamic>,
    (data) => ClinicalListResponse.fromJson(data as Map<String, dynamic>),
  );
  if (envelope.isError || envelope.data == null) {
    throw Exception(envelope.error?.message ?? 'Failed to load resources');
  }
  return envelope.data!;
}
