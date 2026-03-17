import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_spacing.dart';
import '../../../shared/models/formulary_models.dart';
import '../../../shared/widgets/search_field.dart';
import '../../../shared/widgets/severity_badge.dart';
import 'formulary_providers.dart';

// ═══════════════════════════════════════════════════════════════════════════════
// Formulary Screen — 3-pane layout
// ═══════════════════════════════════════════════════════════════════════════════

/// Full formulary screen with left (search), center (detail/interactions),
/// and right (stock) panes.
class FormularyScreen extends ConsumerStatefulWidget {
  const FormularyScreen({super.key});

  @override
  ConsumerState<FormularyScreen> createState() => _FormularyScreenState();
}

class _FormularyScreenState extends ConsumerState<FormularyScreen> {
  bool _showInteractionChecker = false;

  @override
  void initState() {
    super.initState();
    // Trigger initial search on load.
    Future.microtask(() {
      ref.read(medicationSearchProvider.notifier).search();
    });
  }

  @override
  Widget build(BuildContext context) {
    final selectedMed = ref.watch(selectedMedicationProvider);

    return Padding(
      padding: AppSpacing.pagePadding,
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // --- Left pane: Search + Results (300px) ---
          SizedBox(
            width: 300,
            child: _LeftPane(
              onMedicationSelected: (med) {
                ref.read(selectedMedicationProvider.notifier).state = med;
                setState(() => _showInteractionChecker = false);
              },
            ),
          ),
          const SizedBox(width: AppSpacing.md),

          // --- Center pane: Detail or Interaction Checker (fills) ---
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                // Toggle bar
                Row(
                  children: [
                    SegmentedButton<bool>(
                      segments: const [
                        ButtonSegment(
                          value: false,
                          icon: Icon(Icons.info_outline),
                          label: Text('Medication Detail'),
                        ),
                        ButtonSegment(
                          value: true,
                          icon: Icon(Icons.compare_arrows),
                          label: Text('Interaction Checker'),
                        ),
                      ],
                      selected: {_showInteractionChecker},
                      onSelectionChanged: (v) {
                        setState(() => _showInteractionChecker = v.first);
                      },
                    ),
                  ],
                ),
                const SizedBox(height: AppSpacing.md),

                Expanded(
                  child: _showInteractionChecker
                      ? const _InteractionCheckerPane()
                      : selectedMed != null
                          ? _MedicationDetailPane(medication: selectedMed)
                          : const _EmptyCenterPane(),
                ),
              ],
            ),
          ),
          const SizedBox(width: AppSpacing.md),

          // --- Right pane: Stock info (250px) ---
          SizedBox(
            width: 250,
            child: selectedMed != null
                ? _StockPane(medicationCode: selectedMed.code)
                : const _EmptyStockPane(),
          ),
        ],
      ),
    );
  }
}

// ═══════════════════════════════════════════════════════════════════════════════
// Left Pane — Search + Results
// ═══════════════════════════════════════════════════════════════════════════════

class _LeftPane extends ConsumerWidget {
  const _LeftPane({required this.onMedicationSelected});

  final ValueChanged<MedicationDetail> onMedicationSelected;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final searchState = ref.watch(medicationSearchProvider);
    final formularyInfo = ref.watch(formularyInfoProvider);
    final colorScheme = Theme.of(context).colorScheme;

    // Extract categories from formulary info.
    final categories = formularyInfo.whenOrNull(data: (info) => info?.categories)
        ?? <String>[];

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        // Search field
        SearchField(
          hintText: 'Search medications...',
          onChanged: (query) {
            ref.read(medicationSearchProvider.notifier).search(query: query);
          },
        ),
        const SizedBox(height: AppSpacing.sm),

        // Category filter
        if (categories.isNotEmpty)
          DropdownButtonFormField<String?>(
            value: searchState.category,
            decoration: const InputDecoration(
              labelText: 'Category',
              border: OutlineInputBorder(),
              isDense: true,
              contentPadding:
                  EdgeInsets.symmetric(horizontal: 12, vertical: 10),
            ),
            items: [
              const DropdownMenuItem<String?>(
                value: null,
                child: Text('All Categories'),
              ),
              ...categories.map((c) => DropdownMenuItem<String?>(
                    value: c,
                    child: Text(c, overflow: TextOverflow.ellipsis),
                  )),
            ],
            onChanged: (v) {
              ref
                  .read(medicationSearchProvider.notifier)
                  .search(category: () => v);
            },
          ),
        const SizedBox(height: AppSpacing.sm),

