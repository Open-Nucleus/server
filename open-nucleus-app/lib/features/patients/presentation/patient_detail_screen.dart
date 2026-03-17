import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/constants/api_paths.dart';
import '../../../core/extensions/date_extensions.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_spacing.dart';
import '../../../core/theme/app_typography.dart';
import '../../../shared/models/patient_models.dart';
import '../../../shared/providers/dio_provider.dart';
import '../../../shared/widgets/confirm_dialog.dart';
import '../../../shared/widgets/data_table_card.dart';
import '../../../shared/widgets/empty_state.dart';
import '../../../shared/widgets/error_state.dart';
import '../../../shared/widgets/loading_skeleton.dart';
import '../../../shared/widgets/pagination_controls.dart';
import 'patient_detail_providers.dart';

/// Full patient detail screen with demographics panel and tabbed clinical data.
///
/// Layout: fixed-width left panel (280 px) with demographics and quick actions,
/// plus a right panel filling remaining space with 10 tabs of clinical data.
class PatientDetailScreen extends ConsumerStatefulWidget {
  const PatientDetailScreen({required this.patientId, super.key});

  final String patientId;

  @override
  ConsumerState<PatientDetailScreen> createState() =>
      _PatientDetailScreenState();
}

class _PatientDetailScreenState extends ConsumerState<PatientDetailScreen>
    with TickerProviderStateMixin {
  late final TabController _tabController;

  static const _tabs = <String>[
    'Overview',
    'Encounters',
    'Vitals',
    'Conditions',
    'Medications',
    'Allergies',
    'Immunizations',
    'Procedures',
    'Consent',
    'History',
  ];

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: _tabs.length, vsync: this);
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  // ---------------------------------------------------------------------------
  // Build
  // ---------------------------------------------------------------------------

  @override
  Widget build(BuildContext context) {
    final bundleAsync = ref.watch(patientDetailProvider(widget.patientId));

    return bundleAsync.when(
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (err, _) => ErrorState(
        message: 'Failed to load patient',
        details: err.toString(),
        onRetry: () => ref.invalidate(patientDetailProvider(widget.patientId)),
      ),
      data: (bundle) => Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // ── Left Panel: Demographics ──────────────────────────────────
          SizedBox(
            width: 280,
            child: _DemographicsPanel(
              patient: bundle.patient,
              patientId: widget.patientId,
              onErase: () => _handleErase(context),
            ),
          ),

          // ── Vertical divider ─────────────────────────────────────────
          VerticalDivider(width: 1, thickness: 1),

          // ── Right Panel: Tabbed Content ──────────────────────────────
          Expanded(
            child: Column(
              children: [
                TabBar(
                  controller: _tabController,
                  isScrollable: true,
                  tabAlignment: TabAlignment.start,
                  tabs: _tabs.map((t) => Tab(text: t)).toList(),
                ),
                Expanded(
                  child: TabBarView(
                    controller: _tabController,
                    children: [
                      _OverviewTab(bundle: bundle),
                      _EncountersTab(patientId: widget.patientId),
                      _VitalsTab(patientId: widget.patientId),
                      _ConditionsTab(patientId: widget.patientId),
                      _MedicationsTab(patientId: widget.patientId),
                      _AllergiesTab(patientId: widget.patientId),
                      _ImmunizationsTab(patientId: widget.patientId),
                      _ProceduresTab(patientId: widget.patientId),
                      _ConsentTab(patientId: widget.patientId),
                      _HistoryTab(patientId: widget.patientId),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  // ---------------------------------------------------------------------------
  // Erase (crypto-erasure)
  // ---------------------------------------------------------------------------

  Future<void> _handleErase(BuildContext context) async {
    final confirmed = await ConfirmDialog.show(
      context: context,
      title: 'Erase Patient Data',
      message:
          'This will permanently destroy the encryption key and purge all '
          'index data for this patient. This action CANNOT be undone.\n\n'
          'Patient ID: ${widget.patientId}',
      confirmLabel: 'Erase',
      destructive: true,
    );

    if (!confirmed || !mounted) return;

    try {
      final dio = ref.read(dioProvider);
      await dio.delete(ApiPaths.patientErase(widget.patientId));

      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Patient data erased')),
      );
      context.go('/patients');
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Erase failed: $e')),
      );
    }
  }
}

// =============================================================================
// Demographics Panel (Left)
// =============================================================================

class _DemographicsPanel extends StatelessWidget {
  const _DemographicsPanel({
    required this.patient,
    required this.patientId,
    required this.onErase,
  });

  final Map<String, dynamic> patient;
  final String patientId;
  final VoidCallback onErase;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final name = _extractName(patient);
    final gender = _extractGender(patient);
    final (birthDate, age) = _extractBirthDateAndAge(patient);
    final active = patient['active'] as bool? ?? true;
    final siteId =
        (patient['meta'] as Map<String, dynamic>?)?['source'] as String? ??
            'Unknown';

    return Container(
      color: colorScheme.surfaceContainerLowest,
      child: ListView(
        padding: AppSpacing.pagePadding,
        children: [
          // ── Gender Icon ─────────────────────────────────────────────
          Center(
            child: CircleAvatar(
              radius: 36,
              backgroundColor: colorScheme.primaryContainer,
              child: Icon(
                gender.toLowerCase() == 'female'
                    ? Icons.female
                    : gender.toLowerCase() == 'male'
                        ? Icons.male
                        : Icons.person,
                size: 36,
                color: colorScheme.onPrimaryContainer,
              ),
            ),
          ),
          const SizedBox(height: AppSpacing.md),

          // ── Name ────────────────────────────────────────────────────
          Text(
            name,
            style: AppTypography.h2.copyWith(color: colorScheme.onSurface),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: AppSpacing.xs),

          // ── Gender text ─────────────────────────────────────────────
          Text(
            gender,
            style:
                AppTypography.bodyMedium.copyWith(color: colorScheme.onSurfaceVariant),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: AppSpacing.md),

          // ── DOB + Age ───────────────────────────────────────────────
          _InfoRow(
            icon: Icons.cake_outlined,
            label: 'Date of Birth',
            value: birthDate,
          ),
          const SizedBox(height: AppSpacing.xs),
          _InfoRow(icon: Icons.timelapse, label: 'Age', value: age),
          const SizedBox(height: AppSpacing.sm),

          // ── Patient ID (copyable) ───────────────────────────────────
          _CopyableId(patientId: patientId),
          const SizedBox(height: AppSpacing.sm),

          // ── Active badge ────────────────────────────────────────────
          Row(
            children: [
              Icon(Icons.circle, size: 10, color: active ? AppColors.statusActive : AppColors.statusInactive),
              const SizedBox(width: 6),
              Text(
                active ? 'Active' : 'Inactive',
                style: AppTypography.labelMedium.copyWith(
                  color: active ? AppColors.statusActive : AppColors.statusInactive,
                ),
              ),
            ],
          ),
          const SizedBox(height: AppSpacing.xs),

          // ── Site ID ─────────────────────────────────────────────────
          _InfoRow(icon: Icons.location_on_outlined, label: 'Site', value: siteId),

          const SizedBox(height: AppSpacing.lg),
          const Divider(),
          const SizedBox(height: AppSpacing.md),

          // ── Quick Actions ───────────────────────────────────────────
          Text(
            'Actions',
            style: AppTypography.labelLarge.copyWith(color: colorScheme.onSurfaceVariant),
          ),
          const SizedBox(height: AppSpacing.sm),

          OutlinedButton.icon(
            onPressed: () {
              // TODO: navigate to patient edit form
            },
            icon: const Icon(Icons.edit_outlined, size: 18),
            label: const Text('Edit'),
          ),
          const SizedBox(height: AppSpacing.xs),
          OutlinedButton.icon(
            onPressed: () {
              // TODO: navigate to full history
            },
            icon: const Icon(Icons.history, size: 18),
            label: const Text('History'),
          ),
          const SizedBox(height: AppSpacing.xs),
          OutlinedButton.icon(
            onPressed: onErase,
            icon: Icon(Icons.delete_forever, size: 18, color: colorScheme.error),
            label: Text('Erase', style: TextStyle(color: colorScheme.error)),
            style: OutlinedButton.styleFrom(
              side: BorderSide(color: colorScheme.error.withOpacity(0.5)),
            ),
          ),
        ],
      ),
    );
  }
}

// ─────────────────────────────────────────────────────────────────────────────
// Small helper widgets for demographics
// ─────────────────────────────────────────────────────────────────────────────

class _InfoRow extends StatelessWidget {
  const _InfoRow({
    required this.icon,
    required this.label,
    required this.value,
  });

  final IconData icon;
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    return Row(
      children: [
        Icon(icon, size: 16, color: colorScheme.onSurfaceVariant),
        const SizedBox(width: 8),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                label,
                style: AppTypography.labelSmall
                    .copyWith(color: colorScheme.onSurfaceVariant),
              ),
              Text(
                value,
                style:
                    AppTypography.bodyMedium.copyWith(color: colorScheme.onSurface),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

class _CopyableId extends StatelessWidget {
  const _CopyableId({required this.patientId});

  final String patientId;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    return InkWell(
      borderRadius: BorderRadius.circular(AppSpacing.borderRadiusSm),
      onTap: () {
        Clipboard.setData(ClipboardData(text: patientId));
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Patient ID copied'),
            duration: Duration(seconds: 1),
          ),
        );
      },
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
        decoration: BoxDecoration(
          color: colorScheme.surfaceContainerHighest.withOpacity(0.5),
          borderRadius: BorderRadius.circular(AppSpacing.borderRadiusSm),
        ),
        child: Row(
          children: [
            Icon(Icons.badge_outlined, size: 16, color: colorScheme.onSurfaceVariant),
            const SizedBox(width: 8),
            Expanded(
              child: Text(
                patientId,
                style: AppTypography.code.copyWith(
                  color: colorScheme.onSurface,
                  fontSize: 12,
                ),
                overflow: TextOverflow.ellipsis,
              ),
            ),
            Icon(Icons.copy, size: 14, color: colorScheme.onSurfaceVariant),
          ],
        ),
      ),
    );
  }
}

// =============================================================================
// Tab 1: Overview
// =============================================================================

class _OverviewTab extends StatelessWidget {
  const _OverviewTab({required this.bundle});

  final PatientBundle bundle;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    // Filter active conditions.
    final activeConditions = bundle.conditions.where((c) {
      final cs = (c['clinicalStatus'] as Map<String, dynamic>?);
      final coding = (cs?['coding'] as List?)?.firstOrNull as Map<String, dynamic>?;
      return coding?['code'] == 'active';
    }).toList();

    // Filter active medications.
    final activeMeds = bundle.medicationRequests.where((m) {
      final status = m['status'] as String?;
      return status == 'active';
    }).toList();

    // All allergies.
    final allergies = bundle.allergyIntolerances;

    // Recent encounters (up to 3).
    final recentEncounters = bundle.encounters.take(3).toList();

    return SingleChildScrollView(
      padding: AppSpacing.pagePadding,
      child: Wrap(
        spacing: AppSpacing.md,
        runSpacing: AppSpacing.md,
        children: [
          _SummaryCard(
            title: 'Active Conditions',
            count: activeConditions.length,
            icon: Icons.medical_information_outlined,
            color: AppColors.warning,
            items: activeConditions.take(5).map(_extractCodeDisplay).toList(),
          ),
          _SummaryCard(
            title: 'Current Medications',
            count: activeMeds.length,
            icon: Icons.medication_outlined,
            color: AppColors.info,
            items: activeMeds.take(5).map((m) {
              final med = m['medicationCodeableConcept'] as Map<String, dynamic>?;
              return _extractCodeableConceptDisplay(med);
            }).toList(),
          ),
          _SummaryCard(
            title: 'Active Allergies',
            count: allergies.length,
            icon: Icons.warning_amber_outlined,
            color: AppColors.error,
            items: allergies.take(5).map((a) {
              final code = a['code'] as Map<String, dynamic>?;
              return _extractCodeableConceptDisplay(code);
            }).toList(),
          ),
          _SummaryCard(
            title: 'Recent Encounters',
            count: bundle.encounters.length,
            icon: Icons.event_note_outlined,
            color: AppColors.secondary,
            items: recentEncounters.map((e) {
              final classCode = e['class'] as Map<String, dynamic>?;
              final status = e['status'] as String? ?? '';
              final display = classCode?['display'] as String? ??
                  classCode?['code'] as String? ??
                  'Encounter';
              return '$display ($status)';
            }).toList(),
          ),
        ],
      ),
    );
  }
}

class _SummaryCard extends StatelessWidget {
  const _SummaryCard({
    required this.title,
    required this.count,
    required this.icon,
    required this.color,
    required this.items,
  });

  final String title;
  final int count;
  final IconData icon;
  final Color color;
  final List<String> items;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return SizedBox(
      width: 320,
      child: Card(
        elevation: 1,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(AppSpacing.borderRadiusLg),
        ),
        child: Padding(
          padding: AppSpacing.cardPadding,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisSize: MainAxisSize.min,
            children: [
              Row(
                children: [
                  Container(
                    padding: const EdgeInsets.all(8),
                    decoration: BoxDecoration(
                      color: color.withOpacity(0.12),
                      borderRadius:
                          BorderRadius.circular(AppSpacing.borderRadiusMd),
                    ),
                    child: Icon(icon, size: 20, color: color),
                  ),
                  const SizedBox(width: AppSpacing.sm),
                  Expanded(
                    child: Text(
                      title,
                      style: AppTypography.h4
                          .copyWith(color: colorScheme.onSurface),
                    ),
                  ),
                  Container(
                    padding:
                        const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                    decoration: BoxDecoration(
                      color: color.withOpacity(0.12),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Text(
                      '$count',
                      style: AppTypography.labelMedium.copyWith(color: color),
                    ),
                  ),
                ],
              ),
              if (items.isNotEmpty) ...[
                const SizedBox(height: AppSpacing.sm),
                const Divider(height: 1),
                const SizedBox(height: AppSpacing.sm),
                ...items.map(
                  (item) => Padding(
                    padding: const EdgeInsets.only(bottom: 4),
                    child: Row(
                      children: [
                        Icon(Icons.circle, size: 6, color: colorScheme.onSurfaceVariant),
                        const SizedBox(width: 8),
                        Expanded(
                          child: Text(
                            item,
                            style: AppTypography.bodySmall
                                .copyWith(color: colorScheme.onSurface),
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ] else ...[
                const SizedBox(height: AppSpacing.sm),
                Text(
                  'None recorded',
                  style: AppTypography.bodySmall
                      .copyWith(color: colorScheme.onSurfaceVariant),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}

// =============================================================================
// Tab 2: Encounters
// =============================================================================

class _EncountersTab extends ConsumerWidget {
  const _EncountersTab({required this.patientId});

  final String patientId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final asyncData = ref.watch(patientEncountersProvider(patientId));

    return asyncData.when(
      loading: () => LoadingSkeleton.table(rows: 6, cols: 4),
      error: (err, _) => ErrorState(
        message: 'Failed to load encounters',
        details: err.toString(),
        onRetry: () =>
            ref.invalidate(patientEncountersProvider(patientId)),
      ),
      data: (data) => _buildTable(context, data),
    );
  }

  Widget _buildTable(BuildContext context, dynamic data) {
    final resources = data.resources as List<Map<String, dynamic>>;

    return Padding(
      padding: AppSpacing.pagePadding,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Expanded(
            child: DataTableCard(
              title: 'Encounters',
              actions: [
                FilledButton.icon(
                  onPressed: () {
                    // TODO: open new encounter form
                  },
                  icon: const Icon(Icons.add, size: 18),
                  label: const Text('New Encounter'),
                ),
              ],
              emptyTitle: 'No encounters recorded',
              emptyIcon: Icons.event_note_outlined,
              columns: const [
                DataColumn(label: Text('Date')),
                DataColumn(label: Text('Status')),
                DataColumn(label: Text('Class')),
                DataColumn(label: Text('Duration')),
              ],
              rows: resources.map((e) {
                final period = e['period'] as Map<String, dynamic>?;
                final start = period?['start'] as String? ?? '';
                final end = period?['end'] as String?;
                final status = _extractStatus(e);
                final classCode = e['class'] as Map<String, dynamic>?;
                final classDisplay = classCode?['display'] as String? ??
                    classCode?['code'] as String? ??
                    '';
                final duration = _computeDuration(start, end);

                return DataRow(cells: [
                  DataCell(Text(_formatDateTime(start))),
                  DataCell(_StatusChip(status: status)),
                  DataCell(Text(classDisplay)),
                  DataCell(Text(duration)),
                ]);
              }).toList(),
            ),
          ),
          if (data.totalPages > 1)
            PaginationControls(
              currentPage: data.page,
              totalPages: data.totalPages,
              totalItems: data.total,
              rowsPerPage: data.perPage,
              onPageChanged: (_) {
                // TODO: implement page change
              },
            ),
        ],
      ),
    );
  }
}

// =============================================================================
// Tab 3: Vitals (Observations)
// =============================================================================

class _VitalsTab extends ConsumerWidget {
  const _VitalsTab({required this.patientId});

  final String patientId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final asyncData = ref.watch(patientObservationsProvider(patientId));

    return asyncData.when(
      loading: () => LoadingSkeleton.table(rows: 6, cols: 4),
      error: (err, _) => ErrorState(
        message: 'Failed to load vitals',
        details: err.toString(),
        onRetry: () =>
            ref.invalidate(patientObservationsProvider(patientId)),
      ),
      data: (data) => _buildTable(context, data),
    );
  }

  Widget _buildTable(BuildContext context, dynamic data) {
    final resources = data.resources as List<Map<String, dynamic>>;

    return Padding(
      padding: AppSpacing.pagePadding,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Expanded(
            child: DataTableCard(
              title: 'Observations / Vitals',
              actions: [
                FilledButton.icon(
                  onPressed: () {
                    // TODO: open record vital form
                  },
                  icon: const Icon(Icons.add, size: 18),
                  label: const Text('Record Vital'),
                ),
              ],
              emptyTitle: 'No vitals recorded',
              emptyIcon: Icons.monitor_heart_outlined,
              columns: const [
                DataColumn(label: Text('Date')),
                DataColumn(label: Text('Code / Display')),
                DataColumn(label: Text('Value')),
                DataColumn(label: Text('Status')),
              ],
              rows: resources.map((o) {
                final effectiveDateTime =
                    o['effectiveDateTime'] as String? ?? '';
                final codeDisplay = _extractCodeDisplay(o);
                final value = _extractObservationValue(o);
                final status = _extractStatus(o);

                return DataRow(cells: [
                  DataCell(Text(_formatDateTime(effectiveDateTime))),
                  DataCell(Text(codeDisplay)),
                  DataCell(Text(value)),
                  DataCell(_StatusChip(status: status)),
                ]);
              }).toList(),
            ),
          ),
          if (data.totalPages > 1)
            PaginationControls(
              currentPage: data.page,
              totalPages: data.totalPages,
              totalItems: data.total,
              rowsPerPage: data.perPage,
              onPageChanged: (_) {},
            ),
        ],
      ),
    );
  }
}

// =============================================================================
// Tab 4: Conditions
// =============================================================================

class _ConditionsTab extends ConsumerWidget {
  const _ConditionsTab({required this.patientId});

  final String patientId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final asyncData = ref.watch(patientConditionsProvider(patientId));

    return asyncData.when(
      loading: () => LoadingSkeleton.table(rows: 6, cols: 4),
      error: (err, _) => ErrorState(
        message: 'Failed to load conditions',
        details: err.toString(),
        onRetry: () =>
            ref.invalidate(patientConditionsProvider(patientId)),
      ),
      data: (data) => _buildTable(context, data),
    );
  }

  Widget _buildTable(BuildContext context, dynamic data) {
    final resources = data.resources as List<Map<String, dynamic>>;

    return Padding(
      padding: AppSpacing.pagePadding,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Expanded(
            child: DataTableCard(
              title: 'Conditions',
              actions: [
                FilledButton.icon(
                  onPressed: () {
                    // TODO: open add condition form
                  },
                  icon: const Icon(Icons.add, size: 18),
                  label: const Text('Add Condition'),
                ),
              ],
              emptyTitle: 'No conditions recorded',
              emptyIcon: Icons.medical_information_outlined,
              columns: const [
                DataColumn(label: Text('Code / Display')),
                DataColumn(label: Text('Clinical Status')),
                DataColumn(label: Text('Verification')),
                DataColumn(label: Text('Onset')),
              ],
              rows: resources.map((c) {
                final codeDisplay = _extractCodeDisplay(c);
                final clinicalStatus = _extractNestedCodeDisplay(
                    c['clinicalStatus'] as Map<String, dynamic>?);
                final verificationStatus = _extractNestedCodeDisplay(
                    c['verificationStatus'] as Map<String, dynamic>?);
                final onset = c['onsetDateTime'] as String? ??
                    c['onsetString'] as String? ??
                    '';

                return DataRow(cells: [
                  DataCell(Text(codeDisplay)),
                  DataCell(_ClinicalStatusBadge(status: clinicalStatus)),
                  DataCell(Text(verificationStatus)),
                  DataCell(Text(_formatDateTime(onset))),
                ]);
              }).toList(),
            ),
          ),
          if (data.totalPages > 1)
            PaginationControls(
              currentPage: data.page,
              totalPages: data.totalPages,
              totalItems: data.total,
              rowsPerPage: data.perPage,
              onPageChanged: (_) {},
            ),
        ],
      ),
    );
  }
}

// =============================================================================
// Tab 5: Medications
// =============================================================================

class _MedicationsTab extends ConsumerWidget {
  const _MedicationsTab({required this.patientId});

  final String patientId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final asyncData = ref.watch(patientMedicationsProvider(patientId));

    return asyncData.when(
      loading: () => LoadingSkeleton.table(rows: 6, cols: 4),
      error: (err, _) => ErrorState(
        message: 'Failed to load medications',
        details: err.toString(),
        onRetry: () =>
            ref.invalidate(patientMedicationsProvider(patientId)),
      ),
      data: (data) => _buildTable(context, data),
    );
  }

  Widget _buildTable(BuildContext context, dynamic data) {
    final resources = data.resources as List<Map<String, dynamic>>;

    return Padding(
      padding: AppSpacing.pagePadding,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Expanded(
            child: DataTableCard(
              title: 'Medication Requests',
              actions: [
                FilledButton.icon(
                  onPressed: () {
                    // TODO: open prescribe form
                  },
                  icon: const Icon(Icons.add, size: 18),
                  label: const Text('Prescribe'),
                ),
              ],
              emptyTitle: 'No medications prescribed',
              emptyIcon: Icons.medication_outlined,
              columns: const [
                DataColumn(label: Text('Medication')),
                DataColumn(label: Text('Status')),
                DataColumn(label: Text('Intent')),
                DataColumn(label: Text('Dosage')),
              ],
              rows: resources.map((m) {
                final medConcept =
                    m['medicationCodeableConcept'] as Map<String, dynamic>?;
                final medName = _extractCodeableConceptDisplay(medConcept);
                final status = _extractStatus(m);
                final intent = m['intent'] as String? ?? '';
                final dosage = _extractDosageText(m);

                return DataRow(cells: [
                  DataCell(Text(medName)),
                  DataCell(_StatusChip(status: status)),
                  DataCell(Text(intent)),
                  DataCell(
                    SizedBox(
                      width: 200,
                      child: Text(dosage, overflow: TextOverflow.ellipsis),
                    ),
                  ),
                ]);
              }).toList(),
            ),
          ),
          if (data.totalPages > 1)
            PaginationControls(
              currentPage: data.page,
              totalPages: data.totalPages,
              totalItems: data.total,
              rowsPerPage: data.perPage,
              onPageChanged: (_) {},
            ),
        ],
      ),
    );
  }
}

// =============================================================================
// Tab 6: Allergies
// =============================================================================

class _AllergiesTab extends ConsumerWidget {
  const _AllergiesTab({required this.patientId});

  final String patientId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final asyncData = ref.watch(patientAllergiesProvider(patientId));

    return asyncData.when(
      loading: () => LoadingSkeleton.table(rows: 6, cols: 4),
      error: (err, _) => ErrorState(
        message: 'Failed to load allergies',
        details: err.toString(),
        onRetry: () =>
            ref.invalidate(patientAllergiesProvider(patientId)),
      ),
      data: (data) => _buildTable(context, data),
    );
  }

  Widget _buildTable(BuildContext context, dynamic data) {
    final resources = data.resources as List<Map<String, dynamic>>;

    return Padding(
      padding: AppSpacing.pagePadding,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Expanded(
            child: DataTableCard(
              title: 'Allergy Intolerances',
              actions: [
                FilledButton.icon(
                  onPressed: () {
                    // TODO: open add allergy form
                  },
                  icon: const Icon(Icons.add, size: 18),
                  label: const Text('Add Allergy'),
                ),
              ],
              emptyTitle: 'No allergies recorded',
              emptyIcon: Icons.warning_amber_outlined,
              columns: const [
                DataColumn(label: Text('Substance')),
                DataColumn(label: Text('Type')),
                DataColumn(label: Text('Clinical Status')),
                DataColumn(label: Text('Criticality')),
              ],
              rows: resources.map((a) {
                final code = a['code'] as Map<String, dynamic>?;
                final substance = _extractCodeableConceptDisplay(code);
                final type = a['type'] as String? ?? '';
                final clinicalStatus = _extractNestedCodeDisplay(
                    a['clinicalStatus'] as Map<String, dynamic>?);
                final criticality = a['criticality'] as String? ?? '';

                return DataRow(cells: [
                  DataCell(Text(substance)),
                  DataCell(Text(type)),
                  DataCell(_ClinicalStatusBadge(status: clinicalStatus)),
                  DataCell(_CriticalityBadge(criticality: criticality)),
                ]);
              }).toList(),
            ),
          ),
          if (data.totalPages > 1)
            PaginationControls(
              currentPage: data.page,
              totalPages: data.totalPages,
              totalItems: data.total,
              rowsPerPage: data.perPage,
              onPageChanged: (_) {},
            ),
        ],
      ),
    );
  }
}

