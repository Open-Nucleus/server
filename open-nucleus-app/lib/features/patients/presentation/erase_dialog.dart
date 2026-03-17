import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/theme/app_spacing.dart';
import '../../../shared/models/app_exception.dart';
import '../../../shared/providers/dio_provider.dart';
import '../data/patient_api.dart';

// ---------------------------------------------------------------------------
// Provider
// ---------------------------------------------------------------------------

final _patientApiProvider = Provider<PatientApi>((ref) {
  return PatientApi(ref.watch(dioProvider));
});

// ---------------------------------------------------------------------------
// Erase Dialog
// ---------------------------------------------------------------------------

/// Confirmation dialog for crypto-erasure of a patient record.
///
/// Displays a warning explaining the irreversible nature of the operation
/// and requires the user to type "DELETE" before enabling the erase button.
/// On success, navigates back to the patient list with a success message.
class EraseDialog extends ConsumerStatefulWidget {
  const EraseDialog({
    required this.patientId,
    required this.patientName,
    super.key,
  });

  final String patientId;
  final String patientName;

  /// Shows the erase dialog and returns `true` if the patient was erased.
  static Future<bool> show(
    BuildContext context, {
    required String patientId,
    required String patientName,
  }) async {
    final result = await showDialog<bool>(
      context: context,
      barrierDismissible: false,
      builder: (_) => EraseDialog(
        patientId: patientId,
        patientName: patientName,
      ),
    );
    return result ?? false;
  }

  @override
  ConsumerState<EraseDialog> createState() => _EraseDialogState();
}

class _EraseDialogState extends ConsumerState<EraseDialog> {
  final _confirmCtrl = TextEditingController();
  bool _erasing = false;
  String? _errorMessage;

  bool get _canErase =>
      _confirmCtrl.text.trim().toUpperCase() == 'DELETE' && !_erasing;

  @override
  void dispose() {
    _confirmCtrl.dispose();
    super.dispose();
  }

  Future<void> _onErase() async {
    if (!_canErase) return;

    setState(() {
      _erasing = true;
      _errorMessage = null;
    });

    try {
      final api = ref.read(_patientApiProvider);
      final envelope = await api.erasePatient(widget.patientId);

      if (!mounted) return;

      if (envelope.isSuccess) {
        Navigator.of(context).pop(true);
        // Navigate to patient list and show success
        context.go('/patients');
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              'Patient "${widget.patientName}" has been crypto-erased.',
            ),
            backgroundColor: Colors.green,
          ),
        );
      } else {
        setState(() =>
            _errorMessage = envelope.error?.message ?? 'Erase failed');
      }
    } on DioException catch (e) {
      if (mounted) {
        final appErr = e.error;
        setState(() => _errorMessage = appErr is AppException
            ? appErr.message
            : (e.message ?? 'Network error'));
      }
    } catch (e) {
      if (mounted) {
        setState(() => _errorMessage = 'Unexpected error: $e');
      }
    } finally {
      if (mounted) setState(() => _erasing = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return AlertDialog(
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(AppSpacing.borderRadiusLg),
      ),
      title: Row(
        children: [
          Icon(Icons.warning_amber_rounded, color: colorScheme.error, size: 28),
          const SizedBox(width: AppSpacing.sm),
          Text(
            'Crypto-Erase Patient',
            style: TextStyle(color: colorScheme.error),
          ),
        ],
      ),
      content: SizedBox(
        width: 480,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Warning text
            Container(
              padding: AppSpacing.cardPadding,
              decoration: BoxDecoration(
                color: colorScheme.errorContainer,
                borderRadius:
                    BorderRadius.circular(AppSpacing.borderRadiusMd),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'THIS ACTION IS IRREVERSIBLE',
                    style: TextStyle(
                      fontWeight: FontWeight.w700,
                      color: colorScheme.onErrorContainer,
                    ),
                  ),
                  const SizedBox(height: AppSpacing.sm),
                  Text(
                    'Crypto-erasure permanently destroys the encryption key '
                    'for this patient. All encrypted FHIR resources in Git will '
                    'become permanently unreadable. The SQLite search index '
                    'entries will also be purged.',
                    style: TextStyle(
                      color: colorScheme.onErrorContainer,
                      fontSize: 13,
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: AppSpacing.lg),

            // Patient name
            Text.rich(
              TextSpan(
                text: 'Patient: ',
                style: TextStyle(color: colorScheme.onSurfaceVariant),
                children: [
                  TextSpan(
                    text: widget.patientName,
                    style: TextStyle(
                      fontWeight: FontWeight.w600,
                      color: colorScheme.onSurface,
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: AppSpacing.sm),
            Text(
              'ID: ${widget.patientId}',
              style: TextStyle(
                fontFamily: 'monospace',
                fontSize: 12,
                color: colorScheme.onSurfaceVariant,
              ),
            ),
            const SizedBox(height: AppSpacing.lg),

            // Confirmation field
            Text(
              'Type DELETE to confirm:',
              style: TextStyle(
                fontWeight: FontWeight.w600,
                color: colorScheme.onSurface,
              ),
            ),
            const SizedBox(height: AppSpacing.sm),
            TextField(
              controller: _confirmCtrl,
              onChanged: (_) => setState(() {}),
              decoration: InputDecoration(
                border: const OutlineInputBorder(),
                hintText: 'DELETE',
                hintStyle: TextStyle(color: colorScheme.onSurfaceVariant),
              ),
              autofocus: true,
            ),

            // Error message
            if (_errorMessage != null) ...[
              const SizedBox(height: AppSpacing.md),
              Text(
                _errorMessage!,
                style: TextStyle(color: colorScheme.error, fontSize: 13),
              ),
            ],
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: _erasing ? null : () => Navigator.of(context).pop(false),
          child: const Text('Cancel'),
        ),
        FilledButton(
          onPressed: _canErase ? _onErase : null,
          style: FilledButton.styleFrom(
            backgroundColor: colorScheme.error,
            foregroundColor: colorScheme.onError,
            disabledBackgroundColor: colorScheme.error.withOpacity(0.3),
          ),
          child: _erasing
              ? const SizedBox(
                  width: 18,
                  height: 18,
                  child: CircularProgressIndicator(
                    strokeWidth: 2,
                    color: Colors.white,
                  ),
                )
              : const Text('Erase Patient'),
        ),
      ],
    );
  }
}
