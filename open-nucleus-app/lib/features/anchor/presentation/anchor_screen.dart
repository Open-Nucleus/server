import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_spacing.dart';
import '../../../shared/models/anchor_models.dart';
import '../../../shared/widgets/data_table_card.dart';
import '../../../shared/widgets/error_state.dart';
import '../../../shared/widgets/json_viewer.dart';
import 'anchor_providers.dart';

/// Anchor / Integrity screen with status cards, history, DID, credentials,
/// and backend tabs.
class AnchorScreen extends ConsumerStatefulWidget {
  const AnchorScreen({super.key});

  @override
  ConsumerState<AnchorScreen> createState() => _AnchorScreenState();
}

class _AnchorScreenState extends ConsumerState<AnchorScreen>
    with SingleTickerProviderStateMixin {
  late final TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 4, vsync: this);
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: AppSpacing.pagePadding,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          // ── Status Cards ───────────────────────────────────────────
          _buildStatusSection(),
          const SizedBox(height: AppSpacing.md),

          // ── Tabs ───────────────────────────────────────────────────
          TabBar(
            controller: _tabController,
            tabs: const [
              Tab(text: 'History'),
              Tab(text: 'DID'),
              Tab(text: 'Credentials'),
              Tab(text: 'Backends'),
            ],
          ),
          const SizedBox(height: AppSpacing.sm),

          // ── Tab Content ────────────────────────────────────────────
          Expanded(
            child: TabBarView(
              controller: _tabController,
              children: [
                _HistoryTab(),
                _DIDTab(),
                _CredentialsTab(),
                _BackendsTab(),
              ],
            ),
          ),
        ],
      ),
    );
  }

  // ── Status Section ──────────────────────────────────────────────────────

  Widget _buildStatusSection() {
    final statusAsync = ref.watch(anchorStatusProvider);

    return statusAsync.when(
      loading: () => const SizedBox(
        height: 100,
        child: Center(child: CircularProgressIndicator()),
      ),
      error: (err, _) => SizedBox(
        height: 100,
        child: ErrorState(
          message: 'Failed to load anchor status',
          details: err.toString(),
          onRetry: () => ref.invalidate(anchorStatusProvider),
        ),
      ),
      data: (status) => Row(
        children: [
          Expanded(child: _StatusCard.state(status.state)),
          const SizedBox(width: AppSpacing.sm),
          Expanded(child: _StatusCard.merkleRoot(context, status.merkleRoot)),
          const SizedBox(width: AppSpacing.sm),
          Expanded(
            child: _StatusCard.lastAnchor(
              status.lastAnchorTime,
              status.lastAnchorId,
            ),
          ),
          const SizedBox(width: AppSpacing.sm),
          Expanded(
            child: _StatusCard.queue(
              status.queueDepth,
              status.pendingCommits,
            ),
          ),
        ],
      ),
    );
  }
}

// =============================================================================
// Status Cards
// =============================================================================

class _StatusCard extends StatelessWidget {
  final String title;
  final Widget content;

  const _StatusCard({required this.title, required this.content});