// =============================================================================
// Tab 7: Immunizations
// =============================================================================

class _ImmunizationsTab extends ConsumerWidget {
  const _ImmunizationsTab({required this.patientId});

  final String patientId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final asyncData = ref.watch(patientImmunizationsProvider(patientId));

    return asyncData.when(
      loading: () => LoadingSkeleton.table(rows: 6, cols: 3),
      error: (err, _) => ErrorState(
        message: 'Failed to load immunizations',
        details: err.toString(),
        onRetry: () =>
            ref.invalidate(patientImmunizationsProvider(patientId)),
      ),
      data: (data) => _buildTable(context, data),
    );
  }

  Widget _buildTable(BuildContext context, dynamic data) {
    final resources = data.resources as List<Map<String, dynamic>>;

    return Padding(
      padding: AppSpacing.pagePadding,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Expanded(
            child: DataTableCard(
              title: 'Immunizations',
              actions: [
                FilledButton.icon(
                  onPressed: () {
                    // TODO: open record immunization form
                  },
                  icon: const Icon(Icons.add, size: 18),
                  label: const Text('Record Immunization'),
                ),
              ],
              emptyTitle: 'No immunizations recorded',
              emptyIcon: Icons.vaccines_outlined,
              columns: const [
                DataColumn(label: Text('Vaccine')),
                DataColumn(label: Text('Date')),
                DataColumn(label: Text('Status')),
              ],
              rows: resources.map((i) {
                final vaccineCode =
                    i['vaccineCode'] as Map<String, dynamic>?;
                final vaccine = _extractCodeableConceptDisplay(vaccineCode);
                final date = i['occurrenceDateTime'] as String? ??
                    i['occurrenceString'] as String? ??
                    '';
                final status = _extractStatus(i);

                return DataRow(cells: [
                  DataCell(Text(vaccine)),
                  DataCell(Text(_formatDateTime(date))),
                  DataCell(_StatusChip(status: status)),
                ]);
              }).toList(),
            ),
          ),
          if (data.totalPages > 1)
            PaginationControls(
              currentPage: data.page,
              totalPages: data.totalPages,
              totalItems: data.total,
              rowsPerPage: data.perPage,
              onPageChanged: (_) {},
            ),
        ],
      ),
    );
  }
}

