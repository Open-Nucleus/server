import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../core/theme/app_spacing.dart';

/// Displays a JSON value with syntax highlighting, collapsible sections for
/// objects and arrays, and a copy-to-clipboard button.
class JsonViewer extends StatelessWidget {
  const JsonViewer({
    required this.data,
    this.initiallyExpanded = true,
    super.key,
  });

  /// The JSON-compatible value to display: a [Map], [List], or primitive.
  final dynamic data;

  /// Whether top-level objects/arrays are initially expanded.
  final bool initiallyExpanded;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        // ── Copy Button ────────────────────────────────────────────────
        Align(
          alignment: Alignment.topRight,
          child: IconButton(
            icon: const Icon(Icons.copy, size: 18),
            tooltip: 'Copy JSON',
            onPressed: () {
              final text = const JsonEncoder.withIndent('  ').convert(data);
              Clipboard.setData(ClipboardData(text: text));
              ScaffoldMessenger.of(context)
                ..hideCurrentSnackBar()
                ..showSnackBar(
                  const SnackBar(
                    content: Text('JSON copied to clipboard'),
                    duration: Duration(seconds: 2),
                    behavior: SnackBarBehavior.floating,
                  ),
                );
            },
          ),
        ),

        // ── JSON Tree ──────────────────────────────────────────────────
        Container(
          padding: const EdgeInsets.all(AppSpacing.sm),
          decoration: BoxDecoration(
            color: colorScheme.surfaceContainerLowest,
            borderRadius: BorderRadius.circular(AppSpacing.borderRadiusMd),
            border: Border.all(color: colorScheme.outlineVariant),
          ),
          child: SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            child: _JsonNode(
              value: data,
              initiallyExpanded: initiallyExpanded,
            ),
          ),
        ),
      ],
    );
  }
}

// ─────────────────────────────────────────────────────────────────────────────
// Internal tree node
// ─────────────────────────────────────────────────────────────────────────────

class _JsonNode extends StatefulWidget {
  const _JsonNode({
    required this.value,
    this.keyName,
    this.initiallyExpanded = true,
  });

  final dynamic value;
  final String? keyName;
  final bool initiallyExpanded;

  @override
  State<_JsonNode> createState() => _JsonNodeState();
}

class _JsonNodeState extends State<_JsonNode> {
  late bool _expanded;

  @override
  void initState() {
    super.initState();
    _expanded = widget.initiallyExpanded;
  }

  @override
  Widget build(BuildContext context) {
    final value = widget.value;

    if (value is Map<String, dynamic>) {
      return _buildCollapsible(
        context,
        openBracket: '{',
        closeBracket: '}',
        children: value.entries
            .map((e) => _JsonNode(
                  keyName: e.key,
                  value: e.value,
                  initiallyExpanded: false,
                ))
            .toList(),
        childCount: value.length,
      );
    }

    if (value is List) {
      return _buildCollapsible(
        context,
        openBracket: '[',
        closeBracket: ']',
        children: value
            .asMap()
            .entries
            .map((e) => _JsonNode(
                  keyName: '${e.key}',
                  value: e.value,
                  initiallyExpanded: false,
                ))
            .toList(),
        childCount: value.length,
      );
    }

    // Primitive value
    return _buildPrimitive(context);
  }

  Widget _buildCollapsible(
    BuildContext context, {
    required String openBracket,
    required String closeBracket,
    required List<Widget> children,
    required int childCount,
  }) {
    final keyWidget = widget.keyName != null
        ? Text.rich(
            TextSpan(children: [
              TextSpan(
                text: '"${widget.keyName}"',
                style: const TextStyle(
                  color: Color(0xFF9C27B0),
                  fontFamily: 'monospace',
                  fontSize: 13,
                ),
              ),
              const TextSpan(
                text: ': ',
                style: TextStyle(
                  fontFamily: 'monospace',
                  fontSize: 13,
                ),
              ),
            ]),
          )
        : const SizedBox.shrink();

    if (!_expanded) {
      return InkWell(
        onTap: () => setState(() => _expanded = true),
        borderRadius: BorderRadius.circular(4),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.chevron_right, size: 16),
            keyWidget,
            Text(
              '$openBracket...$closeBracket ($childCount)',
              style: TextStyle(
                fontFamily: 'monospace',
                fontSize: 13,
                color: Theme.of(context).colorScheme.onSurfaceVariant,
              ),
            ),
          ],
        ),
      );
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        InkWell(
          onTap: () => setState(() => _expanded = false),
          borderRadius: BorderRadius.circular(4),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(Icons.expand_more, size: 16),
              keyWidget,
              Text(
                openBracket,
                style: const TextStyle(fontFamily: 'monospace', fontSize: 13),
              ),
            ],
          ),
        ),
        Padding(
          padding: const EdgeInsets.only(left: 20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: children,
          ),
        ),
        Text(
          closeBracket,
          style: const TextStyle(fontFamily: 'monospace', fontSize: 13),
        ),
      ],
    );
  }

  Widget _buildPrimitive(BuildContext context) {
    final value = widget.value;
    Color valueColor;
    String valueText;

    if (value is String) {
      valueColor = const Color(0xFF2E7D32); // green
      valueText = '"$value"';
    } else if (value is num) {
      valueColor = const Color(0xFF1565C0); // blue
      valueText = '$value';
    } else if (value is bool) {
      valueColor = const Color(0xFFE65100); // orange
      valueText = '$value';
    } else {
      valueColor = const Color(0xFF757575); // grey
      valueText = 'null';
    }

    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 1),
      child: Text.rich(
        TextSpan(
          children: [
            if (widget.keyName != null) ...[
              TextSpan(
                text: '"${widget.keyName}"',
                style: const TextStyle(
                  color: Color(0xFF9C27B0),
                  fontFamily: 'monospace',
                  fontSize: 13,
                ),
              ),
              const TextSpan(
                text: ': ',
                style: TextStyle(fontFamily: 'monospace', fontSize: 13),
              ),
            ],
            TextSpan(
              text: valueText,
              style: TextStyle(
                color: valueColor,
                fontFamily: 'monospace',
                fontSize: 13,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
