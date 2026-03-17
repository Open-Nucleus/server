import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';

import '../../../core/theme/app_spacing.dart';
import '../../../shared/models/app_exception.dart';
import '../../../shared/providers/dio_provider.dart';
import '../data/clinical_api.dart';

// ---------------------------------------------------------------------------
// Provider
// ---------------------------------------------------------------------------

final clinicalApiProvider = Provider<ClinicalApi>((ref) {
  return ClinicalApi(ref.watch(dioProvider));
});

// ═══════════════════════════════════════════════════════════════════════════════
// 1. Encounter Form Dialog
// ═══════════════════════════════════════════════════════════════════════════════

class EncounterFormDialog extends ConsumerStatefulWidget {
  const EncounterFormDialog({required this.patientId, super.key});

  final String patientId;

  static Future<Map<String, dynamic>?> show(
      BuildContext context, String patientId) {
    return showDialog<Map<String, dynamic>>(
      context: context,
      builder: (_) => EncounterFormDialog(patientId: patientId),
    );
  }

  @override
  ConsumerState<EncounterFormDialog> createState() =>
      _EncounterFormDialogState();
}

class _EncounterFormDialogState extends ConsumerState<EncounterFormDialog> {
  final _formKey = GlobalKey<FormState>();
  String _status = 'planned';
  final _classCodeCtrl = TextEditingController();
  final _classSystemCtrl = TextEditingController(
      text: 'http://terminology.hl7.org/CodeSystem/v3-ActCode');
  DateTime? _periodStart;
  DateTime? _periodEnd;
  bool _saving = false;

  @override
  void dispose() {
    _classCodeCtrl.dispose();
    _classSystemCtrl.dispose();
    super.dispose();
  }

  Future<void> _pickDateTime({required bool isStart}) async {
    final now = DateTime.now();
    final date = await showDatePicker(
      context: context,
      initialDate: now,
      firstDate: DateTime(2000),
      lastDate: DateTime(2100),
    );
    if (date == null || !mounted) return;

    final time = await showTimePicker(
      context: context,
      initialTime: TimeOfDay.fromDateTime(now),
    );
    if (time == null || !mounted) return;

    final dt = DateTime(date.year, date.month, date.day, time.hour, time.minute);
    setState(() {
      if (isStart) {
        _periodStart = dt;
      } else {
        _periodEnd = dt;
      }
    });
  }