  factory _StatusCard.state(String state) {
    final Color color;
    switch (state.toLowerCase()) {
      case 'active':
      case 'anchored':
        color = AppColors.statusActive;
        break;
      case 'error':
      case 'failed':
        color = AppColors.statusError;
        break;
      default:
        color = AppColors.statusInactive;
    }

    return _StatusCard(
      title: 'State',
      content: Row(
        children: [
          Container(
            width: 12,
            height: 12,
            decoration: BoxDecoration(
              color: color,
              shape: BoxShape.circle,
              boxShadow: [
                BoxShadow(
                  color: color.withOpacity(0.4),
                  blurRadius: 4,
                  spreadRadius: 1,
                ),
              ],
            ),
          ),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              state,
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w600,
                color: color,
              ),
              overflow: TextOverflow.ellipsis,
            ),
          ),
        ],
      ),
    );
  }

  factory _StatusCard.merkleRoot(BuildContext context, String merkleRoot) {
    final display = merkleRoot.length > 16
        ? '${merkleRoot.substring(0, 16)}...'
        : merkleRoot;

    return _StatusCard(
      title: 'Merkle Root',
      content: Row(
        children: [
          Expanded(
            child: Text(
              display,
              style: const TextStyle(
                fontSize: 13,
                fontFamily: 'monospace',
                fontWeight: FontWeight.w500,
              ),
              overflow: TextOverflow.ellipsis,
            ),
          ),
          InkWell(
            onTap: () {
              Clipboard.setData(ClipboardData(text: merkleRoot));
              ScaffoldMessenger.of(context)
                ..hideCurrentSnackBar()
                ..showSnackBar(
                  const SnackBar(
                    content: Text('Merkle root copied'),
                    duration: Duration(seconds: 2),
                    behavior: SnackBarBehavior.floating,
                  ),
                );
            },
            child: const Icon(Icons.copy, size: 16),
          ),
        ],
      ),
    );
  }

  factory _StatusCard.lastAnchor(String timestamp, String anchorId) {
    final display =
        anchorId.length > 12 ? '${anchorId.substring(0, 12)}...' : anchorId;

    return _StatusCard(
      title: 'Last Anchor',
      content: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            timestamp,
            style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w500),
            overflow: TextOverflow.ellipsis,
          ),
          const SizedBox(height: 2),
          Text(
            display,
            style: const TextStyle(
              fontSize: 11,
              fontFamily: 'monospace',
              color: AppColors.statusInactive,
            ),
            overflow: TextOverflow.ellipsis,
          ),
        ],
      ),
    );
  }

  factory _StatusCard.queue(int queueDepth, int pendingCommits) {
    return _StatusCard(
      title: 'Queue',
      content: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            '$queueDepth queued',
            style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w500),
          ),
          const SizedBox(height: 2),
          Text(
            '$pendingCommits pending commits',
            style: const TextStyle(
              fontSize: 11,
              color: AppColors.statusInactive,
            ),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Card(
      elevation: 1,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(AppSpacing.borderRadiusLg),
      ),
      child: Padding(
        padding: AppSpacing.cardPadding,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              title,
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.w500,
                color: colorScheme.onSurfaceVariant,
              ),
            ),
            const SizedBox(height: AppSpacing.sm),
            content,
          ],
        ),
      ),
    );
  }
}

// =============================================================================
// History Tab
// =============================================================================

class _HistoryTab extends ConsumerWidget {
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final historyAsync = ref.watch(anchorHistoryProvider);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        // ── Verify Commit Button ─────────────────────────────────────
        Align(
          alignment: Alignment.topRight,
          child: Padding(
            padding: const EdgeInsets.only(bottom: AppSpacing.sm),
            child: FilledButton.icon(
              onPressed: () => _showVerifyCommitDialog(context, ref),
              icon: const Icon(Icons.verified_user, size: 18),
              label: const Text('Verify Commit'),
            ),
          ),
        ),

