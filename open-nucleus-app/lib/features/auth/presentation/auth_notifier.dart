import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../shared/models/auth_models.dart';
import '../data/auth_repository.dart';

// ---------------------------------------------------------------------------
// AuthState
// ---------------------------------------------------------------------------

/// Sealed state hierarchy for the auth lifecycle.
///
/// [AuthInitial] — app just started, no auth check yet.
/// [AuthLoading] — login / refresh / check in progress.
/// [Authenticated] — user is logged in with valid tokens.
/// [AuthError] — login or refresh failed.
abstract class AuthState {
  const AuthState();
}

class AuthInitial extends AuthState {
  const AuthInitial();
}

class AuthLoading extends AuthState {
  const AuthLoading();
}

class Authenticated extends AuthState {
  final LoginResponse loginResponse;
  final String keypairFingerprint;

  const Authenticated({
    required this.loginResponse,
    required this.keypairFingerprint,
  });
}

class AuthError extends AuthState {
  final String message;

  const AuthError(this.message);
}

// ---------------------------------------------------------------------------
// AuthNotifier
// ---------------------------------------------------------------------------

/// Manages authentication state and token lifecycle.
///
/// Talks to [AuthRepository] for persistence and API calls.
class AuthNotifier extends StateNotifier<AuthState> {
  final AuthRepository _repository;

  AuthNotifier(this._repository) : super(const AuthInitial());

  /// Convenience accessor used by [AuthInterceptor] to inject the bearer
  /// token into outgoing requests.
  String? get accessToken => _repository.getAccessToken();

  // ── Login ────────────────────────────────────────────────────────────

  /// Performs a full login flow: sends credentials, persists tokens, and
  /// transitions to [Authenticated] on success.
  Future<void> login({
    required LoginRequest request,
    required String keypairFingerprint,
  }) async {
    state = const AuthLoading();

    try {
      final response = await _repository.login(request);

      state = Authenticated(
        loginResponse: response,
        keypairFingerprint: keypairFingerprint,
      );
    } catch (e) {
      state = AuthError(e.toString());
    }
  }

  // ── Logout ───────────────────────────────────────────────────────────

  /// Logs out on the server and resets to [AuthInitial].
  Future<void> logout() async {
    state = const AuthLoading();

    try {
      await _repository.logout();
    } catch (_) {
      // Best-effort server logout; local state is cleared regardless.
    }

    state = const AuthInitial();
  }

  // ── Check existing auth ──────────────────────────────────────────────

  /// Loads a previously saved session from secure storage.
  ///
  /// If tokens exist, transitions to [Authenticated]; otherwise stays
  /// at [AuthInitial].
  Future<void> checkAuth({required String keypairFingerprint}) async {
    state = const AuthLoading();

    try {
      final saved = await _repository.loadSavedAuth();

      if (saved != null) {
        state = Authenticated(
          loginResponse: saved,
          keypairFingerprint: keypairFingerprint,
        );
      } else {
        state = const AuthInitial();
      }
    } catch (e) {
      state = AuthError(e.toString());
    }
  }

  // ── Token refresh ────────────────────────────────────────────────────

  /// Attempts to refresh the access token using the stored refresh token.
  ///
  /// Returns `true` if the refresh succeeded.
  Future<bool> refreshToken() async {
    try {
      final newToken = await _repository.refreshToken();
      return newToken != null;
    } catch (_) {
      return false;
    }
  }
}
