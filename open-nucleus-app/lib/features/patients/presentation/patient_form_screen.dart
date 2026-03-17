import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';

import '../../../core/theme/app_spacing.dart';
import '../../../shared/models/app_exception.dart';
import '../../../shared/models/patient_models.dart';
import '../../../shared/providers/dio_provider.dart';
import '../data/patient_api.dart';

// ---------------------------------------------------------------------------
// Providers
// ---------------------------------------------------------------------------

final _patientApiProvider = Provider<PatientApi>((ref) {
  return PatientApi(ref.watch(dioProvider));
});

// ---------------------------------------------------------------------------
// Patient Form Screen
// ---------------------------------------------------------------------------

/// Patient create/edit form with a card-based layout (max 800px, centered).
///
/// When [patientId] is `null` the form is in **create** mode; when provided
/// it loads the existing patient and switches to **edit** mode.
class PatientFormScreen extends ConsumerStatefulWidget {
  const PatientFormScreen({this.patientId, super.key});

  /// If non-null, the form operates in edit mode for this patient.
  final String? patientId;

  @override
  ConsumerState<PatientFormScreen> createState() => _PatientFormScreenState();
}

class _PatientFormScreenState extends ConsumerState<PatientFormScreen> {
  final _formKey = GlobalKey<FormState>();

  // --- Demographics ---
  final _familyNameCtrl = TextEditingController();
  final _givenNamesCtrl = TextEditingController();
  String _gender = 'unknown';
  DateTime? _birthDate;
  bool _active = true;

  // --- Identifier ---
  final _identifierSystemCtrl = TextEditingController();
  final _identifierValueCtrl = TextEditingController();

  // --- Contact ---
  final _phoneCtrl = TextEditingController();
  final _emailCtrl = TextEditingController();

  // --- Address ---
  final _addressLineCtrl = TextEditingController();
  final _cityCtrl = TextEditingController();
  final _stateCtrl = TextEditingController();
  final _postalCodeCtrl = TextEditingController();
  final _countryCtrl = TextEditingController();

  bool _loading = false;
  bool _loadingPatient = false;
  String? _errorMessage;

  bool get _isEditing => widget.patientId != null;

  @override
  void initState() {
    super.initState();
    if (_isEditing) {
      _loadExistingPatient();
    }
  }

  @override
  void dispose() {
    _familyNameCtrl.dispose();
    _givenNamesCtrl.dispose();
    _identifierSystemCtrl.dispose();
    _identifierValueCtrl.dispose();
    _phoneCtrl.dispose();
    _emailCtrl.dispose();
    _addressLineCtrl.dispose();
    _cityCtrl.dispose();
    _stateCtrl.dispose();
    _postalCodeCtrl.dispose();
    _countryCtrl.dispose();
    super.dispose();
  }

  Future<void> _loadExistingPatient() async {
    setState(() => _loadingPatient = true);
    try {
      final api = ref.read(_patientApiProvider);
      final envelope = await api.getPatient(widget.patientId!);
      if (envelope.isSuccess && envelope.data != null) {
        _populateFromFhir(envelope.data!.patient);
      }
    } catch (e) {
      setState(() => _errorMessage = 'Failed to load patient: $e');
    } finally {
      setState(() => _loadingPatient = false);
    }
  }

