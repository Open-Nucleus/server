import { Outlet } from '@tanstack/react-router';

/**
 * Authenticated layout shell.
 * The beforeLoad guard in the router handles redirect-to-login;
 * this component just renders the Outlet (future: sidebar + topbar wrapper).
 */
export default function AuthLayout() {
  return <Outlet />;
}
