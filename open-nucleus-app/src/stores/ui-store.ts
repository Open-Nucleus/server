import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export interface UIState {
  sidebarExpanded: boolean;
  toggleSidebar: () => void;
  pageTitle: string;
  setPageTitle: (title: string) => void;
  theme: 'light' | 'dark';
  toggleTheme: () => void;
  unacknowledgedAlerts: number;
  setAlertCount: (count: number) => void;
}

/** Apply or remove the "dark" class on <html>. */
function applyThemeClass(theme: 'light' | 'dark') {
  if (theme === 'dark') {
    document.documentElement.classList.add('dark');
  } else {
    document.documentElement.classList.remove('dark');
  }
}

export const useUIStore = create<UIState>()(
  persist(
    (set, get) => ({
      sidebarExpanded: true,
      toggleSidebar: () => set({ sidebarExpanded: !get().sidebarExpanded }),

      pageTitle: 'Dashboard',
      setPageTitle: (title: string) => set({ pageTitle: title }),

      theme: 'light',
      toggleTheme: () => {
        const next = get().theme === 'light' ? 'dark' : 'light';
        applyThemeClass(next);
        set({ theme: next });
      },

      unacknowledgedAlerts: 0,
      setAlertCount: (count: number) => set({ unacknowledgedAlerts: count }),
    }),
    {
      name: 'nucleus-ui',
      partialize: (state) => ({
        theme: state.theme,
        sidebarExpanded: state.sidebarExpanded,
      }),
      onRehydrateStorage: () => {
        return (state) => {
          if (state) {
            applyThemeClass(state.theme);
          }
        };
      },
    },
  ),
);
