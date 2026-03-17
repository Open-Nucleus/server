import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import '../../../shared/models/auth_models.dart';
import 'auth_api.dart';

/// Keys used in [FlutterSecureStorage] for auth persistence.
abstract final class _StorageKeys {
  static const accessToken = 'auth_access_token';
  static const refreshToken = 'auth_refresh_token';
  static const roleCode = 'auth_role_code';
  static const roleDisplay = 'auth_role_display';
  static const rolePermissions = 'auth_role_permissions';
  static const siteId = 'auth_site_id';
  static const nodeId = 'auth_node_id';
  static const expiresAt = 'auth_expires_at';
  static const practitionerId = 'auth_practitioner_id';
}

/// Wraps [AuthApi] and [FlutterSecureStorage] to provide a cohesive
/// auth persistence layer.
///
/// Tokens and role metadata are persisted to secure storage so that the
/// user can survive an app restart without re-authenticating (as long as
/// the token hasn't expired).
class AuthRepository {
  final AuthApi _api;
  final FlutterSecureStorage _storage;

  String? _accessToken;
  String? _refreshTokenValue;

  AuthRepository(this._api, this._storage);

  /// The in-memory access token, or `null` if not authenticated.
  String? get accessToken => _accessToken;

  /// The in-memory refresh token, or `null` if not authenticated.
  String? get refreshTokenValue => _refreshTokenValue;

  // ── Login ────────────────────────────────────────────────────────────

  /// Authenticates with the backend and persists the resulting tokens.
  ///
  /// Returns the [LoginResponse] on success.
  Future<LoginResponse> login(LoginRequest request) async {
    final envelope = await _api.login(request);
    final data = envelope.data!;

    _accessToken = data.token;
    _refreshTokenValue = data.refreshToken;

    await _persistAuth(data, request.practitionerId);

    return data;
  }

  // ── Logout ───────────────────────────────────────────────────────────

  /// Logs out on the server, then wipes all persisted auth data.
  Future<void> logout() async {
    if (_accessToken != null) {
      try {
        await _api.logout(_accessToken!);
      } catch (_) {
        // Best-effort: even if the server call fails we clear local state.
      }
    }

    _accessToken = null;
    _refreshTokenValue = null;
    await _clearStorage();
  }

  // ── Refresh ──────────────────────────────────────────────────────────

  /// Attempts to exchange the current refresh token for a new token pair.
  ///
  /// Returns the new access token on success, `null` on failure.
  Future<String?> refreshToken() async {
    final rt = _refreshTokenValue;
    if (rt == null || rt.isEmpty) return null;

    try {
      final envelope = await _api.refresh(rt);
      final data = envelope.data!;

      _accessToken = data.token;
      _refreshTokenValue = data.refreshToken;

      await _storage.write(key: _StorageKeys.accessToken, value: data.token);
      await _storage.write(
          key: _StorageKeys.refreshToken, value: data.refreshToken);
      await _storage.write(
          key: _StorageKeys.expiresAt, value: data.expiresAt);

      return data.token;
    } catch (_) {
      return null;
    }
  }

  // ── Restore from secure storage ──────────────────────────────────────

  /// Tries to load a previously saved auth session from secure storage.
  ///
  /// Returns a [LoginResponse] if tokens exist, `null` otherwise.
  Future<LoginResponse?> loadSavedAuth() async {
    final token = await _storage.read(key: _StorageKeys.accessToken);
    final refresh = await _storage.read(key: _StorageKeys.refreshToken);

    if (token == null || refresh == null) return null;

    _accessToken = token;
    _refreshTokenValue = refresh;

    final roleCode =
        await _storage.read(key: _StorageKeys.roleCode) ?? '';
    final roleDisplay =
        await _storage.read(key: _StorageKeys.roleDisplay) ?? '';
    final permsRaw =
        await _storage.read(key: _StorageKeys.rolePermissions) ?? '';
    final siteId = await _storage.read(key: _StorageKeys.siteId) ?? '';
    final nodeId = await _storage.read(key: _StorageKeys.nodeId) ?? '';
    final expiresAt =
        await _storage.read(key: _StorageKeys.expiresAt) ?? '';

    return LoginResponse(
      token: token,
      expiresAt: expiresAt,
      refreshToken: refresh,
      role: RoleDTO(
        code: roleCode,
        display: roleDisplay,
        permissions:
            permsRaw.isEmpty ? [] : permsRaw.split(','),
      ),
      siteId: siteId,
      nodeId: nodeId,
    );
  }

  /// Returns the stored access token directly from memory.
  String? getAccessToken() => _accessToken;

  // ── Private helpers ──────────────────────────────────────────────────

  Future<void> _persistAuth(
      LoginResponse data, String practitionerId) async {
    await _storage.write(
        key: _StorageKeys.accessToken, value: data.token);
    await _storage.write(
        key: _StorageKeys.refreshToken, value: data.refreshToken);
    await _storage.write(
        key: _StorageKeys.roleCode, value: data.role.code);
    await _storage.write(
        key: _StorageKeys.roleDisplay, value: data.role.display);
    await _storage.write(
        key: _StorageKeys.rolePermissions,
        value: data.role.permissions.join(','));
    await _storage.write(key: _StorageKeys.siteId, value: data.siteId);
    await _storage.write(key: _StorageKeys.nodeId, value: data.nodeId);
    await _storage.write(
        key: _StorageKeys.expiresAt, value: data.expiresAt);
    await _storage.write(
        key: _StorageKeys.practitionerId, value: practitionerId);
  }

  Future<void> _clearStorage() async {
    for (final key in [
      _StorageKeys.accessToken,
      _StorageKeys.refreshToken,
      _StorageKeys.roleCode,
      _StorageKeys.roleDisplay,
      _StorageKeys.rolePermissions,
      _StorageKeys.siteId,
      _StorageKeys.nodeId,
      _StorageKeys.expiresAt,
      _StorageKeys.practitionerId,
    ]) {
      await _storage.delete(key: key);
    }
  }
}
