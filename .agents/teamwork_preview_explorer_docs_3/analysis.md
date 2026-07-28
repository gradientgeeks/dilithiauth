# Mobile Sidebar Drawer & Accessibility Analysis (Docs.tsx & index.css)

## 1. Executive Summary

This report evaluates the mobile navigation responsiveness and accessibility (a11y) implementation of the documentation page (`frontend/src/pages/Docs.tsx`) and global styling (`frontend/src/index.css`) in DilithiAuth. 

Currently, **`Docs.tsx` lacks a mobile sidebar drawer and mobile navigation state (`mobileNavOpen`)**. Responsive design relies on a single `@media (max-width: 640px)` media query in `index.css`, missing standard breakpoints for mobile (<768px), tablet (768px-1023px), and desktop (≥1024px). Accessibility compliance is low, missing key `aria-*` attributes, Escape key handlers, focus trap mechanisms, and body scroll locking.

This report provides a concrete architectural design and code proposal following the standard **Stripe / Google Developers documentation navigation pattern**.

---

## 2. Detailed Audit of Existing Codebase

### A. Mobile Navigation Logic in `Docs.tsx`
* **State Management**: `Docs.tsx` contains state for language selection (`selectedLang`), code stage (`activeCodeStage`), endpoint selection (`selectedEndpoint`), endpoint tabs (`endpointTab`), form inputs (`clientId`, `redirectUri`, `subject`), and token strategy (`tokenStrategy`). **It has ZERO state variables for mobile navigation drawer state (`mobileNavOpen`)**.
* **Navigation Components**: 
  * Navigation header (`<Navbar />` imported at line 546) provides high-level page navigation.
  * In-page quick jump navigation (lines 574–587) uses a static horizontal tab list (`.tab-header`). On small viewports, tabs scroll off-screen (`overflow-x: auto`).
  * Endpoint selection sidebar (lines 850–906) uses a 260px grid column (`.detail-grid`) that collapses vertically under 640px, stacking 6 endpoint buttons above the playground controls.
* **Overlays & Drawers**: No backdrop overlay, slide-over drawer panel, or mobile toggle hamburger buttons exist in `Docs.tsx` or `Navbar.tsx`.

### B. Responsive Breakpoint Evaluation (`index.css`)
* **Existing Breakpoints**:
  * Line 750: `@media (max-width: 640px)` is the **only** media query in the entire stylesheet.
  * It adjusts `.detail-grid` column count to 1, reduces `.hero-section h1` font size, reduces `nav` padding, and adjusts `.otp-box` size.
* **Missing Breakpoints**:
  * `<768px (Mobile)`: No rules for hiding/showing mobile navigation elements or drawer triggers.
  * `768px–1023px (Tablet/Medium)`: No rules for scaling Section 3 interactive playground or tab headers. On 768px viewports, the `260px 1fr` grid compresses response inspector preview blocks heavily.
  * `≥1024px (Desktop)`: No sticky sidebar TOC layout for documentation navigation. Content is locked in a single column container (`max-width: 980px`).

### C. Accessibility (a11y) Assessment
1. **ARIA Attributes**:
   * `aria-expanded`: **0 occurrences**. Buttons that switch active stages (language selection lines 617-636, code stage tabs 738-762, endpoint selection 853-905) do not specify `aria-expanded` or `aria-selected`.
   * `aria-controls`: **0 occurrences**. Interactive tabs do not link to their target panels via `aria-controls`.
   * `aria-label` / `aria-labelledby`: Copy buttons (lines 660, 698, 789, 1022) use plain text "Copy" without explicit `aria-label` describing what is being copied. Form inputs in Section 3 (lines 924-959) use `<label className="label">` but lack `id` and `htmlFor` association.
2. **Keyboard Navigation & Focus Management**:
   * No `onKeyDown` or `useEffect` keyboard listeners exist to handle the `Escape` key.
   * Modals (e.g., CSS `.modal-backdrop` in index.css:612) and dropdowns lack focus trapping (`focus-trap`), allowing keyboard focus to cycle into hidden off-screen background elements.
3. **Body Scroll-Lock**:
   * No scroll locking (`document.body.style.overflow = 'hidden'`) is implemented when overlays or dialogs are active.

