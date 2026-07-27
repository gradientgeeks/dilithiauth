import axios from 'axios';
const apiBaseURL = import.meta.env.VITE_API_BASE_URL || 'https://dilithiauthapis.gradientgeeks.tech/api/v1/developer';
const api = axios.create({
    baseURL: apiBaseURL,
    withCredentials: true,
});
// CSRF token from the most recent /me (or /bff/me) response. State-changing
// requests (create/rotate/delete app) must echo it back via X-CSRF-Token —
// the server rejects those requests without it, even with a valid session
// cookie, since SameSite=Lax alone is not treated as sufficient CSRF defense.
let csrfToken;
api.interceptors.request.use(config => {
    if (csrfToken && ['post', 'put', 'delete', 'patch'].includes((config.method || '').toLowerCase())) {
        config.headers = config.headers || {};
        config.headers['X-CSRF-Token'] = csrfToken;
    }
    return config;
});
export const auth = {
    register: (email, name, password) => api.post('/register', { email, name, password }),
    verifyEmail: async (email, code) => {
        const res = await api.post('/verify-email', { email, code });
        if (res.data.csrf_token)
            csrfToken = res.data.csrf_token;
        return res;
    },
    resendCode: (email) => api.post('/resend-code', { email }),
    login: async (email, password) => {
        const res = await api.post('/login', { email, password });
        if (res.data.csrf_token)
            csrfToken = res.data.csrf_token;
        return res;
    },
    logout: () => api.post('/logout'),
    me: async () => {
        const res = await axios.get((import.meta.env.VITE_API_BASE_URL || 'https://dilithiauthapis.gradientgeeks.tech/api/v1/developer').replace('/api/v1/developer', '') + '/api/v1/bff/me', { withCredentials: true });
        csrfToken = res.data.csrf_token;
        return res;
    },
};
export const apps = {
    list: () => api.get('/apps'),
    get: (id) => api.get(`/apps/${id}`),
    create: (data) => api.post('/apps', data),
    rotateSecret: (id) => api.post(`/apps/${id}/rotate-secret`),
    delete: (id) => api.delete(`/apps/${id}`),
};