  Future<void> _onSave() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() => _saving = true);

    final fhir = <String, dynamic>{
      'resourceType': 'Encounter',
      'status': _status,
      'subject': {'reference': 'Patient/${widget.patientId}'},
    };

    if (_classCodeCtrl.text.trim().isNotEmpty) {
      fhir['class'] = {
        'system': _classSystemCtrl.text.trim(),
        'code': _classCodeCtrl.text.trim(),
      };
    }

    if (_periodStart != null || _periodEnd != null) {
      fhir['period'] = {
        if (_periodStart != null) 'start': _periodStart!.toIso8601String(),
        if (_periodEnd != null) 'end': _periodEnd!.toIso8601String(),
      };
    }

    try {
      final api = ref.read(clinicalApiProvider);
      final envelope =
          await api.createEncounter(widget.patientId, fhir);
      if (mounted) {
        if (envelope.isSuccess) {
          Navigator.of(context).pop(envelope.data?.resource);
        } else {
          _showError(envelope.error?.message ?? 'Failed to create encounter');
        }
      }
    } on DioException catch (e) {
      if (mounted) {
        final appErr = e.error;
        _showError(appErr is AppException
            ? appErr.message
            : (e.message ?? 'Network error'));
      }
    } catch (e) {
      if (mounted) _showError('Unexpected error: $e');
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  void _showError(String msg) {
    ScaffoldMessenger.of(context)
        .showSnackBar(SnackBar(content: Text(msg)));
  }

  @override
  Widget build(BuildContext context) {
    final df = DateFormat('yyyy-MM-dd HH:mm');

    return AlertDialog(
      title: const Text('New Encounter'),
      content: SizedBox(
        width: 500,
        child: Form(
          key: _formKey,
          child: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                DropdownButtonFormField<String>(
                  value: _status,
                  decoration: const InputDecoration(
                    labelText: 'Status *',
                    border: OutlineInputBorder(),
                  ),
                  items: const [
                    DropdownMenuItem(value: 'planned', child: Text('Planned')),
                    DropdownMenuItem(value: 'arrived', child: Text('Arrived')),
                    DropdownMenuItem(value: 'triaged', child: Text('Triaged')),
                    DropdownMenuItem(
                        value: 'in-progress', child: Text('In Progress')),
                    DropdownMenuItem(
                        value: 'finished', child: Text('Finished')),
                    DropdownMenuItem(
                        value: 'cancelled', child: Text('Cancelled')),
                  ],
                  onChanged: (v) => setState(() => _status = v ?? 'planned'),
                ),
                const SizedBox(height: AppSpacing.md),
                Row(
                  children: [
                    Expanded(
                      child: TextFormField(
                        controller: _classCodeCtrl,
                        decoration: const InputDecoration(
                          labelText: 'Class Code',
                          hintText: 'e.g. AMB, EMER, IMP',
                          border: OutlineInputBorder(),
                        ),
                      ),
                    ),
                    const SizedBox(width: AppSpacing.sm),
                    Expanded(
                      child: TextFormField(
                        controller: _classSystemCtrl,
                        decoration: const InputDecoration(
                          labelText: 'Class System',
                          border: OutlineInputBorder(),
                        ),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: AppSpacing.md),
                Row(
                  children: [
                    Expanded(
                      child: _DateTimeField(
                        label: 'Period Start',
                        value:
                            _periodStart != null ? df.format(_periodStart!) : null,
                        onTap: () => _pickDateTime(isStart: true),
                      ),
                    ),
                    const SizedBox(width: AppSpacing.sm),
                    Expanded(
                      child: _DateTimeField(
                        label: 'Period End',
                        value:
                            _periodEnd != null ? df.format(_periodEnd!) : null,
                        onTap: () => _pickDateTime(isStart: false),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      ),
      actions: _dialogActions(_saving, _onSave),
    );
  }
}

// ═══════════════════════════════════════════════════════════════════════════════
// 2. Observation Form Dialog
// ═══════════════════════════════════════════════════════════════════════════════

/// Common LOINC vital-sign codes and their default units.
const _loincVitals = <String, ({String display, String unit, String unitCode})>{
  '8867-4': (display: 'Heart Rate', unit: '/min', unitCode: '/min'),
  '8480-6': (display: 'Systolic Blood Pressure', unit: 'mmHg', unitCode: 'mm[Hg]'),
  '8462-4': (display: 'Diastolic Blood Pressure', unit: 'mmHg', unitCode: 'mm[Hg]'),
  '8310-5': (display: 'Body Temperature', unit: 'Cel', unitCode: 'Cel'),
  '9279-1': (display: 'Respiratory Rate', unit: '/min', unitCode: '/min'),
  '2708-6': (display: 'SpO2', unit: '%', unitCode: '%'),
  '29463-7': (display: 'Body Weight', unit: 'kg', unitCode: 'kg'),
  '8302-2': (display: 'Body Height', unit: 'cm', unitCode: 'cm'),
};

class ObservationFormDialog extends ConsumerStatefulWidget {
  const ObservationFormDialog({required this.patientId, super.key});

  final String patientId;

  static Future<Map<String, dynamic>?> show(
      BuildContext context, String patientId) {
    return showDialog<Map<String, dynamic>>(
      context: context,
      builder: (_) => ObservationFormDialog(patientId: patientId),
    );
  }

  @override
  ConsumerState<ObservationFormDialog> createState() =>
      _ObservationFormDialogState();
}

class _ObservationFormDialogState extends ConsumerState<ObservationFormDialog> {
  final _formKey = GlobalKey<FormState>();
  String? _selectedCode;
  final _valueCtrl = TextEditingController();
  String _status = 'final';
  bool _saving = false;

  String get _unit =>
      _selectedCode != null ? _loincVitals[_selectedCode]!.unit : '';

  @override
  void dispose() {
    _valueCtrl.dispose();
    super.dispose();
  }

  Future<void> _onSave() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() => _saving = true);

    final vital = _loincVitals[_selectedCode]!;
    final fhir = <String, dynamic>{
      'resourceType': 'Observation',
      'status': _status,
      'category': [
        {
          'coding': [
            {
              'system':
                  'http://terminology.hl7.org/CodeSystem/observation-category',
              'code': 'vital-signs',
              'display': 'Vital Signs',
            }
          ]
        }
      ],
      'code': {
        'coding': [
          {
            'system': 'http://loinc.org',
            'code': _selectedCode,
            'display': vital.display,
          }
        ],
        'text': vital.display,
      },
      'subject': {'reference': 'Patient/${widget.patientId}'},
      'effectiveDateTime': DateTime.now().toIso8601String(),
      'valueQuantity': {
        'value': double.tryParse(_valueCtrl.text.trim()) ?? 0,
        'unit': vital.unit,
        'system': 'http://unitsofmeasure.org',
        'code': vital.unitCode,
      },
    };

    try {
      final api = ref.read(clinicalApiProvider);
      final envelope =
          await api.createObservation(widget.patientId, fhir);
      if (mounted) {
        if (envelope.isSuccess) {
          Navigator.of(context).pop(envelope.data?.resource);
        } else {
          _showError(
              envelope.error?.message ?? 'Failed to create observation');
        }
      }
    } on DioException catch (e) {
      if (mounted) {
        final appErr = e.error;
        _showError(appErr is AppException
            ? appErr.message
            : (e.message ?? 'Network error'));
      }
    } catch (e) {
      if (mounted) _showError('Unexpected error: $e');
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  void _showError(String msg) {
    ScaffoldMessenger.of(context)
        .showSnackBar(SnackBar(content: Text(msg)));
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('New Observation'),
      content: SizedBox(
        width: 500,
        child: Form(
          key: _formKey,
          child: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                DropdownButtonFormField<String>(
                  value: _selectedCode,
                  decoration: const InputDecoration(
                    labelText: 'Vital Sign Code *',
                    border: OutlineInputBorder(),
                  ),
                  items: _loincVitals.entries
                      .map((e) => DropdownMenuItem(
                            value: e.key,
                            child: Text('${e.value.display} (${e.key})'),
                          ))
                      .toList(),
                  onChanged: (v) => setState(() => _selectedCode = v),
                  validator: (v) =>
                      v == null ? 'Please select a vital sign' : null,
                ),
                const SizedBox(height: AppSpacing.md),
                Row(
                  children: [
                    Expanded(
                      child: TextFormField(
                        controller: _valueCtrl,
                        decoration: InputDecoration(
                          labelText: 'Value *',
                          border: const OutlineInputBorder(),
                          suffixText: _unit,
                        ),
                        keyboardType: const TextInputType.numberWithOptions(
                            decimal: true),
                        validator: (v) {
                          if (v == null || v.trim().isEmpty) {
                            return 'Value is required';
                          }
                          if (double.tryParse(v.trim()) == null) {
                            return 'Must be a number';
                          }
                          return null;
                        },
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: AppSpacing.md),
                DropdownButtonFormField<String>(
                  value: _status,
                  decoration: const InputDecoration(
                    labelText: 'Status',
                    border: OutlineInputBorder(),
                  ),
                  items: const [
                    DropdownMenuItem(
                        value: 'registered', child: Text('Registered')),
                    DropdownMenuItem(
                        value: 'preliminary', child: Text('Preliminary')),
                    DropdownMenuItem(value: 'final', child: Text('Final')),
                    DropdownMenuItem(value: 'amended', child: Text('Amended')),
                  ],
                  onChanged: (v) => setState(() => _status = v ?? 'final'),
                ),
              ],
            ),
          ),
        ),
      ),
      actions: _dialogActions(_saving, _onSave),
    );
  }
}

// ═══════════════════════════════════════════════════════════════════════════════
// 3. Condition Form Dialog
// ═══════════════════════════════════════════════════════════════════════════════

class ConditionFormDialog extends ConsumerStatefulWidget {
  const ConditionFormDialog({required this.patientId, super.key});

  final String patientId;

  static Future<Map<String, dynamic>?> show(
      BuildContext context, String patientId) {
    return showDialog<Map<String, dynamic>>(
      context: context,
      builder: (_) => ConditionFormDialog(patientId: patientId),
    );
  }

  @override
  ConsumerState<ConditionFormDialog> createState() =>
      _ConditionFormDialogState();
}

class _ConditionFormDialogState extends ConsumerState<ConditionFormDialog> {
  final _formKey = GlobalKey<FormState>();
  final _displayCtrl = TextEditingController();
  final _snomedCodeCtrl = TextEditingController();
  String _clinicalStatus = 'active';
  String _verificationStatus = 'confirmed';
  bool _saving = false;

  @override
  void dispose() {
    _displayCtrl.dispose();
    _snomedCodeCtrl.dispose();
    super.dispose();
  }

  Future<void> _onSave() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() => _saving = true);

    final fhir = <String, dynamic>{
      'resourceType': 'Condition',
      'clinicalStatus': {
        'coding': [
          {
            'system':
                'http://terminology.hl7.org/CodeSystem/condition-clinical',
            'code': _clinicalStatus,
          }
        ]
      },
      'verificationStatus': {
        'coding': [
          {
            'system':
                'http://terminology.hl7.org/CodeSystem/condition-ver-status',
            'code': _verificationStatus,
          }
        ]
      },
      'code': {
        'text': _displayCtrl.text.trim(),
        if (_snomedCodeCtrl.text.trim().isNotEmpty)
          'coding': [
            {
              'system': 'http://snomed.info/sct',
              'code': _snomedCodeCtrl.text.trim(),
              'display': _displayCtrl.text.trim(),
            }
          ],
      },
      'subject': {'reference': 'Patient/${widget.patientId}'},
      'recordedDate': DateTime.now().toIso8601String(),
    };

    try {
      final api = ref.read(clinicalApiProvider);
      final envelope =
          await api.createCondition(widget.patientId, fhir);
      if (mounted) {
        if (envelope.isSuccess) {
          Navigator.of(context).pop(envelope.data?.resource);
        } else {
          _showError(envelope.error?.message ?? 'Failed to create condition');
        }
      }
    } on DioException catch (e) {
      if (mounted) {
        final appErr = e.error;
        _showError(appErr is AppException
            ? appErr.message
            : (e.message ?? 'Network error'));
      }
    } catch (e) {
      if (mounted) _showError('Unexpected error: $e');
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  void _showError(String msg) {
    ScaffoldMessenger.of(context)
        .showSnackBar(SnackBar(content: Text(msg)));
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('New Condition'),
      content: SizedBox(
        width: 500,
        child: Form(
          key: _formKey,
          child: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                TextFormField(
                  controller: _displayCtrl,
                  decoration: const InputDecoration(
                    labelText: 'Condition Name *',
                    border: OutlineInputBorder(),
                  ),
                  validator: (v) => (v == null || v.trim().isEmpty)
                      ? 'Condition name is required'
                      : null,
                ),
                const SizedBox(height: AppSpacing.md),
                TextFormField(
                  controller: _snomedCodeCtrl,
                  decoration: const InputDecoration(
                    labelText: 'SNOMED Code',
                    hintText: 'e.g. 38341003',
                    border: OutlineInputBorder(),
                  ),
                ),
                const SizedBox(height: AppSpacing.md),
                DropdownButtonFormField<String>(
                  value: _clinicalStatus,
                  decoration: const InputDecoration(
                    labelText: 'Clinical Status',
                    border: OutlineInputBorder(),
                  ),
                  items: const [
                    DropdownMenuItem(value: 'active', child: Text('Active')),
                    DropdownMenuItem(
                        value: 'inactive', child: Text('Inactive')),
                    DropdownMenuItem(
                        value: 'resolved', child: Text('Resolved')),
                  ],
                  onChanged: (v) =>
                      setState(() => _clinicalStatus = v ?? 'active'),
                ),
                const SizedBox(height: AppSpacing.md),
                DropdownButtonFormField<String>(
                  value: _verificationStatus,
                  decoration: const InputDecoration(
                    labelText: 'Verification Status',
                    border: OutlineInputBorder(),
                  ),
                  items: const [
                    DropdownMenuItem(
                        value: 'confirmed', child: Text('Confirmed')),
                    DropdownMenuItem(
                        value: 'unconfirmed', child: Text('Unconfirmed')),
                    DropdownMenuItem(
                        value: 'provisional', child: Text('Provisional')),
                  ],
                  onChanged: (v) =>
                      setState(() => _verificationStatus = v ?? 'confirmed'),
                ),
              ],
            ),
          ),
        ),
      ),
      actions: _dialogActions(_saving, _onSave),
    );
  }
}

// ═══════════════════════════════════════════════════════════════════════════════
// 4. Medication Request Form Dialog
// ═══════════════════════════════════════════════════════════════════════════════

class MedicationRequestFormDialog extends ConsumerStatefulWidget {
  const MedicationRequestFormDialog({required this.patientId, super.key});

  final String patientId;

  static Future<Map<String, dynamic>?> show(
      BuildContext context, String patientId) {
    return showDialog<Map<String, dynamic>>(
      context: context,
      builder: (_) => MedicationRequestFormDialog(patientId: patientId),
    );
  }

  @override
  ConsumerState<MedicationRequestFormDialog> createState() =>
      _MedicationRequestFormDialogState();
}

class _MedicationRequestFormDialogState
    extends ConsumerState<MedicationRequestFormDialog> {
  final _formKey = GlobalKey<FormState>();
  final _medNameCtrl = TextEditingController();
  final _rxNormCodeCtrl = TextEditingController();
  String _status = 'active';
  String _intent = 'order';
  final _dosageTextCtrl = TextEditingController();
  final _doseValueCtrl = TextEditingController();
  final _doseUnitCtrl = TextEditingController();
  final _frequencyCtrl = TextEditingController();
  String _route = 'oral';
  bool _saving = false;

  static const _routes = [
    'oral',
    'intravenous',
    'intramuscular',
    'subcutaneous',
    'topical',
    'inhalation',
    'rectal',
    'ophthalmic',
    'otic',
    'nasal',
  ];

  @override
  void dispose() {
    _medNameCtrl.dispose();
    _rxNormCodeCtrl.dispose();
    _dosageTextCtrl.dispose();
    _doseValueCtrl.dispose();
    _doseUnitCtrl.dispose();
    _frequencyCtrl.dispose();
    super.dispose();
  }

  Future<void> _onSave() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() => _saving = true);

    final fhir = <String, dynamic>{
      'resourceType': 'MedicationRequest',
      'status': _status,
      'intent': _intent,
      'medicationCodeableConcept': {
        'text': _medNameCtrl.text.trim(),
        if (_rxNormCodeCtrl.text.trim().isNotEmpty)
          'coding': [
            {
              'system': 'http://www.nlm.nih.gov/research/umls/rxnorm',
              'code': _rxNormCodeCtrl.text.trim(),
              'display': _medNameCtrl.text.trim(),
            }
          ],
      },
      'subject': {'reference': 'Patient/${widget.patientId}'},
      'authoredOn': DateTime.now().toIso8601String(),
    };

    // Dosage instruction
    final dosage = <String, dynamic>{};
    if (_dosageTextCtrl.text.trim().isNotEmpty) {
      dosage['text'] = _dosageTextCtrl.text.trim();
    }
    if (_doseValueCtrl.text.trim().isNotEmpty) {
      dosage['doseAndRate'] = [
        {
          'doseQuantity': {
            'value': double.tryParse(_doseValueCtrl.text.trim()) ?? 0,
            'unit': _doseUnitCtrl.text.trim(),
          }
        }
      ];
    }
    if (_frequencyCtrl.text.trim().isNotEmpty) {
      dosage['timing'] = {
        'code': {'text': _frequencyCtrl.text.trim()},
      };
    }
    dosage['route'] = {
      'text': _route,
    };
    if (dosage.isNotEmpty) {
      fhir['dosageInstruction'] = [dosage];
    }

    try {
      final api = ref.read(clinicalApiProvider);
      final envelope =
          await api.createMedicationRequest(widget.patientId, fhir);
      if (mounted) {
        if (envelope.isSuccess) {
          Navigator.of(context).pop(envelope.data?.resource);
        } else {
          _showError(envelope.error?.message ??
              'Failed to create medication request');
        }
      }
    } on DioException catch (e) {
      if (mounted) {
        final appErr = e.error;
        _showError(appErr is AppException
            ? appErr.message
            : (e.message ?? 'Network error'));
      }
    } catch (e) {
      if (mounted) _showError('Unexpected error: $e');
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  void _showError(String msg) {
    ScaffoldMessenger.of(context)
        .showSnackBar(SnackBar(content: Text(msg)));
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('New Medication Request'),
      content: SizedBox(
        width: 550,
        child: Form(
          key: _formKey,
          child: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                TextFormField(
                  controller: _medNameCtrl,
                  decoration: const InputDecoration(
                    labelText: 'Medication Name *',
                    border: OutlineInputBorder(),
                  ),
                  validator: (v) => (v == null || v.trim().isEmpty)
                      ? 'Medication name is required'
                      : null,
                ),
                const SizedBox(height: AppSpacing.md),
                TextFormField(
                  controller: _rxNormCodeCtrl,
                  decoration: const InputDecoration(
                    labelText: 'RxNorm Code',
                    border: OutlineInputBorder(),
                  ),
                ),
                const SizedBox(height: AppSpacing.md),
                Row(
                  children: [
                    Expanded(
                      child: DropdownButtonFormField<String>(
                        value: _status,
                        decoration: const InputDecoration(
                          labelText: 'Status',
                          border: OutlineInputBorder(),
                        ),
                        items: const [
                          DropdownMenuItem(
                              value: 'active', child: Text('Active')),
                          DropdownMenuItem(
                              value: 'stopped', child: Text('Stopped')),
                          DropdownMenuItem(
                              value: 'completed', child: Text('Completed')),
                        ],
                        onChanged: (v) =>
                            setState(() => _status = v ?? 'active'),
                      ),
                    ),
                    const SizedBox(width: AppSpacing.sm),
                    Expanded(
                      child: DropdownButtonFormField<String>(
                        value: _intent,
                        decoration: const InputDecoration(
                          labelText: 'Intent',
                          border: OutlineInputBorder(),
                        ),
                        items: const [
                          DropdownMenuItem(
                              value: 'order', child: Text('Order')),
                          DropdownMenuItem(
                              value: 'plan', child: Text('Plan')),
                        ],
                        onChanged: (v) =>
                            setState(() => _intent = v ?? 'order'),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: AppSpacing.md),
                TextFormField(
                  controller: _dosageTextCtrl,
                  decoration: const InputDecoration(
                    labelText: 'Dosage Instructions',
                    hintText: 'e.g. Take 1 tablet twice daily',
                    border: OutlineInputBorder(),
                  ),
                ),
                const SizedBox(height: AppSpacing.md),
                Row(
                  children: [
                    Expanded(
                      child: TextFormField(
                        controller: _doseValueCtrl,
                        decoration: const InputDecoration(
                          labelText: 'Dose Value',
                          border: OutlineInputBorder(),
                        ),
                        keyboardType: const TextInputType.numberWithOptions(
                            decimal: true),
                      ),
                    ),
                    const SizedBox(width: AppSpacing.sm),
                    Expanded(
                      child: TextFormField(
                        controller: _doseUnitCtrl,
                        decoration: const InputDecoration(
                          labelText: 'Dose Unit',
                          hintText: 'e.g. mg, mL',
                          border: OutlineInputBorder(),
                        ),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: AppSpacing.md),
                TextFormField(
                  controller: _frequencyCtrl,
                  decoration: const InputDecoration(
                    labelText: 'Frequency',
                    hintText: 'e.g. BID, TID, Q8H',
                    border: OutlineInputBorder(),
                  ),
                ),
                const SizedBox(height: AppSpacing.md),
                DropdownButtonFormField<String>(
                  value: _route,
                  decoration: const InputDecoration(
                    labelText: 'Route',
                    border: OutlineInputBorder(),
                  ),
                  items: _routes
                      .map((r) => DropdownMenuItem(
                            value: r,
                            child: Text(r[0].toUpperCase() + r.substring(1)),
                          ))
                      .toList(),
                  onChanged: (v) => setState(() => _route = v ?? 'oral'),
                ),
              ],
            ),
          ),
        ),
      ),
      actions: _dialogActions(_saving, _onSave),
    );
  }
}

// ═══════════════════════════════════════════════════════════════════════════════
// 5. Allergy Form Dialog
// ═══════════════════════════════════════════════════════════════════════════════

class AllergyFormDialog extends ConsumerStatefulWidget {
  const AllergyFormDialog({required this.patientId, super.key});

  final String patientId;

  static Future<Map<String, dynamic>?> show(
      BuildContext context, String patientId) {
    return showDialog<Map<String, dynamic>>(
      context: context,
      builder: (_) => AllergyFormDialog(patientId: patientId),
    );
  }

  @override
  ConsumerState<AllergyFormDialog> createState() => _AllergyFormDialogState();
}

class _AllergyFormDialogState extends ConsumerState<AllergyFormDialog> {
  final _formKey = GlobalKey<FormState>();
  final _substanceCtrl = TextEditingController();
  final _snomedCodeCtrl = TextEditingController();
  String _type = 'allergy';
  String _clinicalStatus = 'active';
  String _criticality = 'low';
  bool _saving = false;

  @override
  void dispose() {
    _substanceCtrl.dispose();
    _snomedCodeCtrl.dispose();
    super.dispose();
  }

  Future<void> _onSave() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() => _saving = true);

    final fhir = <String, dynamic>{
      'resourceType': 'AllergyIntolerance',
      'type': _type,
      'criticality': _criticality,
      'clinicalStatus': {
        'coding': [
          {
            'system':
                'http://terminology.hl7.org/CodeSystem/allergyintolerance-clinical',
            'code': _clinicalStatus,
          }
        ]
      },
      'code': {
        'text': _substanceCtrl.text.trim(),
        if (_snomedCodeCtrl.text.trim().isNotEmpty)
          'coding': [
            {
              'system': 'http://snomed.info/sct',
              'code': _snomedCodeCtrl.text.trim(),
              'display': _substanceCtrl.text.trim(),
            }
          ],
      },
      'patient': {'reference': 'Patient/${widget.patientId}'},
      'recordedDate': DateTime.now().toIso8601String(),
    };

    try {
      final api = ref.read(clinicalApiProvider);
      final envelope =
          await api.createAllergyIntolerance(widget.patientId, fhir);
      if (mounted) {
        if (envelope.isSuccess) {
          Navigator.of(context).pop(envelope.data?.resource);
        } else {
          _showError(envelope.error?.message ??
              'Failed to create allergy intolerance');
        }
      }
    } on DioException catch (e) {
      if (mounted) {
        final appErr = e.error;
        _showError(appErr is AppException
            ? appErr.message
            : (e.message ?? 'Network error'));
      }
    } catch (e) {
      if (mounted) _showError('Unexpected error: $e');
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  void _showError(String msg) {
    ScaffoldMessenger.of(context)
        .showSnackBar(SnackBar(content: Text(msg)));
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('New Allergy / Intolerance'),
      content: SizedBox(
        width: 500,
        child: Form(
          key: _formKey,
          child: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                TextFormField(
                  controller: _substanceCtrl,
                  decoration: const InputDecoration(
                    labelText: 'Substance *',
                    border: OutlineInputBorder(),
                  ),
                  validator: (v) => (v == null || v.trim().isEmpty)
                      ? 'Substance is required'
                      : null,
                ),
                const SizedBox(height: AppSpacing.md),
                TextFormField(
                  controller: _snomedCodeCtrl,
                  decoration: const InputDecoration(
                    labelText: 'SNOMED Code',
                    border: OutlineInputBorder(),
                  ),
                ),
                const SizedBox(height: AppSpacing.md),
                DropdownButtonFormField<String>(
                  value: _type,
                  decoration: const InputDecoration(
                    labelText: 'Type',
                    border: OutlineInputBorder(),
                  ),
                  items: const [
                    DropdownMenuItem(
                        value: 'allergy', child: Text('Allergy')),
                    DropdownMenuItem(
                        value: 'intolerance', child: Text('Intolerance')),
                  ],
                  onChanged: (v) =>
                      setState(() => _type = v ?? 'allergy'),
                ),
                const SizedBox(height: AppSpacing.md),
                DropdownButtonFormField<String>(
                  value: _clinicalStatus,
                  decoration: const InputDecoration(
                    labelText: 'Clinical Status',
                    border: OutlineInputBorder(),
                  ),
                  items: const [
                    DropdownMenuItem(value: 'active', child: Text('Active')),
                    DropdownMenuItem(
                        value: 'inactive', child: Text('Inactive')),
                    DropdownMenuItem(
                        value: 'resolved', child: Text('Resolved')),
                  ],
                  onChanged: (v) =>
                      setState(() => _clinicalStatus = v ?? 'active'),
                ),
                const SizedBox(height: AppSpacing.md),
                DropdownButtonFormField<String>(
                  value: _criticality,
                  decoration: const InputDecoration(
                    labelText: 'Criticality',
                    border: OutlineInputBorder(),
                  ),
                  items: const [
                    DropdownMenuItem(value: 'low', child: Text('Low')),
                    DropdownMenuItem(value: 'high', child: Text('High')),
                    DropdownMenuItem(
                        value: 'unable-to-assess',
                        child: Text('Unable to Assess')),
                  ],
                  onChanged: (v) =>
                      setState(() => _criticality = v ?? 'low'),
                ),
              ],
            ),
          ),
        ),
      ),
      actions: _dialogActions(_saving, _onSave),
    );
  }
}

// ═══════════════════════════════════════════════════════════════════════════════
// 6. Immunization Form Dialog
// ═══════════════════════════════════════════════════════════════════════════════

class ImmunizationFormDialog extends ConsumerStatefulWidget {
  const ImmunizationFormDialog({required this.patientId, super.key});

  final String patientId;

  static Future<Map<String, dynamic>?> show(
      BuildContext context, String patientId) {
    return showDialog<Map<String, dynamic>>(
      context: context,
      builder: (_) => ImmunizationFormDialog(patientId: patientId),
    );
  }

  @override
  ConsumerState<ImmunizationFormDialog> createState() =>
      _ImmunizationFormDialogState();
}

class _ImmunizationFormDialogState
    extends ConsumerState<ImmunizationFormDialog> {
  final _formKey = GlobalKey<FormState>();
  final _vaccineNameCtrl = TextEditingController();
  final _cvxCodeCtrl = TextEditingController();
  DateTime? _occurrenceDate;
  String _status = 'completed';
  bool _saving = false;

  @override
  void dispose() {
    _vaccineNameCtrl.dispose();
    _cvxCodeCtrl.dispose();
    super.dispose();
  }

  Future<void> _pickDate() async {
    final picked = await showDatePicker(
      context: context,
      initialDate: DateTime.now(),
      firstDate: DateTime(2000),
      lastDate: DateTime.now(),
    );
    if (picked != null) setState(() => _occurrenceDate = picked);
  }

  Future<void> _onSave() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() => _saving = true);

    final fhir = <String, dynamic>{
      'resourceType': 'Immunization',
      'status': _status,
      'vaccineCode': {
        'text': _vaccineNameCtrl.text.trim(),
        if (_cvxCodeCtrl.text.trim().isNotEmpty)
          'coding': [
            {
              'system': 'http://hl7.org/fhir/sid/cvx',
              'code': _cvxCodeCtrl.text.trim(),
              'display': _vaccineNameCtrl.text.trim(),
            }
          ],
      },
      'patient': {'reference': 'Patient/${widget.patientId}'},
      'occurrenceDateTime': _occurrenceDate != null
          ? DateFormat('yyyy-MM-dd').format(_occurrenceDate!)
          : DateFormat('yyyy-MM-dd').format(DateTime.now()),
    };

    try {
      final api = ref.read(clinicalApiProvider);
      final envelope =
          await api.createImmunization(widget.patientId, fhir);
      if (mounted) {
        if (envelope.isSuccess) {
          Navigator.of(context).pop(envelope.data?.resource);
        } else {
          _showError(
              envelope.error?.message ?? 'Failed to create immunization');
        }
      }
    } on DioException catch (e) {
      if (mounted) {
        final appErr = e.error;
        _showError(appErr is AppException
            ? appErr.message
            : (e.message ?? 'Network error'));
      }
    } catch (e) {
      if (mounted) _showError('Unexpected error: $e');
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  void _showError(String msg) {
    ScaffoldMessenger.of(context)
        .showSnackBar(SnackBar(content: Text(msg)));
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('New Immunization'),
      content: SizedBox(
        width: 500,
        child: Form(
          key: _formKey,
          child: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                TextFormField(
                  controller: _vaccineNameCtrl,
                  decoration: const InputDecoration(
                    labelText: 'Vaccine Name *',
                    border: OutlineInputBorder(),
                  ),
                  validator: (v) => (v == null || v.trim().isEmpty)
                      ? 'Vaccine name is required'
                      : null,
                ),
                const SizedBox(height: AppSpacing.md),
                TextFormField(
                  controller: _cvxCodeCtrl,
                  decoration: const InputDecoration(
                    labelText: 'CVX Code',
                    border: OutlineInputBorder(),
                  ),
                ),
                const SizedBox(height: AppSpacing.md),
                _DateTimeField(
                  label: 'Date',
                  value: _occurrenceDate != null
                      ? DateFormat('yyyy-MM-dd').format(_occurrenceDate!)
                      : null,
                  onTap: _pickDate,
                ),
                const SizedBox(height: AppSpacing.md),
                DropdownButtonFormField<String>(
                  value: _status,
                  decoration: const InputDecoration(
                    labelText: 'Status',
                    border: OutlineInputBorder(),
                  ),
                  items: const [
                    DropdownMenuItem(
                        value: 'completed', child: Text('Completed')),
                    DropdownMenuItem(
                        value: 'not-done', child: Text('Not Done')),
                  ],
                  onChanged: (v) =>
                      setState(() => _status = v ?? 'completed'),
                ),
              ],
            ),
          ),
        ),
      ),
      actions: _dialogActions(_saving, _onSave),
    );
  }
}

