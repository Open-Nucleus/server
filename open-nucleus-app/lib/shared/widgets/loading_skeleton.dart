import 'package:flutter/material.dart';

import '../../core/theme/app_spacing.dart';

/// Skeleton loading placeholder with a shimmer effect.
///
/// Provides three factory constructors for common layouts:
/// - [LoadingSkeleton.listTile] -- a list-item skeleton with avatar + lines
/// - [LoadingSkeleton.card] -- a card-shaped skeleton
/// - [LoadingSkeleton.table] -- a grid of rectangular cells
class LoadingSkeleton extends StatefulWidget {
  const LoadingSkeleton({
    this.width,
    this.height = 16,
    this.borderRadius = AppSpacing.borderRadiusSm,
    super.key,
  });

  /// A list-tile skeleton: a circle on the left with two text lines.
  factory LoadingSkeleton.listTile({int count = 3, Key? key}) {
    return _ListTileSkeleton(count: count, key: key);
  }

  /// A card-shaped skeleton.
  factory LoadingSkeleton.card({
    double width = double.infinity,
    double height = 120,
    Key? key,
  }) {
    return LoadingSkeleton(
      width: width,
      height: height,
      borderRadius: AppSpacing.borderRadiusLg,
      key: key,
    );
  }

  /// A table skeleton with [rows] rows and [cols] columns.
  factory LoadingSkeleton.table({int rows = 5, int cols = 4, Key? key}) {
    return _TableSkeleton(rows: rows, cols: cols, key: key);
  }

  final double? width;
  final double height;
  final double borderRadius;

  @override
  State<LoadingSkeleton> createState() => _LoadingSkeletonState();
}

class _LoadingSkeletonState extends State<LoadingSkeleton>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;
  late final Animation<double> _animation;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1500),
    )..repeat();
    _animation = Tween<double>(begin: -1.0, end: 2.0).animate(
      CurvedAnimation(parent: _controller, curve: Curves.easeInOut),
    );
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final baseColor = colorScheme.surfaceContainerHighest;
    final highlightColor = colorScheme.surface;

    return AnimatedBuilder(
      animation: _animation,
      builder: (context, child) {
        return Container(
          width: widget.width,
          height: widget.height,
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(widget.borderRadius),
            gradient: _shimmerGradient(
              _animation.value,
              baseColor,
              highlightColor,
            ),
          ),
        );
      },
    );
  }
}

// ─────────────────────────────────────────────────────────────────────────────
// Shared shimmer gradient helper
// ─────────────────────────────────────────────────────────────────────────────

LinearGradient _shimmerGradient(
  double animValue,
  Color baseColor,
  Color highlightColor,
) {
  return LinearGradient(
    begin: Alignment.centerLeft,
    end: Alignment.centerRight,
    colors: [baseColor, highlightColor, baseColor],
    stops: [
      (animValue - 0.3).clamp(0.0, 1.0),
      animValue.clamp(0.0, 1.0),
      (animValue + 0.3).clamp(0.0, 1.0),
    ],
  );
}

// ─────────────────────────────────────────────────────────────────────────────
// List-tile variant
// ─────────────────────────────────────────────────────────────────────────────

class _ListTileSkeleton extends LoadingSkeleton {
  final int count;
  const _ListTileSkeleton({required this.count, super.key})
      : super(height: 0);

  @override
  State<LoadingSkeleton> createState() => _ListTileSkeletonState();
}

class _ListTileSkeletonState extends State<_ListTileSkeleton>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;
  late final Animation<double> _animation;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1500),
    )..repeat();
    _animation = Tween<double>(begin: -1.0, end: 2.0).animate(
      CurvedAnimation(parent: _controller, curve: Curves.easeInOut),
    );
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final baseColor = colorScheme.surfaceContainerHighest;
    final highlightColor = colorScheme.surface;

    return Column(
      children: List.generate(
        (widget as _ListTileSkeleton).count,
        (index) => Padding(
          padding: const EdgeInsets.symmetric(
            vertical: AppSpacing.sm,
            horizontal: AppSpacing.md,
          ),
          child: _ShimmerRow(
            animation: _animation,
            baseColor: baseColor,
            highlightColor: highlightColor,
          ),
        ),
      ),
    );
  }
}

class _ShimmerRow extends AnimatedWidget {
  const _ShimmerRow({
    required Animation<double> animation,
    required this.baseColor,
    required this.highlightColor,
  }) : super(listenable: animation);

  final Color baseColor;
  final Color highlightColor;

  Animation<double> get _animation => listenable as Animation<double>;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        // Circle avatar placeholder
        Container(
          width: 40,
          height: 40,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            gradient: _shimmerGradient(_animation.value, baseColor, highlightColor),
          ),
        ),
        const SizedBox(width: AppSpacing.sm),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Container(
                width: double.infinity,
                height: 14,
                decoration: BoxDecoration(
                  borderRadius:
                      BorderRadius.circular(AppSpacing.borderRadiusSm),
                  gradient: _shimmerGradient(
                      _animation.value, baseColor, highlightColor),
                ),
              ),
              const SizedBox(height: AppSpacing.xs),
              Container(
                width: 160,
                height: 12,
                decoration: BoxDecoration(
                  borderRadius:
                      BorderRadius.circular(AppSpacing.borderRadiusSm),
                  gradient: _shimmerGradient(
                      _animation.value, baseColor, highlightColor),
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

// ─────────────────────────────────────────────────────────────────────────────
// Table variant
// ─────────────────────────────────────────────────────────────────────────────

class _TableSkeleton extends LoadingSkeleton {
  final int rows;
  final int cols;
  const _TableSkeleton({required this.rows, required this.cols, super.key})
      : super(height: 0);

  @override
  State<LoadingSkeleton> createState() => _TableSkeletonState();
}

class _TableSkeletonState extends State<_TableSkeleton>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;
  late final Animation<double> _animation;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1500),
    )..repeat();
    _animation = Tween<double>(begin: -1.0, end: 2.0).animate(
      CurvedAnimation(parent: _controller, curve: Curves.easeInOut),
    );
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final ts = widget as _TableSkeleton;
    final colorScheme = Theme.of(context).colorScheme;
    final baseColor = colorScheme.surfaceContainerHighest;
    final highlightColor = colorScheme.surface;

    return _TableShimmer(
      animation: _animation,
      rows: ts.rows,
      cols: ts.cols,
      baseColor: baseColor,
      highlightColor: highlightColor,
    );
  }
}

class _TableShimmer extends AnimatedWidget {
  const _TableShimmer({
    required Animation<double> animation,
    required this.rows,
    required this.cols,
    required this.baseColor,
    required this.highlightColor,
  }) : super(listenable: animation);

  final int rows;
  final int cols;
  final Color baseColor;
  final Color highlightColor;

  Animation<double> get _animation => listenable as Animation<double>;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.all(AppSpacing.md),
      child: Column(
        children: List.generate(
          rows,
          (r) => Padding(
            padding: const EdgeInsets.only(bottom: AppSpacing.sm),
            child: Row(
              children: List.generate(
                cols,
                (c) => Expanded(
                  child: Padding(
                    padding: EdgeInsets.only(
                      right: c < cols - 1 ? AppSpacing.sm : 0,
                    ),
                    child: Container(
                      height: 16,
                      decoration: BoxDecoration(
                        borderRadius: BorderRadius.circular(
                            AppSpacing.borderRadiusSm),
                        gradient: _shimmerGradient(
                            _animation.value, baseColor, highlightColor),
                      ),
                    ),
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
