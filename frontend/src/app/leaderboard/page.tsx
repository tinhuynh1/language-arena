'use client';

import { useEffect, useState, useCallback } from 'react';
import { api, type LeaderboardEntry, type LeaderboardResponse } from '@/lib/api';

const MEDAL_COLORS = ['#ffd700', '#c0c0c0', '#cd7f32'];
const MEDAL_GLOW = ['rgba(255,215,0,0.5)', 'rgba(192,192,192,0.5)', 'rgba(205,127,50,0.5)'];
const MEDAL_BG = ['rgba(255,215,0,0.05)', 'rgba(192,192,192,0.05)', 'rgba(205,127,50,0.05)'];
const MEDAL_LABELS = ['Gold', 'Silver', 'Bronze'];
const PER_PAGE = 10;

function reactionColor(ms: number) {
  if (ms <= 0) return 'var(--color-text-muted)';
  if (ms < 500) return '#00ff88';
  if (ms < 1000) return '#ffd700';
  return '#ff3548';
}

function formatMs(ms: number) {
  if (ms <= 0) return '—';
  return `${ms.toLocaleString()}ms`;
}

function SkeletonRow() {
  return (
    <div className="grid grid-cols-12 gap-2 px-5 py-4 border border-[var(--color-border-default)] rounded-sm bg-[rgba(255,255,255,0.02)]">
      <div className="col-span-1 skeleton h-4 opacity-50" />
      <div className="col-span-4 skeleton h-4 opacity-50" />
      <div className="col-span-3 skeleton h-4 opacity-50" />
      <div className="col-span-2 skeleton h-4 opacity-50" />
      <div className="col-span-2 skeleton h-4 opacity-50" />
    </div>
  );
}