// =============================================================================
// Tab 8: Procedures
// =============================================================================

class _ProceduresTab extends ConsumerWidget {
  const _ProceduresTab({required this.patientId});

  final String patientId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final asyncData = ref.watch(patientProceduresProvider(patientId));

    return asyncData.when(
      loading: () => LoadingSkeleton.table(rows: 6, cols: 3),
      error: (err, _) => ErrorState(
        message: 'Failed to load procedures',
        details: err.toString(),
        onRetry: () =>
            ref.invalidate(patientProceduresProvider(patientId)),
      ),
      data: (data) => _buildTable(context, data),
    );
  }

  Widget _buildTable(BuildContext context, dynamic data) {
    final resources = data.resources as List<Map<String, dynamic>>;

    return Padding(
      padding: AppSpacing.pagePadding,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Expanded(
            child: DataTableCard(
              title: 'Procedures',
              actions: [
                FilledButton.icon(
                  onPressed: () {
                    // TODO: open record procedure form
                  },
                  icon: const Icon(Icons.add, size: 18),
                  label: const Text('Record Procedure'),
                ),
              ],
              emptyTitle: 'No procedures recorded',
              emptyIcon: Icons.local_hospital_outlined,
              columns: const [
                DataColumn(label: Text('Procedure')),
                DataColumn(label: Text('Date')),
                DataColumn(label: Text('Status')),
              ],
              rows: resources.map((p) {
                final codeDisplay = _extractCodeDisplay(p);
                final performed = p['performedDateTime'] as String? ??
                    (p['performedPeriod']
                        as Map<String, dynamic>?)?['start'] as String? ??
                    '';
                final status = _extractStatus(p);

                return DataRow(cells: [
                  DataCell(Text(codeDisplay)),
                  DataCell(Text(_formatDateTime(performed))),
                  DataCell(_StatusChip(status: status)),
                ]);
              }).toList(),
            ),
          ),
          if (data.totalPages > 1)
            PaginationControls(
              currentPage: data.page,
              totalPages: data.totalPages,
              totalItems: data.total,
              rowsPerPage: data.perPage,
              onPageChanged: (_) {},
            ),
        ],
      ),
    );
  }
}

