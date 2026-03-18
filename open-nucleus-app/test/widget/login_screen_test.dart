import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:open_nucleus_app/features/auth/presentation/login_screen.dart';

void main() {
  Widget buildTestApp() {
    return const ProviderScope(
      child: MaterialApp(
        home: LoginScreen(),
      ),
    );
  }

  group('LoginScreen', () {
    testWidgets('renders "Open Nucleus" title', (tester) async {
      await tester.pumpWidget(buildTestApp());
      await tester.pumpAndSettle();

      expect(find.text('Open Nucleus'), findsOneWidget);
    });

    testWidgets('renders "Electronic Health Record" subtitle', (tester) async {
      await tester.pumpWidget(buildTestApp());
      await tester.pumpAndSettle();

      expect(find.text('Electronic Health Record'), findsOneWidget);
    });

    testWidgets('renders server URL text field', (tester) async {
      await tester.pumpWidget(buildTestApp());
      await tester.pumpAndSettle();

      expect(find.widgetWithText(TextField, 'Server URL'), findsOneWidget);
    });

    testWidgets('renders practitioner ID text field', (tester) async {
      await tester.pumpWidget(buildTestApp());
      await tester.pumpAndSettle();

      expect(
          find.widgetWithText(TextField, 'Practitioner ID'), findsOneWidget);
    });

    testWidgets('renders login button', (tester) async {
      await tester.pumpWidget(buildTestApp());
      await tester.pumpAndSettle();

      expect(find.widgetWithText(FilledButton, 'Login'), findsOneWidget);
    });

    testWidgets('login button is disabled when practitioner ID is empty',
        (tester) async {
      await tester.pumpWidget(buildTestApp());
      await tester.pumpAndSettle();

      // Find the FilledButton
      final loginButton = tester.widget<FilledButton>(
        find.widgetWithText(FilledButton, 'Login'),
      );

      // onPressed should be null (disabled) since no connection test has
      // been performed and practitioner ID is empty.
      expect(loginButton.onPressed, isNull);
    });

    testWidgets('renders connection test button', (tester) async {
      await tester.pumpWidget(buildTestApp());
      await tester.pumpAndSettle();

      expect(find.text('Test'), findsOneWidget);
    });

    testWidgets('renders device keypair section', (tester) async {
      await tester.pumpWidget(buildTestApp());
      await tester.pumpAndSettle();

      expect(find.text('Device Keypair'), findsOneWidget);
    });

    testWidgets('server URL defaults to https://localhost:8080',
        (tester) async {
      await tester.pumpWidget(buildTestApp());
      await tester.pumpAndSettle();

      final serverUrlField = tester.widget<TextField>(
        find.widgetWithText(TextField, 'Server URL'),
      );

      expect(serverUrlField.controller?.text, equals('https://localhost:8080'));
    });
  });
}
