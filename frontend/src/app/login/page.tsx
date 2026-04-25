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
    <div className="relative min-h-screen flex items-center justify-center px-6 overflow-hidden bg-[var(--color-bg-secondary)]">
      {/* Soft blobs */}
      <div className="bg-blob w-[400px] h-[400px] opacity-[0.1]"
           style={{ background: '#4F46E5', top: '-15%', left: '-10%' }} />
      <div className="bg-blob w-[350px] h-[350px] opacity-[0.07]"
           style={{ background: '#0D9488', bottom: '-10%', right: '-10%', animationDelay: '-5s' }} />

      <div className="relative z-10 w-full max-w-sm animate-fade-in-up">
        {/* Logo */}
        <div className="text-center mb-10">
          <Link href="/" aria-label="Go to homepage">
            <div className="inline-block font-heading font-bold text-2xl tracking-tight mb-5 hover:opacity-80 transition-opacity">
              <span className="text-gradient-primary">LinguaLeap</span>
            </div>
          </Link>
          <h1 className="font-heading font-bold text-2xl tracking-tight">
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
              <label htmlFor="username" className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
                {t('login.label.username')}
              </label>
              <input
                id="username"
                name="username"
                type="text"
                autoComplete="username"
                className="input-field"
                placeholder="Your name"
                value={username}
                onChange={e => setUsername(e.target.value)}
                required
                minLength={3}
              />
            </div>
          )}

          <div>
            <label htmlFor="email" className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              {t('login.label.email')}
            </label>
            <input
              id="email"
              name="email"
              type="email"
              autoComplete="email"
              className="input-field"
              placeholder="you@example.com"
              value={email}
              onChange={e => setEmail(e.target.value)}
              required
            />
          </div>

          <div>
            <label htmlFor="password" className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
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
              className="text-sm text-[var(--color-accent-red)] bg-red-50 border border-red-200 px-4 py-3 rounded-[var(--radius-sm)]"
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
            className="font-heading font-medium text-[var(--color-primary)] hover:underline focus-visible:underline cursor-pointer"
          >
            {isRegister ? t('login.toggle.signin') : t('login.toggle.register')}
          </button>
        </p>
      </div>
    </div>
  );
}
