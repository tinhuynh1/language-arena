# 🎯 Interview Questions: Frontend

> Câu hỏi phỏng vấn frontend qua dự án / Frontend interview prep

---

## React & Next.js

### Q1: Next.js App Router vs Pages Router — khác biệt gì?

**Answer:**
- **App Router** (dự án dùng): File-based routing trong `app/` folder, Server Components mặc định, layouts nested, streaming
- **Pages Router**: `pages/` folder, Client Components mặc định, `getServerSideProps`/`getStaticProps`
- Dự án chọn App Router vì: server-side rendering mặc định tốt cho SEO, layout system phù hợp cho shared Header
- 📄 Xem: `app/layout.tsx`, `app/page.tsx`

### Q2: Giải thích React Context pattern trong useAuth?

**Answer:**
```tsx
// 1. Create context
const AuthContext = createContext(null);

// 2. Provider wraps app (in layout.tsx)
<AuthProvider>{children}</AuthProvider>

// 3. Consumer hook
const { user, login, logout } = useAuth();
```
- **Tại sao Context**: Auth state cần access ở mọi page (Header, protected pages)
- **Trade-off**: Context re-renders all consumers when value changes
- **Khi nào KHÔNG dùng Context**: Frequently updating data (dùng Zustand/Jotai thay)

### Q3: Custom hooks vs useEffect trực tiếp?

**Answer:**
- Custom hook (`useGame`, `useWebSocket`) encapsulate logic phức tạp
- Benefits: Reusable, testable, separation of concerns
- `useGame` quản lý state machine (lobby → playing → game_over) — nếu viết trực tiếp trong component sẽ rất messy
- 📄 Xem: `hooks/useGame.ts` (~200+ lines of logic, tách ra khỏi UI)

### Q4: WebSocket state management — tại sao không dùng Redux?

**Answer:**
- Game state thay đổi rất nhanh (mỗi round, mỗi hit)
- Redux overhead: action → reducer → store → re-render — quá nhiều boilerplate
- Custom hook `useGame` + `useState` đủ cho use case này
- WS messages đến → `useGame` update state → component re-render
- Nếu app lớn hơn (10+ screens, shared state phức tạp): consider Zustand (simpler than Redux)

---

## State Management

### Q5: Component nào là "smart" vs "dumb"?

**Answer:**
| Smart (Container) | Dumb (Presentational) |
|---|---|
| `play/page.tsx` — owns `useGame` hook | `GameCanvas.tsx` — renders targets from props |
| `leaderboard/page.tsx` — fetches data | `Countdown.tsx` — pure animation |
| `login/page.tsx` — handles form + API | `GameOverScreen.tsx` — displays results |
| `layout.tsx` — provides auth context | `Header.tsx` — renders nav from auth state |

- **Principle**: Smart components own data/logic, dumb components only render
- **Benefit**: Dumb components dễ test, dễ reuse

### Q6: Tại sao lưu JWT token trong localStorage?

**Answer:**
- **Pros**: Simple, persists across page refresh, easy to attach to API calls
- **Cons**: Vulnerable to XSS attacks (malicious script can read localStorage)
- **Better alternative**: httpOnly cookies (browser manages, cannot be read by JS)
- **Trade-off**: httpOnly cookies phức tạp hơn (CSRF protection needed, same-site config)
- Dự án chọn localStorage vì simplicity, nhưng production nên dùng httpOnly

---

## Performance

### Q7: Next.js performance optimizations trong dự án?

**Answer:**
- **Standalone output**: `next.config.js: output: 'standalone'` → smaller Docker image
- **Server-side rendering**: SEO + faster first paint
- **No external images**: UI effects built entirely with CSS (zero image requests)
- **WebSocket**: Real-time data via WS, không polling API mỗi giây
- **Motion sensitivity**: `prefers-reduced-motion` → skip animations for users who prefer

### Q8: React rendering optimization patterns?

**Answer:**
- **Conditional rendering**: Game state machine → chỉ render component phù hợp với state hiện tại
- **Key prop**: Giúp React identify elements khi list changes (player list, leaderboard)
- **Potential improvements**: `React.memo()` cho GameCanvas, `useMemo` cho expensive calculations

---

## Accessibility

### Q9: Accessibility measures trong dự án?

**Answer:**
- ✅ `focus-visible` states trên tất cả interactive elements
- ✅ `prefers-reduced-motion` respected — animations wrapped với `motion-safe:`
- ✅ Semantic HTML: `<button>` thay vì `<div onClick>`
- ✅ ARIA attributes trên dynamic content
- 📄 Xem: `components/game/GameCanvas.tsx` — target elements là `<button>` với focus states