// =============================================================================
// Tab 9: Consent
// =============================================================================

class _ConsentTab extends ConsumerWidget {
  const _ConsentTab({required this.patientId});

  final String patientId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final asyncData = ref.watch(patientConsentsProvider(patientId));

    return asyncData.when(
      loading: () => LoadingSkeleton.table(rows: 6, cols: 5),
      error: (err, _) => ErrorState(
        message: 'Failed to load consents',
        details: err.toString(),
        onRetry: () =>
            ref.invalidate(patientConsentsProvider(patientId)),
      ),
      data: (data) => _buildTable(context, ref, data),
    );
  }

  Widget _buildTable(BuildContext context, WidgetRef ref, dynamic data) {
    final consents = data.consents as List;

    return Padding(
      padding: AppSpacing.pagePadding,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Expanded(
            child: DataTableCard(
              title: 'Consent Records',
              actions: [
                FilledButton.icon(
                  onPressed: () {
                    // TODO: open grant consent form
                  },
                  icon: const Icon(Icons.add, size: 18),
                  label: const Text('Grant Consent'),
                ),
              ],
              emptyTitle: 'No consent records',
              emptyIcon: Icons.handshake_outlined,
              columns: const [
                DataColumn(label: Text('Scope')),
                DataColumn(label: Text('Performer')),
                DataColumn(label: Text('Status')),
                DataColumn(label: Text('Period')),
                DataColumn(label: Text('Category')),
                DataColumn(label: Text('Actions')),
              ],
              rows: consents.map<DataRow>((c) {
                final scope = c.scopeCode;
                final performer = c.performerId;
                final status = c.status;
                final periodStart = c.periodStart ?? '';
                final periodEnd = c.periodEnd ?? '';
                final period = periodStart.isNotEmpty
                    ? '${_formatDateTime(periodStart)} - ${_formatDateTime(periodEnd)}'
                    : '';
                final category = c.category ?? '';

                return DataRow(cells: [
                  DataCell(Text(scope)),
                  DataCell(Text(
                    performer,
                    style: AppTypography.code.copyWith(fontSize: 11),
                  )),
                  DataCell(_StatusChip(status: status)),
                  DataCell(Text(period)),
                  DataCell(Text(category)),
                  DataCell(
                    status == 'active'
                        ? TextButton(
                            onPressed: () async {
                              final confirmed = await ConfirmDialog.show(
                                context: context,
                                title: 'Revoke Consent',
                                message:
                                    'Are you sure you want to revoke this consent grant?',
                                confirmLabel: 'Revoke',
                                destructive: true,
                              );
                              if (confirmed) {
                                // TODO: call revoke endpoint
                              }
                            },
                            child: Text(
                              'Revoke',
                              style: TextStyle(
                                color: Theme.of(context).colorScheme.error,
                              ),
                            ),
                          )
                        : const SizedBox.shrink(),
                  ),
                ]);
              }).toList(),
            ),
          ),
        ],
      ),
    );
  }
}

