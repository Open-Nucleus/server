import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/extensions/date_extensions.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_spacing.dart';
import '../../../shared/models/patient_models.dart';
import '../../../shared/widgets/empty_state.dart';
import '../../../shared/widgets/error_state.dart';
import '../../../shared/widgets/loading_skeleton.dart';
import '../../../shared/widgets/pagination_controls.dart';
import '../../../shared/widgets/search_field.dart';
import '../../../shared/widgets/severity_badge.dart';
import 'patient_list_providers.dart';

/// Full patient list screen with search, filters, data table, and pagination.
///
/// Supports debounced search (via blind indexes), expandable filter panel,
/// and keyboard shortcut Ctrl+N to create a new patient.
class PatientListScreen extends ConsumerStatefulWidget {
  const PatientListScreen({super.key});

  @override
  ConsumerState<PatientListScreen> createState() => _PatientListScreenState();
}

class _PatientListScreenState extends ConsumerState<PatientListScreen> {
  bool _filtersExpanded = false;

  // Filter form state
  String? _selectedGender;
  String? _selectedStatus;
  bool? _hasAlerts;
  String? _siteIdText;
  DateTime? _birthDateFrom;
  DateTime? _birthDateTo;

  @override
  Widget build(BuildContext context) {
    final listState = ref.watch(patientListProvider);
    final searchAsync = ref.watch(patientSearchProvider);
    final searchQuery = ref.watch(patientSearchQueryProvider);
    final isSearching = searchQuery.trim().isNotEmpty;

    return CallbackShortcuts(
      bindings: {
        const SingleActivator(LogicalKeyboardKey.keyN, control: true): () {
          GoRouter.of(context).go('/patients/new');
        },
      },
      child: Focus(
        autofocus: true,
        child: Padding(
          padding: AppSpacing.pagePadding,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              // ── Header Row ─────────────────────────────────────────
              _HeaderRow(
                onSearchChanged: (query) {
                  ref.read(patientSearchQueryProvider.notifier).state = query;
                },
              ),
              const SizedBox(height: AppSpacing.md),

              // ── Filter Panel ───────────────────────────────────────
              _FilterPanel(
                expanded: _filtersExpanded,
                onToggle: () =>
                    setState(() => _filtersExpanded = !_filtersExpanded),
                selectedGender: _selectedGender,
                selectedStatus: _selectedStatus,
                hasAlerts: _hasAlerts,
                siteId: _siteIdText,
                birthDateFrom: _birthDateFrom,
                birthDateTo: _birthDateTo,
                onGenderChanged: (v) => setState(() => _selectedGender = v),
                onStatusChanged: (v) => setState(() => _selectedStatus = v),
                onHasAlertsChanged: (v) => setState(() => _hasAlerts = v),
                onSiteIdChanged: (v) => setState(() => _siteIdText = v),
                onBirthDateFromChanged: (v) =>
                    setState(() => _birthDateFrom = v),
                onBirthDateToChanged: (v) =>
                    setState(() => _birthDateTo = v),
                onApply: _applyFilters,
                onClear: _clearFilters,
                hasActiveFilters:
                    listState.filters.hasActiveFilters,
              ),
              const SizedBox(height: AppSpacing.sm),

              // ── Content ────────────────────────────────────────────
              Expanded(
                child: isSearching
                    ? _buildSearchResults(searchAsync)
                    : _buildPatientList(listState),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildPatientList(PatientListState listState) {
    if (listState.isLoading && listState.patients.isEmpty) {
      return LoadingSkeleton.table(rows: 8, cols: 6);
    }

    if (listState.error != null && listState.patients.isEmpty) {
      return ErrorState(
        message: 'Failed to load patients',
        details: listState.error,
        onRetry: () => ref.read(patientListProvider.notifier).fetch(),
      );
    }

    if (listState.patients.isEmpty) {
      return EmptyState(
        icon: Icons.people_outlined,
        title: 'No patients found',
        subtitle: 'Create your first patient record to get started.',
        actionLabel: 'Create Patient',
        onAction: () => GoRouter.of(context).go('/patients/new'),
      );
    }

    return Column(
      children: [
        Expanded(
          child: _PatientDataTable(
            patients: listState.patients,
            isLoading: listState.isLoading,
          ),
        ),
        PaginationControls(
          currentPage: listState.page,
          totalPages: listState.totalPages,
          totalItems: listState.totalItems,
          rowsPerPage: listState.perPage,
          onPageChanged: (page) {
            ref.read(patientListProvider.notifier).goToPage(page);
          },
          onRowsPerPageChanged: (perPage) {
            ref.read(patientListProvider.notifier).setPerPage(perPage);
          },
        ),
      ],
    );
  }

  Widget _buildSearchResults(AsyncValue<List<PatientSummary>?> searchAsync) {
    return searchAsync.when(
      loading: () => LoadingSkeleton.table(rows: 5, cols: 6),
      error: (error, _) => ErrorState(
        message: 'Search failed',
        details: error.toString(),
      ),
      data: (results) {
        if (results == null || results.isEmpty) {
          return const EmptyState(
            icon: Icons.search_off,
            title: 'No results found',
            subtitle: 'Try a different search term.',
          );
        }
        return _PatientDataTable(
          patients: results,
          isLoading: false,
        );
      },
    );
  }

  void _applyFilters() {
    final filters = PatientListFilters(
      gender: _selectedGender,
      status: _selectedStatus,
      hasAlerts: _hasAlerts,
      siteId: _siteIdText?.isNotEmpty == true ? _siteIdText : null,
      birthDateFrom: _birthDateFrom?.toFhirDate,
      birthDateTo: _birthDateTo?.toFhirDate,
    );
    ref.read(patientListProvider.notifier).applyFilters(filters);
  }

  void _clearFilters() {
    setState(() {
      _selectedGender = null;
      _selectedStatus = null;
      _hasAlerts = null;
      _siteIdText = null;
      _birthDateFrom = null;
      _birthDateTo = null;
    });
    ref.read(patientListProvider.notifier).clearFilters();
  }
}

// ---------------------------------------------------------------------------
// Header row: Title + Search + New Patient button
// ---------------------------------------------------------------------------

class _HeaderRow extends StatelessWidget {
  const _HeaderRow({required this.onSearchChanged});

  final ValueChanged<String> onSearchChanged;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Row(
      children: [
        Text(
          'Patients',
          style: TextStyle(
            fontSize: 24,
            fontWeight: FontWeight.w700,
            color: colorScheme.onSurface,
          ),
        ),
        const SizedBox(width: AppSpacing.lg),
        Expanded(
          child: SizedBox(
            height: 40,
            child: SearchField(
              hintText: 'Search patients by name, DOB...',
              shortcutHint: 'Ctrl+K',
              onChanged: onSearchChanged,
            ),
          ),
        ),
        const SizedBox(width: AppSpacing.md),
        FilledButton.icon(
          onPressed: () => GoRouter.of(context).go('/patients/new'),
          icon: const Icon(Icons.add, size: 18),
          label: const Text('New Patient'),
        ),
      ],
    );
  }
}

// ---------------------------------------------------------------------------
// Expandable filter panel
// ---------------------------------------------------------------------------

class _FilterPanel extends StatelessWidget {
  const _FilterPanel({
    required this.expanded,
    required this.onToggle,
    required this.selectedGender,
    required this.selectedStatus,
    required this.hasAlerts,
    required this.siteId,
    required this.birthDateFrom,
    required this.birthDateTo,
    required this.onGenderChanged,
    required this.onStatusChanged,
    required this.onHasAlertsChanged,
    required this.onSiteIdChanged,
    required this.onBirthDateFromChanged,
    required this.onBirthDateToChanged,
    required this.onApply,
    required this.onClear,
    required this.hasActiveFilters,
  });

  final bool expanded;
  final VoidCallback onToggle;
  final String? selectedGender;
  final String? selectedStatus;
  final bool? hasAlerts;
  final String? siteId;
  final DateTime? birthDateFrom;
  final DateTime? birthDateTo;
  final ValueChanged<String?> onGenderChanged;
  final ValueChanged<String?> onStatusChanged;
  final ValueChanged<bool?> onHasAlertsChanged;
  final ValueChanged<String?> onSiteIdChanged;
  final ValueChanged<DateTime?> onBirthDateFromChanged;
  final ValueChanged<DateTime?> onBirthDateToChanged;
  final VoidCallback onApply;
  final VoidCallback onClear;
  final bool hasActiveFilters;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(AppSpacing.borderRadiusMd),
        side: BorderSide(color: colorScheme.outline.withOpacity(0.3)),
      ),
      child: Column(
        children: [
          // ── Toggle Row ─────────────────────────────────────────────
          InkWell(
            onTap: onToggle,
            borderRadius: BorderRadius.circular(AppSpacing.borderRadiusMd),
            child: Padding(
              padding: const EdgeInsets.symmetric(
                horizontal: AppSpacing.md,
                vertical: AppSpacing.sm,
              ),
              child: Row(
                children: [
                  Icon(
                    Icons.filter_list,
                    size: 18,
                    color: hasActiveFilters
                        ? colorScheme.primary
                        : colorScheme.onSurfaceVariant,
                  ),
                  const SizedBox(width: AppSpacing.sm),
                  Text(
                    'Filters',
                    style: TextStyle(
                      fontSize: 13,
                      fontWeight: FontWeight.w500,
                      color: hasActiveFilters
                          ? colorScheme.primary
                          : colorScheme.onSurfaceVariant,
                    ),
                  ),
                  if (hasActiveFilters) ...[
                    const SizedBox(width: AppSpacing.xs),
                    Container(
                      width: 8,
                      height: 8,
                      decoration: BoxDecoration(
                        color: colorScheme.primary,
                        shape: BoxShape.circle,
                      ),
                    ),
                  ],
                  const Spacer(),
                  Icon(
                    expanded ? Icons.expand_less : Icons.expand_more,
                    size: 20,
                    color: colorScheme.onSurfaceVariant,
                  ),
                ],
              ),
            ),
          ),

          // ── Filter Controls ────────────────────────────────────────
          if (expanded)
            Padding(
              padding: const EdgeInsets.fromLTRB(
                AppSpacing.md,
                0,
                AppSpacing.md,
                AppSpacing.md,
              ),
              child: Column(
                children: [
                  const Divider(height: 1),
                  const SizedBox(height: AppSpacing.md),
                  Wrap(
                    spacing: AppSpacing.md,
                    runSpacing: AppSpacing.md,
                    children: [
                      // Gender
                      _FilterDropdown<String>(
                        label: 'Gender',
                        value: selectedGender,
                        items: const [
                          DropdownMenuItem(value: 'male', child: Text('Male')),
                          DropdownMenuItem(
                              value: 'female', child: Text('Female')),
                          DropdownMenuItem(
                              value: 'other', child: Text('Other')),
                          DropdownMenuItem(
                              value: 'unknown', child: Text('Unknown')),
                        ],
                        onChanged: onGenderChanged,
                      ),

                      // Status
                      _FilterDropdown<String>(
                        label: 'Status',
                        value: selectedStatus,
                        items: const [
                          DropdownMenuItem(
                              value: 'active', child: Text('Active')),
                          DropdownMenuItem(
                              value: 'inactive', child: Text('Inactive')),
                        ],
                        onChanged: onStatusChanged,
                      ),

                      // DOB From
                      _DateFilterButton(
                        label: 'DOB From',
                        value: birthDateFrom,
                        onChanged: onBirthDateFromChanged,
                      ),

                      // DOB To
                      _DateFilterButton(
                        label: 'DOB To',
                        value: birthDateTo,
                        onChanged: onBirthDateToChanged,
                      ),

                      // Site ID
                      SizedBox(
                        width: 160,
                        child: TextField(
                          decoration: InputDecoration(
                            labelText: 'Site ID',
                            isDense: true,
                            contentPadding: const EdgeInsets.symmetric(
                              horizontal: AppSpacing.sm,
                              vertical: AppSpacing.sm,
                            ),
                            border: OutlineInputBorder(
                              borderRadius: BorderRadius.circular(
                                  AppSpacing.borderRadiusMd),
                            ),
                          ),
                          style: const TextStyle(fontSize: 13),
                          onChanged: onSiteIdChanged,
                        ),
                      ),

                      // Has Alerts
                      Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Checkbox(
                            value: hasAlerts ?? false,
                            tristate: true,
                            onChanged: onHasAlertsChanged,
                          ),
                          const Text(
                            'Has Alerts',
                            style: TextStyle(fontSize: 13),
                          ),
                        ],
                      ),
                    ],
                  ),
                  const SizedBox(height: AppSpacing.md),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.end,
                    children: [
                      TextButton(
                        onPressed: onClear,
                        child: const Text('Clear All'),
                      ),
                      const SizedBox(width: AppSpacing.sm),
                      FilledButton(
                        onPressed: onApply,
                        child: const Text('Apply Filters'),
                      ),
                    ],
                  ),
                ],
              ),
            ),
        ],
      ),
    );
  }
}

