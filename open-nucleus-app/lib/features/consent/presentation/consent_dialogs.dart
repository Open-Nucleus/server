import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/theme/app_spacing.dart';
import '../../../shared/models/consent_models.dart';
import '../../../shared/widgets/json_viewer.dart';
import 'consent_providers.dart';

// =============================================================================
// Grant Consent Dialog
// =============================================================================

/// A dialog form for granting consent to a performer on a patient's data.
///
/// On successful submission the dialog closes and returns the
/// [ConsentGrantResponse] to the caller.
class GrantConsentDialog extends ConsumerStatefulWidget {
  /// The patient for whom consent is being granted.
  final String patientId;

  const GrantConsentDialog({required this.patientId, super.key});

  /// Convenience method to show the dialog and return the result.
  static Future<ConsentGrantResponse?> show({
    required BuildContext context,
    required String patientId,
  }) {
    return showDialog<ConsentGrantResponse>(
      context: context,
      builder: (_) => GrantConsentDialog(patientId: patientId),
    );
  }

  @override
  ConsumerState<GrantConsentDialog> createState() =>
      _GrantConsentDialogState();
}

class _GrantConsentDialogState extends ConsumerState<GrantConsentDialog> {
  final _formKey = GlobalKey<FormState>();
  final _performerIdController = TextEditingController();
  final _categoryController = TextEditingController();

  String _scope = 'treatment';
  DateTime? _periodStart;
  DateTime? _periodEnd;
  bool _submitting = false;
  String? _errorMsg;

  static const _scopeOptions = [
    'treatment',
    'research',
    'emergency',
    'operations',
  ];

  @override
  void dispose() {
    _performerIdController.dispose();
    _categoryController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return AlertDialog(
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(AppSpacing.borderRadiusLg),
      ),
      title: const Text('Grant Consent'),
      content: SizedBox(
        width: 440,
        child: Form(
          key: _formKey,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              // ── Performer ID ─────────────────────────────────────────
              TextFormField(
                controller: _performerIdController,
                decoration: const InputDecoration(
                  labelText: 'Performer ID *',
                  hintText: 'Practitioner or device ID',
                  border: OutlineInputBorder(),
                ),
                validator: (v) =>
                    (v == null || v.trim().isEmpty) ? 'Required' : null,
              ),
              const SizedBox(height: AppSpacing.md),

              // ── Scope ────────────────────────────────────────────────
              DropdownButtonFormField<String>(
                value: _scope,
                decoration: const InputDecoration(
                  labelText: 'Scope *',
                  border: OutlineInputBorder(),
                ),
                items: _scopeOptions
                    .map((s) => DropdownMenuItem(
                          value: s,
                          child: Text(s[0].toUpperCase() + s.substring(1)),
                        ))
                    .toList(),
                onChanged: (v) {
                  if (v != null) setState(() => _scope = v);
                },
              ),
              const SizedBox(height: AppSpacing.md),

              // ── Period Start ─────────────────────────────────────────
              _DatePickerField(
                label: 'Period Start *',
                value: _periodStart,
                onPicked: (d) => setState(() => _periodStart = d),
              ),
              const SizedBox(height: AppSpacing.md),

              // ── Period End ───────────────────────────────────────────
              _DatePickerField(
                label: 'Period End *',
                value: _periodEnd,
                onPicked: (d) => setState(() => _periodEnd = d),
              ),
              const SizedBox(height: AppSpacing.md),

              // ── Category ─────────────────────────────────────────────
              TextFormField(
                controller: _categoryController,
                decoration: const InputDecoration(
                  labelText: 'Category (optional)',
                  hintText: 'e.g. clinical-data',
                  border: OutlineInputBorder(),
                ),
              ),

              if (_errorMsg != null) ...[
                const SizedBox(height: AppSpacing.md),
                Text(
                  _errorMsg!,
                  style: TextStyle(color: colorScheme.error, fontSize: 13),
                ),
              ],
            ],
          ),
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text('Cancel'),
        ),
        FilledButton(
          onPressed: _submitting ? null : _submit,
          child: _submitting
              ? const SizedBox(
                  width: 18,
                  height: 18,
                  child: CircularProgressIndicator(strokeWidth: 2),
                )
              : const Text('Grant'),
        ),
      ],
    );
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    if (_periodStart == null || _periodEnd == null) {
      setState(() => _errorMsg = 'Please select both start and end dates.');
      return;
    }

    setState(() {
      _submitting = true;
      _errorMsg = null;
    });

    try {
      final api = ref.read(consentApiProvider);
      final envelope = await api.grantConsent(
        patientId: widget.patientId,
        performerId: _performerIdController.text.trim(),
        scope: _scope,
        periodStart: _periodStart!.toIso8601String(),
        periodEnd: _periodEnd!.toIso8601String(),
        category: _categoryController.text.trim().isNotEmpty
            ? _categoryController.text.trim()
            : null,
      );

      // Invalidate the consent list so the caller sees fresh data.
      ref.invalidate(patientConsentsProvider(widget.patientId));

      if (mounted) {
        Navigator.of(context).pop(envelope.data);
      }
    } catch (e) {
      setState(() {
        _submitting = false;
        _errorMsg = e.toString();
      });
    }
  }
}