export default function LeaderboardPage() {
  const [entries, setEntries] = useState<LeaderboardEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);

  const fetchPage = useCallback((p: number) => {
    setLoading(true);
    api.leaderboard.get(p, PER_PAGE)
      .then((res: LeaderboardResponse) => {
        setEntries(res.entries);
        setTotal(res.total);
        setPage(res.page);
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    fetchPage(1);
  }, [fetchPage]);

  const totalPages = Math.max(1, Math.ceil(total / PER_PAGE));
  const isFirstPage = page === 1;
  const top3 = isFirstPage ? entries.slice(0, 3) : [];
  const rest = isFirstPage ? entries.slice(3) : entries;

  return (
    <div className="relative min-h-screen px-4 sm:px-6 py-12 overflow-hidden">
      {/* Background grid */}
      <div className="absolute inset-0 opacity-[0.03]" aria-hidden="true" style={{
        backgroundImage: `linear-gradient(rgba(0,212,255,0.4) 1px, transparent 1px),
                          linear-gradient(90deg, rgba(0,212,255,0.4) 1px, transparent 1px)`,
        backgroundSize: '60px 60px',
      }} />

      {/* Ambient orbs */}
      <div className="orb w-[500px] h-[500px] opacity-[0.05] motion-safe:animate-float"
           style={{ background: '#ffd700', top: '-10%', right: '-10%' }} />
      <div className="orb w-[400px] h-[400px] opacity-[0.03] motion-safe:animate-float"
           style={{ background: '#00d4ff', bottom: '-5%', left: '-10%', animationDelay: '-4s' }} />

      {/* Scanline overlay */}
      <div className="absolute inset-0 pointer-events-none overflow-hidden z-[2]" style={{ opacity: 0.015 }}>
        <div className="motion-safe:animate-[scanline_10s_linear_infinite]" style={{
          width: '100%',
          height: '200%',
          backgroundImage: 'repeating-linear-gradient(0deg, transparent, transparent 2px, rgba(0,255,136,0.4) 2px, rgba(0,255,136,0.4) 4px)'
        }} />
      </div>

      <div className="relative z-10 max-w-4xl mx-auto">
        {/* Header */}
        <div className="text-center mb-14 motion-safe:animate-fade-in-up">
          <div className="inline-block px-3 py-1 mb-4 text-[10px] font-heading uppercase tracking-[0.3em] rounded-sm"
               style={{ background: 'rgba(0,212,255,0.1)', color: '#00d4ff', border: '1px solid rgba(0,212,255,0.2)' }}>
            Global Rankings
          </div>
          <h1 className="font-heading font-bold text-4xl sm:text-5xl lg:text-6xl uppercase tracking-wider mb-2">
            Leader<span className="text-glow-cyan" style={{ color: '#00d4ff' }}>board</span>
          </h1>
          <p className="text-sm font-heading tracking-widest text-[var(--color-text-muted)] mt-4">
            TOP SNIPERS RANKED BY FASTEST REACTION TIME
          </p>
        </div>

        {loading ? (
          <div className="space-y-3 relative z-10">
            {Array.from({ length: 8 }).map((_, i) => <SkeletonRow key={i} />)}
          </div>
        ) : entries.length === 0 ? (
          <div className="text-center py-24 text-[var(--color-text-muted)] relative z-10 card" style={{ background: 'rgba(255,255,255,0.01)' }}>
            <div className="text-5xl mb-4 opacity-50" aria-hidden="true">◎</div>
            <p className="font-heading text-xl uppercase tracking-wider mb-2" style={{ color: '#00d4ff' }}>No players yet</p>
            <p className="text-sm text-[var(--color-text-secondary)]">The arena is empty. Be the first to claim #1!</p>
          </div>
        ) : (
          <div className="relative z-10">
            {/* Podium — top 3 (only on first page) */}
            {top3.length > 0 && (
              <div className="grid grid-cols-3 gap-3 sm:gap-6 mb-12 motion-safe:animate-fade-in-up delay-100 items-end px-2 sm:px-10" aria-label="Top 3 players">
                {[
                  top3[1] ?? null,
                  top3[0] ?? null,
                  top3[2] ?? null,
                ].map((entry, slot) => {
                  if (!entry) return <div key={slot} />;
                  const idx = (entry.rank - 1) as 0 | 1 | 2;
                  const color = MEDAL_COLORS[idx] ?? '#888';
                  const glow = MEDAL_GLOW[idx] ?? 'rgba(136,136,136,0.5)';
                  const bg = MEDAL_BG[idx] ?? 'transparent';
                  
                  const heights = ['h-28', 'h-40', 'h-24'];
                  const orderHeight = slot === 1 ? heights[1] : slot === 0 ? heights[0] : heights[2];
                  
                  return (
                    <div key={entry.user_id} className="flex flex-col items-center gap-4 transition-transform duration-300 hover:-translate-y-2">
                      <div className="flex flex-col items-center text-center">
                        <div className="relative mb-3">
                          <div className="w-14 h-14 sm:w-16 sm:h-16 flex items-center justify-center border-2 font-heading font-bold text-2xl relative z-10"
                               style={{ 
                                 borderColor: color, 
                                 color: '#fff',
                                 background: `linear-gradient(135deg, ${bg}, rgba(0,0,0,0.8))`,
                                 boxShadow: `0 0 20px ${glow}`,
                                 borderRadius: '2px' 
                               }}>
                            {entry.username[0]?.toUpperCase()}
                          </div>
                          <div className="absolute -bottom-2 -right-2 w-6 h-6 flex items-center justify-center font-heading font-bold text-xs z-20"
                               style={{ background: color, color: '#000', borderRadius: '50%', boxShadow: `0 0 10px ${color}` }}>
                            {entry.rank}
                          </div>
                        </div>
                        <div className="font-heading font-bold text-base uppercase tracking-wider truncate w-full px-1" style={{ color, textShadow: `0 0 10px ${glow}` }}>
                          {entry.username}
                        </div>
                        <div className="font-mono font-bold text-sm mt-1" style={{ color: reactionColor(entry.avg_reaction_ms), textShadow: `0 0 10px ${reactionColor(entry.avg_reaction_ms)}33` }}>
                          {formatMs(entry.avg_reaction_ms)} <span className="text-[10px] text-[var(--color-text-muted)] tracking-widest">AVG</span>
                        </div>
                      </div>

                      <div
                        className={`w-full ${orderHeight} flex items-end justify-center pb-4 tracking-widest relative overflow-hidden`}
                        style={{ 
                          background: `linear-gradient(to top, ${bg}, transparent)`, 
                          borderTop: `2px solid ${color}`,
                          boxShadow: `inset 0 20px 20px -20px ${color}, 0 -5px 15px -10px ${color}`,
                          borderRadius: '2px'
                        }}
                        aria-label={`Rank ${entry.rank} — ${MEDAL_LABELS[idx]}`}
                      >
                        <div className="absolute inset-0 opacity-20" style={{
                          backgroundImage: `linear-gradient(transparent 1px, ${color} 1px)`,
                          backgroundSize: '100% 4px',
                        }} />
                        <span className="font-heading font-bold text-2xl relative z-10 opacity-30" style={{ color }}>
                          0{entry.rank}
                        </span>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}

            {/* Table Header */}
            <div className="grid grid-cols-12 gap-3 px-5 py-3 mb-2 text-[10px] font-heading uppercase tracking-[0.2em] text-[var(--color-text-muted)] border-b border-[rgba(255,255,255,0.05)] sticky top-0 bg-[rgba(8,12,20,0.8)] backdrop-blur-md z-20"
                 role="row">
              <div className="col-span-1 hidden sm:block" role="columnheader">Rank</div>
              <div className="col-span-2 sm:col-span-1 block sm:hidden" role="columnheader">#</div>
              <div className="col-span-4 sm:col-span-3" role="columnheader">Player</div>
              <div className="col-span-3 sm:col-span-3 text-right" role="columnheader">Avg Reaction</div>
              <div className="col-span-2 text-right hidden sm:block" role="columnheader">Games</div>
              <div className="col-span-3 sm:col-span-3 text-right" role="columnheader">Best</div>
            </div>

            {/* Rows */}
            <div className="space-y-2 relative" role="table" aria-label="Leaderboard Ranking">
              {rest.map((entry, i) => (
                <div
                  key={entry.user_id}
                  role="row"
                  className="grid grid-cols-12 gap-3 items-center px-5 py-3.5 motion-safe:animate-fade-in-up transition-all hover:scale-[1.01] hover:bg-[rgba(255,255,255,0.03)] group"
                  style={{ 
                    animationDelay: `${i * 0.04}s`,
                    background: 'rgba(255,255,255,0.01)',
                    border: '1px solid rgba(255,255,255,0.04)',
                    borderRadius: '2px',
                  }}
                >
                  <div className="col-span-2 sm:col-span-1 font-heading font-bold text-sm text-[var(--color-text-secondary)] group-hover:text-white transition-colors" role="cell">
                    {entry.rank.toString().padStart(2, '0')}
                  </div>
                  
                  <div className="col-span-4 sm:col-span-3 flex items-center gap-3 overflow-hidden" role="cell">
                    <div className="w-6 h-6 hidden sm:flex items-center justify-center shrink-0 text-[10px] font-heading font-bold"
                         style={{ background: 'rgba(0,212,255,0.1)', color: '#00d4ff', border: '1px solid rgba(0,212,255,0.2)' }}>
                      {entry.username[0]?.toUpperCase()}
                    </div>
                    <span className="font-heading font-bold text-sm uppercase tracking-wider truncate text-[var(--color-text-primary)]">
                      {entry.username}
                    </span>
                  </div>
                  
                  <div className="col-span-3 sm:col-span-3 text-right font-mono font-bold text-sm sm:text-base" style={{ color: reactionColor(entry.avg_reaction_ms) }} role="cell">
                    {formatMs(entry.avg_reaction_ms)}
                  </div>
                  
                  <div className="col-span-2 text-right font-mono text-xs text-[var(--color-text-secondary)] hidden sm:block" role="cell">
                    {entry.games_played}
                  </div>
                  
                  <div
                    className="col-span-3 sm:col-span-3 text-right font-mono font-bold text-xs"
                    style={{ color: reactionColor(entry.best_reaction_ms) }}
                    role="cell"
                  >
                    {formatMs(entry.best_reaction_ms)}
                  </div>
                </div>
              ))}
            </div>

            {/* Pagination */}
            {totalPages > 1 && (
              <div className="flex items-center justify-center gap-4 mt-8 motion-safe:animate-fade-in-up">
                <button
                  onClick={() => fetchPage(page - 1)}
                  disabled={page <= 1}
                  className="px-4 py-2 font-heading font-bold text-sm uppercase tracking-wider border transition-all cursor-pointer disabled:opacity-30 disabled:cursor-not-allowed"
                  style={{
                    borderColor: 'rgba(0,212,255,0.3)',
                    color: '#00d4ff',
                    background: 'rgba(0,212,255,0.05)',
                    borderRadius: '3px',
                  }}
                >
                  ← Prev
                </button>
                <span className="font-mono text-sm text-[var(--color-text-muted)]">
                  Page <span className="text-[var(--color-text-primary)] font-bold">{page}</span> / {totalPages}
                </span>
                <button
                  onClick={() => fetchPage(page + 1)}
                  disabled={page >= totalPages}
                  className="px-4 py-2 font-heading font-bold text-sm uppercase tracking-wider border transition-all cursor-pointer disabled:opacity-30 disabled:cursor-not-allowed"
                  style={{
                    borderColor: 'rgba(0,212,255,0.3)',
                    color: '#00d4ff',
                    background: 'rgba(0,212,255,0.05)',
                    borderRadius: '3px',
                  }}
                >
                  Next →
                </button>
              </div>
            )}

            {/* Total count */}
            <div className="text-center mt-4 text-xs font-mono text-[var(--color-text-muted)]">
              {total} snipers ranked
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