  void _populateFromFhir(Map<String, dynamic> patient) {
    // Name
    final names = patient['name'] as List<dynamic>?;
    if (names != null && names.isNotEmpty) {
      final name = names.first as Map<String, dynamic>;
      _familyNameCtrl.text = name['family'] as String? ?? '';
      final given = (name['given'] as List<dynamic>?)
          ?.map((g) => g as String)
          .toList();
      if (given != null) _givenNamesCtrl.text = given.join(', ');
    }

    // Gender
    _gender = patient['gender'] as String? ?? 'unknown';

    // Birth date
    final bd = patient['birthDate'] as String?;
    if (bd != null) {
      _birthDate = DateTime.tryParse(bd);
    }

    // Active
    _active = patient['active'] as bool? ?? true;

    // Identifier
    final identifiers = patient['identifier'] as List<dynamic>?;
    if (identifiers != null && identifiers.isNotEmpty) {
      final ident = identifiers.first as Map<String, dynamic>;
      _identifierSystemCtrl.text = ident['system'] as String? ?? '';
      _identifierValueCtrl.text = ident['value'] as String? ?? '';
    }

    // Telecom
    final telecoms = patient['telecom'] as List<dynamic>?;
    if (telecoms != null) {
      for (final t in telecoms) {
        final telecom = t as Map<String, dynamic>;
        final system = telecom['system'] as String?;
        final value = telecom['value'] as String?;
        if (system == 'phone' && value != null) _phoneCtrl.text = value;
        if (system == 'email' && value != null) _emailCtrl.text = value;
      }
    }

    // Address
    final addresses = patient['address'] as List<dynamic>?;
    if (addresses != null && addresses.isNotEmpty) {
      final addr = addresses.first as Map<String, dynamic>;
      final lines = addr['line'] as List<dynamic>?;
      if (lines != null && lines.isNotEmpty) {
        _addressLineCtrl.text = lines.first as String;
      }
      _cityCtrl.text = addr['city'] as String? ?? '';
      _stateCtrl.text = addr['state'] as String? ?? '';
      _postalCodeCtrl.text = addr['postalCode'] as String? ?? '';
      _countryCtrl.text = addr['country'] as String? ?? '';
    }

    setState(() {});
  }

  Map<String, dynamic> _buildFhirPatient() {
    final fhir = <String, dynamic>{
      'resourceType': 'Patient',
      'active': _active,
      'gender': _gender,
      'name': [
        {
          'use': 'official',
          'family': _familyNameCtrl.text.trim(),
          if (_givenNamesCtrl.text.trim().isNotEmpty)
            'given': _givenNamesCtrl.text
                .split(',')
                .map((s) => s.trim())
                .where((s) => s.isNotEmpty)
                .toList(),
        }
      ],
    };

    // Birth date
    if (_birthDate != null) {
      fhir['birthDate'] = DateFormat('yyyy-MM-dd').format(_birthDate!);
    }

    // Identifier
    if (_identifierValueCtrl.text.trim().isNotEmpty) {
      fhir['identifier'] = [
        {
          if (_identifierSystemCtrl.text.trim().isNotEmpty)
            'system': _identifierSystemCtrl.text.trim(),
          'value': _identifierValueCtrl.text.trim(),
        }
      ];
    }

    // Telecom
    final telecoms = <Map<String, dynamic>>[];
    if (_phoneCtrl.text.trim().isNotEmpty) {
      telecoms.add({
        'system': 'phone',
        'value': _phoneCtrl.text.trim(),
        'use': 'mobile',
      });
    }
    if (_emailCtrl.text.trim().isNotEmpty) {
      telecoms.add({
        'system': 'email',
        'value': _emailCtrl.text.trim(),
      });
    }
    if (telecoms.isNotEmpty) fhir['telecom'] = telecoms;

    // Address
    if (_addressLineCtrl.text.trim().isNotEmpty ||
        _cityCtrl.text.trim().isNotEmpty ||
        _countryCtrl.text.trim().isNotEmpty) {
      fhir['address'] = [
        {
          if (_addressLineCtrl.text.trim().isNotEmpty)
            'line': [_addressLineCtrl.text.trim()],
          if (_cityCtrl.text.trim().isNotEmpty)
            'city': _cityCtrl.text.trim(),
          if (_stateCtrl.text.trim().isNotEmpty)
            'state': _stateCtrl.text.trim(),
          if (_postalCodeCtrl.text.trim().isNotEmpty)
            'postalCode': _postalCodeCtrl.text.trim(),
          if (_countryCtrl.text.trim().isNotEmpty)
            'country': _countryCtrl.text.trim(),
        }
      ];
    }

    // Preserve ID for edits
    if (_isEditing) {
      fhir['id'] = widget.patientId;
    }

    return fhir;
  }