class _FilterDropdown<T> extends StatelessWidget {
  const _FilterDropdown({
    required this.label,
    required this.value,
    required this.items,
    required this.onChanged,
  });

  final String label;
  final T? value;
  final List<DropdownMenuItem<T>> items;
  final ValueChanged<T?> onChanged;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: 140,
      child: DropdownButtonFormField<T>(
        value: value,
        decoration: InputDecoration(
          labelText: label,
          isDense: true,
          contentPadding: const EdgeInsets.symmetric(
            horizontal: AppSpacing.sm,
            vertical: AppSpacing.sm,
          ),
          border: OutlineInputBorder(
            borderRadius: BorderRadius.circular(AppSpacing.borderRadiusMd),
          ),
        ),
        style: const TextStyle(fontSize: 13),
        items: items,
        onChanged: onChanged,
      ),
    );
  }
}

class _DateFilterButton extends StatelessWidget {
  const _DateFilterButton({
    required this.label,
    required this.value,
    required this.onChanged,
  });

  final String label;
  final DateTime? value;
  final ValueChanged<DateTime?> onChanged;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return SizedBox(
      width: 160,
      child: OutlinedButton.icon(
        onPressed: () async {
          final picked = await showDatePicker(
            context: context,
            initialDate: value ?? DateTime(2000),
            firstDate: DateTime(1900),
            lastDate: DateTime.now(),
          );
          onChanged(picked);
        },
        icon: const Icon(Icons.calendar_today, size: 14),
        label: Text(
          value != null ? value!.toDisplayDate : label,
          style: TextStyle(
            fontSize: 12,
            color: value != null
                ? colorScheme.onSurface
                : colorScheme.onSurfaceVariant,
          ),
        ),
        style: OutlinedButton.styleFrom(
          padding: const EdgeInsets.symmetric(
            horizontal: AppSpacing.sm,
            vertical: AppSpacing.sm,
          ),
        ),
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Patient data table
// ---------------------------------------------------------------------------

class _PatientDataTable extends StatelessWidget {
  const _PatientDataTable({
    required this.patients,
    required this.isLoading,
  });

  final List<PatientSummary> patients;
  final bool isLoading;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Card(
      elevation: 1,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(AppSpacing.borderRadiusLg),
      ),
      child: Stack(
        children: [
          SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            child: SingleChildScrollView(
              child: DataTable(
                showCheckboxColumn: false,
                headingRowColor: WidgetStateProperty.all(
                  colorScheme.surfaceContainerHighest.withOpacity(0.5),
                ),
                columns: const [
                  DataColumn(label: Text('Name')),
                  DataColumn(label: Text('DOB')),
                  DataColumn(label: Text('Gender')),
                  DataColumn(label: Text('Site')),
                  DataColumn(label: Text('Last Updated')),
                  DataColumn(label: Text('Alerts')),
                ],
                rows: patients
                    .map((patient) => _buildRow(context, patient))
                    .toList(),
              ),
            ),
          ),

          // Loading overlay
          if (isLoading)
            Positioned.fill(
              child: Container(
                color: colorScheme.surface.withOpacity(0.5),
                child: const Center(
                  child: CircularProgressIndicator(),
                ),
              ),
            ),
        ],
      ),
    );
  }

