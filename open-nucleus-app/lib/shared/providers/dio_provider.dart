import 'dart:io';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/config/app_config.dart';
import '../../features/auth/presentation/auth_providers.dart';
import '../models/app_exception.dart';

// ---------------------------------------------------------------------------
// AppConfig provider
// ---------------------------------------------------------------------------

/// Provides the current [AppConfig]. Override this in tests or at startup
/// to point at a different server URL.
final appConfigProvider = Provider<AppConfig>((_) => const AppConfig());

// ---------------------------------------------------------------------------
// Dio provider
// ---------------------------------------------------------------------------

/// Creates a fully configured [Dio] instance with auth, error mapping,
/// logging, and retry interceptors.
final dioProvider = Provider<Dio>((ref) {
  final config = ref.watch(appConfigProvider);

  final dio = Dio(
    BaseOptions(
      baseUrl: config.serverUrl,
      connectTimeout: const Duration(seconds: 10),
      receiveTimeout: const Duration(seconds: 30),
      headers: {
        HttpHeaders.contentTypeHeader: ContentType.json.toString(),
        HttpHeaders.acceptHeader: ContentType.json.toString(),
      },
    ),
  );

  // Order matters: interceptors run in the order they are added for requests,
  // and in reverse order for responses / errors.
  dio.interceptors.addAll([
    AuthInterceptor(ref),
    RetryInterceptor(dio),
    LoggingInterceptor(),
    ErrorInterceptor(),
  ]);

  return dio;
});

// ---------------------------------------------------------------------------
// 1. AuthInterceptor
// ---------------------------------------------------------------------------

/// Injects `Authorization: Bearer <token>` into every outgoing request.
/// On a 401 response it attempts a single token refresh and retries.
class AuthInterceptor extends Interceptor {
  final Ref _ref;

  AuthInterceptor(this._ref);

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    final notifier = _ref.read(authNotifierProvider.notifier);
    final token = notifier.accessToken;

    if (token != null && token.isNotEmpty) {
      options.headers[HttpHeaders.authorizationHeader] = 'Bearer $token';
    }

    handler.next(options);
  }

  @override
  void onError(DioException err, ErrorInterceptorHandler handler) async {
    if (err.response?.statusCode == 401) {
      // Attempt a single refresh.
      final notifier = _ref.read(authNotifierProvider.notifier);
      final refreshed = await notifier.refreshToken();

      if (refreshed) {
        // Retry the original request with the new token.
        final opts = err.requestOptions;
        opts.headers[HttpHeaders.authorizationHeader] =
            'Bearer ${notifier.accessToken}';

        try {
          final cloneResponse = await Dio(
            BaseOptions(
              baseUrl: opts.baseUrl,
              connectTimeout: opts.connectTimeout,
              receiveTimeout: opts.receiveTimeout,
            ),
          ).fetch(opts);
          return handler.resolve(cloneResponse);
        } on DioException catch (retryErr) {
          return handler.next(retryErr);
        }
      }
    }

    handler.next(err);
  }
}

// ---------------------------------------------------------------------------
// 2. ErrorInterceptor
// ---------------------------------------------------------------------------

/// Maps [DioException]s to [AppException] so callers can handle errors
/// uniformly without depending on Dio types.
class ErrorInterceptor extends Interceptor {
  @override
  void onError(DioException err, ErrorInterceptorHandler handler) {
    final response = err.response;

    String code;
    String message;
    int? statusCode = response?.statusCode;
    dynamic details;

    // Try to extract the backend error envelope first.
    if (response?.data is Map<String, dynamic>) {
      final body = response!.data as Map<String, dynamic>;
      final errorBody = body['error'] as Map<String, dynamic>?;

      if (errorBody != null) {
        code = (errorBody['code'] as String?) ?? _codeFromType(err.type);
        message = (errorBody['message'] as String?) ?? err.message ?? 'Unknown error';
        details = errorBody['details'];
      } else {
        code = _codeFromType(err.type);
        message = err.message ?? 'Unknown error';
      }
    } else {
      code = _codeFromType(err.type);
      message = err.message ?? 'Unknown error';
    }

    final appException = AppException(
      code: code,
      message: message,
      statusCode: statusCode,
      details: details,
    );

    handler.next(
      DioException(
        requestOptions: err.requestOptions,
        response: err.response,
        type: err.type,
        error: appException,
        message: appException.message,
      ),
    );
  }

  static String _codeFromType(DioExceptionType type) {
    switch (type) {
      case DioExceptionType.connectionTimeout:
        return 'CONNECTION_TIMEOUT';
      case DioExceptionType.sendTimeout:
        return 'SEND_TIMEOUT';
      case DioExceptionType.receiveTimeout:
        return 'RECEIVE_TIMEOUT';
      case DioExceptionType.connectionError:
        return 'CONNECTION_ERROR';
      case DioExceptionType.cancel:
        return 'REQUEST_CANCELLED';
      case DioExceptionType.badCertificate:
        return 'BAD_CERTIFICATE';
      case DioExceptionType.badResponse:
        return 'BAD_RESPONSE';
      case DioExceptionType.unknown:
        return 'UNKNOWN';
    }
  }
}

// ---------------------------------------------------------------------------
// 3. LoggingInterceptor
// ---------------------------------------------------------------------------

/// Prints request method + path and response status to the console.
class LoggingInterceptor extends Interceptor {
  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    print('[HTTP] --> ${options.method} ${options.path}');
    handler.next(options);
  }

  @override
  void onResponse(Response response, ResponseInterceptorHandler handler) {
    print(
      '[HTTP] <-- ${response.statusCode} ${response.requestOptions.method} '
      '${response.requestOptions.path}',
    );
    handler.next(response);
  }

  @override
  void onError(DioException err, ErrorInterceptorHandler handler) {
    print(
      '[HTTP] <-- ERROR ${err.response?.statusCode ?? 'N/A'} '
      '${err.requestOptions.method} ${err.requestOptions.path}: '
      '${err.message}',
    );
    handler.next(err);
  }
}

// ---------------------------------------------------------------------------
// 4. RetryInterceptor
// ---------------------------------------------------------------------------

/// Retries requests that fail due to connection timeout, up to [maxRetries].
class RetryInterceptor extends Interceptor {
  final Dio _dio;
  final int maxRetries;

  RetryInterceptor(this._dio, {this.maxRetries = 2});

  @override
  void onError(DioException err, ErrorInterceptorHandler handler) async {
    if (err.type == DioExceptionType.connectionTimeout) {
      final retryCount =
          (err.requestOptions.extra['_retryCount'] as int?) ?? 0;

      if (retryCount < maxRetries) {
        err.requestOptions.extra['_retryCount'] = retryCount + 1;

        print(
          '[HTTP] Retry ${retryCount + 1}/$maxRetries for '
          '${err.requestOptions.method} ${err.requestOptions.path}',
        );

        try {
          final response = await _dio.fetch(err.requestOptions);
          return handler.resolve(response);
        } on DioException catch (retryErr) {
          return handler.next(retryErr);
        }
      }
    }

    handler.next(err);
  }
}
