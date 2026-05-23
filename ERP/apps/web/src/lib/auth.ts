import { create } from "zustand";

export interface AuthUser {
  id: string;
  full_name: string;
  role: string;
  location_id: string;
}

interface AuthState {
  token: string | null;
  user: AuthUser | null;
  setAuth: (token: string, user: AuthUser) => void;
  logout: () => void;
}

function loadUser(): AuthUser | null {
  try {
    const raw = localStorage.getItem("cs_user");
    return raw ? JSON.parse(raw) : null;
  } catch {
    return null;
  }
}

export const useAuth = create<AuthState>((set) => ({
  token: localStorage.getItem("cs_token"),
  user: loadUser(),
  setAuth: (token, user) => {
    localStorage.setItem("cs_token", token);
    localStorage.setItem("cs_user", JSON.stringify(user));
    set({ token, user });
  },
  logout: () => {
    localStorage.removeItem("cs_token");
    localStorage.removeItem("cs_user");
    set({ token: null, user: null });
  },
}));
