'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useAuth } from '@/hooks/useAuth';
import { useLocale } from '@/i18n/LocaleProvider';

export default function LoginPage() {
  const [isRegister, setIsRegister] = useState(false);
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const { login, register, user, loading: authLoading } = useAuth();
  const { t } = useLocale();
  const router = useRouter();

  useEffect(() => {
    if (!authLoading && user) {
      router.replace('/play');
    }
  }, [user, authLoading, router]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      if (isRegister) {
        await register(username, email, password);
      } else {
        await login(email, password);
      }
      router.push('/play');
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : t('login.error.default'));
    } finally {
      setLoading(false);
    }
  };

  const switchMode = () => {
    setIsRegister(v => !v);
    setError('');
  };

  return (
    <div className="relative min-h-screen flex items-center justify-center px-6 overflow-hidden">
      {/* Ambient orbs */}
      <div className="orb w-[500px] h-[500px] opacity-[0.06] animate-float"
           style={{ background: '#00ff88', top: '-20%', left: '-15%' }} />
      <div className="orb w-[400px] h-[400px] opacity-[0.05] animate-float"
           style={{ background: '#00d4ff', bottom: '-15%', right: '-10%', animationDelay: '-5s' }} />

      {/* Grid */}
      <div className="absolute inset-0 opacity-[0.02]" aria-hidden="true" style={{
        backgroundImage: `linear-gradient(rgba(0,255,136,0.4) 1px, transparent 1px),
                          linear-gradient(90deg, rgba(0,255,136,0.4) 1px, transparent 1px)`,
        backgroundSize: '60px 60px',
      }} />

      <div className="relative z-10 w-full max-w-sm animate-fade-in-up">
        {/* Logo */}
        <div className="text-center mb-10">
          <Link href="/" aria-label="Go to homepage">
            <div className="inline-block font-heading font-bold text-3xl tracking-wider mb-5 hover:opacity-80 transition-opacity">
              <span className="text-gradient-neon text-glow">LINGO</span>
              <span className="text-[var(--color-text-primary)]"> SNIPER</span>
            </div>
          </Link>
          <h1 className="font-heading font-bold text-2xl uppercase tracking-wider">
            {isRegister ? t('login.title.register') : t('login.title.login')}
          </h1>
          <p className="text-sm text-[var(--color-text-muted)] mt-2">
            {isRegister ? t('login.subtitle.register') : t('login.subtitle.login')}
          </p>
        </div>

        {/* Form card */}
        <form onSubmit={handleSubmit} className="card space-y-5">
          {isRegister && (
            <div>
              <label htmlFor="username" className="block text-xs font-heading uppercase tracking-wider text-[var(--color-text-muted)] mb-2">
                {t('login.label.username')}
              </label>
              <input
                id="username"
                name="username"
                type="text"
                autoComplete="username"
                className="input-field"
                placeholder="SniperElite"
                value={username}
                onChange={e => setUsername(e.target.value)}
                required
                minLength={3}
              />
            </div>
          )}

          <div>
            <label htmlFor="email" className="block text-xs font-heading uppercase tracking-wider text-[var(--color-text-muted)] mb-2">
              {t('login.label.email')}
            </label>
            <input
              id="email"
              name="email"
              type="email"
              autoComplete="email"
              className="input-field"
              placeholder="you@arena.com"
              value={email}
              onChange={e => setEmail(e.target.value)}
              required
            />
          </div>

          <div>
            <label htmlFor="password" className="block text-xs font-heading uppercase tracking-wider text-[var(--color-text-muted)] mb-2">
              {t('login.label.password')}
            </label>
            <input
              id="password"
              name="password"
              type="password"
              autoComplete={isRegister ? 'new-password' : 'current-password'}
              className="input-field"
              placeholder="••••••••"
              value={password}
              onChange={e => setPassword(e.target.value)}
              required
              minLength={6}
            />
          </div>

          {error && (
            <div
              role="alert"
              className="text-sm text-[var(--color-accent-red)] bg-[rgba(255,53,72,0.08)] border border-[rgba(255,53,72,0.25)] px-4 py-3"
              style={{ borderRadius: '3px' }}
            >
              {error}
            </div>
          )}

          <button
            type="submit"
            disabled={loading}
            className="btn-primary w-full"
            aria-busy={loading}
          >
            {loading
              ? t('login.loading')
              : isRegister
                ? t('login.submit.register')
                : t('login.submit.login')}
          </button>
        </form>

        {/* Toggle */}
        <p className="text-center text-sm text-[var(--color-text-muted)] mt-6">
          {isRegister ? t('login.toggle.hasAccount') : t('login.toggle.noAccount')}{' '}
          <button
            onClick={switchMode}
            className="font-heading uppercase tracking-wider text-[var(--color-accent-neon)] hover:underline focus-visible:underline"
          >
            {isRegister ? t('login.toggle.signin') : t('login.toggle.register')}
          </button>
        </p>
      </div>
    </div>
  );
}
