import { DataProvider, fetchUtils } from 'react-admin';

const API_BASE = '/api/v1/admin';

const httpClient = (url: string, options: fetchUtils.Options = {}) => {
  const token = localStorage.getItem('admin_token') || '';
  const headers = (options.headers as Headers) || new Headers();
  headers.set('X-Admin-Token', token);
  return fetchUtils.fetchJson(url, { ...options, headers });
};

// Map resource name to API path segment
const resourcePath = (resource: string): string => {
  const map: Record<string, string> = {
    modules: 'content/modules',
    themes: 'content/themes',
    mnemonics: 'content/mnemonics',
    tests: 'content/tests',
    promo_codes: 'promo-codes',
    users: 'users',
  };
  return map[resource] || resource;
};

// The field used as the unique record ID for each resource
const idField = (resource: string): string => {
  if (resource === 'promo_codes') return 'code';
  if (resource === 'users') return 'telegram_id';
  return 'id';
};

// Add an 'id' alias so react-admin can identify records
const withId = (resource: string, record: any): any => {
  const field = idField(resource);
  return { ...record, id: record[field] };
};

const dataProvider: DataProvider = {
  getList: async (resource, params) => {
    const path = resourcePath(resource);
    let url = `${API_BASE}/${path}`;

    // Users endpoint supports server-side pagination
    if (resource === 'users') {
      const { page, perPage } = params.pagination;
      const offset = (page - 1) * perPage;
      const q = new URLSearchParams({
        limit: String(perPage),
        offset: String(offset),
      });
      if (params.filter?.role) q.set('role', params.filter.role);
      if (params.filter?.subscription_status) q.set('subscription_status', params.filter.subscription_status);
      url += `?${q}`;
      const { json } = await httpClient(url);
      return {
        data: (json.users || []).map((r: any) => withId(resource, r)),
        total: json.total || 0,
      };
    }

    const { json } = await httpClient(url);
    const records: any[] = json.data || [];
    return {
      data: records.map((r) => withId(resource, r)),
      total: json.total ?? records.length,
    };
  },

  getOne: async (resource, params) => {
    const path = resourcePath(resource);
    const { json } = await httpClient(`${API_BASE}/${path}/${params.id}`);
    return { data: withId(resource, json) };
  },

  getMany: async (resource, params) => {
    const path = resourcePath(resource);
    const results = await Promise.all(
      params.ids.map((id) => httpClient(`${API_BASE}/${path}/${id}`).then(({ json }) => withId(resource, json)))
    );
    return { data: results };
  },

  getManyReference: async (resource, params) => {
    const path = resourcePath(resource);
    const { json } = await httpClient(`${API_BASE}/${path}`);
    const allRecords: any[] = (json.data || json.users || []).map((r: any) => withId(resource, r));
    const filtered = allRecords.filter(
      (r) => r[params.target] === params.id
    );
    return { data: filtered, total: filtered.length };
  },

  create: async (resource, params) => {
    const path = resourcePath(resource);
    const { json } = await httpClient(`${API_BASE}/${path}`, {
      method: 'POST',
      body: JSON.stringify(params.data),
    });
    return { data: withId(resource, json) };
  },

  update: async (resource, params) => {
    const path = resourcePath(resource);
    const { json } = await httpClient(`${API_BASE}/${path}/${params.id}`, {
      method: 'PUT',
      body: JSON.stringify(params.data),
    });
    return { data: withId(resource, json) };
  },

  updateMany: async (resource, params) => {
    const path = resourcePath(resource);
    await Promise.all(
      params.ids.map((id) =>
        httpClient(`${API_BASE}/${path}/${id}`, {
          method: 'PUT',
          body: JSON.stringify(params.data),
        })
      )
    );
    return { data: params.ids };
  },

  delete: async (resource, params) => {
    const path = resourcePath(resource);
    await httpClient(`${API_BASE}/${path}/${params.id}`, { method: 'DELETE' });
    return { data: { id: params.id } as any };
  },

  deleteMany: async (resource, params) => {
    const path = resourcePath(resource);
    await Promise.all(
      params.ids.map((id) =>
        httpClient(`${API_BASE}/${path}/${id}`, { method: 'DELETE' })
      )
    );
    return { data: params.ids };
  },
};

export default dataProvider;
