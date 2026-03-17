import 'package:flutter/material.dart';

import '../../core/theme/app_spacing.dart';

/// Pagination controls showing page info, previous/next buttons, and a
/// rows-per-page selector.
class PaginationControls extends StatelessWidget {
  const PaginationControls({
    required this.currentPage,
    required this.totalPages,
    required this.totalItems,
    required this.rowsPerPage,
    required this.onPageChanged,
    this.onRowsPerPageChanged,
    this.rowsPerPageOptions = const [10, 25, 50, 100],
    super.key,
  });

  final int currentPage;
  final int totalPages;
  final int totalItems;
  final int rowsPerPage;
  final ValueChanged<int> onPageChanged;
  final ValueChanged<int>? onRowsPerPageChanged;
  final List<int> rowsPerPageOptions;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final hasPrevious = currentPage > 1;
    final hasNext = currentPage < totalPages;

    return Padding(
      padding: const EdgeInsets.symmetric(
        horizontal: AppSpacing.md,
        vertical: AppSpacing.sm,
      ),
      child: Row(
        children: [
          // ── Page Info ──────────────────────────────────────────────
          Text(
            'Page $currentPage of $totalPages, $totalItems total',
            style: TextStyle(
              fontSize: 13,
              color: colorScheme.onSurfaceVariant,
            ),
          ),

          const Spacer(),

          // ── Rows Per Page ──────────────────────────────────────────
          if (onRowsPerPageChanged != null) ...[
            Text(
              'Rows per page:',
              style: TextStyle(
                fontSize: 13,
                color: colorScheme.onSurfaceVariant,
              ),
            ),
            const SizedBox(width: AppSpacing.xs),
            DropdownButton<int>(
              value: rowsPerPage,
              underline: const SizedBox.shrink(),
              isDense: true,
              items: rowsPerPageOptions
                  .map((n) => DropdownMenuItem(value: n, child: Text('$n')))
                  .toList(),
              onChanged: (value) {
                if (value != null) onRowsPerPageChanged!(value);
              },
            ),
            const SizedBox(width: AppSpacing.md),
          ],

          // ── Previous / Next ────────────────────────────────────────
          IconButton(
            icon: const Icon(Icons.chevron_left),
            onPressed: hasPrevious ? () => onPageChanged(currentPage - 1) : null,
            tooltip: 'Previous page',
            iconSize: 20,
          ),
          IconButton(
            icon: const Icon(Icons.chevron_right),
            onPressed: hasNext ? () => onPageChanged(currentPage + 1) : null,
            tooltip: 'Next page',
            iconSize: 20,
          ),
        ],
      ),
    );
  }
}