// ═══════════════════════════════════════════════════════════════════════════════
// 7. Procedure Form Dialog
// ═══════════════════════════════════════════════════════════════════════════════

class ProcedureFormDialog extends ConsumerStatefulWidget {
  const ProcedureFormDialog({required this.patientId, super.key});

  final String patientId;

  static Future<Map<String, dynamic>?> show(
      BuildContext context, String patientId) {
    return showDialog<Map<String, dynamic>>(
      context: context,
      builder: (_) => ProcedureFormDialog(patientId: patientId),
    );
  }

  @override
  ConsumerState<ProcedureFormDialog> createState() =>
      _ProcedureFormDialogState();
}

class _ProcedureFormDialogState extends ConsumerState<ProcedureFormDialog> {
  final _formKey = GlobalKey<FormState>();
  final _nameCtrl = TextEditingController();
  final _snomedCodeCtrl = TextEditingController();
  DateTime? _performedDate;
  String _status = 'completed';
  bool _saving = false;

  @override
  void dispose() {
    _nameCtrl.dispose();
    _snomedCodeCtrl.dispose();
    super.dispose();
  }

  Future<void> _pickDate() async {
    final picked = await showDatePicker(
      context: context,
      initialDate: DateTime.now(),
      firstDate: DateTime(2000),
      lastDate: DateTime.now(),
    );
    if (picked != null) setState(() => _performedDate = picked);
  }

