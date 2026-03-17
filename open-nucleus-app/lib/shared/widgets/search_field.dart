import 'dart:async';

import 'package:flutter/material.dart';

import '../../core/theme/app_spacing.dart';

/// A text field with a search icon, debounced [onChanged] callback, and
/// optional keyboard shortcut hint.
///
/// The [onChanged] fires 300 ms after the user stops typing. A clear button
/// appears when the field has text.
class SearchField extends StatefulWidget {
  const SearchField({
    this.hintText = 'Search...',
    this.onChanged,
    this.shortcutHint,
    this.debounceMs = 300,
    this.controller,
    super.key,
  });

  final String hintText;
  final ValueChanged<String>? onChanged;

  /// Optional hint shown at the end of the field, e.g. "Ctrl+K".
  final String? shortcutHint;
  final int debounceMs;
  final TextEditingController? controller;

  @override
  State<SearchField> createState() => _SearchFieldState();
}

class _SearchFieldState extends State<SearchField> {
  late final TextEditingController _controller;
  Timer? _debounce;

  @override
  void initState() {
    super.initState();
    _controller = widget.controller ?? TextEditingController();
  }

  @override
  void dispose() {
    _debounce?.cancel();
    if (widget.controller == null) _controller.dispose();
    super.dispose();
  }

  void _onTextChanged(String value) {
    _debounce?.cancel();
    _debounce = Timer(Duration(milliseconds: widget.debounceMs), () {
      widget.onChanged?.call(value);
    });
    setState(() {}); // Refresh clear button visibility
  }

  void _clear() {
    _controller.clear();
    widget.onChanged?.call('');
    setState(() {});
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return TextField(
      controller: _controller,
      onChanged: _onTextChanged,
      style: const TextStyle(fontSize: 14),
      decoration: InputDecoration(
        hintText: widget.hintText,
        hintStyle: TextStyle(
          fontSize: 13,
          color: colorScheme.onSurfaceVariant,
        ),
        prefixIcon: Icon(
          Icons.search,
          size: 20,
          color: colorScheme.onSurfaceVariant,
        ),
        suffixIcon: _controller.text.isNotEmpty
            ? IconButton(
                icon: Icon(
                  Icons.clear,
                  size: 18,
                  color: colorScheme.onSurfaceVariant,
                ),
                onPressed: _clear,
              )
            : widget.shortcutHint != null
                ? Padding(
                    padding: const EdgeInsets.only(right: AppSpacing.sm),
                    child: Center(
                      widthFactor: 1.0,
                      child: Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 6,
                          vertical: 2,
                        ),
                        decoration: BoxDecoration(
                          color: colorScheme.surfaceContainerHighest,
                          borderRadius: BorderRadius.circular(4),
                        ),
                        child: Text(
                          widget.shortcutHint!,
                          style: TextStyle(
                            fontSize: 11,
                            color: colorScheme.onSurfaceVariant,
                          ),
                        ),
                      ),
                    ),
                  )
                : null,
        isDense: true,
        contentPadding: const EdgeInsets.symmetric(
          vertical: AppSpacing.sm,
          horizontal: AppSpacing.sm,
        ),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(AppSpacing.borderRadiusMd),
          borderSide: BorderSide(color: colorScheme.outline),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(AppSpacing.borderRadiusMd),
          borderSide: BorderSide(color: colorScheme.outline),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(AppSpacing.borderRadiusMd),
          borderSide: BorderSide(color: colorScheme.primary, width: 2),
        ),
      ),
    );
  }
}
