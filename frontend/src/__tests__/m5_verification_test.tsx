import React from 'react'
import { renderToString } from 'react-dom/server'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import SettingsTab from '../components/app-detail/SettingsTab'
import SnippetsTab from '../components/app-detail/SnippetsTab'
import DangerZoneTab from '../components/app-detail/DangerZoneTab'
import SecretDisplayModal from '../components/app-detail/SecretDisplayModal'
import DeleteAppModal from '../components/app-detail/DeleteAppModal'
import AppDetail from '../pages/AppDetail'
import type { OAuthApp } from '../api'

export default async function runM5TestSuite() {
  console.log('===========================================================')
  console.log('  EMPIRICAL VERIFICATION: MILESTONE 5 APP DETAIL REDESIGN')
  console.log('===========================================================\n')

  let passed = 0
  let failed = 0

  function assert(condition: boolean, testName: string, detail: string = '') {
    if (condition) {
      console.log(`[PASS] ${testName}`)
      passed++
    } else {
      console.error(`[FAIL] ${testName}`)
      if (detail) console.error(`       Detail: ${detail}`)
      failed++
    }
  }

  const mockConfidentialApp: OAuthApp = {
    ID: 'app_m5_confidential_123',
    DeveloperID: 'dev_m5',
    Name: 'Production Microservice IdP',
    Description: 'Quantum-safe backend service authentication provider',
    HomepageURL: 'https://service.example.com',
    ClientID: 'client_m5_conf_98765',
    RedirectURIs: ['https://service.example.com/oauth/callback', 'https://service.example.com/api/login'],
    IsPublic: false,
    CreatedAt: '2026-07-28T21:00:00Z',
    SecretRotatedAt: '2026-07-28T21:00:00Z',
  }

  const mockPublicApp: OAuthApp = {
    ID: 'app_m5_public_456',
    DeveloperID: 'dev_m5',
    Name: 'React SPA Client App',
    Description: 'Public single page web application',
    HomepageURL: 'https://spa.example.com',
    ClientID: 'client_m5_pub_12345',
    RedirectURIs: ['https://spa.example.com/callback'],
    IsPublic: true,
    CreatedAt: '2026-07-28T21:00:00Z',
    SecretRotatedAt: null,
  }

  // --- SECTION 1: SETTINGS TAB COMPONENT ---
  console.log('--- SECTION 1: SettingsTab Component ---')
  const settingsHtml = renderToString(
    React.createElement(SettingsTab, {
      app: mockConfidentialApp,
      onSave: async () => {},
      saving: false,
      saveSuccess: true,
      saveError: null,
    })
  )
  assert(settingsHtml.includes('General Settings'), '1.1 Renders General Settings section header')
  assert(settingsHtml.includes('Production Microservice IdP'), '1.2 Displays application name value')
  assert(settingsHtml.includes('https://service.example.com'), '1.3 Displays homepage URL')
  assert(settingsHtml.includes('Authorized Redirect URIs'), '1.4 Renders Authorized Redirect URIs manager label')
  assert(settingsHtml.includes('https://service.example.com/oauth/callback'), '1.5 Displays redirect URI row 1')
  assert(settingsHtml.includes('https://service.example.com/api/login'), '1.6 Displays redirect URI row 2')
  assert(settingsHtml.includes('Add Redirect URI'), '1.7 Renders "Add Redirect URI" button')
  assert(settingsHtml.includes('client_m5_conf_98765'), '1.8 Displays Client ID in read-only input')
  assert(settingsHtml.includes('Confidential Client (Client Secret)'), '1.9 Renders Confidential Client type badge')
  assert(settingsHtml.includes('Settings updated successfully!'), '1.10 Displays success toast banner')

  const publicSettingsHtml = renderToString(
    React.createElement(SettingsTab, {
      app: mockPublicApp,
      onSave: async () => {},
      saving: false,
      saveSuccess: false,
      saveError: null,
    })
  )
  assert(publicSettingsHtml.includes('Public SPA (PKCE Mode)'), '1.11 Renders Public SPA type badge')

  // --- SECTION 2: SNIPPETS TAB COMPONENT ---
  console.log('\n--- SECTION 2: SnippetsTab Component ---')
  const snippetsHtml = renderToString(
    React.createElement(SnippetsTab, { app: mockConfidentialApp })
  )
  assert(snippetsHtml.includes('OAuth 2.0 / DilithiAuth Integration Snippets'), '2.1 Renders SnippetsTab title')
  assert(snippetsHtml.includes('/token/thin'), '2.2 Features /token/thin endpoint strategy button')
  assert(snippetsHtml.includes('/token/full'), '2.3 Features /token/full endpoint strategy button')
  assert(snippetsHtml.includes('/token/cwt'), '2.4 Features /token/cwt endpoint strategy button')
  assert(snippetsHtml.includes('cURL'), '2.5 Features cURL language tab')
  assert(snippetsHtml.includes('JS / Fetch'), '2.6 Features JS / Fetch language tab')
  assert(snippetsHtml.includes('Go'), '2.7 Features Go language tab')
  assert(snippetsHtml.includes('Python'), '2.8 Features Python language tab')
  assert(snippetsHtml.includes('Copy Code'), '2.9 Renders Copy Code action button')
  assert(snippetsHtml.includes('/.well-known/jwks.json'), '2.10 Renders JWKS verification info banner')

  // --- SECTION 3: DANGER ZONE TAB COMPONENT ---
  console.log('\n--- SECTION 3: DangerZoneTab Component ---')
  const dangerHtml = renderToString(
    React.createElement(DangerZoneTab, {
      app: mockConfidentialApp,
      onOpenRotateModal: () => {},
      onOpenDeleteModal: () => {},
      rotating: false,
      deleting: false,
    })
  )
  assert(dangerHtml.includes('Danger Zone'), '3.1 Renders Danger Zone title header')
  assert(dangerHtml.includes('Rotate Client Secret'), '3.2 Renders "Rotate Client Secret" button for confidential app')
  assert(dangerHtml.includes('Delete Application'), '3.3 Renders "Delete Application" button')

  const publicDangerHtml = renderToString(
    React.createElement(DangerZoneTab, {
      app: mockPublicApp,
      onOpenRotateModal: () => {},
      onOpenDeleteModal: () => {},
      rotating: false,
      deleting: false,
    })
  )
  assert(publicDangerHtml.includes('Public applications do not use client secrets'), '3.4 Displays public application secret note')

  // --- SECTION 4: SECRET DISPLAY MODAL ---
  console.log('\n--- SECTION 4: SecretDisplayModal Component ---')
  const secretModalHtml = renderToString(
    React.createElement(SecretDisplayModal, {
      clientId: 'client_m5_conf_98765',
      clientSecret: 'sec_ml_dsa_44_rotated_secret_999xxx',
      onClose: () => {},
    })
  )
  assert(secretModalHtml.includes('New Client Secret Generated'), '4.1 Renders secret modal heading')
  assert(secretModalHtml.includes('client_m5_conf_98765'), '4.2 Displays Client ID')
  assert(secretModalHtml.includes('sec_ml_dsa_44_rotated_secret_999xxx'), '4.3 Displays rotated secret value')
  assert(secretModalHtml.includes('Save this secret now'), '4.4 Displays Amber warning banner')
  assert(secretModalHtml.includes('Saved My Secret'), '4.5 Renders "I\'ve Saved My Secret" dismiss CTA')

  // --- SECTION 5: DELETE APP MODAL ---
  console.log('\n--- SECTION 5: DeleteAppModal Component ---')
  const deleteModalHtml = renderToString(
    React.createElement(DeleteAppModal, {
      appName: 'Production Microservice IdP',
      loading: false,
      onConfirm: () => {},
      onCancel: () => {},
    })
  )
  assert(deleteModalHtml.includes('Delete Application'), '5.1 Renders Delete Application title')
  assert(deleteModalHtml.includes('To confirm deletion, please type'), '5.2 Renders type-name confirmation instruction')
  assert(deleteModalHtml.includes('Production Microservice IdP'), '5.3 Renders target application name')
  assert(deleteModalHtml.includes('Permanently Delete Application'), '5.4 Renders confirm deletion CTA')

  // --- SECTION 6: APP DETAIL PAGE LAYOUT & ZERO NAVBAR ---
  console.log('\n--- SECTION 6: AppDetail Page Router Integration ---')
  const appDetailHtml = renderToString(
    React.createElement(
      MemoryRouter,
      { initialEntries: ['/apps/app_m5_confidential_123'] },
      React.createElement(
        Routes,
        null,
        React.createElement(Route, { path: '/apps/:id', element: React.createElement(AppDetail) })
      )
    )
  )

  assert(!appDetailHtml.toLowerCase().includes('nav-logo') && !appDetailHtml.includes('DilithiAuth Identity Provider'), '6.1 Contains ZERO Navbar component rendering')

  // --- SUMMARY ---
  console.log('\n===========================================================')
  console.log(`  VERIFICATION RESULTS: ${passed} PASSED, ${failed} FAILED`)
  console.log('===========================================================')

  if (failed > 0) {
    throw new Error(`${failed} M5 verification tests failed`)
  }
}