        // ── History Table ────────────────────────────────────────────
        Expanded(
          child: historyAsync.when(
            loading: () =>
                const Center(child: CircularProgressIndicator()),
            error: (err, _) => ErrorState(
              message: 'Failed to load anchor history',
              details: err.toString(),
              onRetry: () => ref.invalidate(anchorHistoryProvider),
            ),
            data: (history) => SingleChildScrollView(
              child: DataTableCard(
                title: 'Anchor History',
                emptyTitle: 'No anchors yet',
                emptySubtitle:
                    'Anchor records will appear here after the first anchoring.',
                columns: const [
                  DataColumn(label: Text('Anchor ID')),
                  DataColumn(label: Text('Merkle Root')),
                  DataColumn(label: Text('Git Head')),
                  DataColumn(label: Text('State')),
                  DataColumn(label: Text('Timestamp')),
                  DataColumn(label: Text('Backend')),
                  DataColumn(label: Text('Tx ID')),
                ],
                rows: history.records.map((r) {
                  return DataRow(cells: [
                    DataCell(Text(
                      _truncate(r.anchorId, 12),
                      style: const TextStyle(
                          fontFamily: 'monospace', fontSize: 12),
                    )),
                    DataCell(Text(
                      _truncate(r.merkleRoot, 12),
                      style: const TextStyle(
                          fontFamily: 'monospace', fontSize: 12),
                    )),
                    DataCell(Text(
                      _truncate(r.gitHead, 12),
                      style: const TextStyle(
                          fontFamily: 'monospace', fontSize: 12),
                    )),
                    DataCell(_StateBadge(state: r.state)),
                    DataCell(Text(r.timestamp, style: const TextStyle(fontSize: 12))),
                    DataCell(Text(r.backend, style: const TextStyle(fontSize: 12))),
                    DataCell(Text(
                      _truncate(r.txId, 12),
                      style: const TextStyle(
                          fontFamily: 'monospace', fontSize: 12),
                    )),
                  ]);
                }).toList(),
              ),
            ),
          ),
        ),
      ],
    );
  }

  Future<void> _showVerifyCommitDialog(
      BuildContext context, WidgetRef ref) async {
    final controller = TextEditingController();
    AnchorVerifyResponse? result;
    String? errorMsg;

    await showDialog(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setDialogState) {
          return AlertDialog(
            shape: RoundedRectangleBorder(
              borderRadius:
                  BorderRadius.circular(AppSpacing.borderRadiusLg),
            ),
            title: const Text('Verify Commit'),
            content: SizedBox(
              width: 480,
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  TextField(
                    controller: controller,
                    decoration: const InputDecoration(
                      labelText: 'Commit Hash',
                      hintText: 'Enter a git commit hash',
                      border: OutlineInputBorder(),
                    ),
                    style: const TextStyle(
                        fontFamily: 'monospace', fontSize: 13),
                  ),
                  if (result != null) ...[
                    const SizedBox(height: AppSpacing.md),
                    _VerifyResult(result: result!),
                  ],
                  if (errorMsg != null) ...[
                    const SizedBox(height: AppSpacing.md),
                    Text(
                      errorMsg!,
                      style: TextStyle(
                        color: Theme.of(ctx).colorScheme.error,
                        fontSize: 13,
                      ),
                    ),
                  ],
                ],
              ),
            ),
            actions: [
              TextButton(
                onPressed: () => Navigator.of(ctx).pop(),
                child: const Text('Close'),
              ),
              FilledButton(
                onPressed: () async {
                  final hash = controller.text.trim();
                  if (hash.isEmpty) return;

                  try {
                    final api = ref.read(anchorApiProvider);
                    final envelope = await api.verify(hash);
                    setDialogState(() {
                      result = envelope.data;
                      errorMsg = null;
                    });
                  } catch (e) {
                    setDialogState(() {
                      result = null;
                      errorMsg = e.toString();
                    });
                  }
                },
                child: const Text('Verify'),
              ),
            ],
          );
        },
      ),
    );
  }
}

// =============================================================================
// DID Tab
// =============================================================================

class _DIDTab extends ConsumerStatefulWidget {
  @override
  ConsumerState<_DIDTab> createState() => _DIDTabState();
}

class _DIDTabState extends ConsumerState<_DIDTab> {
  final _resolveController = TextEditingController();
  DIDDocumentResponse? _resolvedDID;
  String? _resolveError;
  bool _resolving = false;