// =============================================================================
// Consent VC Dialog
// =============================================================================

/// A dialog that displays a consent's Verifiable Credential, or issues one if
/// not yet present.
class ConsentVCDialog extends ConsumerStatefulWidget {
  final String consentId;

  const ConsentVCDialog({required this.consentId, super.key});

  /// Convenience method to show the dialog.
  static Future<void> show({
    required BuildContext context,
    required String consentId,
  }) {
    return showDialog<void>(
      context: context,
      builder: (_) => ConsentVCDialog(consentId: consentId),
    );
  }

  @override
  ConsumerState<ConsentVCDialog> createState() => _ConsentVCDialogState();
}

class _ConsentVCDialogState extends ConsumerState<ConsentVCDialog> {
  ConsentVCResponse? _vcResponse;
  bool _issuing = false;
  String? _errorMsg;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return AlertDialog(
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(AppSpacing.borderRadiusLg),
      ),
      title: const Text('Consent Verifiable Credential'),
      content: SizedBox(
        width: 520,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              'Consent ID: ${widget.consentId}',
              style: const TextStyle(
                fontFamily: 'monospace',
                fontSize: 13,
              ),
            ),
            const SizedBox(height: AppSpacing.md),

            if (_vcResponse != null) ...[
              const Text(
                'Verifiable Credential',
                style: TextStyle(fontWeight: FontWeight.w600, fontSize: 14),
              ),
              const SizedBox(height: AppSpacing.sm),
              SizedBox(
                height: 300,
                child: SingleChildScrollView(
                  child: JsonViewer(
                      data: _vcResponse!.verifiableCredential),
                ),
              ),
            ] else ...[
              Text(
                'No VC issued yet. Click "Issue VC" to create one.',
                style: TextStyle(
                  color: colorScheme.onSurfaceVariant,
                  fontSize: 13,
                ),
              ),
            ],

            if (_errorMsg != null) ...[
              const SizedBox(height: AppSpacing.md),
              Text(
                _errorMsg!,
                style: TextStyle(color: colorScheme.error, fontSize: 13),
              ),
            ],
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text('Close'),
        ),
        FilledButton.icon(
          onPressed: _issuing ? null : _issueVC,
          icon: _issuing
              ? const SizedBox(
                  width: 18,
                  height: 18,
                  child: CircularProgressIndicator(strokeWidth: 2),
                )
              : const Icon(Icons.verified, size: 18),
          label: const Text('Issue VC'),
        ),
      ],
    );
  }

  Future<void> _issueVC() async {
    setState(() {
      _issuing = true;
      _errorMsg = null;
    });

    try {
      final api = ref.read(consentApiProvider);
      final envelope = await api.issueVC(widget.consentId);
      setState(() {
        _vcResponse = envelope.data;
        _issuing = false;
      });
    } catch (e) {
      setState(() {
        _issuing = false;
        _errorMsg = e.toString();
      });
    }
  }
}

// =============================================================================
// Internal Widgets
// =============================================================================

/// A read-only text field that opens a date picker when tapped.
class _DatePickerField extends StatelessWidget {
  final String label;
  final DateTime? value;
  final ValueChanged<DateTime> onPicked;

  const _DatePickerField({
    required this.label,
    required this.value,
    required this.onPicked,
  });

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: () async {
        final picked = await showDatePicker(
          context: context,
          initialDate: value ?? DateTime.now(),
          firstDate: DateTime(2020),
          lastDate: DateTime(2040),
        );
        if (picked != null) onPicked(picked);
      },
      child: InputDecorator(
        decoration: InputDecoration(
          labelText: label,
          border: const OutlineInputBorder(),
          suffixIcon: const Icon(Icons.calendar_today, size: 18),
        ),
        child: Text(
          value != null
              ? '${value!.year}-${value!.month.toString().padLeft(2, '0')}-${value!.day.toString().padLeft(2, '0')}'
              : 'Select date',
          style: TextStyle(
            color: value != null
                ? Theme.of(context).colorScheme.onSurface
                : Theme.of(context).colorScheme.onSurfaceVariant,
          ),
        ),
      ),
    );
  }
}
