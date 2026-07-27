import { jsx as _jsx } from "react/jsx-runtime";
import { createContext, useContext, useEffect, useState } from 'react';
import { auth } from './api';
const AuthContext = createContext(null);
export function AuthProvider({ children }) {
    const [developer, setDeveloper] = useState(null);
    const [loading, setLoading] = useState(true);
    useEffect(() => {
        auth.me()
            .then(r => setDeveloper(r.data))
            .catch(() => setDeveloper(null))
            .finally(() => setLoading(false));
    }, []);
    const login = async (email, password) => {
        const r = await auth.login(email, password);
        setDeveloper(r.data);
    };
    const register = async (email, name, password) => {
        const r = await auth.register(email, name, password);
        setDeveloper(r.data);
    };
    const logout = async () => {
        await auth.logout();
        setDeveloper(null);
    };
    return (_jsx(AuthContext.Provider, { value: { developer, loading, login, register, logout }, children: children }));
}
export function useAuth() {
    const ctx = useContext(AuthContext);
    if (!ctx)
        throw new Error('useAuth must be used within AuthProvider');
    return ctx;
}