  @override
  void dispose() {
    _resolveController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final nodeDIDAsync = ref.watch(nodeDIDProvider);
    final colorScheme = Theme.of(context).colorScheme;

    return SingleChildScrollView(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          // ── Node DID ─────────────────────────────────────────────────
          Card(
            elevation: 1,
            shape: RoundedRectangleBorder(
              borderRadius:
                  BorderRadius.circular(AppSpacing.borderRadiusLg),
            ),
            child: Padding(
              padding: AppSpacing.cardPadding,
              child: nodeDIDAsync.when(
                loading: () => const SizedBox(
                  height: 200,
                  child: Center(child: CircularProgressIndicator()),
                ),
                error: (err, _) => ErrorState(
                  message: 'Failed to load Node DID',
                  details: err.toString(),
                  onRetry: () => ref.invalidate(nodeDIDProvider),
                ),
                data: (did) => Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Node DID Document',
                      style: TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.w600,
                        color: colorScheme.onSurface,
                      ),
                    ),
                    const SizedBox(height: AppSpacing.sm),

                    // DID string (copyable)
                    Row(
                      children: [
                        const Text(
                          'DID: ',
                          style: TextStyle(
                            fontSize: 13,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                        Expanded(
                          child: SelectableText(
                            did.id,
                            style: const TextStyle(
                              fontFamily: 'monospace',
                              fontSize: 13,
                            ),
                          ),
                        ),
                        IconButton(
                          icon: const Icon(Icons.copy, size: 16),
                          tooltip: 'Copy DID',
                          onPressed: () {
                            Clipboard.setData(ClipboardData(text: did.id));
                            ScaffoldMessenger.of(context)
                              ..hideCurrentSnackBar()
                              ..showSnackBar(
                                const SnackBar(
                                  content: Text('DID copied to clipboard'),
                                  duration: Duration(seconds: 2),
                                  behavior: SnackBarBehavior.floating,
                                ),
                              );
                          },
                        ),
                      ],
                    ),
                    const SizedBox(height: AppSpacing.sm),

                    // Verification methods
                    if (did.verificationMethod.isNotEmpty) ...[
                      Text(
                        'Verification Methods',
                        style: TextStyle(
                          fontSize: 14,
                          fontWeight: FontWeight.w600,
                          color: colorScheme.onSurface,
                        ),
                      ),
                      const SizedBox(height: AppSpacing.xs),
                      ...did.verificationMethod.map((vm) => Padding(
                            padding: const EdgeInsets.only(
                                bottom: AppSpacing.xs),
                            child: Card(
                              elevation: 0,
                              color: colorScheme.surfaceContainerHighest
                                  .withOpacity(0.3),
                              child: Padding(
                                padding: const EdgeInsets.all(
                                    AppSpacing.sm),
                                child: Column(
                                  crossAxisAlignment:
                                      CrossAxisAlignment.start,
                                  children: [
                                    _kvRow('ID', vm.id),
                                    _kvRow('Type', vm.type),
                                    _kvRow('Controller', vm.controller),
                                    _kvRow('Public Key',
                                        vm.publicKeyMultibase),
                                  ],
                                ),
                              ),
                            ),
                          )),
                    ],

                    const SizedBox(height: AppSpacing.sm),

                    // Full JSON
                    ExpansionTile(
                      title: const Text(
                        'Full DID Document (JSON)',
                        style: TextStyle(fontSize: 14),
                      ),
                      children: [
                        JsonViewer(data: did.toJson()),
                      ],
                    ),
                  ],
                ),
              ),
            ),
          ),

          const SizedBox(height: AppSpacing.md),

