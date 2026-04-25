'use client';

import type { LeaderboardPlayer } from '@/hooks/useWebSocket';
import { useLocale } from '@/i18n/LocaleProvider';

interface LiveLeaderboardProps {
  players: LeaderboardPlayer[];
  round: number;
}

export default function LiveLeaderboard({ players, round }: LiveLeaderboardProps) {
  const { t } = useLocale();

  if (!players || players.length === 0) return null;

  return (
    <div className="absolute top-20 right-4 z-30 w-64">
      <div className="text-xs font-heading font-medium text-[var(--color-text-muted)] mb-2 text-right">
        {t('live.ranking')}
      </div>
      <div className="space-y-1">
        {players.map((p, i) => (
          <div
            key={p.username}
            className="flex items-center justify-between px-3 py-2 text-sm rounded-[var(--radius-sm)]"
            style={{
              background: i === 0 ? 'rgba(79, 70, 229, 0.06)' : 'var(--color-bg-card)',
              borderLeft: i === 0 ? '2px solid var(--color-primary)' : '2px solid transparent',
              border: i !== 0 ? '1px solid var(--color-border-default)' : undefined,
            }}
          >
            <div className="flex items-center gap-2">
              <span className="font-mono text-xs" style={{
                color: i === 0 ? '#F59E0B' : i === 1 ? '#94A3B8' : i === 2 ? '#EA580C' : 'var(--color-text-muted)',
              }}>
                #{p.rank}
              </span>
              <span className="font-heading font-bold truncate max-w-[80px]" style={{
                color: i === 0 ? 'var(--color-primary)' : 'var(--color-text-primary)',
              }}>
                {p.username}
              </span>
            </div>
            <div className="flex items-center gap-2">
              <span className="font-mono font-bold text-[var(--color-primary)]">
                {p.correct_count}
              </span>
              {p.avg_reaction_ms > 0 && (
                <span className="font-mono text-xs text-[var(--color-text-muted)]">
                  {p.avg_reaction_ms}ms
                </span>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
