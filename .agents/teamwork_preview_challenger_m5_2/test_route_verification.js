import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter, Routes, Route, Navigate } from 'react-router-dom';
import Navbar from '../../frontend/src/components/Navbar';
import { createContext, useContext } from 'react';
const AuthContext = createContext(null);
function useAuth() {
    const ctx = useContext(AuthContext);
    if (!ctx)
        throw new Error('useAuth must be used within AuthProvider');
    return ctx;
}
function RequireAuth({ children, developer, loading }) {
    if (loading)
        return _jsx("div", { style: { padding: '2rem' }, children: "Loading\u2026" });
    return developer ? children : _jsx(Navigate, { to: "/login", replace: true });
}
function RequireGuest({ children, developer, loading }) {
    if (loading)
        return _jsx("div", { style: { padding: '2rem' }, children: "Loading\u2026" });
    return developer ? _jsx(Navigate, { to: "/apps", replace: true }) : children;
}
function MockDocs() {
    return (_jsxs("div", { id: "docs-page", children: [_jsx(Navbar, {}), _jsx("h1", { children: "Docs Page Content" })] }));
}
function MockLanding() {
    return _jsx("div", { id: "landing-page", children: "Landing Page Content" });
}
function MockDashboard() {
    return _jsx("div", { id: "dashboard-page", children: "Dashboard Page Content" });
}
function MockApp({ initialRoute, developer, loading = false }) {
    const authValue = {
        developer,
        loading,
        login: async () => { },
        register: async () => { },
        logout: async () => { },
    };
    return (_jsx(AuthContext.Provider, { value: authValue, children: _jsx(MemoryRouter, { initialEntries: [initialRoute], children: _jsxs(Routes, { children: [_jsx(Route, { path: "/", element: _jsx(RequireGuest, { developer: developer, loading: loading, children: _jsx(MockLanding, {}) }) }), _jsx(Route, { path: "/docs", element: _jsx(MockDocs, {}) }), _jsx(Route, { path: "/apps", element: _jsx(RequireAuth, { developer: developer, loading: loading, children: _jsx(MockDashboard, {}) }) })] }) }) }));
}
function runTests() {
    console.log('=== EMPIRICAL ROUTE & NAVBAR VERIFICATION ===\n');
    let passed = 0;
    let failed = 0;
    function assert(condition, testName, detail = '') {
        if (condition) {
            console.log(`[PASS] ${testName}`);
            passed++;
        }
        else {
            console.error(`[FAIL] ${testName} - ${detail}`);
            failed++;
        }
    }
    // Test 1: Guest User accessing /docs
    const guestDocsHtml = renderToString(_jsx(MockApp, { initialRoute: "/docs", developer: null }));
    assert(guestDocsHtml.includes('id="docs-page"') && guestDocsHtml.includes('Docs Page Content'), 'Test 1: Guest user can access /docs route', `HTML output: ${guestDocsHtml}`);
    // Test 2: Guest User Navbar active highlighting for /docs
    assert(guestDocsHtml.includes('class="active"') && guestDocsHtml.includes('SDK &amp; Docs'), 'Test 2: Navbar highlights /docs with active class for guest user', `HTML output: ${guestDocsHtml}`);
    // Test 3: Guest User Navbar links present
    assert(guestDocsHtml.includes('Overview') && guestDocsHtml.includes('Sign in') && guestDocsHtml.includes('Get Started'), 'Test 3: Navbar displays guest action links (Overview, Sign in, Get Started)', `HTML output: ${guestDocsHtml}`);
    // Test 4: Authenticated User accessing /docs
    const authDev = { id: 'dev_123', email: 'alice@example.com', name: 'Alice' };
    const authDocsHtml = renderToString(_jsx(MockApp, { initialRoute: "/docs", developer: authDev }));
    assert(authDocsHtml.includes('id="docs-page"') && authDocsHtml.includes('Docs Page Content'), 'Test 4: Authenticated user can access /docs route without redirect', `HTML output: ${authDocsHtml}`);
    // Test 5: Authenticated User Navbar active highlighting for /docs
    assert(authDocsHtml.includes('class="active"') && authDocsHtml.includes('SDK &amp; Docs'), 'Test 5: Navbar highlights /docs with active class for authenticated user', `HTML output: ${authDocsHtml}`);
    // Test 6: Authenticated User Navbar email pill and actions
    assert(authDocsHtml.includes('alice@example.com') && authDocsHtml.includes('Applications') && authDocsHtml.includes('Sign out'), 'Test 6: Navbar displays authenticated user email pill and links (Applications, Sign out)', `HTML output: ${authDocsHtml}`);
    // Test 7: Active class highlighting on /apps route for Authenticated user
    const authAppsHtml = renderToString(_jsx(MockApp, { initialRoute: "/apps", developer: authDev }));
    assert(authAppsHtml.includes('Applications') && authAppsHtml.includes('class="active"'), 'Test 7: Navbar highlights /apps with active class when on /apps route', `HTML output: ${authAppsHtml}`);
    // Test 8: Verify active class is NOT on /docs link when on /apps route
    const docsActiveOnApps = authAppsHtml.includes('href="/docs" class="active"') || authAppsHtml.includes('class="active" href="/docs"');
    assert(!docsActiveOnApps, 'Test 8: Navbar /docs link is NOT highlighted active when on /apps route', `HTML output: ${authAppsHtml}`);
    console.log(`\nSummary: ${passed} passed, ${failed} failed.`);
    if (failed > 0) {
        process.exit(1);
    }
}
runTests();