          // ── Resolve DID ──────────────────────────────────────────────
          Card(
            elevation: 1,
            shape: RoundedRectangleBorder(
              borderRadius:
                  BorderRadius.circular(AppSpacing.borderRadiusLg),
            ),
            child: Padding(
              padding: AppSpacing.cardPadding,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Resolve DID',
                    style: TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.w600,
                      color: colorScheme.onSurface,
                    ),
                  ),
                  const SizedBox(height: AppSpacing.sm),
                  Row(
                    children: [
                      Expanded(
                        child: TextField(
                          controller: _resolveController,
                          decoration: const InputDecoration(
                            labelText: 'DID String',
                            hintText: 'did:nucleus:...',
                            border: OutlineInputBorder(),
                          ),
                          style: const TextStyle(
                            fontFamily: 'monospace',
                            fontSize: 13,
                          ),
                        ),
                      ),
                      const SizedBox(width: AppSpacing.sm),
                      FilledButton(
                        onPressed: _resolving ? null : _resolveDID,
                        child: _resolving
                            ? const SizedBox(
                                width: 18,
                                height: 18,
                                child: CircularProgressIndicator(
                                    strokeWidth: 2),
                              )
                            : const Text('Resolve'),
                      ),
                    ],
                  ),
                  if (_resolveError != null) ...[
                    const SizedBox(height: AppSpacing.sm),
                    Text(
                      _resolveError!,
                      style: TextStyle(
                        color: colorScheme.error,
                        fontSize: 13,
                      ),
                    ),
                  ],
                  if (_resolvedDID != null) ...[
                    const SizedBox(height: AppSpacing.md),
                    JsonViewer(data: _resolvedDID!.toJson()),
                  ],
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Future<void> _resolveDID() async {
    final did = _resolveController.text.trim();
    if (did.isEmpty) return;

    setState(() {
      _resolving = true;
      _resolveError = null;
      _resolvedDID = null;
    });

    try {
      final api = ref.read(anchorApiProvider);
      final envelope = await api.resolveDID(did);
      setState(() {
        _resolvedDID = envelope.data;
        _resolving = false;
      });
    } catch (e) {
      setState(() {
        _resolveError = e.toString();
        _resolving = false;
      });
    }
  }

