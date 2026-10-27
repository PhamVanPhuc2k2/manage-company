import { create } from "zustand";

export type CurrentUser = {
  id: string;
  employee_id: string;
  email: string;
  full_name: string;
  role: string[];
  permissions: string[];
  scope: string;
  organization_id?: string;
  organization_name?: string;
};

type AuthState = {
  accessToken: string | null;
  expiresAt: number | null;
  user: CurrentUser | null;

  loading: boolean;

  setToken: (token: string, expires: string | Date) => void;
  setUser: (user: CurrentUser | null) => void;
  setLoading: (loading: boolean) => void;
  clear: () => void;
};

export const useAuthStore = create<AuthState>((set) => ({
  accessToken: null,
  expiresAt: null,
  user: null,
  loading: true,

  setToken: (token, expires) =>
    set({ accessToken: token, expiresAt: new Date(expires).getTime() }),
  setUser: (user) => set({ user }),
  setLoading: (loading) => set({ loading }),
  clear: () => set({ accessToken: null, expiresAt: null, user: null }),
}));

export function isTokenExpiringSoon(): boolean {
  const { accessToken, expiresAt } = useAuthStore.getState();
  if (!accessToken || !expiresAt) return true;
  return Date.now() >= expiresAt - 60_000;
}