        // Results
        Expanded(
          child: searchState.loading
              ? const Center(child: CircularProgressIndicator())
              : searchState.results.isEmpty
                  ? Center(
                      child: Text(
                        searchState.errorMessage ?? 'No medications found',
                        style: TextStyle(color: colorScheme.onSurfaceVariant),
                        textAlign: TextAlign.center,
                      ),
                    )
                  : ListView.builder(
                      itemCount: searchState.results.length,
                      itemBuilder: (context, index) {
                        final med = searchState.results[index];
                        return _MedicationCard(
                          medication: med,
                          onTap: () => onMedicationSelected(med),
                        );
                      },
                    ),
        ),

        // Pagination
        if (searchState.totalPages > 1)
          Padding(
            padding: const EdgeInsets.only(top: AppSpacing.sm),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                IconButton(
                  icon: const Icon(Icons.chevron_left),
                  onPressed: searchState.page > 1
                      ? () => ref
                          .read(medicationSearchProvider.notifier)
                          .loadPage(searchState.page - 1)
                      : null,
                  iconSize: 20,
                ),
                Text(
                  '${searchState.page} / ${searchState.totalPages}',
                  style: TextStyle(
                    fontSize: 12,
                    color: colorScheme.onSurfaceVariant,
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.chevron_right),
                  onPressed: searchState.page < searchState.totalPages
                      ? () => ref
                          .read(medicationSearchProvider.notifier)
                          .loadPage(searchState.page + 1)
                      : null,
                  iconSize: 20,
                ),
              ],
            ),
          ),
      ],
    );
  }
}

// ── Medication Card ──────────────────────────────────────────────────────────

class _MedicationCard extends StatelessWidget {
  const _MedicationCard({
    required this.medication,
    required this.onTap,
  });

  final MedicationDetail medication;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Card(
      elevation: 0,
      margin: const EdgeInsets.only(bottom: AppSpacing.xs),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(AppSpacing.borderRadiusMd),
        side: BorderSide(color: colorScheme.outlineVariant),
      ),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(AppSpacing.borderRadiusMd),
        child: Padding(
          padding: const EdgeInsets.all(AppSpacing.sm),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                medication.display,
                style: TextStyle(
                  fontWeight: FontWeight.w600,
                  fontSize: 13,
                  color: colorScheme.onSurface,
                ),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
              const SizedBox(height: 2),
              Text(
                '${medication.code}  |  ${medication.form}  |  ${medication.route}',
                style: TextStyle(
                  fontSize: 11,
                  color: colorScheme.onSurfaceVariant,
                ),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
              const SizedBox(height: 4),
              Wrap(
                spacing: 4,
                children: [
                  if (medication.whoEssential)
                    _SmallBadge(
                      label: 'WHO Essential',
                      color: AppColors.severityInfo,
                    ),
                  _SmallBadge(
                    label: medication.available ? 'Available' : 'Unavailable',
                    color: medication.available
                        ? AppColors.statusActive
                        : AppColors.statusError,
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _SmallBadge extends StatelessWidget {
  const _SmallBadge({required this.label, required this.color});

  final String label;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      decoration: BoxDecoration(
        color: color.withOpacity(0.12),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: color.withOpacity(0.35)),
      ),
      child: Text(
        label,
        style: TextStyle(fontSize: 10, fontWeight: FontWeight.w600, color: color),
      ),
    );
  }
}

// ═══════════════════════════════════════════════════════════════════════════════
// Center Pane — Medication Detail
// ═══════════════════════════════════════════════════════════════════════════════

class _EmptyCenterPane extends StatelessWidget {
  const _EmptyCenterPane();

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.medication_outlined, size: 48, color: colorScheme.outlineVariant),
          const SizedBox(height: AppSpacing.sm),
          Text(
            'Select a medication to view details',
            style: TextStyle(color: colorScheme.onSurfaceVariant),
          ),
        ],
      ),
    );
  }
}

class _MedicationDetailPane extends StatelessWidget {
  const _MedicationDetailPane({required this.medication});

