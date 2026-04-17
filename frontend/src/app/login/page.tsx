'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useAuth } from '@/hooks/useAuth';

export default function LoginPage() {
  const [isRegister, setIsRegister] = useState(false);
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const { login, register } = useAuth();
  const router = useRouter();

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
      setError(err instanceof Error ? err.message : 'Something went wrong');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center px-6">
      <div className="w-full max-w-md">
        {/* Header */}
        <div className="text-center mb-8">
          <Link href="/" className="inline-block mb-6">
            <div className="font-heading font-bold text-4xl tracking-wider">
              <span style={{ color: '#00ff88' }}>LINGO</span> SNIPER
            </div>
          </Link>
          <h1 className="font-heading font-bold text-2xl uppercase tracking-wider">
            {isRegister ? 'Create Account' : 'Sign In'}
          </h1>
          <p className="text-sm text-[var(--color-text-muted)] mt-1">
            {isRegister ? 'Join the arena and start training' : 'Welcome back, soldier'}
          </p>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="card space-y-4">
          {isRegister && (
            <div>
              <label htmlFor="username" className="block text-xs font-heading uppercase tracking-wider text-[var(--color-text-muted)] mb-1">
                Username
              </label>
              <input
                id="username"
                type="text"
                className="input-field"
                placeholder="SniperElite"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                required
                minLength={3}
              />
            </div>
          )}

          <div>
            <label htmlFor="email" className="block text-xs font-heading uppercase tracking-wider text-[var(--color-text-muted)] mb-1">
              Email
            </label>
            <input
              id="email"
              type="email"
              className="input-field"
              placeholder="sniper@arena.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </div>

          <div>
            <label htmlFor="password" className="block text-xs font-heading uppercase tracking-wider text-[var(--color-text-muted)] mb-1">
              Password
            </label>
            <input
              id="password"
              type="password"
              className="input-field"
              placeholder="••••••••"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={6}
            />
          </div>

          {error && (
            <div className="text-sm text-[var(--color-accent-red)] bg-[rgba(255,53,72,0.1)] border border-[rgba(255,53,72,0.3)] px-3 py-2"
                 style={{ borderRadius: '2px' }}>
              {error}
            </div>
          )}

          <button type="submit" disabled={loading} className="btn-primary w-full">
            {loading ? 'Loading...' : isRegister ? 'CREATE ACCOUNT' : 'SIGN IN'}
          </button>
        </form>

        {/* Toggle */}
        <div className="text-center mt-4">
          <button
            onClick={() => { setIsRegister(!isRegister); setError(''); }}
            className="text-sm text-[var(--color-text-muted)] hover:text-[var(--color-accent-neon)] transition-colors font-heading"
          >
            {isRegister ? 'Already have an account? Sign In' : "Don't have an account? Register"}
          </button>
        </div>
      </div>
    </div>
  );
}
