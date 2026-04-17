'use client';

import Link from 'next/link';
import { useAuth } from '@/hooks/useAuth';

export default function Header() {
  const { user, logout, loading } = useAuth();

  return (
    <header className="fixed top-0 left-0 right-0 z-50 border-b border-[var(--color-border-default)]"
            style={{ background: 'rgba(10, 14, 23, 0.9)', backdropFilter: 'blur(8px)' }}>
      <div className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
        {/* Logo */}
        <Link href="/" className="flex items-center gap-2 group">
          <div className="w-8 h-8 flex items-center justify-center border border-[var(--color-accent-neon)] text-[var(--color-accent-neon)] font-heading font-bold text-sm"
               style={{ borderRadius: '2px' }}>
            LS
          </div>
          <span className="font-heading font-bold text-xl tracking-wider group-hover:text-[var(--color-accent-neon)] transition-colors">
            LINGO SNIPER
          </span>
        </Link>

        {/* Nav */}
        <nav className="flex items-center gap-6">
          <Link href="/play" className="text-sm font-heading uppercase tracking-wider text-[var(--color-text-secondary)] hover:text-[var(--color-accent-neon)] transition-colors">
            Play
          </Link>
          <Link href="/leaderboard" className="text-sm font-heading uppercase tracking-wider text-[var(--color-text-secondary)] hover:text-[var(--color-accent-neon)] transition-colors">
            Leaderboard
          </Link>

          {!loading && (
            user ? (
              <div className="flex items-center gap-4">
                <Link href="/dashboard" className="text-sm font-heading uppercase tracking-wider text-[var(--color-text-secondary)] hover:text-[var(--color-accent-neon)] transition-colors">
                  {user.username}
                </Link>
                <button onClick={logout} className="text-sm font-heading uppercase tracking-wider text-[var(--color-text-muted)] hover:text-[var(--color-accent-red)] transition-colors">
                  Logout
                </button>
              </div>
            ) : (
              <Link href="/login" className="btn-primary text-xs py-2 px-4">
                Sign In
              </Link>
            )
          )}
        </nav>
      </div>
    </header>
  );
}
