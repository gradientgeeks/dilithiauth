import React from 'react'
import { renderToString } from 'react-dom/server'
import { MemoryRouter } from 'react-router-dom'
import SidebarLayout from '../components/layout/SidebarLayout'
import { AuthProvider } from '../AuthContext'

declare global {
  var __MOCK_DEVELOPER__: any
}

export default async function runM2ChallengeSuite() {
  console.log('===========================================================')
  console.log('  ADVERSARIAL CHALLENGE SUITE: MILESTONE 2 (SIDEBAR LAYOUT)')
  console.log('===========================================================\n')

  let totalTests = 0
  let passed = 0
  let failed = 0

  function testResult(name: string, isPass: boolean, detail: string = '') {
    totalTests++
    if (isPass) {
      console.log(`[PASS] Test ${totalTests}: ${name}`)
      passed++
    } else {
      console.log(`[FAIL] Test ${totalTests}: ${name}`)
      if (detail) console.log(`       Detail: ${detail}`)
      failed++
    }
  }

  // --------------------------------------------------------------------------
  // TEST GROUP 1: Missing User Email / Name / Whitespace Name in AuthContext
  // --------------------------------------------------------------------------
  console.log('--- GROUP 1: Missing User Email / Name / Whitespace Edge Cases ---')

  // Case 1.1: Missing name & email (both empty string)
  globalThis.__MOCK_DEVELOPER__ = { id: 'dev_empty', email: '', name: '' }
  const htmlEmpty = renderToString(
    React.createElement(AuthProvider, null,
      React.createElement(MemoryRouter, { initialEntries: ['/apps'] },
        React.createElement(SidebarLayout)
      )
    )
  )
  testResult(
    'Initial fallback to "U" when name and email are empty strings',
    htmlEmpty.includes('>U<'),
    'Avatar initial should default to "U"'
  )
  testResult(
    'Display name fallback to "Developer" when name is empty string',
    htmlEmpty.includes('Developer'),
    'Name should fallback to "Developer"'
  )

  // Case 1.2: Whitespace-only name ("   ")
  globalThis.__MOCK_DEVELOPER__ = { id: 'dev_space', email: 'bob@example.com', name: '   ' }
  const htmlSpaceName = renderToString(
    React.createElement(AuthProvider, null,
      React.createElement(MemoryRouter, { initialEntries: ['/apps'] },
        React.createElement(SidebarLayout)
      )
    )
  )
  // Check if initial uses email 'B' instead of space
  const initialIsB = htmlSpaceName.includes('>B<')
  testResult(
    'Avatar initial falls back to email initial "B" when name is whitespace-only',
    initialIsB,
    `Expected >B< in HTML for initial when name is whitespace-only`
  )
  // Check if rendered name falls back to "Developer" or renders blank whitespace
  const rendersDeveloperForSpaceName = !htmlSpaceName.includes('<span class="text-xs font-semibold text-[var(--color-txt)] truncate">   </span>')
  testResult(
    'Rendered name falls back to "Developer" when name is whitespace-only ("   ")',
    rendersDeveloperForSpaceName,
    'BUG DETECTED: `developer.name || "Developer"` evaluates whitespace string "   " as truthy, rendering blank spaces instead of "Developer"!'
  )

  // Case 1.3: Undefined email / undefined name
  globalThis.__MOCK_DEVELOPER__ = { id: 'dev_undef', email: undefined, name: undefined }
  const htmlUndef = renderToString(
    React.createElement(AuthProvider, null,
      React.createElement(MemoryRouter, { initialEntries: ['/apps'] },
        React.createElement(SidebarLayout)
      )
    )
  )
  testResult(
    'Avatar initial defaults to "U" when name and email are undefined',
    htmlUndef.includes('>U<'),
    'Avatar initial should default to "U"'
  )
  // Check tablet title attribute for undefined email
  const titleHasUndefinedStr = htmlUndef.includes('title="Developer (undefined)"')
  testResult(
    'Tablet icon rail title tooltip does NOT output literal string "undefined"',
    !titleHasUndefinedStr,
    'BUG DETECTED: Tablet rail title attribute outputs title="Developer (undefined)" when developer.email is undefined!'
  )

  // Case 1.4: Emoji / Multi-byte unicode character in name
  globalThis.__MOCK_DEVELOPER__ = { id: 'dev_emoji', email: 'emoji@example.com', name: '😊 Alice' }
  const htmlEmoji = renderToString(
    React.createElement(AuthProvider, null,
      React.createElement(MemoryRouter, { initialEntries: ['/apps'] },
        React.createElement(SidebarLayout)
      )
    )
  )
  // Check avatar initial for emoji
  const hasBrokenSurrogate = htmlEmoji.includes('>\uD83D<')
  testResult(
    'Avatar initial handles multi-byte emoji character without producing broken surrogate unit',
    !hasBrokenSurrogate,
    'BUG DETECTED: charAt(0) on emoji "😊" extracts high surrogate unit "\\uD83D", rendering broken character!'
  )


  // --------------------------------------------------------------------------
  // TEST GROUP 2: Long Email Overflow in Profile Footer
  // --------------------------------------------------------------------------
  console.log('\n--- GROUP 2: Long Email Overflow in Profile Footer ---')

  const superLongEmail = 'alexander.christopher.montgomery.the.third@quantum.dilithiauth.identity.provider.example.org'
  const superLongName = 'Alexander Christopher Montgomery The Third'
  globalThis.__MOCK_DEVELOPER__ = { id: 'dev_long', email: superLongEmail, name: superLongName }

  const htmlLong = renderToString(
    React.createElement(AuthProvider, null,
      React.createElement(MemoryRouter, { initialEntries: ['/apps'] },
        React.createElement(SidebarLayout)
      )
    )
  )

  testResult(
    'Long email is rendered in HTML output',
    htmlLong.includes(superLongEmail),
    'Long email string must be rendered'
  )

  // Check CSS truncation classes in Desktop Sidebar
  const hasDesktopTruncateSpan = htmlLong.includes(`class="text-[0.75rem] text-[var(--color-sub)] truncate">${superLongEmail}</span>`)
  testResult(
    'Desktop profile footer email span has "truncate" class',
    hasDesktopTruncateSpan,
    'Email span must have "truncate" class'
  )

  const hasDesktopMinW0FlexCol = htmlLong.includes('class="flex flex-col min-w-0"')
  testResult(
    'Desktop profile footer text container has "min-w-0" flex item constraint',
    hasDesktopMinW0FlexCol,
    'Text container must have "min-w-0" for CSS truncate to work'
  )


  // --------------------------------------------------------------------------
  // TEST GROUP 3: Disabled Link Behavior ("Settings (Coming Soon)")
  // --------------------------------------------------------------------------
  console.log('\n--- GROUP 3: Disabled Link Behavior ("Settings (Coming Soon)") ---')

  testResult(
    'Settings item renders "Soon" badge in sidebar',
    htmlLong.includes('Soon'),
    'Badge "Soon" should be present'
  )

  // Check if Settings element is non-interactive div and has appropriate styling/accessibility
  const settingsHasCursorNotAllowed = htmlLong.includes('cursor-not-allowed')
  testResult(
    'Settings item has "cursor-not-allowed" class',
    settingsHasCursorNotAllowed,
    'Settings item should indicate non-interactive status with cursor-not-allowed'
  )

  const settingsHasAriaDisabled = htmlLong.includes('aria-disabled="true"')
  testResult(
    'Settings item has `aria-disabled="true"` attribute for assistive technology',
    settingsHasAriaDisabled,
    'WARNING/FINDING: Disabled link item does not specify aria-disabled="true" for screen readers'
  )

  const settingsHasRoleLinkOrButton = htmlLong.includes('role="link"') || htmlLong.includes('role="button"')
  testResult(
    'Settings item has explicit ARIA role ("link" or "button")',
    settingsHasRoleLinkOrButton,
    'FINDING: Settings item is rendered as plain <div> without role="link" or role="button"'
  )


  // --------------------------------------------------------------------------
  // TEST GROUP 4: Narrow Viewport Layout & Mobile Drawer Structure
  // --------------------------------------------------------------------------
  console.log('\n--- GROUP 4: Narrow Viewport (<320px) & Layout Checks ---')

  testResult(
    'Mobile Header contains brand name "DilithiAuth"',
    htmlLong.includes('DilithiAuth'),
    'Brand header must render'
  )

  testResult(
    'Mobile Header contains algorithm badge "ML-DSA-44"',
    htmlLong.includes('ML-DSA-44'),
    'Algorithm badge must render in header'
  )

  const headerHasTruncateOrMinW0 = htmlLong.includes('header') && (htmlLong.includes('truncate') || htmlLong.includes('min-w-0'))
  testResult(
    'Mobile Header text elements have min-w-0 or text-overflow truncation for screens <320px',
    headerHasTruncateOrMinW0,
    'FINDING: Mobile header brand text lacks min-w-0/truncate, which can cause horizontal overflow on viewports <320px (e.g. Galaxy Fold at 280px)'
  )


  // --------------------------------------------------------------------------
  // SUMMARY
  // --------------------------------------------------------------------------
  console.log('\n===========================================================')
  console.log(`  EMPIRICAL CHALLENGE SUMMARY: ${passed} PASSED, ${failed} FAILED (Out of ${totalTests} checks)`)
  console.log('===========================================================')

  return { totalTests, passed, failed }
}