  Widget _kvRow(String key, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 1),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 100,
            child: Text(
              key,
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.w600,
                color: Theme.of(context).colorScheme.onSurfaceVariant,
              ),
            ),
          ),
          Expanded(
            child: SelectableText(
              value,
              style: const TextStyle(
                fontFamily: 'monospace',
                fontSize: 12,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

// =============================================================================
// Credentials Tab
// =============================================================================

class _CredentialsTab extends ConsumerWidget {
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final credAsync = ref.watch(credentialListProvider);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        // ── Action Buttons ──────────────────────────────────────────
        Padding(
          padding: const EdgeInsets.only(bottom: AppSpacing.sm),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.end,
            children: [
              FilledButton.icon(
                onPressed: () =>
                    _showIssueCredentialDialog(context, ref),
                icon: const Icon(Icons.add, size: 18),
                label: const Text('Issue Credential'),
              ),
              const SizedBox(width: AppSpacing.sm),
              OutlinedButton.icon(
                onPressed: () =>
                    _showVerifyCredentialDialog(context, ref),
                icon: const Icon(Icons.verified, size: 18),
                label: const Text('Verify Credential'),
              ),
            ],
          ),
        ),

        // ── Credential Table ────────────────────────────────────────
        Expanded(
          child: credAsync.when(
            loading: () =>
                const Center(child: CircularProgressIndicator()),
            error: (err, _) => ErrorState(
              message: 'Failed to load credentials',
              details: err.toString(),
              onRetry: () => ref.invalidate(credentialListProvider),
            ),
            data: (credList) => SingleChildScrollView(
              child: DataTableCard(
                title: 'Verifiable Credentials',
                emptyTitle: 'No credentials issued',
                emptySubtitle:
                    'Issue a credential to create an integrity proof.',
                columns: const [
                  DataColumn(label: Text('ID')),
                  DataColumn(label: Text('Type')),
                  DataColumn(label: Text('Issuer')),
                  DataColumn(label: Text('Issuance Date')),
                  DataColumn(label: Text('Expiration')),
                ],
                rows: credList.credentials.map((c) {
                  return DataRow(cells: [
                    DataCell(Text(
                      _truncate(c.id, 16),
                      style: const TextStyle(
                          fontFamily: 'monospace', fontSize: 12),
                    )),
                    DataCell(Text(
                      c.type.join(', '),
                      style: const TextStyle(fontSize: 12),
                    )),
                    DataCell(Text(
                      _truncate(c.issuer, 20),
                      style: const TextStyle(fontSize: 12),
                    )),
                    DataCell(Text(
                      c.issuanceDate,
                      style: const TextStyle(fontSize: 12),
                    )),
                    DataCell(Text(
                      c.expirationDate ?? '-',
                      style: const TextStyle(fontSize: 12),
                    )),
                  ]);
                }).toList(),
              ),
            ),
          ),
        ),
      ],
    );
  }

  Future<void> _showIssueCredentialDialog(
      BuildContext context, WidgetRef ref) async {
    final anchorIdController = TextEditingController();
    final typesController = TextEditingController();
    final claimsController = TextEditingController();
    CredentialResponse? result;
    String? errorMsg;

    await showDialog(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setDialogState) {
          return AlertDialog(
            shape: RoundedRectangleBorder(
              borderRadius:
                  BorderRadius.circular(AppSpacing.borderRadiusLg),
            ),
            title: const Text('Issue Credential'),
            content: SizedBox(
              width: 480,
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  TextField(
                    controller: anchorIdController,
                    decoration: const InputDecoration(
                      labelText: 'Anchor ID *',
                      hintText: 'Enter the anchor ID',
                      border: OutlineInputBorder(),
                    ),
                  ),
                  const SizedBox(height: AppSpacing.sm),
                  TextField(
                    controller: typesController,
                    decoration: const InputDecoration(
                      labelText: 'Types (comma-separated, optional)',
                      hintText: 'e.g. IntegrityCredential',
                      border: OutlineInputBorder(),
                    ),
                  ),
                  const SizedBox(height: AppSpacing.sm),
                  TextField(
                    controller: claimsController,
                    maxLines: 3,
                    decoration: const InputDecoration(
                      labelText: 'Additional Claims JSON (optional)',
                      hintText: '{"key": "value"}',
                      border: OutlineInputBorder(),
                    ),
                    style: const TextStyle(
                        fontFamily: 'monospace', fontSize: 13),
                  ),
                  if (result != null) ...[
                    const SizedBox(height: AppSpacing.md),
                    SizedBox(
                      height: 200,
                      child: SingleChildScrollView(
                        child: JsonViewer(data: result!.toJson()),
                      ),
                    ),
                  ],
                  if (errorMsg != null) ...[
                    const SizedBox(height: AppSpacing.md),
                    Text(
                      errorMsg!,
                      style: TextStyle(
                        color: Theme.of(ctx).colorScheme.error,
                        fontSize: 13,
                      ),
                    ),
                  ],
                ],
              ),
            ),
            actions: [
              TextButton(
                onPressed: () => Navigator.of(ctx).pop(),
                child: const Text('Close'),
              ),
              FilledButton(
                onPressed: () async {
                  final anchorId = anchorIdController.text.trim();
                  if (anchorId.isEmpty) return;

                  final typesRaw = typesController.text.trim();
                  final types = typesRaw.isNotEmpty
                      ? typesRaw
                          .split(',')
                          .map((t) => t.trim())
                          .where((t) => t.isNotEmpty)
                          .toList()
                      : null;

                  Map<String, String>? claims;
                  final claimsRaw = claimsController.text.trim();
                  if (claimsRaw.isNotEmpty) {
                    try {
                      final parsed =
                          json.decode(claimsRaw) as Map<String, dynamic>;
                      claims =
                          parsed.map((k, v) => MapEntry(k, v.toString()));
                    } catch (_) {
                      setDialogState(
                          () => errorMsg = 'Invalid JSON in claims field');
                      return;
                    }
                  }

                  try {
                    final api = ref.read(anchorApiProvider);
                    final envelope = await api.issueCredential(
                      IssueCredentialRequest(
                        anchorId: anchorId,
                        types: types,
                        additionalClaims: claims,
                      ),
                    );
                    setDialogState(() {
                      result = envelope.data;
                      errorMsg = null;
                    });
                    ref.invalidate(credentialListProvider);
                  } catch (e) {
                    setDialogState(() {
                      result = null;
                      errorMsg = e.toString();
                    });
                  }
                },
                child: const Text('Issue'),
              ),
            ],
          );
        },
      ),
    );
  }

  Future<void> _showVerifyCredentialDialog(
      BuildContext context, WidgetRef ref) async {
    final controller = TextEditingController();
    CredentialVerificationResponse? result;
    String? errorMsg;

    await showDialog(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setDialogState) {
          return AlertDialog(
            shape: RoundedRectangleBorder(
              borderRadius:
                  BorderRadius.circular(AppSpacing.borderRadiusLg),
            ),
            title: const Text('Verify Credential'),
            content: SizedBox(
              width: 520,
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  TextField(
                    controller: controller,
                    maxLines: 10,
                    decoration: const InputDecoration(
                      labelText: 'Credential JSON',
                      hintText: 'Paste the full credential JSON here',
                      border: OutlineInputBorder(),
                    ),
                    style: const TextStyle(
                        fontFamily: 'monospace', fontSize: 12),
                  ),
                  if (result != null) ...[
                    const SizedBox(height: AppSpacing.md),
                    Card(
                      color: result!.valid
                          ? AppColors.statusActive.withOpacity(0.1)
                          : AppColors.statusError.withOpacity(0.1),
                      child: Padding(
                        padding: AppSpacing.cardPadding,
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Row(
                              children: [
                                Icon(
                                  result!.valid
                                      ? Icons.verified
                                      : Icons.error,
                                  color: result!.valid
                                      ? AppColors.statusActive
                                      : AppColors.statusError,
                                  size: 20,
                                ),
                                const SizedBox(width: 8),
                                Text(
                                  result!.valid
                                      ? 'Credential Verified'
                                      : 'Verification Failed',
                                  style: TextStyle(
                                    fontWeight: FontWeight.w600,
                                    color: result!.valid
                                        ? AppColors.statusActive
                                        : AppColors.statusError,
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: AppSpacing.xs),
                            Text('Issuer: ${result!.issuer}',
                                style: const TextStyle(fontSize: 13)),
                            Text('Message: ${result!.message}',
                                style: const TextStyle(fontSize: 13)),
                          ],
                        ),
                      ),
                    ),
                  ],
                  if (errorMsg != null) ...[
                    const SizedBox(height: AppSpacing.md),
                    Text(
                      errorMsg!,
                      style: TextStyle(
                        color: Theme.of(ctx).colorScheme.error,
                        fontSize: 13,
                      ),
                    ),
                  ],
                ],
              ),
            ),
            actions: [
              TextButton(
                onPressed: () => Navigator.of(ctx).pop(),
                child: const Text('Close'),
              ),
              FilledButton(
                onPressed: () async {
                  final rawJson = controller.text.trim();
                  if (rawJson.isEmpty) return;

                  Map<String, dynamic> credJson;
                  try {
                    credJson =
                        json.decode(rawJson) as Map<String, dynamic>;
                  } catch (_) {
                    setDialogState(
                        () => errorMsg = 'Invalid JSON format');
                    return;
                  }

                  try {
                    final api = ref.read(anchorApiProvider);
                    final envelope =
                        await api.verifyCredential(credJson);
                    setDialogState(() {
                      result = envelope.data;
                      errorMsg = null;
                    });
                  } catch (e) {
                    setDialogState(() {
                      result = null;
                      errorMsg = e.toString();
                    });
                  }
                },
                child: const Text('Verify'),
              ),
            ],
          );
        },
      ),
    );
  }
}