---

## 3. Recommended Stripe / Google Developers Mobile Navigation Pattern

To bring DilithiAuth's documentation UI up to modern enterprise standards (similar to Stripe Docs and Google Cloud / Web.dev documentation):

### Key Features
1. **Floating/Sticky Mobile Menu Toggle Bar (<768px)**:
   * A sticky bar displayed below the primary `<Navbar />` on mobile screens containing a menu trigger button (`<button aria-expanded={mobileNavOpen} aria-controls="docs-drawer">`).
   * Displays the current active section title and a breadcrumb icon.
2. **Slide-Over Sidebar Drawer & Translucent Backdrop**:
   * A fixed `<aside>` panel sliding in smoothly from the left (`transform: translateX(-100%)` to `translateX(0)`).
   * A dark translucent backdrop (`bg-black/75`) overlaying the main content.
3. **Smooth Animations & Transitions**:
   * GPU-accelerated CSS transitions (`transition: transform 300ms cubic-bezier(0.16, 1, 0.3, 1)`).
4. **Full Accessibility Integration**:
   * `Escape` key keydown listener closing the drawer automatically.
   * Body scroll lock preventing background scrolling while drawer is open.
   * Focus trapping inside drawer while open, returning focus to trigger button upon close.

---

## 4. Proposed Implementation Architecture & Code Snippets

### A. Recommended `Docs.tsx` Component State & Hooks
```tsx
// State additions in Docs.tsx
const [mobileNavOpen, setMobileNavOpen] = useState(false)
const drawerRef = useRef<HTMLDivElement>(null)
const triggerBtnRef = useRef<HTMLButtonElement>(null)

// 1. Lock Body Scroll when Drawer Open
useEffect(() => {
  if (mobileNavOpen) {
    document.body.style.overflow = 'hidden'
  } else {
    document.body.style.overflow = ''
  }
  return () => {
    document.body.style.overflow = ''
  }
}, [mobileNavOpen])

// 2. Escape Key Listener & Focus Trap
useEffect(() => {
  const handleKeyDown = (e: KeyboardEvent) => {
    if (e.key === 'Escape' && mobileNavOpen) {
      setMobileNavOpen(false)
      triggerBtnRef.current?.focus()
    }
  }
  window.addEventListener('keydown', handleKeyDown)
  return () => window.removeEventListener('keydown', handleKeyDown)
}, [mobileNavOpen])
```

### B. Proposed Mobile Trigger Bar & Drawer Markup
```tsx
{/* Mobile Sub-Header & Sticky Navigation Trigger (<768px) */}
<div className="docs-mobile-bar">
  <button
    ref={triggerBtnRef}
    className="btn btn-outline btn-sm flex items-center gap-2"
    onClick={() => setMobileNavOpen(!mobileNavOpen)}
    aria-expanded={mobileNavOpen}
    aria-controls="docs-sidebar-drawer"
    aria-label="Toggle documentation navigation menu"
  >
    <Menu size={16} />
    <span>Documentation Menu</span>
  </button>
  <span className="text-muted text-xs">Section: {selectedEndpoint}</span>
</div>

{/* Backdrop Overlay */}
{mobileNavOpen && (
  <div 
    className="drawer-backdrop" 
    onClick={() => setMobileNavOpen(false)}
    aria-hidden="true"
  />
)}

{/* Slide-over Mobile Sidebar Drawer */}
<aside
  id="docs-sidebar-drawer"
  ref={drawerRef}
  className={`docs-drawer ${mobileNavOpen ? 'open' : ''}`}
  role="dialog"
  aria-modal="true"
  aria-label="Documentation Navigation Drawer"
>
  <div className="drawer-header">
    <div className="flex items-center gap-2 font-semibold">
      <BookOpen size={18} /> Documentation Sections
    </div>
    <button
      className="btn btn-outline btn-sm p-1"
      onClick={() => setMobileNavOpen(false)}
      aria-label="Close navigation drawer"
    >
      <X size={18} />
    </button>
  </div>

  <div className="drawer-body">
    <nav className="drawer-nav">
      <div className="drawer-section-title">Quick Jump</div>
      <a href="#quickstart" onClick={() => setMobileNavOpen(false)} className="drawer-nav-item">
        <Terminal size={15} /> 1. SDK Quickstart
      </a>
      <a href="#code-snippets" onClick={() => setMobileNavOpen(false)} className="drawer-nav-item">
        <Code size={15} /> 2. Code Snippets
      </a>
      <a href="#api-reference" onClick={() => setMobileNavOpen(false)} className="drawer-nav-item">
        <Key size={15} /> 3. API Reference
      </a>
      <a href="#token-strategies" onClick={() => setMobileNavOpen(false)} className="drawer-nav-item">
        <Cpu size={15} /> 4. PQC Token Strategies
      </a>
    </nav>
  </div>
</aside>
```