// =============================================================================
// Tab 10: History (Timeline)
// =============================================================================

class _HistoryTab extends ConsumerWidget {
  const _HistoryTab({required this.patientId});

  final String patientId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final asyncData = ref.watch(patientHistoryProvider(patientId));

    return asyncData.when(
      loading: () => LoadingSkeleton.listTile(count: 8),
      error: (err, _) => ErrorState(
        message: 'Failed to load history',
        details: err.toString(),
        onRetry: () =>
            ref.invalidate(patientHistoryProvider(patientId)),
      ),
      data: (data) {
        final entries = data.entries;
        if (entries.isEmpty) {
          return const EmptyState(
            icon: Icons.history,
            title: 'No history',
            subtitle: 'No changes have been recorded for this patient yet.',
          );
        }
        return _HistoryTimeline(entries: entries);
      },
    );
  }
}

class _HistoryTimeline extends StatelessWidget {
  const _HistoryTimeline({required this.entries});

  final List<HistoryEntry> entries;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return ListView.builder(
      padding: AppSpacing.pagePadding,
      itemCount: entries.length,
      itemBuilder: (context, index) {
        final entry = entries[index];
        final isLast = index == entries.length - 1;

        return IntrinsicHeight(
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // ── Timeline line + dot ──────────────────────────────────
              SizedBox(
                width: 40,
                child: Column(
                  children: [
                    Container(
                      width: 12,
                      height: 12,
                      decoration: BoxDecoration(
                        color: _operationColor(entry.operation),
                        shape: BoxShape.circle,
                        border: Border.all(
                          color: colorScheme.surface,
                          width: 2,
                        ),
                        boxShadow: [
                          BoxShadow(
                            color: _operationColor(entry.operation)
                                .withOpacity(0.3),
                            blurRadius: 4,
                          ),
                        ],
                      ),
                    ),
                    if (!isLast)
                      Expanded(
                        child: Container(
                          width: 2,
                          color: colorScheme.outlineVariant,
                        ),
                      ),
                  ],
                ),
              ),

              // ── Entry content ────────────────────────────────────────
              Expanded(
                child: Padding(
                  padding: const EdgeInsets.only(bottom: AppSpacing.md),
                  child: Card(
                    elevation: 0,
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
                          // Timestamp + operation badge
                          Row(
                            children: [
                              Text(
                                _formatDateTime(entry.timestamp),
                                style: AppTypography.caption,
                              ),
                              const SizedBox(width: AppSpacing.sm),
                              _OperationBadge(operation: entry.operation),
                              const Spacer(),
                              Text(
                                entry.resourceType,
                                style: AppTypography.labelSmall.copyWith(
                                  color: colorScheme.primary,
                                ),
                              ),
                            ],
                          ),
                          const SizedBox(height: AppSpacing.xs),

                          // Message
                          Text(
                            entry.message,
                            style: AppTypography.bodySmall
                                .copyWith(color: colorScheme.onSurface),
                          ),

                          // Author + commit
                          const SizedBox(height: AppSpacing.xs),
                          Row(
                            children: [
                              Icon(Icons.person_outline,
                                  size: 12,
                                  color: colorScheme.onSurfaceVariant),
                              const SizedBox(width: 4),
                              Text(
                                entry.author,
                                style: AppTypography.caption,
                              ),
                              const SizedBox(width: AppSpacing.md),
                              Icon(Icons.tag,
                                  size: 12,
                                  color: colorScheme.onSurfaceVariant),
                              const SizedBox(width: 4),
                              Text(
                                entry.commitHash.length >= 7
                                    ? entry.commitHash.substring(0, 7)
                                    : entry.commitHash,
                                style: AppTypography.code.copyWith(
                                  fontSize: 11,
                                  color: colorScheme.onSurfaceVariant,
                                ),
                              ),
                            ],
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
              ),
            ],
          ),
        );
      },
    );
  }

