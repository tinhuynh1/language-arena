'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useAuth } from '@/hooks/useAuth';
import { api, type User } from '@/lib/api';
import { useLocale } from '@/i18n/LocaleProvider';

function reactionColor(ms: number) {
  if (ms <= 0) return 'var(--color-accent)';
  if (ms < 500) return 'var(--color-secondary)';
  if (ms < 1000) return 'var(--color-accent-gold)';
  return 'var(--color-accent-red)';
}

function StatCard({ value, label, color }: { value: string; label: string; color: string }) {
  return (
    <div className="card text-center py-5">
      <div className="font-heading font-bold text-3xl sm:text-4xl mb-1" style={{ color }}>{value}</div>
      <div className="text-xs font-heading text-[var(--color-text-muted)] mt-1">{label}</div>
    </div>
  );
}

export default function DashboardPage() {
  const { user: authUser } = useAuth();
  const { t } = useLocale();
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

  const user = stats?.user ?? authUser;

  return (
    <div className="relative min-h-screen px-6 py-12 overflow-hidden">
      {/* Soft blobs */}
      <div className="bg-blob w-[350px] h-[350px] opacity-[0.08]"
           style={{ background: '#4F46E5', top: '-5%', right: '-10%' }} />
      <div className="bg-blob w-[300px] h-[300px] opacity-[0.06]"
           style={{ background: '#0D9488', bottom: '10%', left: '-10%', animationDelay: '-6s' }} />

      <div className="relative z-10 max-w-2xl mx-auto">
        <div className="mb-10 animate-fade-in-up">
          <div className="badge mb-4">{t('dashboard.badge')}</div>
          <h1 className="font-heading font-bold text-4xl sm:text-5xl tracking-tight">
            {t('dashboard.title')}
          </h1>
        </div>

        {loading ? (
          <div className="space-y-4">
            <div className="skeleton h-32 rounded-[var(--radius-md)]" />
            <div className="grid grid-cols-3 gap-4">
              {[1, 2, 3].map(i => <div key={i} className="skeleton h-24 rounded-[var(--radius-md)]" />)}
            </div>
          </div>
        ) : (
          <>
            {/* Player card */}
            <div className="card mb-6 animate-fade-in-up delay-100">
              <div className="flex items-center gap-5">
                {/* Avatar */}
                <div
                  className="w-16 h-16 flex-shrink-0 flex items-center justify-center rounded-full bg-[rgba(79,70,229,0.08)] border-2 border-[var(--color-primary)] font-heading font-bold text-2xl text-[var(--color-primary)]"
                  aria-label={`Avatar for ${user.username}`}
                >
                  {user.username[0]?.toUpperCase()}
                </div>
                <div className="min-w-0">
                  <div className="font-heading font-bold text-2xl tracking-tight truncate text-[var(--color-primary)]">
                    {user.username}
                  </div>
                  <div className="text-sm text-[var(--color-text-muted)] truncate">{user.email}</div>
                </div>
              </div>
            </div>

            {/* Stats grid */}
            <div className="grid grid-cols-3 gap-4 mb-6 animate-fade-in-up delay-200">
              <StatCard
                value={user.avg_reaction_ms > 0 ? `${user.avg_reaction_ms}ms` : '—'}
                label={t('dashboard.avgReaction')}
                color={reactionColor(user.avg_reaction_ms)}
              />
              <StatCard
                value={String(user.games_played)}
                label={t('dashboard.gamesPlayed')}
                color="var(--color-accent-orange)"
              />
              <StatCard
                value={user.best_reaction_ms > 0 ? `${user.best_reaction_ms}ms` : '—'}
                label={t('dashboard.bestReaction')}
                color={reactionColor(user.best_reaction_ms)}
              />
            </div>

            {/* Actions */}
            <div className="flex gap-4 animate-fade-in-up delay-300">
              <Link href="/play" className="btn-primary flex-1 text-center">
                {t('dashboard.playNow')}
              </Link>
              <Link href="/leaderboard" className="btn-secondary flex-1 text-center">
                {t('dashboard.leaderboard')}
              </Link>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
