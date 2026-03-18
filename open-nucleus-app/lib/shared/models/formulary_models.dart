/// Formulary DTOs matching Go `service.MedicationListResponse`,
/// `service.MedicationDetail`, interaction checks, stock management, etc.

/// Paginated medication list from `GET /api/v1/formulary/medications`.
class MedicationListResponse {
  final List<MedicationDetail> medications;
  final int page;
  final int perPage;
  final int total;
  final int totalPages;

  const MedicationListResponse({
    required this.medications,
    required this.page,
    required this.perPage,
    required this.total,
    required this.totalPages,
  });

  factory MedicationListResponse.fromJson(Map<String, dynamic> json) {
    return MedicationListResponse(
      medications: (json['medications'] as List<dynamic>)
          .map((m) => MedicationDetail.fromJson(m as Map<String, dynamic>))
          .toList(),
      page: (json['page'] as num).toInt(),
      perPage: (json['per_page'] as num).toInt(),
      total: (json['total'] as num).toInt(),
      totalPages: (json['total_pages'] as num).toInt(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'medications': medications.map((m) => m.toJson()).toList(),
      'page': page,
      'per_page': perPage,
      'total': total,
      'total_pages': totalPages,
    };
  }
}

/// A single medication entry.
///
/// Matches Go `service.MedicationDetail`.
class MedicationDetail {
  final String code;
  final String display;
  final String form;
  final String route;
  final String category;
  final bool available;
  final bool whoEssential;
  final String therapeuticClass;
  final List<String>? commonFrequencies;
  final String? strength;
  final String? unit;

  const MedicationDetail({
    required this.code,
    required this.display,
    required this.form,
    required this.route,
    required this.category,
    required this.available,
    required this.whoEssential,
    required this.therapeuticClass,
    this.commonFrequencies,
    this.strength,
    this.unit,
  });

  factory MedicationDetail.fromJson(Map<String, dynamic> json) {
    return MedicationDetail(
      code: json['code'] as String,
      display: json['display'] as String,
      form: json['form'] as String,
      route: json['route'] as String,
      category: json['category'] as String,
      available: json['available'] as bool,
      whoEssential: json['who_essential'] as bool,
      therapeuticClass: json['therapeutic_class'] as String,
      commonFrequencies: (json['common_frequencies'] as List<dynamic>?)
          ?.map((f) => f as String)
          .toList(),
      strength: json['strength'] as String?,
      unit: json['unit'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'code': code,
      'display': display,
      'form': form,
      'route': route,
      'category': category,
      'available': available,
      'who_essential': whoEssential,
      'therapeutic_class': therapeuticClass,
      if (commonFrequencies != null) 'common_frequencies': commonFrequencies,
      if (strength != null) 'strength': strength,
      if (unit != null) 'unit': unit,
    };
  }
}

// ---------------------------------------------------------------------------
// Interaction / Safety checks
// ---------------------------------------------------------------------------

/// Body of `POST /api/v1/formulary/check-interactions`.
class CheckInteractionsRequest {
  final List<String> medicationCodes;
  final String patientId;
  final List<String>? allergyCodes;
  final String? siteId;

  const CheckInteractionsRequest({
    required this.medicationCodes,
    required this.patientId,
    this.allergyCodes,
    this.siteId,
  });

  factory CheckInteractionsRequest.fromJson(Map<String, dynamic> json) {
    return CheckInteractionsRequest(
      medicationCodes: (json['medication_codes'] as List<dynamic>)
          .map((c) => c as String)
          .toList(),
      patientId: json['patient_id'] as String,
      allergyCodes: (json['allergy_codes'] as List<dynamic>?)
          ?.map((c) => c as String)
          .toList(),
      siteId: json['site_id'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'medication_codes': medicationCodes,
      'patient_id': patientId,
      if (allergyCodes != null) 'allergy_codes': allergyCodes,
      if (siteId != null) 'site_id': siteId,
    };
  }
}

/// Response from `POST /api/v1/formulary/check-interactions`.
class CheckInteractionsResponse {
  final List<InteractionDetail> interactions;
  final List<AllergyAlertDTO>? allergyAlerts;
  final List<DosingWarningDTO>? dosingWarnings;
  final StockSummaryDTO? stockSummary;
  final String overallRisk;

  const CheckInteractionsResponse({
    required this.interactions,
    this.allergyAlerts,
    this.dosingWarnings,
    this.stockSummary,
    required this.overallRisk,
  });

