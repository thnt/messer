import dayjs from 'dayjs';
import { get, writable } from 'svelte/store';

const _userStore = writable({
  init: false,
  user: null,
  loading: false,
  error: null,
});

export const userStore = {
  subscribe: _userStore.subscribe,
  auth: async () => {
    try {
      const user = await api.GET('/login');
      _userStore.update(s => ({ ...s, user, init: true }));
    } catch (error) {
      _userStore.update(s => ({ ...s, loading: false, init: true }));
    }
  },
  login: async (username, password) => {
    try {
      _userStore.update(s => ({ ...s, error: null, loading: true }));
      const user = await api.POST('/login', { username, password });
      _userStore.update(s => ({ ...s, error: null, loading: false, user }));
    } catch (error) {
      _userStore.update(s => ({ ...s, loading: false, error }));
    }
  },
  logout: async () => {
    try {
      await api.DELETE('/login');
      _userStore.set({});
    } catch (error) {}
  },
};

const _metricStore = writable({
  total: 0,
  metrics: [],
  recents: [],
  loading: false,
  error: null,
  dashboard: {},
});

export const metricStore = {
  subscribe: _metricStore.subscribe,
  getMetrics: async params => {
    _metricStore.update(s => ({ ...s, loading: true }));
    try {
      const res = await api.GET('/metrics', params);
      _metricStore.update(s => ({
        ...s,
        total: res.total,
        metrics: res.metrics || [],
        loading: false,
        error: null,
      }));
    } catch (error) {
      _metricStore.update(s => ({ ...s, loading: false, error }));
    }
  },
  getRecents: async limit => {
    try {
      const { recents } = get(_metricStore);
      const params = { limit, watch: 30, from: dayjs().unix() - 86400, src };
      if (recents.length) {
        params.from = recents[0].Timestamp + 1;
      }
      const res = await api.GET('/metrics', params);
      if (res.metrics?.length) {
        _metricStore.update(s => ({
          ...s,
          recents: [...res.metrics, ...s.recents].slice(0, limit),
        }));
      }

      setTimeout(() => {
        metricStore.getRecents(limit);
      }, 1000);
    } catch (error) {
      console.log(error);
      setTimeout(() => {
        metricStore.getRecents(limit);
      }, 5000);
    }
  },
  getDashboard: async (src = '') => {
    try {
      const { dashboard } = get(_metricStore);
      const params = { limit: 1, src, watch: 30, from: dayjs().unix() - 86400 };
      if (dashboard?.Timestamp) {
        params.from = dashboard.Timestamp + 1;
      }
      const res = await api.GET('/metrics', params);
      if (res.metrics?.length) {
        _metricStore.update(s => ({
          ...s,
          dashboard: res.metrics[0],
        }));
      }

      setTimeout(() => {
        metricStore.getDashboard(src);
      }, 1000);
    } catch (error) {
      console.log(error);
      setTimeout(() => {
        metricStore.getDashboard(src);
      }, 5000);
    }
  },
};

const api = {
  baseURL: import.meta.env.VITE_API_BASE_URL,
  async request(method, path, params) {
    let url = this.baseURL + path;
    const opts = {
      method,
    };
    if (method === 'GET' || method === 'DELETE') {
      const p = new URLSearchParams(params);
      url += `?${p}`;
    } else {
      opts.headers = {
        'Content-Type': 'application/json',
      };
      opts.body = JSON.stringify(params);
    }
    const res = await fetch(url, opts);

    if (res.status === 200) {
      return await res.json();
    } else if (res.status === 401) {
      _userStore.set({});
    }

    const { error } = await res.json();
    throw new Error(error);
  },
  async GET(path, params) {
    return this.request('GET', path, params);
  },
  async DELETE(path, params) {
    return this.request('DELETE', path, params);
  },
  async POST(path, params) {
    return this.request('POST', path, params);
  },
};