// =============================================================================
// Backends Tab
// =============================================================================

class _BackendsTab extends ConsumerWidget {
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final backendsAsync = ref.watch(backendListProvider);

    return backendsAsync.when(
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (err, _) => ErrorState(
        message: 'Failed to load backends',
        details: err.toString(),
        onRetry: () => ref.invalidate(backendListProvider),
      ),
      data: (backendList) => SingleChildScrollView(
        child: DataTableCard(
          title: 'Anchor Backends',
          emptyTitle: 'No backends configured',
          emptySubtitle:
              'Backend integrations will appear here when configured.',
          columns: const [
            DataColumn(label: Text('Name')),
            DataColumn(label: Text('Available')),
            DataColumn(label: Text('Description')),
          ],
          rows: backendList.backends.map((b) {
            return DataRow(cells: [
              DataCell(Text(
                b.name,
                style: const TextStyle(
                    fontWeight: FontWeight.w600, fontSize: 13),
              )),
              DataCell(_AvailableBadge(available: b.available)),
              DataCell(Text(b.description,
                  style: const TextStyle(fontSize: 13))),
            ]);
          }).toList(),
        ),
      ),
    );
  }
}

// =============================================================================
// Shared Widgets
// =============================================================================

class _StateBadge extends StatelessWidget {
  final String state;
  const _StateBadge({required this.state});