  factory CheckInteractionsResponse.fromJson(Map<String, dynamic> json) {
    return CheckInteractionsResponse(
      interactions: (json['interactions'] as List<dynamic>)
          .map((i) => InteractionDetail.fromJson(i as Map<String, dynamic>))
          .toList(),
      allergyAlerts: (json['allergy_alerts'] as List<dynamic>?)
          ?.map((a) => AllergyAlertDTO.fromJson(a as Map<String, dynamic>))
          .toList(),
      dosingWarnings: (json['dosing_warnings'] as List<dynamic>?)
          ?.map((d) => DosingWarningDTO.fromJson(d as Map<String, dynamic>))
          .toList(),
      stockSummary: json['stock_summary'] != null
          ? StockSummaryDTO.fromJson(
              json['stock_summary'] as Map<String, dynamic>)
          : null,
      overallRisk: json['overall_risk'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'interactions': interactions.map((i) => i.toJson()).toList(),
      if (allergyAlerts != null)
        'allergy_alerts': allergyAlerts!.map((a) => a.toJson()).toList(),
      if (dosingWarnings != null)
        'dosing_warnings': dosingWarnings!.map((d) => d.toJson()).toList(),
      if (stockSummary != null) 'stock_summary': stockSummary!.toJson(),
      'overall_risk': overallRisk,
    };
  }
}

/// A drug-drug interaction detail.
class InteractionDetail {
  final String severity;
  final String type;
  final String description;
  final String medicationA;
  final String medicationB;
  final String source;
  final String? clinicalEffect;
  final String? recommendation;

  const InteractionDetail({
    required this.severity,
    required this.type,
    required this.description,
    required this.medicationA,
    required this.medicationB,
    required this.source,
    this.clinicalEffect,
    this.recommendation,
  });

  factory InteractionDetail.fromJson(Map<String, dynamic> json) {
    return InteractionDetail(
      severity: json['severity'] as String,
      type: json['type'] as String,
      description: json['description'] as String,
      medicationA: json['medication_a'] as String,
      medicationB: json['medication_b'] as String,
      source: json['source'] as String,
      clinicalEffect: json['clinical_effect'] as String?,
      recommendation: json['recommendation'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'severity': severity,
      'type': type,
      'description': description,
      'medication_a': medicationA,
      'medication_b': medicationB,
      'source': source,
      if (clinicalEffect != null) 'clinical_effect': clinicalEffect,
      if (recommendation != null) 'recommendation': recommendation,
    };
  }
}

/// An allergy conflict alert.
class AllergyAlertDTO {
  final String severity;
  final String allergyCode;
  final String medicationCode;
  final String description;
  final String? crossReactivityClass;

  const AllergyAlertDTO({
    required this.severity,
    required this.allergyCode,
    required this.medicationCode,
    required this.description,
    this.crossReactivityClass,
  });

  factory AllergyAlertDTO.fromJson(Map<String, dynamic> json) {
    return AllergyAlertDTO(
      severity: json['severity'] as String,
      allergyCode: json['allergy_code'] as String,
      medicationCode: json['medication_code'] as String,
      description: json['description'] as String,
      crossReactivityClass: json['cross_reactivity_class'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'severity': severity,
      'allergy_code': allergyCode,
      'medication_code': medicationCode,
      'description': description,
      if (crossReactivityClass != null)
        'cross_reactivity_class': crossReactivityClass,
    };
  }
}

/// A dosing-related warning.
class DosingWarningDTO {
  final String medicationCode;
  final String warning;
  final String severity;

  const DosingWarningDTO({
    required this.medicationCode,
    required this.warning,
    required this.severity,
  });

  factory DosingWarningDTO.fromJson(Map<String, dynamic> json) {
    return DosingWarningDTO(
      medicationCode: json['medication_code'] as String,
      warning: json['warning'] as String,
      severity: json['severity'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'medication_code': medicationCode,
      'warning': warning,
      'severity': severity,
    };
  }
}

/// Stock availability summary for a set of medications.
class StockSummaryDTO {
  final List<StockItemDTO> items;

  const StockSummaryDTO({required this.items});

  factory StockSummaryDTO.fromJson(Map<String, dynamic> json) {
    return StockSummaryDTO(
      items: (json['items'] as List<dynamic>)
          .map((i) => StockItemDTO.fromJson(i as Map<String, dynamic>))
          .toList(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'items': items.map((i) => i.toJson()).toList(),
    };
  }
}

/// Stock level for a single medication at a site.
class StockItemDTO {
  final String medicationCode;
  final bool available;
  final int quantity;
  final String unit;

  const StockItemDTO({
    required this.medicationCode,
    required this.available,
    required this.quantity,
    required this.unit,
  });

  factory StockItemDTO.fromJson(Map<String, dynamic> json) {
    return StockItemDTO(
      medicationCode: json['medication_code'] as String,
      available: json['available'] as bool,
      quantity: (json['quantity'] as num).toInt(),
      unit: json['unit'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'medication_code': medicationCode,
      'available': available,
      'quantity': quantity,
      'unit': unit,
    };
  }
}

// ---------------------------------------------------------------------------
// Stock management
// ---------------------------------------------------------------------------

/// Response from `GET /api/v1/formulary/stock/{site_id}/{medication_code}`.
class StockLevelResponse {
  final String siteId;
  final String medicationCode;
  final int quantity;
  final String unit;
  final String lastUpdated;
  final String? earliestExpiry;
  final double dailyConsumptionRate;

  const StockLevelResponse({
    required this.siteId,
    required this.medicationCode,
    required this.quantity,
    required this.unit,
    required this.lastUpdated,
    this.earliestExpiry,
    required this.dailyConsumptionRate,
  });