  final MedicationDetail medication;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return SingleChildScrollView(
      child: Card(
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
              // Header
              Row(
                children: [
                  Icon(Icons.medication, color: colorScheme.primary),
                  const SizedBox(width: AppSpacing.sm),
                  Expanded(
                    child: Text(
                      medication.display,
                      style: TextStyle(
                        fontSize: 20,
                        fontWeight: FontWeight.w700,
                        color: colorScheme.onSurface,
                      ),
                    ),
                  ),
                  if (medication.whoEssential)
                    Tooltip(
                      message: 'WHO Essential Medicine',
                      child: Container(
                        padding: const EdgeInsets.symmetric(
                            horizontal: 8, vertical: 4),
                        decoration: BoxDecoration(
                          color: AppColors.severityInfo.withOpacity(0.12),
                          borderRadius: BorderRadius.circular(12),
                          border: Border.all(
                              color: AppColors.severityInfo.withOpacity(0.4)),
                        ),
                        child: const Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Icon(Icons.verified, size: 16,
                                color: AppColors.severityInfo),
                            SizedBox(width: 4),
                            Text(
                              'WHO Essential',
                              style: TextStyle(
                                fontSize: 12,
                                fontWeight: FontWeight.w600,
                                color: AppColors.severityInfo,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                ],
              ),
              const SizedBox(height: AppSpacing.lg),

              // Details grid
              _DetailRow(label: 'Code', value: medication.code),
              _DetailRow(label: 'Form', value: medication.form),
              _DetailRow(label: 'Route', value: medication.route),
              _DetailRow(label: 'Category', value: medication.category),
              _DetailRow(
                  label: 'Therapeutic Class',
                  value: medication.therapeuticClass),
              if (medication.strength != null)
                _DetailRow(label: 'Strength', value: medication.strength!),
              _DetailRow(
                label: 'Available',
                value: medication.available ? 'Yes' : 'No',
                valueColor: medication.available
                    ? AppColors.statusActive
                    : AppColors.statusError,
              ),

              // Common frequencies
              if (medication.commonFrequencies != null &&
                  medication.commonFrequencies!.isNotEmpty) ...[
                const SizedBox(height: AppSpacing.lg),
                Text(
                  'Common Frequencies',
                  style: TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.w600,
                    color: colorScheme.onSurface,
                  ),
                ),
                const SizedBox(height: AppSpacing.sm),
                Wrap(
                  spacing: AppSpacing.sm,
                  runSpacing: AppSpacing.xs,
                  children: medication.commonFrequencies!
                      .map((f) => Chip(
                            label: Text(f, style: const TextStyle(fontSize: 12)),
                            visualDensity: VisualDensity.compact,
                          ))
                      .toList(),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}

class _DetailRow extends StatelessWidget {
  const _DetailRow({
    required this.label,
    required this.value,
    this.valueColor,
  });

  final String label;
  final String value;
  final Color? valueColor;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Padding(
      padding: const EdgeInsets.only(bottom: AppSpacing.sm),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 140,
            child: Text(
              label,
              style: TextStyle(
                fontSize: 13,
                color: colorScheme.onSurfaceVariant,
              ),
            ),
          ),
          Expanded(
            child: Text(
              value,
              style: TextStyle(
                fontSize: 13,
                fontWeight: FontWeight.w500,
                color: valueColor ?? colorScheme.onSurface,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

// ═══════════════════════════════════════════════════════════════════════════════
// Center Pane — Interaction Checker
// ═══════════════════════════════════════════════════════════════════════════════

class _InteractionCheckerPane extends ConsumerStatefulWidget {
  const _InteractionCheckerPane();

  @override
  ConsumerState<_InteractionCheckerPane> createState() =>
      _InteractionCheckerPaneState();
}

class _InteractionCheckerPaneState
    extends ConsumerState<_InteractionCheckerPane> {
  final _searchCtrl = TextEditingController();
  List<MedicationDetail> _searchResults = [];
  bool _searching = false;

  @override
  void dispose() {
    _searchCtrl.dispose();
    super.dispose();
  }

  Future<void> _searchMedication(String query) async {
    if (query.length < 2) {
      setState(() => _searchResults = []);
      return;
    }

    setState(() => _searching = true);
    try {
      final api = ref.read(formularyApiProvider);
      final envelope = await api.searchMedications(query: query, perPage: 10);
      if (mounted && envelope.isSuccess && envelope.data != null) {
        setState(() => _searchResults = envelope.data!.medications);
      }
    } catch (_) {
      // Silently fail search suggestions.
    } finally {
      if (mounted) setState(() => _searching = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final checkerState = ref.watch(interactionCheckerProvider);

    return SingleChildScrollView(
      child: Card(
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
              Text(
                'Drug Interaction Checker',
                style: TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.w700,
                  color: colorScheme.onSurface,
                ),
              ),
              const SizedBox(height: AppSpacing.md),

              // Search + Add medication
              Row(
                children: [
                  Expanded(
                    child: TextField(
                      controller: _searchCtrl,
                      decoration: const InputDecoration(
                        labelText: 'Search medication to add',
                        border: OutlineInputBorder(),
                        prefixIcon: Icon(Icons.search),
                        isDense: true,
                      ),
                      onChanged: _searchMedication,
                    ),
                  ),
                ],
              ),

              // Search suggestions
              if (_searchResults.isNotEmpty || _searching)
                Container(
                  constraints: const BoxConstraints(maxHeight: 200),
                  margin: const EdgeInsets.only(top: 4),
                  decoration: BoxDecoration(
                    border: Border.all(color: colorScheme.outlineVariant),
                    borderRadius:
                        BorderRadius.circular(AppSpacing.borderRadiusMd),
                  ),
                  child: _searching
                      ? const Padding(
                          padding: EdgeInsets.all(AppSpacing.md),
                          child: Center(
                              child: SizedBox(
                            width: 20,
                            height: 20,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )),
                        )
                      : ListView.builder(
                          shrinkWrap: true,
                          itemCount: _searchResults.length,
                          itemBuilder: (context, i) {
                            final med = _searchResults[i];
                            return ListTile(
                              dense: true,
                              title: Text(med.display,
                                  style: const TextStyle(fontSize: 13)),
                              subtitle: Text(med.code,
                                  style: const TextStyle(fontSize: 11)),
                              onTap: () {
                                ref
                                    .read(interactionCheckerProvider.notifier)
                                    .addMedication(med);
                                _searchCtrl.clear();
                                setState(() => _searchResults = []);
                              },
                            );
                          },
                        ),
                ),
              const SizedBox(height: AppSpacing.md),

              // Selected medication chips
              Text(
                'Selected Medications (${checkerState.selectedMedications.length})',
                style: TextStyle(
                  fontWeight: FontWeight.w600,
                  color: colorScheme.onSurface,
                ),
              ),
              const SizedBox(height: AppSpacing.sm),
              Wrap(
                spacing: AppSpacing.sm,
                runSpacing: AppSpacing.xs,
                children: checkerState.selectedMedications
                    .map((med) => Chip(
                          label: Text(med.display,
                              style: const TextStyle(fontSize: 12)),
                          deleteIcon: const Icon(Icons.close, size: 16),
                          onDeleted: () => ref
                              .read(interactionCheckerProvider.notifier)
                              .removeMedication(med.code),
                        ))
                    .toList(),
              ),

              if (checkerState.selectedMedications.isEmpty)
                Padding(
                  padding: const EdgeInsets.symmetric(vertical: AppSpacing.md),
                  child: Text(
                    'Add at least 2 medications to check interactions',
                    style: TextStyle(
                      color: colorScheme.onSurfaceVariant,
                      fontSize: 13,
                    ),
                  ),
                ),

              const SizedBox(height: AppSpacing.md),

              // Check button
              Row(
                children: [
                  FilledButton.icon(
                    onPressed: checkerState.selectedMedications.length >= 2 &&
                            !checkerState.loading
                        ? () => ref
                            .read(interactionCheckerProvider.notifier)
                            .checkInteractions()
                        : null,
                    icon: checkerState.loading
                        ? const SizedBox(
                            width: 16,
                            height: 16,
                            child: CircularProgressIndicator(
                                strokeWidth: 2, color: Colors.white),
                          )
                        : const Icon(Icons.compare_arrows),
                    label: const Text('Check Interactions'),
                  ),
                  const SizedBox(width: AppSpacing.sm),
                  if (checkerState.selectedMedications.isNotEmpty)
                    OutlinedButton(
                      onPressed: () =>
                          ref.read(interactionCheckerProvider.notifier).clear(),
                      child: const Text('Clear All'),
                    ),
                ],
              ),

              // Error
              if (checkerState.errorMessage != null) ...[
                const SizedBox(height: AppSpacing.md),
                Text(
                  checkerState.errorMessage!,
                  style: TextStyle(color: colorScheme.error, fontSize: 13),
                ),
              ],

              // Results
              if (checkerState.result != null) ...[
                const SizedBox(height: AppSpacing.lg),
                _InteractionResults(result: checkerState.result!),
              ],
            ],
          ),
        ),
      ),
    );
  }
}

class _InteractionResults extends StatelessWidget {
  const _InteractionResults({required this.result});

  final CheckInteractionsResponse result;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Overall risk
        Container(
          padding: AppSpacing.cardPadding,
          decoration: BoxDecoration(
            color: _riskColor(result.overallRisk).withOpacity(0.1),
            borderRadius: BorderRadius.circular(AppSpacing.borderRadiusMd),
            border:
                Border.all(color: _riskColor(result.overallRisk).withOpacity(0.4)),
          ),
          child: Row(
            children: [
              Icon(
                _riskIcon(result.overallRisk),
                color: _riskColor(result.overallRisk),
              ),
              const SizedBox(width: AppSpacing.sm),
              Text(
                'Overall Risk: ${result.overallRisk.toUpperCase()}',
                style: TextStyle(
                  fontWeight: FontWeight.w700,
                  color: _riskColor(result.overallRisk),
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: AppSpacing.md),

        // Interactions list
        if (result.interactions.isNotEmpty) ...[
          Text(
            'Interactions (${result.interactions.length})',
            style: TextStyle(
              fontWeight: FontWeight.w600,
              color: colorScheme.onSurface,
            ),
          ),
          const SizedBox(height: AppSpacing.sm),
          ...result.interactions.map((i) => _InteractionTile(interaction: i)),
        ] else
          Text(
            'No drug-drug interactions found.',
            style: TextStyle(
              color: colorScheme.onSurfaceVariant,
              fontSize: 13,
            ),
          ),

        // Allergy alerts
        if (result.allergyAlerts != null &&
            result.allergyAlerts!.isNotEmpty) ...[
          const SizedBox(height: AppSpacing.lg),
          Text(
            'Allergy Alerts (${result.allergyAlerts!.length})',
            style: TextStyle(
              fontWeight: FontWeight.w600,
              color: colorScheme.error,
            ),
          ),
          const SizedBox(height: AppSpacing.sm),
          ...result.allergyAlerts!.map((a) => ListTile(
                dense: true,
                leading: Icon(Icons.warning, color: colorScheme.error, size: 20),
                title: Text(a.description, style: const TextStyle(fontSize: 13)),
                subtitle: Text(
                  'Medication: ${a.medicationCode} | Allergy: ${a.allergyCode}',
                  style: const TextStyle(fontSize: 11),
                ),
                trailing: SeverityBadge(severity: a.severity),
              )),
        ],
      ],
    );
  }

  static Color _riskColor(String risk) {
    switch (risk.toLowerCase()) {
      case 'critical':
        return AppColors.severityCritical;
      case 'high':
        return AppColors.severityHigh;
      case 'moderate':
        return AppColors.severityWarning;
      case 'low':
        return AppColors.severityLow;
      default:
        return AppColors.severityInfo;
    }
  }

  static IconData _riskIcon(String risk) {
    switch (risk.toLowerCase()) {
      case 'critical':
      case 'high':
        return Icons.error;
      case 'moderate':
        return Icons.warning_amber;
      default:
        return Icons.check_circle;
    }
  }
}

class _InteractionTile extends StatelessWidget {
  const _InteractionTile({required this.interaction});

  final InteractionDetail interaction;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Card(
      elevation: 0,
      margin: const EdgeInsets.only(bottom: AppSpacing.sm),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(AppSpacing.borderRadiusMd),
        side: BorderSide(color: colorScheme.outlineVariant),
      ),
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.sm),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                SeverityBadge(severity: interaction.severity),
                const SizedBox(width: AppSpacing.sm),
                Expanded(
                  child: Text(
                    '${interaction.medicationA} + ${interaction.medicationB}',
                    style: TextStyle(
                      fontWeight: FontWeight.w600,
                      fontSize: 13,
                      color: colorScheme.onSurface,
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 4),
            Text(
              interaction.description,
              style: TextStyle(
                fontSize: 12,
                color: colorScheme.onSurfaceVariant,
              ),
            ),
            if (interaction.clinicalEffect != null) ...[
              const SizedBox(height: 2),
              Text(
                'Effect: ${interaction.clinicalEffect}',
                style: const TextStyle(fontSize: 11),
              ),
            ],
            if (interaction.recommendation != null) ...[
              const SizedBox(height: 2),
              Text(
                'Recommendation: ${interaction.recommendation}',
                style: TextStyle(
                  fontSize: 11,
                  fontWeight: FontWeight.w500,
                  color: colorScheme.primary,
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

// ═══════════════════════════════════════════════════════════════════════════════
// Right Pane — Stock Info
// ═══════════════════════════════════════════════════════════════════════════════

class _EmptyStockPane extends StatelessWidget {
  const _EmptyStockPane();

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.inventory_2_outlined,
              size: 36, color: colorScheme.outlineVariant),
          const SizedBox(height: AppSpacing.sm),
          Text(
            'Stock info will appear\nwhen a medication\nis selected',
            style: TextStyle(
              color: colorScheme.onSurfaceVariant,
              fontSize: 12,
            ),
            textAlign: TextAlign.center,
          ),
        ],
      ),
    );
  }
}

class _StockPane extends ConsumerWidget {
  const _StockPane({required this.medicationCode});

  final String medicationCode;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    // Use a default siteId — in a real app this would come from the auth context.
    final stockAsync = ref.watch(stockInfoProvider(
      (siteId: 'default', medicationCode: medicationCode),
    ));
    final colorScheme = Theme.of(context).colorScheme;

    return SingleChildScrollView(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          // Stock header
          Text(
            'Stock Info',
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.w700,
              color: colorScheme.onSurface,
            ),
          ),
          const SizedBox(height: AppSpacing.md),

          stockAsync.when(
            loading: () =>
                const Center(child: CircularProgressIndicator()),
            error: (err, _) => Card(
              elevation: 0,
              child: Padding(
                padding: AppSpacing.cardPadding,
                child: Text(
                  'Unable to load stock data',
                  style: TextStyle(
                    color: colorScheme.onSurfaceVariant,
                    fontSize: 12,
                  ),
                ),
              ),
            ),
            data: (data) {
              final level = data.level;
              final prediction = data.prediction;

              if (level == null && prediction == null) {
                return Card(
                  elevation: 0,
                  shape: RoundedRectangleBorder(
                    borderRadius:
                        BorderRadius.circular(AppSpacing.borderRadiusMd),
                    side: BorderSide(color: colorScheme.outlineVariant),
                  ),
                  child: Padding(
                    padding: AppSpacing.cardPadding,
                    child: Text(
                      'No stock data available for this medication.',
                      style: TextStyle(
                        color: colorScheme.onSurfaceVariant,
                        fontSize: 12,
                      ),
                    ),
                  ),
                );
              }

              return Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  // Current stock
                  if (level != null)
                    _StockInfoCard(
                      children: [
                        _StockRow(
                          label: 'Quantity',
                          value: '${level.quantity} ${level.unit}',
                        ),
                        _StockRow(
                          label: 'Last Updated',
                          value: level.lastUpdated,
                        ),
                        if (level.earliestExpiry != null)
                          _StockRow(
                            label: 'Earliest Expiry',
                            value: level.earliestExpiry!,
                          ),
                        _StockRow(
                          label: 'Daily Consumption',
                          value:
                              '${level.dailyConsumptionRate.toStringAsFixed(1)} / day',
                        ),
                      ],
                    ),
                  const SizedBox(height: AppSpacing.md),

                  // Prediction
                  if (prediction != null) ...[
                    Text(
                      'Prediction',
                      style: TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.w600,
                        color: colorScheme.onSurface,
                      ),
                    ),
                    const SizedBox(height: AppSpacing.sm),
                    _StockInfoCard(
                      children: [
                        _StockRow(
                          label: 'Days Remaining',
                          value: '${prediction.daysRemaining}',
                        ),
                        Row(
                          mainAxisAlignment:
                              MainAxisAlignment.spaceBetween,
                          children: [
                            Text(
                              'Risk Level',
                              style: TextStyle(
                                fontSize: 12,
                                color: colorScheme.onSurfaceVariant,
                              ),
                            ),
                            SeverityBadge(
                                severity: prediction.riskLevel),
                          ],
                        ),
                        const SizedBox(height: 4),
                        if (prediction.earliestExpiry != null)
                          _StockRow(
                            label: 'Earliest Expiry',
                            value: prediction.earliestExpiry!,
                          ),
                        _StockRow(
                          label: 'Expiring Qty',
                          value: '${prediction.expiringQuantity}',
                        ),
                        const SizedBox(height: 4),
                        Text(
                          prediction.recommendedAction,
                          style: TextStyle(
                            fontSize: 11,
                            fontStyle: FontStyle.italic,
                            color: colorScheme.primary,
                          ),
                        ),
                      ],
                    ),
                  ],

                  // Redistribution suggestions
                  const SizedBox(height: AppSpacing.md),
                  _RedistributionSection(
                      medicationCode: medicationCode),
                ],
              );
            },
          ),
        ],
      ),
    );
  }
}

class _StockInfoCard extends StatelessWidget {
  const _StockInfoCard({required this.children});

  final List<Widget> children;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(AppSpacing.borderRadiusMd),
        side: BorderSide(color: colorScheme.outlineVariant),
      ),
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.sm),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: children,
        ),
      ),
    );
  }
}

