import { jsx as _jsx, jsxs as _jsxs, Fragment as _Fragment } from "react/jsx-runtime";
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { LogOut, LayoutGrid, Mail, ShieldCheck, LogIn, UserPlus, BookOpen } from 'lucide-react';
import { useAuth } from '../AuthContext';
export default function Navbar() {
    const { developer, logout } = useAuth();
    const navigate = useNavigate();
    const location = useLocation();
    const handleLogout = async () => {
        await logout();
        navigate('/login');
    };
    const isActive = (path) => location.pathname === path;
    return (_jsxs("nav", { children: [_jsxs(Link, { to: developer ? '/apps' : '/', className: "nav-brand", children: [_jsx("img", { src: "/favicon.svg", alt: "DilithiAuth Logo", width: 28, height: 28 }), _jsxs("div", { children: ["DilithiAuth ", _jsx("span", { className: "nav-brand-badge", children: "PQC IdP" })] })] }), _jsx("div", { className: "nav-actions", children: developer ? (_jsxs(_Fragment, { children: [_jsxs("span", { className: "nav-email-pill", children: [_jsx(Mail, { size: 13, color: "#9ca3af" }), " ", developer.email] }), _jsxs(Link, { to: "/apps", className: isActive('/apps') ? 'active' : '', children: [_jsx(LayoutGrid, { size: 15 }), " Applications"] }), _jsxs(Link, { to: "/docs", className: isActive('/docs') ? 'active' : '', children: [_jsx(BookOpen, { size: 15 }), " SDK & Docs"] }), _jsxs("button", { onClick: handleLogout, children: [_jsx(LogOut, { size: 15 }), " Sign out"] })] })) : (_jsxs(_Fragment, { children: [_jsxs(Link, { to: "/", className: isActive('/') ? 'active' : '', children: [_jsx(ShieldCheck, { size: 15 }), " Overview"] }), _jsxs(Link, { to: "/docs", className: isActive('/docs') ? 'active' : '', children: [_jsx(BookOpen, { size: 15 }), " SDK & Docs"] }), _jsxs(Link, { to: "/login", className: isActive('/login') ? 'active' : '', children: [_jsx(LogIn, { size: 15 }), " Sign in"] }), _jsxs(Link, { to: "/register", className: "nav-register", children: [_jsx(UserPlus, { size: 15 }), " Get Started"] })] })) })] }));
}