  factory StockLevelResponse.fromJson(Map<String, dynamic> json) {
    return StockLevelResponse(
      siteId: json['site_id'] as String,
      medicationCode: json['medication_code'] as String,
      quantity: (json['quantity'] as num).toInt(),
      unit: json['unit'] as String,
      lastUpdated: json['last_updated'] as String,
      earliestExpiry: json['earliest_expiry'] as String?,
      dailyConsumptionRate:
          (json['daily_consumption_rate'] as num).toDouble(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'site_id': siteId,
      'medication_code': medicationCode,
      'quantity': quantity,
      'unit': unit,
      'last_updated': lastUpdated,
      if (earliestExpiry != null) 'earliest_expiry': earliestExpiry,
      'daily_consumption_rate': dailyConsumptionRate,
    };
  }
}

/// Response from stock prediction endpoint.
class StockPredictionResponse {
  final int daysRemaining;
  final String riskLevel;
  final String? earliestExpiry;
  final int expiringQuantity;
  final String recommendedAction;

  const StockPredictionResponse({
    required this.daysRemaining,
    required this.riskLevel,
    this.earliestExpiry,
    required this.expiringQuantity,
    required this.recommendedAction,
  });

  factory StockPredictionResponse.fromJson(Map<String, dynamic> json) {
    return StockPredictionResponse(
      daysRemaining: (json['days_remaining'] as num).toInt(),
      riskLevel: json['risk_level'] as String,
      earliestExpiry: json['earliest_expiry'] as String?,
      expiringQuantity: (json['expiring_quantity'] as num).toInt(),
      recommendedAction: json['recommended_action'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'days_remaining': daysRemaining,
      'risk_level': riskLevel,
      if (earliestExpiry != null) 'earliest_expiry': earliestExpiry,
      'expiring_quantity': expiringQuantity,
      'recommended_action': recommendedAction,
    };
  }
}

/// Formulary metadata from `GET /api/v1/formulary/info`.
class FormularyInfoResponse {
  final String version;
  final int totalMedications;
  final int totalInteractions;
  final String lastUpdated;
  final List<String> categories;
  final bool dosingEngineAvailable;

  const FormularyInfoResponse({
    required this.version,
    required this.totalMedications,
    required this.totalInteractions,
    required this.lastUpdated,
    required this.categories,
    required this.dosingEngineAvailable,
  });

  factory FormularyInfoResponse.fromJson(Map<String, dynamic> json) {
    return FormularyInfoResponse(
      version: json['version'] as String,
      totalMedications: (json['total_medications'] as num).toInt(),
      totalInteractions: (json['total_interactions'] as num).toInt(),
      lastUpdated: json['last_updated'] as String,
      categories: (json['categories'] as List<dynamic>)
          .map((c) => c as String)
          .toList(),
      dosingEngineAvailable: json['dosing_engine_available'] as bool,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'version': version,
      'total_medications': totalMedications,
      'total_interactions': totalInteractions,
      'last_updated': lastUpdated,
      'categories': categories,
      'dosing_engine_available': dosingEngineAvailable,
    };
  }
}

/// Redistribution suggestions for a medication across sites.
class FormularyRedistributionResponse {
  final List<FormularyRedistributionSuggestion> suggestions;

  const FormularyRedistributionResponse({required this.suggestions});

  factory FormularyRedistributionResponse.fromJson(Map<String, dynamic> json) {
    return FormularyRedistributionResponse(
      suggestions: (json['suggestions'] as List<dynamic>)
          .map((s) => FormularyRedistributionSuggestion.fromJson(
              s as Map<String, dynamic>))
          .toList(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'suggestions': suggestions.map((s) => s.toJson()).toList(),
    };
  }
}

/// A single redistribution suggestion between two sites.
class FormularyRedistributionSuggestion {
  final String fromSite;
  final String toSite;
  final int suggestedQuantity;
  final String rationale;
  final int fromSiteQuantity;
  final int toSiteQuantity;

  const FormularyRedistributionSuggestion({
    required this.fromSite,
    required this.toSite,
    required this.suggestedQuantity,
    required this.rationale,
    required this.fromSiteQuantity,
    required this.toSiteQuantity,
  });

  factory FormularyRedistributionSuggestion.fromJson(
      Map<String, dynamic> json) {
    return FormularyRedistributionSuggestion(
      fromSite: json['from_site'] as String,
      toSite: json['to_site'] as String,
      suggestedQuantity: (json['suggested_quantity'] as num).toInt(),
      rationale: json['rationale'] as String,
      fromSiteQuantity: (json['from_site_quantity'] as num).toInt(),
      toSiteQuantity: (json['to_site_quantity'] as num).toInt(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'from_site': fromSite,
      'to_site': toSite,
      'suggested_quantity': suggestedQuantity,
      'rationale': rationale,
      'from_site_quantity': fromSiteQuantity,
      'to_site_quantity': toSiteQuantity,
    };
  }
}
