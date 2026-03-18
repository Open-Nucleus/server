/// Common FHIR code systems and values used throughout the app.
///
/// These mirror the LOINC, encounter-status, condition-clinical, and
/// allergy-intolerance value sets that the Open Nucleus backend validates.
class FhirCodes {
  FhirCodes._();

  // ---------------------------------------------------------------------------
  // LOINC vital-sign codes
  // ---------------------------------------------------------------------------

  static const String loincSystem = 'http://loinc.org';

  static const String heartRate = '8867-4';
  static const String bloodPressureSystolic = '8480-6';
  static const String bloodPressureDiastolic = '8462-4';
  static const String bodyTemperature = '8310-5';
  static const String respiratoryRate = '9279-1';
  static const String oxygenSaturation = '2708-6';
  static const String bodyWeight = '29463-7';
  static const String bodyHeight = '8302-2';
  static const String bmi = '39156-5';

  /// Human-readable display strings keyed by LOINC code.
  static const Map<String, String> vitalDisplayNames = {
    heartRate: 'Heart Rate',
    bloodPressureSystolic: 'Systolic Blood Pressure',
    bloodPressureDiastolic: 'Diastolic Blood Pressure',
    bodyTemperature: 'Body Temperature',
    respiratoryRate: 'Respiratory Rate',
    oxygenSaturation: 'SpO2',
    bodyWeight: 'Weight',
    bodyHeight: 'Height',
    bmi: 'BMI',
  };

  /// Standard units for each vital sign.
  static const Map<String, String> vitalUnits = {
    heartRate: 'bpm',
    bloodPressureSystolic: 'mmHg',
    bloodPressureDiastolic: 'mmHg',
    bodyTemperature: 'Cel',
    respiratoryRate: '/min',
    oxygenSaturation: '%',
    bodyWeight: 'kg',
    bodyHeight: 'cm',
    bmi: 'kg/m2',
  };

  // ---------------------------------------------------------------------------
  // Encounter status (http://hl7.org/fhir/encounter-status)
  // ---------------------------------------------------------------------------

  static const String encounterStatusSystem =
      'http://hl7.org/fhir/encounter-status';

  static const String encounterPlanned = 'planned';
  static const String encounterArrived = 'arrived';
  static const String encounterTriaged = 'triaged';
  static const String encounterInProgress = 'in-progress';
  static const String encounterOnLeave = 'onleave';
  static const String encounterFinished = 'finished';
  static const String encounterCancelled = 'cancelled';

  static const List<String> encounterStatuses = [
    encounterPlanned,
    encounterArrived,
    encounterTriaged,
    encounterInProgress,
    encounterOnLeave,
    encounterFinished,
    encounterCancelled,
  ];

  // ---------------------------------------------------------------------------
  // Condition clinical status
  // (http://terminology.hl7.org/CodeSystem/condition-clinical)
  // ---------------------------------------------------------------------------

  static const String conditionClinicalSystem =
      'http://terminology.hl7.org/CodeSystem/condition-clinical';

  static const String conditionActive = 'active';
  static const String conditionRecurrence = 'recurrence';
  static const String conditionRelapse = 'relapse';
  static const String conditionInactive = 'inactive';
  static const String conditionRemission = 'remission';
  static const String conditionResolved = 'resolved';

  static const List<String> conditionClinicalStatuses = [
    conditionActive,
    conditionRecurrence,
    conditionRelapse,
    conditionInactive,
    conditionRemission,
    conditionResolved,
  ];

  // ---------------------------------------------------------------------------
  // AllergyIntolerance criticality
  // (http://hl7.org/fhir/allergy-intolerance-criticality)
  // ---------------------------------------------------------------------------

  static const String allergyCriticalitySystem =
      'http://hl7.org/fhir/allergy-intolerance-criticality';

  static const String criticalityLow = 'low';
  static const String criticalityHigh = 'high';
  static const String criticalityUnableToAssess = 'unable-to-assess';

  static const List<String> allergyCriticalities = [
    criticalityLow,
    criticalityHigh,
    criticalityUnableToAssess,
  ];
}
