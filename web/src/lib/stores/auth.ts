import { writable } from 'svelte/store';
import { auth as authApi } from '$lib/api/endpoints';
import type { User } from '$lib/api/types';
import { UnauthenticatedError } from '$lib/api/client';

interface AuthState {
  user: User | null;
  loading: boolean;
}

const store = writable<AuthState>({ user: null, loading: true });

export const authStore = {
  subscribe: store.subscribe,

  async refresh() {
    try {
      const { user } = await authApi.me();
      store.set({ user, loading: false });
    } catch (err) {
      if (err instanceof UnauthenticatedError) {
        store.set({ user: null, loading: false });
      } else {
        store.update((s) => ({ ...s, loading: false }));
      }
    }
  },

  async login(email: string, password: string, rememberMe: boolean) {
    const { user } = await authApi.login(email, password, rememberMe);
    store.set({ user, loading: false });
  },

  async logout() {
    try {
      await authApi.logout();
    } finally {
      store.set({ user: null, loading: false });
    }
  }
};
