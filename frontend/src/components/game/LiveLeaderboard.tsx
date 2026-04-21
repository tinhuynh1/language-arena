'use client';

import type { LeaderboardPlayer } from '@/hooks/useWebSocket';

interface LiveLeaderboardProps {
  players: LeaderboardPlayer[];
  round: number;
}

export default function LiveLeaderboard({ players, round }: LiveLeaderboardProps) {
  if (!players || players.length === 0) return null;

  return (
    <div className="absolute top-20 right-4 z-30 w-64">
      <div className="text-xs font-heading uppercase tracking-widest text-[var(--color-text-muted)] mb-2 text-right">
        Live Ranking
      </div>
      <div className="space-y-1">
        {players.map((p, i) => (
          <div
            key={p.username}
            className="flex items-center justify-between px-3 py-2 text-sm"
            style={{
              background: i === 0 ? 'rgba(0, 255, 136, 0.1)' : 'rgba(26, 35, 50, 0.8)',
              borderLeft: i === 0 ? '2px solid #00ff88' : '2px solid transparent',
              borderRadius: '2px',
            }}
          >
            <div className="flex items-center gap-2">
              <span className="font-mono text-xs" style={{
                color: i === 0 ? '#ffd700' : i === 1 ? '#c0c0c0' : i === 2 ? '#cd7f32' : 'var(--color-text-muted)',
              }}>
                #{p.rank}
              </span>
              <span className="font-heading font-bold truncate max-w-[80px]" style={{
                color: i === 0 ? '#00ff88' : 'var(--color-text-primary)',
              }}>
                {p.username}
              </span>
            </div>
            <div className="flex items-center gap-2">
              <span className="font-mono font-bold" style={{ color: '#00ff88' }}>
                {p.correct_count}
              </span>
              {p.avg_reaction_ms > 0 && (
                <span className="font-mono text-xs" style={{ color: 'var(--color-text-muted)' }}>
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
