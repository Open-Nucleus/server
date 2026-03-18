import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_spacing.dart';
import '../../../shared/widgets/data_table_card.dart';
import '../../../shared/widgets/error_state.dart';
import 'supply_providers.dart';

/// Supply chain screen showing inventory, predictions, and redistribution
/// suggestions.
///
/// This is an initial implementation; the main integration point for supply
/// data is the formulary stock panel.
class SupplyScreen extends ConsumerWidget {
  const SupplyScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final inventoryAsync = ref.watch(inventoryProvider);
    final predictionsAsync = ref.watch(predictionsProvider);
    final redistributionAsync = ref.watch(redistributionProvider);

    return Padding(
      padding: AppSpacing.pagePadding,
      child: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            // ── Inventory ────────────────────────────────────────────────
            inventoryAsync.when(
              loading: () => const SizedBox(
                height: 200,
                child: Center(child: CircularProgressIndicator()),
              ),
              error: (err, _) => ErrorState(
                message: 'Failed to load inventory',
                details: err.toString(),
                onRetry: () => ref.invalidate(inventoryProvider),
              ),
              data: (inventory) => DataTableCard(
                title: 'Inventory',
                emptyTitle: 'No inventory data',
                emptySubtitle:
                    'Inventory items will appear once supply tracking is active.',
                columns: const [
                  DataColumn(label: Text('Item')),
                  DataColumn(label: Text('Code')),
                  DataColumn(label: Text('Qty'), numeric: true),
                  DataColumn(label: Text('Unit')),
                  DataColumn(label: Text('Site')),
                  DataColumn(label: Text('Reorder Level'), numeric: true),
                  DataColumn(label: Text('Last Updated')),
                ],
                rows: inventory.items.map((item) {
                  final lowStock = item.quantity <= item.reorderLevel;
                  return DataRow(cells: [
                    DataCell(Text(
                      item.display,
                      style: const TextStyle(
                          fontWeight: FontWeight.w600, fontSize: 13),
                    )),
                    DataCell(Text(
                      item.itemCode,
                      style: const TextStyle(
                          fontFamily: 'monospace', fontSize: 12),
                    )),
                    DataCell(Text(
                      '${item.quantity}',
                      style: TextStyle(
                        fontSize: 13,
                        fontWeight: FontWeight.w600,
                        color: lowStock
                            ? AppColors.statusError
                            : AppColors.statusActive,
                      ),
                    )),
                    DataCell(Text(item.unit,
                        style: const TextStyle(fontSize: 12))),
                    DataCell(Text(item.siteId,
                        style: const TextStyle(fontSize: 12))),
                    DataCell(Text('${item.reorderLevel}',
                        style: const TextStyle(fontSize: 12))),
                    DataCell(Text(item.lastUpdated,
                        style: const TextStyle(fontSize: 12))),
                  ]);
                }).toList(),
              ),
            ),

            const SizedBox(height: AppSpacing.lg),

            // ── Predictions ──────────────────────────────────────────────
            predictionsAsync.when(
              loading: () => const SizedBox(
                height: 100,
                child: Center(child: CircularProgressIndicator()),
              ),
              error: (err, _) => ErrorState(
                message: 'Failed to load predictions',
                details: err.toString(),
                onRetry: () => ref.invalidate(predictionsProvider),
              ),
              data: (predictions) => DataTableCard(
                title: 'Stock-out Predictions',
                emptyTitle: 'No predictions',
                emptySubtitle:
                    'Predictions will appear once enough data is available.',
                columns: const [
                  DataColumn(label: Text('Item')),
                  DataColumn(label: Text('Code')),
                  DataColumn(
                      label: Text('Current Qty'), numeric: true),
                  DataColumn(
                      label: Text('Days Remaining'), numeric: true),
                  DataColumn(label: Text('Risk')),
                  DataColumn(label: Text('Action')),
                ],
                rows: predictions.predictions.map((p) {
                  return DataRow(cells: [
                    DataCell(Text(p.display,
                        style: const TextStyle(
                            fontWeight: FontWeight.w600,
                            fontSize: 13))),
                    DataCell(Text(p.itemCode,
                        style: const TextStyle(
                            fontFamily: 'monospace',
                            fontSize: 12))),
                    DataCell(Text('${p.currentQuantity}',
                        style: const TextStyle(fontSize: 13))),
                    DataCell(Text('${p.predictedDaysRemaining}',
                        style: const TextStyle(fontSize: 13))),
                    DataCell(_RiskBadge(riskLevel: p.riskLevel)),
                    DataCell(Text(p.recommendedAction,
                        style: const TextStyle(fontSize: 12))),
                  ]);
                }).toList(),
              ),
            ),

            const SizedBox(height: AppSpacing.lg),

            // ── Redistribution ───────────────────────────────────────────
            redistributionAsync.when(
              loading: () => const SizedBox(
                height: 100,
                child: Center(child: CircularProgressIndicator()),
              ),
              error: (err, _) => ErrorState(
                message: 'Failed to load redistribution suggestions',
                details: err.toString(),
                onRetry: () =>
                    ref.invalidate(redistributionProvider),
              ),
              data: (redistribution) => DataTableCard(
                title: 'Redistribution Suggestions',
                emptyTitle: 'No suggestions',
                emptySubtitle:
                    'Redistribution suggestions will appear when inter-site balancing is needed.',
                columns: const [
                  DataColumn(label: Text('Item Code')),
                  DataColumn(label: Text('From Site')),
                  DataColumn(label: Text('To Site')),
                  DataColumn(
                      label: Text('Suggested Qty'), numeric: true),
                  DataColumn(label: Text('Rationale')),
                ],
                rows: redistribution.suggestions.map((s) {
                  return DataRow(cells: [
                    DataCell(Text(s.itemCode,
                        style: const TextStyle(
                            fontFamily: 'monospace',
                            fontSize: 12))),
                    DataCell(Text(s.fromSite,
                        style: const TextStyle(fontSize: 13))),
                    DataCell(Text(s.toSite,
                        style: const TextStyle(fontSize: 13))),
                    DataCell(Text('${s.suggestedQuantity}',
                        style: const TextStyle(fontSize: 13))),
                    DataCell(Text(s.rationale,
                        style: const TextStyle(fontSize: 12))),
                  ]);
                }).toList(),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

// =============================================================================
// Shared Widgets
// =============================================================================

class _RiskBadge extends StatelessWidget {
  final String riskLevel;
  const _RiskBadge({required this.riskLevel});

  @override
  Widget build(BuildContext context) {
    final Color color;
    switch (riskLevel.toLowerCase()) {
      case 'critical':
      case 'high':
        color = AppColors.statusError;
        break;
      case 'medium':
      case 'moderate':
        color = AppColors.statusPending;
        break;
      case 'low':
        color = AppColors.statusActive;
        break;
      default:
        color = AppColors.statusInactive;
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: color.withOpacity(0.15),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withOpacity(0.4)),
      ),
      child: Text(
        riskLevel,
        style: TextStyle(
          fontSize: 11,
          fontWeight: FontWeight.w600,
          color: color,
        ),
      ),
    );
  }
}
