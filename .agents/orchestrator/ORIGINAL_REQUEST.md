# Original User Request

## 2026-07-28T00:15:25+05:30

Redesign and fix responsiveness issues in the DilithiAuth React Frontend SPA Documentation page (`frontend/src/pages/Docs.tsx` & `frontend/src/index.css`) to match Google Developers / Stripe enterprise documentation standards.

Working directory for frontend code: `/home/uttam/projects/DilithiAuth/frontend`
Integrity mode: `development`

Key Requirements:
R1. Google Developers / Stripe-Style Layout Architecture (Sticky top navbar, sticky left sidebar navigation, fluid main content column with controlled max-width, right TOC for desktop).
R2. Responsive Overflow & Code Block Protection (`<pre><code>`, tables, API parameter lists with max-width: 100%, overflow-x: auto, word-break: break-word). Ensure fluid responsiveness across Mobile (320px-767px), Tablet (768px-1023px), and Desktop (1024px+).
R3. Mobile Sidebar Drawer & Accessibility (<768px collapsible sidebar drawer triggered by floating/sticky button or top header menu toggle).

Acceptance Criteria:
- `/docs` page renders without any horizontal page overflow down to 320px.
- Code blocks and API tables scroll horizontally inside their cards.
- Left sidebar navigation remains sticky during scroll on desktop.
- Frontend builds cleanly with `npm run build` (0 TypeScript / Vite errors).
