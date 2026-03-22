import {
  createRouter,
  createRoute,
  createRootRoute,
  redirect,
} from '@tanstack/react-router';

import RootLayout from './App.tsx';
import LoginPage from './pages/login.tsx';
import AuthLayout from './pages/auth-layout.tsx';
import DashboardPage from './pages/dashboard.tsx';
import PatientsListPage from './pages/patients/index.tsx';
import PatientNewPage from './pages/patients/new.tsx';
import PatientDetailPage from './pages/patients/detail.tsx';
import PatientEditPage from './pages/patients/edit.tsx';
import FormularyPage from './pages/formulary.tsx';
import SyncPage from './pages/sync.tsx';
import AlertsPage from './pages/alerts.tsx';
import IntegrityPage from './pages/integrity.tsx';
import SettingsPage from './pages/settings.tsx';

import { useAuthStore } from './stores/auth-store.ts';

/* ================================================================
   Route definitions (code-based, no file-gen needed)
   ================================================================ */

// ---- Root ----
const rootRoute = createRootRoute({
  component: RootLayout,
});

// ---- Login (public) ----
const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/login',
  component: LoginPage,
});

// ---- Catch-all redirect: / -> /dashboard ----
const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  beforeLoad: () => {
    throw redirect({ to: '/dashboard' });
  },
});

// ---- Auth layout (guard) ----
const authRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: 'auth',
  beforeLoad: () => {
    const { token } = useAuthStore.getState();
    if (!token) {
      throw redirect({ to: '/login' });
    }
  },
  component: AuthLayout,
});

// ---- Dashboard ----
const dashboardRoute = createRoute({
  getParentRoute: () => authRoute,
  path: '/dashboard',
  component: DashboardPage,
});

// ---- Patients (list) ----
const patientsRoute = createRoute({
  getParentRoute: () => authRoute,
  path: '/patients',
  component: PatientsListPage,
});

// ---- Patient: new ----
const patientNewRoute = createRoute({
  getParentRoute: () => authRoute,
  path: '/patients/new',
  component: PatientNewPage,
});

// ---- Patient: detail ----
const patientDetailRoute = createRoute({
  getParentRoute: () => authRoute,
  path: '/patients/$id',
  component: PatientDetailPage,
});

// ---- Patient: edit ----
const patientEditRoute = createRoute({
  getParentRoute: () => authRoute,
  path: '/patients/$id/edit',
  component: PatientEditPage,
});

// ---- Formulary ----
const formularyRoute = createRoute({
  getParentRoute: () => authRoute,
  path: '/formulary',
  component: FormularyPage,
});

// ---- Sync ----
const syncRoute = createRoute({
  getParentRoute: () => authRoute,
  path: '/sync',
  component: SyncPage,
});

// ---- Alerts ----
const alertsRoute = createRoute({
  getParentRoute: () => authRoute,
  path: '/alerts',
  component: AlertsPage,
});

// ---- Integrity / Anchor ----
const integrityRoute = createRoute({
  getParentRoute: () => authRoute,
  path: '/integrity',
  component: IntegrityPage,
});

// ---- Settings ----
const settingsRoute = createRoute({
  getParentRoute: () => authRoute,
  path: '/settings',
  component: SettingsPage,
});

/* ================================================================
   Route tree
   ================================================================ */

const routeTree = rootRoute.addChildren([
  indexRoute,
  loginRoute,
  authRoute.addChildren([
    dashboardRoute,
    patientNewRoute,   // must precede $id to avoid shadowing
    patientDetailRoute,
    patientEditRoute,
    patientsRoute,
    formularyRoute,
    syncRoute,
    alertsRoute,
    integrityRoute,
    settingsRoute,
  ]),
]);

/* ================================================================
   Router instance
   ================================================================ */

export const router = createRouter({
  routeTree,
  defaultPreload: 'intent',
});

/* ---- Type registration for useNavigate / Link type-safety ---- */
declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}
