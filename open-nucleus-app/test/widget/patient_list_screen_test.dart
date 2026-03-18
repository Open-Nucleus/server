import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:open_nucleus_app/features/patients/presentation/patient_list_screen.dart';
import 'package:open_nucleus_app/shared/providers/dio_provider.dart';

/// Creates a [Dio] instance with a mock base URL. In tests we don't actually
/// hit the network — we just need the widget tree to build without crashing.
/// The providers will enter loading/error state, which is fine for verifying
/// static UI elements.
Dio _mockDio() {
  return Dio(BaseOptions(baseUrl: 'http://localhost:0'));
}

Widget buildTestApp() {
  return ProviderScope(
    overrides: [
      dioProvider.overrideWithValue(_mockDio()),
    ],
    child: const MaterialApp(
      home: Scaffold(
        body: PatientListScreen(),
      ),
    ),
  );
}

void main() {
  group('PatientListScreen', () {
    testWidgets('renders "Patients" title', (tester) async {
      await tester.pumpWidget(buildTestApp());
      // Pump once to build the widget tree (don't pumpAndSettle as network
      // calls will never complete).
      await tester.pump();

      expect(find.text('Patients'), findsOneWidget);
    });

    testWidgets('renders "New Patient" button', (tester) async {
      await tester.pumpWidget(buildTestApp());
      await tester.pump();

      expect(find.text('New Patient'), findsOneWidget);
    });

    testWidgets('renders search field', (tester) async {
      await tester.pumpWidget(buildTestApp());
      await tester.pump();

      // The SearchField widget uses a TextField with a hint
      expect(
        find.widgetWithText(TextField, 'Search patients by name, DOB...'),
        findsOneWidget,
      );
    });

    testWidgets('renders "New Patient" as a FilledButton with add icon',
        (tester) async {
      await tester.pumpWidget(buildTestApp());
      await tester.pump();

      final buttonFinder = find.widgetWithText(FilledButton, 'New Patient');
      expect(buttonFinder, findsOneWidget);

      // The button should also have an add icon
      expect(find.byIcon(Icons.add), findsOneWidget);
    });

    testWidgets('renders filters section', (tester) async {
      await tester.pumpWidget(buildTestApp());
      await tester.pump();

      expect(find.text('Filters'), findsOneWidget);
    });
  });
}
