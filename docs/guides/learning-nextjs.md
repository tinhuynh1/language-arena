# 📘 Learning Guide: Next.js + React Frontend

> Hướng dẫn học Next.js, React, Tailwind qua dự án / Learn through the project

## Prerequisites
- HTML, CSS, JavaScript cơ bản
- React basics (JSX, components, props, state)

---

## 1. Next.js App Router

### Resources
- https://nextjs.org/docs/app
- Xem: `frontend/src/app/`

### Routing (File-based)
```
src/app/
├── page.tsx           → /
├── login/page.tsx     → /login
├── play/page.tsx      → /play
├── leaderboard/page.tsx → /leaderboard
├── dashboard/page.tsx → /dashboard
└── layout.tsx         → Bọc tất cả pages (Header, providers)
```

### Layout pattern — `layout.tsx`
```tsx
// layout.tsx — wraps ALL pages
export default function RootLayout({ children }) {
  return (
    <html>
      <body>
        <AuthProvider>     {/* Auth context available everywhere */}
          <Header />       {/* Persistent navigation */}
          {children}       {/* Page content changes here */}
        </AuthProvider>
      </body>
    </html>
  );
}
```

---

## 2. React Hooks — Custom Hooks

### `useAuth` — Provider Pattern
```tsx
// Context + Provider pattern cho auth state
const AuthContext = createContext();

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [token, setToken] = useState(localStorage.getItem('token'));

  const login = async (email, password) => { /* API call → set token */ };
  const logout = () => { /* clear token + redirect */ };

  return (
    <AuthContext.Provider value={{ user, token, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export const useAuth = () => useContext(AuthContext);
```

### `useWebSocket` — Connection Manager
- Quản lý vòng đời WS connection
- Auto-reconnect với exponential backoff
- Timeout detection + error UI

### `useGame` — State Machine
- Nhận WS messages → update game state
- State transitions: lobby → countdown → playing → game_over
- Expose actions: createRoom, joinRoom, hitTarget, etc.

**Key pattern**: Single hook quản lý toàn bộ game logic, components chỉ render state.

---

## 3. Tailwind CSS v4

### Resources
- https://tailwindcss.com/docs
- Xem: `frontend/src/styles/globals.css`

### Patterns trong dự án

```css
/* CSS variables cho theme — globals.css */
@theme {
  --color-neon-cyan: #00f0ff;
  --color-neon-pink: #ff006e;
  --color-bg-dark: #0a0a1a;
}

/* Glassmorphism effect */
.glass-panel {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(255, 255, 255, 0.1);
}
```

### Accessibility — Motion Sensitivity
```tsx
// Tất cả animations PHẢI dùng motion-safe prefix
<div className="motion-safe:animate-pulse motion-reduce:animate-none">
```

---

## 4. State Management Pattern

Dự án KHÔNG dùng Redux/Zustand. Thay vào đó:

| State Type | Solution | Lý do / Why |
|-----------|---------|-------------|
| **Auth** | React Context (`useAuth`) | Global, ít thay đổi |
| **Game** | Custom hook (`useGame`) | Complex state machine |
| **WS connection** | Custom hook (`useWebSocket`) | Lifecycle management |
| **Page data** | `useState` + API calls | Local, per-page |
| **Token** | `localStorage` | Persist across refreshes |

---

## 5. Component Design Patterns

### Smart (Container) vs Dumb (Presentational)

| Type | Example | Có state? | Gọi API? |
|------|---------|-----------|----------|
| **Smart** | `play/page.tsx` | ✅ useGame | ✅ WS messages |
| **Dumb** | `GameCanvas.tsx` | ❌ Props only | ❌ Render only |
| **Dumb** | `Countdown.tsx` | ❌ Props only | ❌ Animation only |
| **Dumb** | `GameOverScreen.tsx` | ❌ Props only | ❌ Display only |

### Bài tập / Exercises

1. ✏️ Vẽ component tree cho `/play` page — component nào smart, component nào dumb?
2. ✏️ `useGame` hook xử lý bao nhiêu loại WS message? List tất cả.
3. ✏️ Nếu `useWebSocket` disconnect, UI hiển thị gì? Tìm trong code.
4. ✏️ Thử thêm 1 trang mới `/profile` — cần tạo file gì, ở đâu?