  Color _operationColor(String operation) {
    switch (operation.toLowerCase()) {
      case 'create':
        return AppColors.success;
      case 'update':
        return AppColors.info;
      case 'delete':
      case 'erase':
        return AppColors.error;
      default:
        return AppColors.statusInactive;
    }
  }
}

class _OperationBadge extends StatelessWidget {
  const _OperationBadge({required this.operation});

  final String operation;

  @override
  Widget build(BuildContext context) {
    final (Color bg, Color fg) = switch (operation.toLowerCase()) {
      'create' => (AppColors.success, AppColors.success),
      'update' => (AppColors.info, AppColors.info),
      'delete' || 'erase' => (AppColors.error, AppColors.error),
      _ => (AppColors.statusInactive, AppColors.statusInactive),
    };

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
      decoration: BoxDecoration(
        color: bg.withOpacity(0.12),
        borderRadius: BorderRadius.circular(10),
      ),
      child: Text(
        operation.toUpperCase(),
        style: TextStyle(
          fontSize: 10,
          fontWeight: FontWeight.w600,
          color: fg,
          letterSpacing: 0.5,
        ),
      ),
    );
  }
}

// =============================================================================
// Shared badge widgets
// =============================================================================

class _StatusChip extends StatelessWidget {
  const _StatusChip({required this.status});