  DataRow _buildRow(BuildContext context, PatientSummary patient) {
    final colorScheme = Theme.of(context).colorScheme;

    return DataRow(
      onSelectChanged: (_) {
        GoRouter.of(context).go('/patients/${patient.id}');
      },
      cells: [
        // Name
        DataCell(
          Text(
            patient.displayName.isNotEmpty
                ? patient.displayName
                : 'Unknown',
            style: TextStyle(
              fontWeight: FontWeight.w500,
              color: colorScheme.onSurface,
            ),
          ),
        ),

        // DOB
        DataCell(
          Text(
            patient.birthDate ?? '--',
            style: TextStyle(
              fontSize: 13,
              color: colorScheme.onSurfaceVariant,
            ),
          ),
        ),

        // Gender
        DataCell(
          Text(
            _capitalizeGender(patient.gender),
            style: TextStyle(
              fontSize: 13,
              color: colorScheme.onSurfaceVariant,
            ),
          ),
        ),

        // Site
        DataCell(
          Text(
            patient.siteId ?? '--',
            style: TextStyle(
              fontSize: 13,
              fontFamily: 'monospace',
              color: colorScheme.onSurfaceVariant,
            ),
          ),
        ),

        // Last Updated
        DataCell(
          Text(
            _formatLastUpdated(patient.lastUpdated),
            style: TextStyle(
              fontSize: 13,
              color: colorScheme.onSurfaceVariant,
            ),
          ),
        ),

        // Alert badge
        DataCell(
          patient.hasAlerts
              ? const SeverityBadge(severity: 'warning')
              : Text(
                  '--',
                  style: TextStyle(
                    fontSize: 13,
                    color: colorScheme.onSurfaceVariant,
                  ),
                ),
        ),
      ],
    );
  }

  static String _capitalizeGender(String? gender) {
    if (gender == null || gender.isEmpty) return '--';
    return gender[0].toUpperCase() + gender.substring(1);
  }

  static String _formatLastUpdated(String? lastUpdated) {
    if (lastUpdated == null) return '--';
    try {
      return DateTime.parse(lastUpdated).timeAgo;
    } catch (_) {
      return lastUpdated;
    }
  }
}
