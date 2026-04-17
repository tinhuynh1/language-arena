'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useAuth } from '@/hooks/useAuth';
import { api, type User } from '@/lib/api';

export default function DashboardPage() {
  const { user: authUser } = useAuth();
  const router = useRouter();
  const [stats, setStats] = useState<{ user: User; recent_games: unknown[] } | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!authUser) {
      router.push('/login');
      return;
    }
    api.stats.me()
      .then(setStats)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [authUser, router]);

  if (!authUser) return null;

  const user = stats?.user || authUser;

  return (
    <div className="min-h-screen px-6 py-12">
      <div className="max-w-3xl mx-auto">
        <h1 className="font-heading font-bold text-4xl mb-8 uppercase tracking-wider">
          Dashboard
        </h1>

        {loading ? (
          <div className="text-center py-20">
            <div className="w-10 h-10 border-2 border-[var(--color-accent-neon)] border-t-transparent rounded-full animate-spin mx-auto" />
          </div>
        ) : (
          <>
            {/* Player Card */}
            <div className="card mb-8">
              <div className="flex items-center gap-4 mb-6">
                <div className="w-16 h-16 flex items-center justify-center border-2 border-[var(--color-accent-neon)] font-heading font-bold text-2xl"
                     style={{ borderRadius: '2px', color: '#00ff88' }}>
                  {user.username[0]?.toUpperCase()}
                </div>
                <div>
                  <div className="font-heading font-bold text-xl uppercase tracking-wider">{user.username}</div>
                  <div className="text-sm text-[var(--color-text-muted)]">{user.email}</div>
                </div>
              </div>

              <div className="grid grid-cols-3 gap-4">
                <div className="text-center p-4 bg-[var(--color-bg-primary)]" style={{ borderRadius: '2px' }}>
                  <div className="text-3xl font-heading font-bold" style={{ color: '#00ff88' }}>
                    {user.total_score.toLocaleString()}
                  </div>
                  <div className="text-xs text-[var(--color-text-muted)] font-heading uppercase tracking-wider mt-1">Total Score</div>
                </div>
                <div className="text-center p-4 bg-[var(--color-bg-primary)]" style={{ borderRadius: '2px' }}>
                  <div className="text-3xl font-heading font-bold" style={{ color: '#ff6b35' }}>
                    {user.games_played}
                  </div>
                  <div className="text-xs text-[var(--color-text-muted)] font-heading uppercase tracking-wider mt-1">Games Played</div>
                </div>
                <div className="text-center p-4 bg-[var(--color-bg-primary)]" style={{ borderRadius: '2px' }}>
                  <div className="text-3xl font-heading font-bold font-mono"
                       style={{ color: user.best_reaction_ms < 500 ? '#00ff88' : user.best_reaction_ms < 1000 ? '#ffd700' : '#00d4ff' }}>
                    {user.best_reaction_ms > 0 ? `${user.best_reaction_ms}ms` : '—'}
                  </div>
                  <div className="text-xs text-[var(--color-text-muted)] font-heading uppercase tracking-wider mt-1">Best Reaction</div>
                </div>
              </div>
            </div>

            {/* Quick Actions */}
            <div className="flex gap-4">
              <Link href="/play" className="btn-primary flex-1 text-center">
                PLAY NOW
              </Link>
              <Link href="/leaderboard" className="btn-secondary flex-1 text-center">
                LEADERBOARD
              </Link>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