  final String status;

  @override
  Widget build(BuildContext context) {
    final (Color bg, Color fg) = switch (status.toLowerCase()) {
      'active' || 'completed' || 'final' || 'finished' =>
        (AppColors.success, AppColors.success),
      'inactive' || 'cancelled' || 'entered-in-error' =>
        (AppColors.statusInactive, AppColors.statusInactive),
      'draft' || 'preliminary' || 'planned' || 'in-progress' =>
        (AppColors.info, AppColors.info),
      'on-hold' || 'stopped' =>
        (AppColors.warning, const Color(0xFFF57F17)),
      _ => (AppColors.statusInactive, AppColors.statusInactive),
    };

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: bg.withOpacity(0.12),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: bg.withOpacity(0.3)),
      ),
      child: Text(
        status,
        style: TextStyle(
          fontSize: 11,
          fontWeight: FontWeight.w600,
          color: fg,
        ),
      ),
    );
  }
}

class _ClinicalStatusBadge extends StatelessWidget {
  const _ClinicalStatusBadge({required this.status});

  final String status;

  @override
  Widget build(BuildContext context) {
    final (Color bg, Color fg) = switch (status.toLowerCase()) {
      'active' => (AppColors.error, AppColors.error),
      'recurrence' => (AppColors.warning, const Color(0xFFF57F17)),
      'relapse' => (AppColors.severityHigh, AppColors.severityHigh),
      'inactive' || 'remission' || 'resolved' =>
        (AppColors.success, AppColors.success),
      _ => (AppColors.statusInactive, AppColors.statusInactive),
    };

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: bg.withOpacity(0.12),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: bg.withOpacity(0.3)),
      ),
      child: Text(
        status,
        style: TextStyle(
          fontSize: 11,
          fontWeight: FontWeight.w600,
          color: fg,
        ),
      ),
    );
  }
}

class _CriticalityBadge extends StatelessWidget {
  const _CriticalityBadge({required this.criticality});

  final String criticality;

  @override
  Widget build(BuildContext context) {
    final (Color bg, Color fg, String label) = switch (criticality.toLowerCase()) {
      'high' => (AppColors.error, AppColors.error, 'High'),
      'low' => (AppColors.success, AppColors.success, 'Low'),
      'unable-to-assess' =>
        (AppColors.statusInactive, AppColors.statusInactive, 'Unknown'),
      _ => (AppColors.statusInactive, AppColors.statusInactive, criticality),
    };

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: bg.withOpacity(0.12),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: bg.withOpacity(0.3)),
      ),
      child: Text(
        label,
        style: TextStyle(
          fontSize: 11,
          fontWeight: FontWeight.w600,
          color: fg,
        ),
      ),
    );
  }
}

// =============================================================================
// FHIR value extraction helpers
// =============================================================================

/// Extract display name from a FHIR Patient resource.
String _extractName(Map<String, dynamic> patient) {
  final names = patient['name'] as List<dynamic>?;
  if (names == null || names.isEmpty) return 'Unknown';

  final name = names.first as Map<String, dynamic>;
  final given = (name['given'] as List<dynamic>?)
          ?.map((g) => g as String)
          .join(' ') ??
      '';
  final family = name['family'] as String? ?? '';
  final text = name['text'] as String?;

  if (text != null && text.isNotEmpty) return text;
  return '$given $family'.trim().isEmpty ? 'Unknown' : '$given $family'.trim();
}

