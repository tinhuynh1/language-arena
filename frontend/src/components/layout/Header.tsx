'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useAuth } from '@/hooks/useAuth';
import { useLocale } from '@/i18n/LocaleProvider';

export default function Header() {
  const { user, logout, loading } = useAuth();
  const { locale, setLocale, t } = useLocale();
  const pathname = usePathname();

  const isActive = (href: string) => pathname === href || pathname.startsWith(href + '/');

  const NAV_LINKS = [
    { href: '/play', label: t('nav.play') },
    { href: '/leaderboard', label: t('nav.leaderboard') },
  ];

  return (
    <header
      className="fixed top-0 left-0 right-0 z-50"
      style={{
        background: 'rgba(8, 12, 20, 0.9)',
        backdropFilter: 'blur(16px)',
        borderBottom: '1px solid rgba(0, 255, 136, 0.08)',
      }}
    >
      <div className="max-w-7xl mx-auto px-5 sm:px-8 flex items-center justify-between" style={{ height: '4.5rem' }}>
        {/* Logo */}
        <Link href="/" className="flex items-center gap-3 group" aria-label="Lingo Sniper home">
          <div
            className="w-10 h-10 flex items-center justify-center border-2 border-[var(--color-accent-neon)] text-[var(--color-accent-neon)] font-heading font-bold text-sm transition-all duration-200 group-hover:bg-[rgba(0,255,136,0.1)] group-hover:shadow-[0_0_16px_rgba(0,255,136,0.4)]"
            style={{ borderRadius: '3px' }}
            aria-hidden="true"
          >
            LS
          </div>
          <span className="font-heading font-bold text-xl tracking-widest text-[var(--color-text-primary)] group-hover:text-[var(--color-accent-neon)] transition-colors duration-200">
            LINGO SNIPER
          </span>
        </Link>

        {/* Nav */}
        <nav className="flex items-center gap-1 sm:gap-2" aria-label="Main navigation">
          {NAV_LINKS.map(({ href, label }) => (
            <Link
              key={href}
              href={href}
              className="relative px-4 py-2 text-sm font-heading uppercase tracking-widest transition-colors duration-150"
              style={{ color: isActive(href) ? '#00ff88' : undefined }}
              aria-current={isActive(href) ? 'page' : undefined}
            >
              <span className={isActive(href) ? 'text-[var(--color-accent-neon)] text-glow' : 'text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)]'}>
                {label}
              </span>
              {isActive(href) && (
                <span
                  className="absolute bottom-0 left-4 right-4 h-[2px]"
                  style={{
                    background: 'var(--color-accent-neon)',
                    boxShadow: '0 0 8px rgba(0, 255, 136, 0.5)',
                  }}
                  aria-hidden="true"
                />
              )}
            </Link>
          ))}

          {/* Language Switcher */}
          <div className="flex items-center gap-0.5 ml-2 pl-2 border-l border-[var(--color-border-default)]">
            <button
              onClick={() => setLocale('en')}
              className={`px-1.5 py-1 text-base transition-all duration-150 cursor-pointer rounded-sm ${locale === 'en' ? 'opacity-100 scale-110' : 'opacity-40 hover:opacity-70'}`}
              aria-label="English"
              title="English"
            >
              🇬🇧
            </button>
            <button
              onClick={() => setLocale('vi')}
              className={`px-1.5 py-1 text-base transition-all duration-150 cursor-pointer rounded-sm ${locale === 'vi' ? 'opacity-100 scale-110' : 'opacity-40 hover:opacity-70'}`}
              aria-label="Tiếng Việt"
              title="Tiếng Việt"
            >
              🇻🇳
            </button>
          </div>

          {!loading && (
            user ? (
              <div className="flex items-center gap-1 ml-3 pl-3 border-l border-[var(--color-border-default)]">
                <Link
                  href="/dashboard"
                  className="px-4 py-2 text-sm font-heading uppercase tracking-widest transition-colors duration-150"
                  style={{ color: isActive('/dashboard') ? '#00ff88' : undefined }}
                  aria-current={isActive('/dashboard') ? 'page' : undefined}
                >
                  <span className={isActive('/dashboard') ? 'text-[var(--color-accent-neon)] text-glow' : 'text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)]'}>
                    {user.username}
                  </span>
                </Link>
                <button
                  onClick={logout}
                  className="px-3 py-2 text-sm font-heading uppercase tracking-widest text-[var(--color-text-muted)] hover:text-[var(--color-accent-red)] transition-colors duration-150 cursor-pointer"
                  aria-label="Log out"
                >
                  {t('nav.logout')}
                </button>
              </div>
            ) : (
              <Link href="/login" className="btn-primary ml-4 text-sm py-2.5 px-6">
                {t('nav.signin')}
              </Link>
            )
          )}
        </nav>
      </div>
    </header>
  );
}