  Future<void> _onSave() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() => _saving = true);

    final fhir = <String, dynamic>{
      'resourceType': 'Procedure',
      'status': _status,
      'code': {
        'text': _nameCtrl.text.trim(),
        if (_snomedCodeCtrl.text.trim().isNotEmpty)
          'coding': [
            {
              'system': 'http://snomed.info/sct',
              'code': _snomedCodeCtrl.text.trim(),
              'display': _nameCtrl.text.trim(),
            }
          ],
      },
      'subject': {'reference': 'Patient/${widget.patientId}'},
      'performedDateTime': _performedDate != null
          ? DateFormat('yyyy-MM-dd').format(_performedDate!)
          : DateFormat('yyyy-MM-dd').format(DateTime.now()),
    };

    try {
      final api = ref.read(clinicalApiProvider);
      final envelope =
          await api.createProcedure(widget.patientId, fhir);
      if (mounted) {
        if (envelope.isSuccess) {
          Navigator.of(context).pop(envelope.data?.resource);
        } else {
          _showError(
              envelope.error?.message ?? 'Failed to create procedure');
        }
      }
    } on DioException catch (e) {
      if (mounted) {
        final appErr = e.error;
        _showError(appErr is AppException
            ? appErr.message
            : (e.message ?? 'Network error'));
      }
    } catch (e) {
      if (mounted) _showError('Unexpected error: $e');
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  void _showError(String msg) {
    ScaffoldMessenger.of(context)
        .showSnackBar(SnackBar(content: Text(msg)));
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('New Procedure'),
      content: SizedBox(
        width: 500,
        child: Form(
          key: _formKey,
          child: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                TextFormField(
                  controller: _nameCtrl,
                  decoration: const InputDecoration(
                    labelText: 'Procedure Name *',
                    border: OutlineInputBorder(),
                  ),
                  validator: (v) => (v == null || v.trim().isEmpty)
                      ? 'Procedure name is required'
                      : null,
                ),
                const SizedBox(height: AppSpacing.md),
                TextFormField(
                  controller: _snomedCodeCtrl,
                  decoration: const InputDecoration(
                    labelText: 'SNOMED Code',
                    border: OutlineInputBorder(),
                  ),
                ),
                const SizedBox(height: AppSpacing.md),
                _DateTimeField(
                  label: 'Date',
                  value: _performedDate != null
                      ? DateFormat('yyyy-MM-dd').format(_performedDate!)
                      : null,
                  onTap: _pickDate,
                ),
                const SizedBox(height: AppSpacing.md),
                DropdownButtonFormField<String>(
                  value: _status,
                  decoration: const InputDecoration(
                    labelText: 'Status',
                    border: OutlineInputBorder(),
                  ),
                  items: const [
                    DropdownMenuItem(
                        value: 'completed', child: Text('Completed')),
                    DropdownMenuItem(
                        value: 'in-progress', child: Text('In Progress')),
                    DropdownMenuItem(
                        value: 'preparation', child: Text('Preparation')),
                  ],
                  onChanged: (v) =>
                      setState(() => _status = v ?? 'completed'),
                ),
              ],
            ),
          ),
        ),
      ),
      actions: _dialogActions(_saving, _onSave),
    );
  }
}