class _StockRow extends StatelessWidget {
  const _StockRow({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Padding(
      padding: const EdgeInsets.only(bottom: 4),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(
            label,
            style: TextStyle(
              fontSize: 12,
              color: colorScheme.onSurfaceVariant,
            ),
          ),
          Text(
            value,
            style: TextStyle(
              fontSize: 12,
              fontWeight: FontWeight.w500,
              color: colorScheme.onSurface,
            ),
          ),
        ],
      ),
    );
  }
}

class _RedistributionSection extends ConsumerWidget {
  const _RedistributionSection({required this.medicationCode});

  final String medicationCode;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final colorScheme = Theme.of(context).colorScheme;

    // Create a simple FutureProvider inline for redistribution.
    // We use a separate approach to avoid creating too many providers.
    return FutureBuilder(
      future: _fetchRedistribution(ref),
      builder: (context, snapshot) {
        if (snapshot.connectionState == ConnectionState.waiting) {
          return const SizedBox.shrink();
        }

        final suggestions = snapshot.data;
        if (suggestions == null || suggestions.isEmpty) {
          return const SizedBox.shrink();
        }

        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Redistribution Suggestions',
              style: TextStyle(
                fontSize: 14,
                fontWeight: FontWeight.w600,
                color: colorScheme.onSurface,
              ),
            ),
            const SizedBox(height: AppSpacing.sm),
            ...suggestions.map((s) => Card(
                  elevation: 0,
                  margin: const EdgeInsets.only(bottom: AppSpacing.xs),
                  shape: RoundedRectangleBorder(
                    borderRadius:
                        BorderRadius.circular(AppSpacing.borderRadiusMd),
                    side: BorderSide(color: colorScheme.outlineVariant),
                  ),
                  child: Padding(
                    padding: const EdgeInsets.all(AppSpacing.sm),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            const Icon(Icons.swap_horiz, size: 16),
                            const SizedBox(width: 4),
                            Expanded(
                              child: Text(
                                '${s.fromSite} -> ${s.toSite}',
                                style: TextStyle(
                                  fontSize: 12,
                                  fontWeight: FontWeight.w600,
                                  color: colorScheme.onSurface,
                                ),
                              ),
                            ),
                          ],
                        ),
                        Text(
                          'Qty: ${s.suggestedQuantity} '
                          '(${s.fromSiteQuantity} -> ${s.toSiteQuantity})',
                          style: const TextStyle(fontSize: 11),
                        ),
                        Text(
                          s.rationale,
                          style: TextStyle(
                            fontSize: 11,
                            color: colorScheme.onSurfaceVariant,
                          ),
                        ),
                      ],
                    ),
                  ),
                )),
          ],
        );
      },
    );
  }

  Future<List<FormularyRedistributionSuggestion>?> _fetchRedistribution(
      WidgetRef ref) async {
    try {
      final api = ref.read(formularyApiProvider);
      final envelope =
          await api.getRedistributionSuggestions(medicationCode);
      return envelope.data?.suggestions;
    } catch (_) {
      return null;
    }
  }
}