### C. Proposed Styles for `index.css`
```css
/* Responsive Breakpoints & Mobile Navigation Drawer */

/* Mobile Trigger Sub-header */
.docs-mobile-bar {
  display: none;
  background: var(--bg-card);
  border-bottom: 1px solid var(--border-color);
  padding: 0.6rem 1rem;
  position: sticky;
  top: 56px;
  z-index: 90;
  align-items: center;
  justify-content: space-between;
}

/* Backdrop Overlay */
.drawer-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(4px);
  z-index: 998;
  animation: fadeIn 0.2s ease-out;
}

/* Slide-over Sidebar Drawer */
.docs-drawer {
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  width: 280px;
  max-width: 85vw;
  background: var(--bg-card);
  border-right: 1px solid var(--border-color);
  z-index: 999;
  transform: translateX(-100%);
  transition: transform 0.3s cubic-bezier(0.16, 1, 0.3, 1);
  display: flex;
  flex-direction: column;
  box-shadow: 10px 0 25px rgba(0, 0, 0, 0.5);
}

.docs-drawer.open {
  transform: translateX(0);
}

.drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem;
  border-bottom: 1px solid var(--border-color);
}

.drawer-body {
  padding: 1rem;
  overflow-y: auto;
  flex: 1;
}

.drawer-section-title {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 0.5rem;
}

.drawer-nav-item {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.6rem 0.75rem;
  color: var(--text-sub);
  border-radius: var(--radius-sm);
  font-weight: 500;
  text-decoration: none !important;
  margin-bottom: 0.25rem;
}

.drawer-nav-item:hover, .drawer-nav-item.active {
  background: #1f2937;
  color: var(--text-main);
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

/* Responsive Media Queries */

/* Mobile Breakpoint (<768px) */
@media (max-width: 767px) {
  .docs-mobile-bar {
    display: flex;
  }
  .tab-header {
    display: none; /* Replaced by sticky mobile drawer bar */
  }
  .detail-grid {
    grid-template-columns: 1fr !important;
  }
}

/* Tablet Breakpoint (768px - 1023px) */
@media (min-width: 768px) and (max-width: 1023px) {
  .detail-grid {
    grid-template-columns: 200px 1fr;
    gap: 1rem;
  }
}

/* Desktop Breakpoint (≥1024px) */
@media (min-width: 1024px) {
  .container {
    max-width: 1140px;
  }
}
```

---

## 5. Verification & Testing Strategy

To independently verify the implementation after code updates are applied:

1. **Responsive Testing**:
   * Test at viewports: `375px` (iPhone SE/Mobile), `768px` (iPad/Tablet), `1024px` (Tablet Pro), `1440px` (Desktop).
   * Verify mobile menu toggle bar appears `<768px` and hides `≥768px`.
2. **Drawer Animation & Backdrop**:
   * Verify clicking toggle button slides drawer in smoothly from left with translucent blur overlay.
   * Verify clicking backdrop closes drawer.
3. **Accessibility Verification**:
   * Press `Tab` and verify focus enters drawer when opened.
   * Press `Escape` while drawer is open and verify drawer closes and focus returns to trigger button.
   * Verify page body cannot scroll while drawer is open (`overflow: hidden`).
   * Inspect DOM with Chrome DevTools accessibility pane to ensure `aria-expanded`, `aria-controls`, `role="dialog"`, and `aria-modal="true"` are correctly set.