// ═══════════════════════════════════════════════════════════════════════════════
// Shared helpers
// ═══════════════════════════════════════════════════════════════════════════════

/// Standard Cancel / Save action buttons for all clinical dialogs.
List<Widget> _dialogActions(bool saving, VoidCallback onSave) {
  return [
    Builder(builder: (context) {
      return TextButton(
        onPressed: saving ? null : () => Navigator.of(context).pop(),
        child: const Text('Cancel'),
      );
    }),
    Builder(builder: (context) {
      return FilledButton(
        onPressed: saving ? null : onSave,
        child: saving
            ? const SizedBox(
                width: 18,
                height: 18,
                child: CircularProgressIndicator(
                  strokeWidth: 2,
                  color: Colors.white,
                ),
              )
            : const Text('Save'),
      );
    }),
  ];
}

/// A read-only InputDecorator styled as a date/time picker field.
class _DateTimeField extends StatelessWidget {
  const _DateTimeField({
    required this.label,
    this.value,
    required this.onTap,
  });

  final String label;
  final String? value;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return InkWell(
      onTap: onTap,
      child: InputDecorator(
        decoration: InputDecoration(
          labelText: label,
          border: const OutlineInputBorder(),
          suffixIcon: const Icon(Icons.calendar_today),
        ),
        child: Text(
          value ?? 'Select',
          style: TextStyle(
            color: value != null
                ? colorScheme.onSurface
                : colorScheme.onSurfaceVariant,
          ),
        ),
      ),
    );
  }
}