  Future<void> _onSave() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() {
      _loading = true;
      _errorMessage = null;
    });

    try {
      final api = ref.read(_patientApiProvider);
      final fhir = _buildFhirPatient();

      String? createdId;

      if (_isEditing) {
        // --- Update ---
        final envelope = await api.updatePatient(widget.patientId!, fhir);
        if (!envelope.isSuccess) {
          setState(() => _errorMessage =
              envelope.error?.message ?? 'Failed to update patient');
          return;
        }
        createdId = widget.patientId;
      } else {
        // --- Create ---
        final envelope = await api.createPatient(fhir);
        if (!envelope.isSuccess) {
          setState(() => _errorMessage =
              envelope.error?.message ?? 'Failed to create patient');
          return;
        }
        createdId = envelope.data?.resource?['id'] as String?;

        // Duplicate detection — fire and show warning if matches found
        if (createdId != null) {
          await _checkForDuplicates(createdId);
        }
      }

      if (mounted && createdId != null) {
        context.go('/patients/$createdId');
      } else if (mounted) {
        context.pop();
      }
    } on DioException catch (e) {
      final appErr = e.error;
      if (appErr is AppException) {
        setState(() => _errorMessage = appErr.message);
      } else {
        setState(() => _errorMessage = e.message ?? 'Network error');
      }
    } catch (e) {
      setState(() => _errorMessage = 'Unexpected error: $e');
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _checkForDuplicates(String createdId) async {
    try {
      final api = ref.read(_patientApiProvider);
      final request = MatchPatientsRequest(
        familyName: _familyNameCtrl.text.trim(),
        givenNames: _givenNamesCtrl.text
            .split(',')
            .map((s) => s.trim())
            .where((s) => s.isNotEmpty)
            .toList(),
        gender: _gender,
        birthDateApprox: _birthDate != null
            ? DateFormat('yyyy-MM-dd').format(_birthDate!)
            : '',
        district: _cityCtrl.text.trim(),
        threshold: 0.7,
      );

      final matchEnvelope = await api.matchPatients(request);
      if (matchEnvelope.isSuccess && matchEnvelope.data != null) {
        final matches = matchEnvelope.data!.matches
            .where((m) => m.patientId != createdId)
            .toList();
        if (matches.isNotEmpty && mounted) {
          await _showDuplicateWarning(matches);
        }
      }
    } catch (_) {
      // Duplicate check is non-critical; swallow errors.
    }
  }

  Future<void> _showDuplicateWarning(List<PatientMatch> matches) async {
    await showDialog<void>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Row(
          children: [
            Icon(Icons.warning_amber, color: Colors.orange),
            SizedBox(width: 8),
            Text('Potential Duplicates Found'),
          ],
        ),
        content: SizedBox(
          width: 400,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text(
                'The following existing patients may match the record you '
                'just created:',
              ),
              const SizedBox(height: AppSpacing.md),
              ...matches.map((m) => ListTile(
                    dense: true,
                    leading: const Icon(Icons.person_outline),
                    title: Text('Patient ${m.patientId}'),
                    subtitle: Text(
                      'Confidence: ${(m.confidence * 100).toStringAsFixed(0)}% '
                      '(${m.matchFactors.join(", ")})',
                    ),
                    onTap: () {
                      Navigator.of(ctx).pop();
                      context.go('/patients/${m.patientId}');
                    },
                  )),
            ],
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(),
            child: const Text('Dismiss'),
          ),
        ],
      ),
    );
  }

  Future<void> _pickDate() async {
    final now = DateTime.now();
    final picked = await showDatePicker(
      context: context,
      initialDate: _birthDate ?? now,
      firstDate: DateTime(1900),
      lastDate: now,
    );
    if (picked != null) {
      setState(() => _birthDate = picked);
    }
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    if (_loadingPatient) {
      return const Center(child: CircularProgressIndicator());
    }

    return SingleChildScrollView(
      padding: AppSpacing.pagePadding,
      child: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 800),
          child: Form(
            key: _formKey,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                // --- Header ---
                Row(
                  children: [
                    Icon(
                      _isEditing ? Icons.edit : Icons.person_add,
                      size: 28,
                      color: colorScheme.primary,
                    ),
                    const SizedBox(width: AppSpacing.sm),
                    Text(
                      _isEditing ? 'Edit Patient' : 'New Patient',
                      style: TextStyle(
                        fontSize: 24,
                        fontWeight: FontWeight.w700,
                        color: colorScheme.onSurface,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: AppSpacing.lg),

                // --- Error banner ---
                if (_errorMessage != null) ...[
                  Container(
                    padding: AppSpacing.cardPadding,
                    decoration: BoxDecoration(
                      color: colorScheme.errorContainer,
                      borderRadius:
                          BorderRadius.circular(AppSpacing.borderRadiusMd),
                    ),
                    child: Row(
                      children: [
                        Icon(Icons.error_outline,
                            color: colorScheme.onErrorContainer),
                        const SizedBox(width: AppSpacing.sm),
                        Expanded(
                          child: Text(
                            _errorMessage!,
                            style: TextStyle(
                                color: colorScheme.onErrorContainer),
                          ),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: AppSpacing.md),
                ],

                // --- Demographics card ---
                _SectionCard(
                  title: 'Demographics',
                  icon: Icons.person,
                  children: [
                    TextFormField(
                      controller: _familyNameCtrl,
                      decoration: const InputDecoration(
                        labelText: 'Family Name *',
                        border: OutlineInputBorder(),
                      ),
                      validator: (v) => (v == null || v.trim().isEmpty)
                          ? 'Family name is required'
                          : null,
                    ),
                    const SizedBox(height: AppSpacing.md),
                    TextFormField(
                      controller: _givenNamesCtrl,
                      decoration: const InputDecoration(
                        labelText: 'Given Names',
                        hintText: 'Comma-separated (e.g. John, Michael)',
                        border: OutlineInputBorder(),
                      ),
                    ),
                    const SizedBox(height: AppSpacing.md),
                    Row(
                      children: [
                        Expanded(
                          child: DropdownButtonFormField<String>(
                            value: _gender,
                            decoration: const InputDecoration(
                              labelText: 'Gender',
                              border: OutlineInputBorder(),
                            ),
                            items: const [
                              DropdownMenuItem(
                                  value: 'male', child: Text('Male')),
                              DropdownMenuItem(
                                  value: 'female', child: Text('Female')),
                              DropdownMenuItem(
                                  value: 'other', child: Text('Other')),
                              DropdownMenuItem(
                                  value: 'unknown', child: Text('Unknown')),
                            ],
                            onChanged: (v) =>
                                setState(() => _gender = v ?? 'unknown'),
                          ),
                        ),
                        const SizedBox(width: AppSpacing.md),
                        Expanded(
                          child: InkWell(
                            onTap: _pickDate,
                            child: InputDecorator(
                              decoration: const InputDecoration(
                                labelText: 'Date of Birth',
                                border: OutlineInputBorder(),
                                suffixIcon: Icon(Icons.calendar_today),
                              ),
                              child: Text(
                                _birthDate != null
                                    ? DateFormat('yyyy-MM-dd')
                                        .format(_birthDate!)
                                    : 'Select date',
                                style: TextStyle(
                                  color: _birthDate != null
                                      ? colorScheme.onSurface
                                      : colorScheme.onSurfaceVariant,
                                ),
                              ),
                            ),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: AppSpacing.md),
                    SwitchListTile(
                      title: const Text('Active'),
                      subtitle:
                          const Text('Whether this patient record is active'),
                      value: _active,
                      onChanged: (v) => setState(() => _active = v),
                      contentPadding: EdgeInsets.zero,
                    ),
                  ],
                ),
                const SizedBox(height: AppSpacing.md),

                // --- Identifier card ---
                _SectionCard(
                  title: 'Identifier',
                  icon: Icons.badge,
                  children: [
                    Row(
                      children: [
                        Expanded(
                          child: TextFormField(
                            controller: _identifierSystemCtrl,
                            decoration: const InputDecoration(
                              labelText: 'System',
                              hintText: 'e.g. urn:oid:1.2.3',
                              border: OutlineInputBorder(),
                            ),
                          ),
                        ),
                        const SizedBox(width: AppSpacing.md),
                        Expanded(
                          child: TextFormField(
                            controller: _identifierValueCtrl,
                            decoration: const InputDecoration(
                              labelText: 'Value',
                              hintText: 'e.g. MRN-12345',
                              border: OutlineInputBorder(),
                            ),
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
                const SizedBox(height: AppSpacing.md),

                // --- Contact card ---
                _SectionCard(
                  title: 'Contact',
                  icon: Icons.phone,
                  children: [
                    Row(
                      children: [
                        Expanded(
                          child: TextFormField(
                            controller: _phoneCtrl,
                            decoration: const InputDecoration(
                              labelText: 'Phone Number',
                              prefixIcon: Icon(Icons.phone_outlined),
                              border: OutlineInputBorder(),
                            ),
                            keyboardType: TextInputType.phone,
                          ),
                        ),
                        const SizedBox(width: AppSpacing.md),
                        Expanded(
                          child: TextFormField(
                            controller: _emailCtrl,
                            decoration: const InputDecoration(
                              labelText: 'Email',
                              prefixIcon: Icon(Icons.email_outlined),
                              border: OutlineInputBorder(),
                            ),
                            keyboardType: TextInputType.emailAddress,
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
                const SizedBox(height: AppSpacing.md),

                // --- Address card ---
                _SectionCard(
                  title: 'Address',
                  icon: Icons.location_on,
                  children: [
                    TextFormField(
                      controller: _addressLineCtrl,
                      decoration: const InputDecoration(
                        labelText: 'Address Line',
                        border: OutlineInputBorder(),
                      ),
                    ),
                    const SizedBox(height: AppSpacing.md),
                    Row(
                      children: [
                        Expanded(
                          child: TextFormField(
                            controller: _cityCtrl,
                            decoration: const InputDecoration(
                              labelText: 'City',
                              border: OutlineInputBorder(),
                            ),
                          ),
                        ),
                        const SizedBox(width: AppSpacing.md),
                        Expanded(
                          child: TextFormField(
                            controller: _stateCtrl,
                            decoration: const InputDecoration(
                              labelText: 'State / Province',
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
                          child: TextFormField(
                            controller: _postalCodeCtrl,
                            decoration: const InputDecoration(
                              labelText: 'Postal Code',
                              border: OutlineInputBorder(),
                            ),
                          ),
                        ),
                        const SizedBox(width: AppSpacing.md),
                        Expanded(
                          child: TextFormField(
                            controller: _countryCtrl,
                            decoration: const InputDecoration(
                              labelText: 'Country',
                              border: OutlineInputBorder(),
                            ),
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
                const SizedBox(height: AppSpacing.lg),

                // --- Action buttons ---
                Row(
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: [
                    OutlinedButton(
                      onPressed: _loading ? null : () => context.pop(),
                      child: const Text('Cancel'),
                    ),
                    const SizedBox(width: AppSpacing.md),
                    FilledButton.icon(
                      onPressed: _loading ? null : _onSave,
                      icon: _loading
                          ? const SizedBox(
                              width: 18,
                              height: 18,
                              child: CircularProgressIndicator(
                                strokeWidth: 2,
                                color: Colors.white,
                              ),
                            )
                          : const Icon(Icons.save),
                      label: Text(
                          _isEditing ? 'Update Patient' : 'Create Patient'),
                    ),
                  ],
                ),
                const SizedBox(height: AppSpacing.xl),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Section Card helper
// ---------------------------------------------------------------------------

class _SectionCard extends StatelessWidget {
  const _SectionCard({
    required this.title,
    required this.icon,
    required this.children,
  });

  final String title;
  final IconData icon;
  final List<Widget> children;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(AppSpacing.borderRadiusLg),
        side: BorderSide(color: colorScheme.outlineVariant),
      ),
      child: Padding(
        padding: AppSpacing.cardPadding,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(icon, size: 20, color: colorScheme.primary),
                const SizedBox(width: AppSpacing.sm),
                Text(
                  title,
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.w600,
                    color: colorScheme.onSurface,
                  ),
                ),
              ],
            ),
            const SizedBox(height: AppSpacing.md),
            ...children,
          ],
        ),
      ),
    );
  }
}
