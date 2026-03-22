import { AppShell } from '@/features/shell/app-shell';

/**
 * Authenticated layout — wraps all protected pages in the sidebar + topbar shell.
 * The beforeLoad guard in the router handles redirect-to-login.
 */
export default function AuthLayout() {
  return <AppShell />;
}