/// Extract gender string from a FHIR Patient resource.
String _extractGender(Map<String, dynamic> patient) {
  final gender = patient['gender'] as String?;
  if (gender == null) return 'Unknown';
  return gender[0].toUpperCase() + gender.substring(1);
}

/// Extract birth date and calculated age from a FHIR Patient resource.
(String, String) _extractBirthDateAndAge(Map<String, dynamic> patient) {
  final birthDate = patient['birthDate'] as String?;
  if (birthDate == null || birthDate.isEmpty) return ('Unknown', 'Unknown');

  try {
    final dob = DateTime.parse(birthDate);
    final now = DateTime.now();
    int age = now.year - dob.year;
    if (now.month < dob.month ||
        (now.month == dob.month && now.day < dob.day)) {
      age--;
    }

    final formatted = dob.toDisplayDate;
    return (formatted, '$age years');
  } catch (_) {
    return (birthDate, 'Unknown');
  }
}

/// Extract display text from a FHIR resource's `code` field (CodeableConcept).
String _extractCodeDisplay(Map<String, dynamic> resource) {
  final code = resource['code'] as Map<String, dynamic>?;
  return _extractCodeableConceptDisplay(code);
}

/// Extract display from a FHIR CodeableConcept.
String _extractCodeableConceptDisplay(Map<String, dynamic>? concept) {
  if (concept == null) return 'Unknown';

  // Try top-level text first.
  final text = concept['text'] as String?;
  if (text != null && text.isNotEmpty) return text;

  // Then try first coding display.
  final codings = concept['coding'] as List<dynamic>?;
  if (codings != null && codings.isNotEmpty) {
    final coding = codings.first as Map<String, dynamic>;
    final display = coding['display'] as String?;
    if (display != null && display.isNotEmpty) return display;
    final code = coding['code'] as String?;
    if (code != null && code.isNotEmpty) return code;
  }

  return 'Unknown';
}

/// Extract status string from a FHIR resource.
String _extractStatus(Map<String, dynamic> resource) {
  return resource['status'] as String? ?? 'unknown';
}

/// Extract display from a nested CodeableConcept (e.g. clinicalStatus).
String _extractNestedCodeDisplay(Map<String, dynamic>? concept) {
  return _extractCodeableConceptDisplay(concept);
}

/// Extract observation value + unit.
String _extractObservationValue(Map<String, dynamic> observation) {
  // Quantity value
  final valueQuantity = observation['valueQuantity'] as Map<String, dynamic>?;
  if (valueQuantity != null) {
    final value = valueQuantity['value'];
    final unit = valueQuantity['unit'] as String? ??
        valueQuantity['code'] as String? ??
        '';
    return '$value $unit'.trim();
  }

  // String value
  final valueString = observation['valueString'] as String?;
  if (valueString != null) return valueString;

  // CodeableConcept value
  final valueConcept =
      observation['valueCodeableConcept'] as Map<String, dynamic>?;
  if (valueConcept != null) return _extractCodeableConceptDisplay(valueConcept);

  // Boolean value
  final valueBool = observation['valueBoolean'] as bool?;
  if (valueBool != null) return valueBool ? 'Yes' : 'No';

  // Component values (e.g. blood pressure)
  final components = observation['component'] as List<dynamic>?;
  if (components != null && components.isNotEmpty) {
    return components.take(3).map((comp) {
      final c = comp as Map<String, dynamic>;
      final compCode = _extractCodeDisplay(c);
      final compQuantity = c['valueQuantity'] as Map<String, dynamic>?;
      final v = compQuantity?['value'] ?? '';
      final u = compQuantity?['unit'] as String? ?? '';
      return '$compCode: $v $u'.trim();
    }).join(', ');
  }

  return '-';
}

/// Extract dosage instruction text from a MedicationRequest.
String _extractDosageText(Map<String, dynamic> medicationRequest) {
  final dosageInstructions =
      medicationRequest['dosageInstruction'] as List<dynamic>?;
  if (dosageInstructions == null || dosageInstructions.isEmpty) return '-';

  final first = dosageInstructions.first as Map<String, dynamic>;
  final text = first['text'] as String?;
  if (text != null && text.isNotEmpty) return text;

  // Try to build from structured data.
  final timing = first['timing'] as Map<String, dynamic>?;
  final route = first['route'] as Map<String, dynamic>?;
  final doseQuantity =
      (first['doseAndRate'] as List<dynamic>?)?.firstOrNull as Map<String, dynamic>?;

  final parts = <String>[];
  if (doseQuantity != null) {
    final dose = doseQuantity['doseQuantity'] as Map<String, dynamic>?;
    if (dose != null) {
      parts.add('${dose['value']} ${dose['unit'] ?? ''}');
    }
  }
  if (route != null) {
    parts.add(_extractCodeableConceptDisplay(route));
  }
  if (timing != null) {
    final repeat = timing['repeat'] as Map<String, dynamic>?;
    if (repeat != null) {
      final freq = repeat['frequency'];
      final period = repeat['period'];
      final periodUnit = repeat['periodUnit'] as String? ?? '';
      if (freq != null && period != null) {
        parts.add('${freq}x per $period $periodUnit');
      }
    }
  }

  return parts.isEmpty ? '-' : parts.join(', ');
}

/// Compute a human-readable duration from ISO 8601 start/end strings.
String _computeDuration(String start, String? end) {
  if (start.isEmpty) return '-';
  try {
    final startDt = DateTime.parse(start);
    final endDt = end != null && end.isNotEmpty ? DateTime.parse(end) : null;

    if (endDt == null) return 'Ongoing';

    final diff = endDt.difference(startDt);
    if (diff.inDays > 0) return '${diff.inDays}d ${diff.inHours % 24}h';
    if (diff.inHours > 0) return '${diff.inHours}h ${diff.inMinutes % 60}m';
    return '${diff.inMinutes}m';
  } catch (_) {
    return '-';
  }
}

/// Format an ISO 8601 date/datetime string for display.
String _formatDateTime(String isoString) {
  if (isoString.isEmpty) return '-';
  try {
    final dt = DateTime.parse(isoString);
    return dt.toDisplayDateTime;
  } catch (_) {
    return isoString;
  }
}
