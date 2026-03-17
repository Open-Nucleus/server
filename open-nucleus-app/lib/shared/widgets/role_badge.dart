import 'package:flutter/material.dart';

import '../../core/constants/permissions.dart';

/// A coloured chip that displays the user's role with a role-specific colour.
class RoleBadge extends StatelessWidget {
  const RoleBadge({
    required this.role,
    super.key,
  });

  /// The role identifier matching one of [Permissions.allRoles].
  final String role;

  @override
  Widget build(BuildContext context) {
    final displayName = Permissions.roleDisplayName(role);
    final color = _colorForRole(role);

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: color.withOpacity(0.12),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withOpacity(0.4)),
      ),
      child: Text(
        displayName,
        style: TextStyle(
          fontSize: 11,
          fontWeight: FontWeight.w600,
          color: color,
        ),
      ),
    );
  }

  static Color _colorForRole(String role) {
    switch (role) {
      case Permissions.roleCHW:
        return const Color(0xFF2E7D32); // green
      case Permissions.roleNurse:
        return const Color(0xFF1565C0); // blue
      case Permissions.rolePhysician:
        return const Color(0xFF6A1B9A); // purple
      case Permissions.roleSiteAdmin:
        return const Color(0xFFE65100); // deep orange
      case Permissions.roleRegionalAdmin:
        return const Color(0xFFC62828); // red
      default:
        return const Color(0xFF757575); // grey
    }
  }
}
