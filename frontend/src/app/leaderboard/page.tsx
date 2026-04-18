'use client';

import { useEffect, useState } from 'react';
import { api, type LeaderboardEntry } from '@/lib/api';

const MEDAL_COLORS = ['#ffd700', '#c0c0c0', '#cd7f32'];
const MEDAL_LABELS = ['Gold', 'Silver', 'Bronze'];

function reactionColor(ms: number) {
  if (ms <= 0) return 'var(--color-text-muted)';
  if (ms < 500) return '#00ff88';
  if (ms < 1000) return '#ffd700';
  return '#ff3548';
}

function SkeletonRow() {
  return (
    <div className="grid grid-cols-12 gap-2 px-4 py-3 border border-[var(--color-border-default)] rounded-sm">
      <div className="col-span-1 skeleton h-4" />
      <div className="col-span-4 skeleton h-4" />
      <div className="col-span-3 skeleton h-4" />
      <div className="col-span-2 skeleton h-4" />
      <div className="col-span-2 skeleton h-4" />
    </div>
  );
}

export default function LeaderboardPage() {
  const [entries, setEntries] = useState<LeaderboardEntry[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.leaderboard.get(50)
      .then(setEntries)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  const top3 = entries.slice(0, 3);
  const rest = entries.slice(3);

  return (
    <div className="relative min-h-screen px-6 py-12 overflow-hidden">
      {/* Ambient orb */}
      <div className="orb w-[500px] h-[500px] opacity-[0.04] animate-float"
           style={{ background: '#ffd700', top: '-10%', right: '-10%' }} />

      <div className="relative z-10 max-w-3xl mx-auto">
        {/* Header */}
        <div className="text-center mb-14 animate-fade-in-up">
          <div className="badge mb-4">Global Rankings</div>
          <h1 className="font-heading font-bold text-4xl sm:text-5xl uppercase tracking-wider">
            Leader<span className="text-gradient-neon">board</span>
          </h1>
          <p className="text-sm text-[var(--color-text-muted)] mt-3">
            Top snipers ranked by total score
          </p>
        </div>

        {loading ? (
          <div className="space-y-2">
            {Array.from({ length: 8 }).map((_, i) => <SkeletonRow key={i} />)}
          </div>
        ) : entries.length === 0 ? (
          <div className="text-center py-24 text-[var(--color-text-muted)]">
            <div className="text-5xl mb-4" aria-hidden="true">◎</div>
            <p className="font-heading text-xl uppercase tracking-wider mb-2">No players yet</p>
            <p className="text-sm">Be the first to play and claim #1!</p>
          </div>
        ) : (
          <>
            {/* Podium — top 3 */}
            {top3.length > 0 && (
              <div className="grid grid-cols-3 gap-4 mb-10 animate-fade-in-up delay-100" aria-label="Top 3 players">
                {[
                  /* reorder: 2nd, 1st, 3rd */
                  top3[1] ?? null,
                  top3[0] ?? null,
                  top3[2] ?? null,
                ].map((entry, slot) => {
                  if (!entry) return <div key={slot} />;
                  const idx = (entry.rank - 1) as 0 | 1 | 2;
                  const color = MEDAL_COLORS[idx] ?? '#888';
                  const heights = ['h-24', 'h-32', 'h-20'];
                  const orderHeight = slot === 1 ? heights[0] : slot === 0 ? heights[1] : heights[2];
                  return (
                    <div key={entry.user_id} className="flex flex-col items-center gap-2">
                      {/* Avatar */}
                      <div className="w-12 h-12 flex items-center justify-center border-2 font-heading font-bold text-xl"
                           style={{ borderColor: color, color, borderRadius: '2px' }}>
                        {entry.username[0]?.toUpperCase()}
                      </div>
                      <div className="text-center">
                        <div className="font-heading font-bold text-sm uppercase tracking-wider" style={{ color }}>
                          {entry.username}
                        </div>
                        <div className="font-mono text-xs text-[var(--color-text-secondary)]">
                          {entry.total_score.toLocaleString()} pts
                        </div>
                      </div>
                      {/* Podium block */}
                      <div
                        className={`w-full ${orderHeight} flex items-center justify-center border font-heading font-bold text-lg`}
                        style={{ borderColor: `${color}40`, background: `${color}08`, color, borderRadius: '2px' }}
                        aria-label={`Rank ${entry.rank} — ${MEDAL_LABELS[idx]}`}
                      >
                        #{entry.rank}
                      </div>
                    </div>
                  );
                })}
              </div>
            )}

            {/* Table header */}
            <div className="grid grid-cols-12 gap-2 px-4 py-2 mb-1 text-[10px] font-heading uppercase tracking-widest text-[var(--color-text-muted)]"
                 role="row">
              <div className="col-span-1" role="columnheader">Rank</div>
              <div className="col-span-4" role="columnheader">Player</div>
              <div className="col-span-3 text-right" role="columnheader">Score</div>
              <div className="col-span-2 text-right" role="columnheader">Games</div>
              <div className="col-span-2 text-right" role="columnheader">Best</div>
            </div>

            {/* Rows 4+ */}
            <div className="space-y-1.5" role="table" aria-label="Leaderboard">
              {rest.map((entry, i) => (
                <div
                  key={entry.user_id}
                  role="row"
                  className="card grid grid-cols-12 gap-2 items-center px-4 py-3 animate-fade-in-up"
                  style={{ animationDelay: `${i * 0.04}s` }}
                >
                  <div className="col-span-1 font-heading font-bold text-sm text-[var(--color-text-muted)]" role="cell">
                    #{entry.rank}
                  </div>
                  <div className="col-span-4 font-heading font-bold text-sm uppercase tracking-wider truncate" role="cell">
                    {entry.username}
                  </div>
                  <div className="col-span-3 text-right font-mono font-bold text-sm" style={{ color: '#00ff88' }} role="cell">
                    {entry.total_score.toLocaleString()}
                  </div>
                  <div className="col-span-2 text-right font-mono text-xs text-[var(--color-text-secondary)]" role="cell">
                    {entry.games_played}
                  </div>
                  <div
                    className="col-span-2 text-right font-mono text-xs"
                    style={{ color: reactionColor(entry.best_reaction_ms) }}
                    role="cell"
                  >
                    {entry.best_reaction_ms > 0 ? `${entry.best_reaction_ms}ms` : '—'}
                  </div>
                </div>
              ))}
            </div>
          </>
        )}
      </div>
    </div>
  );
}
