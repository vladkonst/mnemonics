import { AuthProvider } from 'react-admin';

const API_BASE = '/api/v1/admin';

const authProvider: AuthProvider = {
  login: async ({ username: token }: { username: string }) => {
    // Validate token by calling analytics endpoint
    const res = await fetch(`${API_BASE}/analytics/overview`, {
      headers: { 'X-Admin-Token': token },
    });
    if (res.status === 401 || res.status === 403) {
      throw new Error('Неверный токен');
    }
    if (!res.ok) {
      throw new Error('Ошибка сервера');
    }
    localStorage.setItem('admin_token', token);
  },
  logout: async () => {
    localStorage.removeItem('admin_token');
  },
  checkAuth: async () => {
    if (!localStorage.getItem('admin_token')) {
      throw new Error('Not authenticated');
    }
  },
  checkError: async (error: any) => {
    if (error?.status === 401 || error?.status === 403) {
      localStorage.removeItem('admin_token');
      throw new Error('Unauthorized');
    }
  },
  getIdentity: async () => ({
    id: 'admin',
    fullName: 'Administrator',
  }),
  getPermissions: async () => 'admin',
};

export default authProvider;
