'use client';

import { useEffect, useState } from 'react';
import { api, type LeaderboardEntry } from '@/lib/api';

export default function LeaderboardPage() {
  const [entries, setEntries] = useState<LeaderboardEntry[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.leaderboard.get(50)
      .then(setEntries)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  const getRankStyle = (rank: number) => {
    if (rank === 1) return { color: '#ffd700', border: '1px solid rgba(255,215,0,0.3)' };
    if (rank === 2) return { color: '#c0c0c0', border: '1px solid rgba(192,192,192,0.2)' };
    if (rank === 3) return { color: '#cd7f32', border: '1px solid rgba(205,127,50,0.2)' };
    return {};
  };

  return (
    <div className="min-h-screen px-6 py-12">
      <div className="max-w-3xl mx-auto">
        <h1 className="font-heading font-bold text-4xl mb-2 uppercase tracking-wider text-center">
          Leader<span style={{ color: '#00ff88' }}>board</span>
        </h1>
        <p className="text-center text-sm text-[var(--color-text-muted)] mb-10">Top snipers ranked by total score</p>

        {loading ? (
          <div className="text-center py-20">
            <div className="w-10 h-10 border-2 border-[var(--color-accent-neon)] border-t-transparent rounded-full animate-spin mx-auto" />
          </div>
        ) : entries.length === 0 ? (
          <div className="text-center py-20 text-[var(--color-text-muted)]">
            <p className="font-heading text-xl mb-2">No players yet</p>
            <p className="text-sm">Be the first to play and claim #1!</p>
          </div>
        ) : (
          <div className="space-y-2">
            {/* Header */}
            <div className="grid grid-cols-12 gap-2 px-4 py-2 text-xs font-heading uppercase tracking-wider text-[var(--color-text-muted)]">
              <div className="col-span-1">Rank</div>
              <div className="col-span-4">Player</div>
              <div className="col-span-3 text-right">Score</div>
              <div className="col-span-2 text-right">Games</div>
              <div className="col-span-2 text-right">Best Reaction</div>
            </div>

            {entries.map((entry) => (
              <div
                key={entry.user_id}
                className="card grid grid-cols-12 gap-2 items-center px-4 py-3"
                style={getRankStyle(entry.rank)}
              >
                <div className="col-span-1 font-heading font-bold text-lg" style={getRankStyle(entry.rank)}>
                  {entry.rank <= 3 ? ['🥇', '🥈', '🥉'][entry.rank - 1] : `#${entry.rank}`}
                </div>
                <div className="col-span-4 font-heading font-bold text-sm uppercase tracking-wider">
                  {entry.username}
                </div>
                <div className="col-span-3 text-right font-mono font-bold" style={{ color: '#00ff88' }}>
                  {entry.total_score.toLocaleString()}
                </div>
                <div className="col-span-2 text-right font-mono text-sm text-[var(--color-text-secondary)]">
                  {entry.games_played}
                </div>
                <div className="col-span-2 text-right font-mono text-sm"
                     style={{ color: entry.best_reaction_ms < 500 ? '#00ff88' : entry.best_reaction_ms < 1000 ? '#ffd700' : '#ff3548' }}>
                  {entry.best_reaction_ms > 0 ? `${entry.best_reaction_ms}ms` : '—'}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
