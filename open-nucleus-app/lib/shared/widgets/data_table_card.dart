import 'package:flutter/material.dart';

import '../../core/theme/app_spacing.dart';
import 'empty_state.dart';

/// A [Card] wrapper around a [DataTable] with an optional title bar,
/// search field, and action buttons.
///
/// When [rows] is empty, an [EmptyState] widget is displayed instead of
/// the data table.
class DataTableCard extends StatelessWidget {
  const DataTableCard({
    required this.columns,
    required this.rows,
    this.title,
    this.searchHint,
    this.onSearchChanged,
    this.actions,
    this.emptyIcon = Icons.table_chart_outlined,
    this.emptyTitle = 'No data',
    this.emptySubtitle,
    this.sortColumnIndex,
    this.sortAscending = true,
    this.onSort,
    this.showCheckboxColumn = false,
    super.key,
  });

  final List<DataColumn> columns;
  final List<DataRow> rows;
  final String? title;
  final String? searchHint;
  final ValueChanged<String>? onSearchChanged;
  final List<Widget>? actions;
  final IconData emptyIcon;
  final String emptyTitle;
  final String? emptySubtitle;
  final int? sortColumnIndex;
  final bool sortAscending;
  final DataColumnSortCallback? onSort;
  final bool showCheckboxColumn;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final hasHeader =
        title != null || searchHint != null || (actions?.isNotEmpty ?? false);

    return Card(
      elevation: 1,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(AppSpacing.borderRadiusLg),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          // ── Header Row ─────────────────────────────────────────────
          if (hasHeader)
            Padding(
              padding: const EdgeInsets.fromLTRB(
                AppSpacing.md,
                AppSpacing.md,
                AppSpacing.md,
                AppSpacing.sm,
              ),
              child: Row(
                children: [
                  if (title != null)
                    Text(
                      title!,
                      style: TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.w600,
                        color: colorScheme.onSurface,
                      ),
                    ),
                  if (title != null) const SizedBox(width: AppSpacing.md),
                  if (searchHint != null)
                    Expanded(
                      child: SizedBox(
                        height: 36,
                        child: TextField(
                          onChanged: onSearchChanged,
                          decoration: InputDecoration(
                            hintText: searchHint,
                            hintStyle: TextStyle(
                              fontSize: 13,
                              color: colorScheme.onSurfaceVariant,
                            ),
                            prefixIcon: Icon(
                              Icons.search,
                              size: 18,
                              color: colorScheme.onSurfaceVariant,
                            ),
                            isDense: true,
                            contentPadding: const EdgeInsets.symmetric(
                              vertical: AppSpacing.xs,
                              horizontal: AppSpacing.sm,
                            ),
                            border: OutlineInputBorder(
                              borderRadius: BorderRadius.circular(
                                  AppSpacing.borderRadiusMd),
                              borderSide:
                                  BorderSide(color: colorScheme.outline),
                            ),
                          ),
                        ),
                      ),
                    )
                  else
                    const Spacer(),
                  if (actions != null) ...[
                    const SizedBox(width: AppSpacing.sm),
                    ...actions!,
                  ],
                ],
              ),
            ),

          // ── Table or Empty State ───────────────────────────────────
          if (rows.isEmpty)
            Padding(
              padding: const EdgeInsets.all(AppSpacing.xl),
              child: EmptyState(
                icon: emptyIcon,
                title: emptyTitle,
                subtitle: emptySubtitle,
              ),
            )
          else
            SingleChildScrollView(
              scrollDirection: Axis.horizontal,
              child: DataTable(
                columns: columns,
                rows: rows,
                sortColumnIndex: sortColumnIndex,
                sortAscending: sortAscending,
                showCheckboxColumn: showCheckboxColumn,
                headingRowColor: WidgetStateProperty.all(
                  colorScheme.surfaceContainerHighest.withOpacity(0.5),
                ),
              ),
            ),
        ],
      ),
    );
  }
}