  @override
  Widget build(BuildContext context) {
    final Color color;
    switch (state.toLowerCase()) {
      case 'anchored':
      case 'confirmed':
      case 'active':
        color = AppColors.statusActive;
        break;
      case 'pending':
      case 'processing':
        color = AppColors.statusPending;
        break;
      case 'error':
      case 'failed':
        color = AppColors.statusError;
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
        state,
        style: TextStyle(
          fontSize: 11,
          fontWeight: FontWeight.w600,
          color: color,
        ),
      ),
    );
  }
}

class _AvailableBadge extends StatelessWidget {
  final bool available;
  const _AvailableBadge({required this.available});

  @override
  Widget build(BuildContext context) {
    final color = available ? AppColors.statusActive : AppColors.statusError;
    final label = available ? 'Available' : 'Unavailable';

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: color.withOpacity(0.15),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withOpacity(0.4)),
      ),
      child: Text(
        label,
        style: TextStyle(
          fontSize: 11,
          fontWeight: FontWeight.w600,
          color: color,
        ),
      ),
    );
  }
}

class _VerifyResult extends StatelessWidget {
  final AnchorVerifyResponse result;
  const _VerifyResult({required this.result});

  @override
  Widget build(BuildContext context) {
    final color =
        result.verified ? AppColors.statusActive : AppColors.statusError;

    return Card(
      color: color.withOpacity(0.1),
      child: Padding(
        padding: AppSpacing.cardPadding,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(
                  result.verified ? Icons.verified : Icons.error,
                  color: color,
                  size: 20,
                ),
                const SizedBox(width: 8),
                Text(
                  result.verified ? 'Verified' : 'Not Verified',
                  style: TextStyle(
                    fontWeight: FontWeight.w600,
                    color: color,
                  ),
                ),
              ],
            ),
            const SizedBox(height: AppSpacing.xs),
            Text('Anchor ID: ${result.anchorId}',
                style: const TextStyle(fontSize: 13)),
            Text('Merkle Root: ${_truncate(result.merkleRoot, 24)}',
                style: const TextStyle(
                    fontSize: 13, fontFamily: 'monospace')),
            Text('Anchored At: ${result.anchoredAt}',
                style: const TextStyle(fontSize: 13)),
          ],
        ),
      ),
    );
  }
}

/// Truncates a string to the given length, appending "..." if truncated.
String _truncate(String s, int maxLen) {
  if (s.length <= maxLen) return s;
  return '${s.substring(0, maxLen)}...';
}
