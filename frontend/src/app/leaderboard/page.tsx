'use client';

import { useEffect, useState, useCallback } from 'react';
import { api, type LeaderboardEntry, type LeaderboardResponse } from '@/lib/api';
import { useLocale } from '@/i18n/LocaleProvider';

const MEDAL_COLORS = ['#F59E0B', '#94A3B8', '#EA580C'];
const MEDAL_BG = ['rgba(245,158,11,0.08)', 'rgba(148,163,184,0.08)', 'rgba(234,88,12,0.08)'];
const MEDAL_LABELS = ['Gold', 'Silver', 'Bronze'];
const PER_PAGE = 10;

function reactionColor(ms: number) {
  if (ms <= 0) return 'var(--color-text-muted)';
  if (ms < 500) return 'var(--color-secondary)';
  if (ms < 1000) return 'var(--color-accent-gold)';
  return 'var(--color-accent-red)';
}

function formatMs(ms: number) {
  if (ms <= 0) return '—';
  return `${ms.toLocaleString()}ms`;
}

function SkeletonRow() {
  return (
    <div className="grid grid-cols-12 gap-2 px-5 py-4 border border-[var(--color-border-default)] rounded-[var(--radius-sm)] bg-[var(--color-bg-card)]">
      <div className="col-span-1 skeleton h-4 opacity-50" />
      <div className="col-span-4 skeleton h-4 opacity-50" />
      <div className="col-span-3 skeleton h-4 opacity-50" />
      <div className="col-span-2 skeleton h-4 opacity-50" />
      <div className="col-span-2 skeleton h-4 opacity-50" />
    </div>
  );
}

export default function LeaderboardPage() {
  const { t } = useLocale();
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
    // eslint-disable-next-line
    fetchPage(1);
  }, [fetchPage]);

  const totalPages = Math.max(1, Math.ceil(total / PER_PAGE));
  const isFirstPage = page === 1;
  const top3 = isFirstPage ? entries.slice(0, 3) : [];
  const rest = isFirstPage ? entries.slice(3) : entries;

  return (
    <div className="relative min-h-screen px-4 sm:px-6 py-12 overflow-hidden">
      {/* Soft blobs */}
      <div className="bg-blob w-[400px] h-[400px] opacity-[0.08]"
           style={{ background: '#F59E0B', top: '-10%', right: '-10%' }} />
      <div className="bg-blob w-[350px] h-[350px] opacity-[0.06]"
           style={{ background: '#4F46E5', bottom: '-5%', left: '-10%', animationDelay: '-4s' }} />

      <div className="relative z-10 max-w-4xl mx-auto">
        {/* Header */}
        <div className="text-center mb-14 motion-safe:animate-fade-in-up">
          <div className="badge mb-4">{t('leaderboard.badge')}</div>
          <h1 className="font-heading font-bold text-4xl sm:text-5xl lg:text-6xl tracking-tight mb-2">
            {t('leaderboard.title').replace('{accent}', '')}<span className="text-[var(--color-primary)]">{t('leaderboard.titleAccent')}</span>
          </h1>
          <p className="text-sm font-heading text-[var(--color-text-muted)] mt-4">
            {t('leaderboard.subtitle')}
          </p>
        </div>

        {loading ? (
          <div className="space-y-3 relative z-10">
            {Array.from({ length: 8 }).map((_, i) => <SkeletonRow key={i} />)}
          </div>
        ) : entries.length === 0 ? (
          <div className="text-center py-24 text-[var(--color-text-muted)] relative z-10 card">
            <div className="text-5xl mb-4 opacity-30" aria-hidden="true">
              <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" className="mx-auto" strokeLinecap="round"><path d="M6 9H4.5a2.5 2.5 0 0 1 0-5H6" /><path d="M18 9h1.5a2.5 2.5 0 0 0 0-5H18" /><path d="M4 22h16" /><path d="M18 2H6v7a6 6 0 0 0 12 0V2Z" /></svg>
            </div>
            <p className="font-heading text-xl tracking-tight mb-2 text-[var(--color-primary)]">{t('leaderboard.empty.title')}</p>
            <p className="text-sm text-[var(--color-text-secondary)]">{t('leaderboard.empty.desc')}</p>
          </div>
        ) : (
          <div className="relative z-10">
            {/* Podium — top 3 (only on first page) */}
            {top3.length > 0 && (
              <div className="grid grid-cols-3 gap-3 sm:gap-6 mb-12 motion-safe:animate-fade-in-up delay-100 items-end px-2 sm:px-10" aria-label="Top 3 learners">
                {[
                  top3[1] ?? null,
                  top3[0] ?? null,
                  top3[2] ?? null,
                ].map((entry, slot) => {
                  if (!entry) return <div key={slot} />;
                  const idx = (entry.rank - 1) as 0 | 1 | 2;
                  const color = MEDAL_COLORS[idx] ?? '#888';
                  const bg = MEDAL_BG[idx] ?? 'transparent';
                  
                  const heights = ['h-28', 'h-40', 'h-24'];
                  const orderHeight = slot === 1 ? heights[1] : slot === 0 ? heights[0] : heights[2];
                  
                  return (
                    <div key={entry.user_id} className="flex flex-col items-center gap-4 transition-transform duration-300 hover:-translate-y-1">
                      <div className="flex flex-col items-center text-center">
                        <div className="relative mb-3">
                          <div className="w-14 h-14 sm:w-16 sm:h-16 flex items-center justify-center border-2 font-heading font-bold text-2xl relative z-10 rounded-full"
                               style={{ 
                                 borderColor: color, 
                                 color: color,
                                 background: bg,
                               }}>
                            {entry.username[0]?.toUpperCase()}
                          </div>
                          <div className="absolute -bottom-2 -right-2 w-6 h-6 flex items-center justify-center font-heading font-bold text-xs z-20 rounded-full"
                               style={{ background: color, color: '#fff' }}>
                            {entry.rank}
                          </div>
                        </div>
                        <div className="font-heading font-bold text-base tracking-tight truncate w-full px-1" style={{ color }}>
                          {entry.username}
                        </div>
                        <div className="font-mono font-bold text-sm mt-1" style={{ color: reactionColor(entry.avg_reaction_ms) }}>
                          {formatMs(entry.avg_reaction_ms)} <span className="text-[10px] text-[var(--color-text-muted)]">AVG</span>
                        </div>
                      </div>

                      <div
                        className={`w-full ${orderHeight} flex items-end justify-center pb-4 relative overflow-hidden rounded-t-[var(--radius-md)]`}
                        style={{ 
                          background: bg, 
                          borderTop: `2px solid ${color}`,
                        }}
                        aria-label={`Rank ${entry.rank} — ${MEDAL_LABELS[idx]}`}
                      >
                        <span className="font-heading font-bold text-2xl relative z-10 opacity-20" style={{ color }}>
                          0{entry.rank}
                        </span>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}

            {/* Table Header */}
            <div className="grid grid-cols-12 gap-3 px-5 py-3 mb-2 text-xs font-heading font-medium text-[var(--color-text-muted)] border-b border-[var(--color-border-default)] sticky top-0 bg-[rgba(250,251,254,0.92)] backdrop-blur-md z-20"
                 role="row">
              <div className="col-span-1 hidden sm:block" role="columnheader">{t('leaderboard.col.rank')}</div>
              <div className="col-span-2 sm:col-span-1 block sm:hidden" role="columnheader">#</div>
              <div className="col-span-4 sm:col-span-3" role="columnheader">{t('leaderboard.col.player')}</div>
              <div className="col-span-3 sm:col-span-3 text-right" role="columnheader">{t('leaderboard.col.avgReaction')}</div>
              <div className="col-span-2 text-right hidden sm:block" role="columnheader">{t('leaderboard.col.games')}</div>
              <div className="col-span-3 sm:col-span-3 text-right" role="columnheader">{t('leaderboard.col.best')}</div>
            </div>

            {/* Rows */}
            <div className="space-y-2 relative" role="table" aria-label="Leaderboard Ranking">
              {rest.map((entry, i) => (
                <div
                  key={entry.user_id}
                  role="row"
                  className="grid grid-cols-12 gap-3 items-center px-5 py-3.5 motion-safe:animate-fade-in-up transition-all hover:bg-[var(--color-bg-hover)] group rounded-[var(--radius-sm)]"
                  style={{ 
                    animationDelay: `${i * 0.04}s`,
                    background: 'var(--color-bg-card)',
                    border: '1px solid var(--color-border-default)',
                  }}
                >
                  <div className="col-span-2 sm:col-span-1 font-heading font-bold text-sm text-[var(--color-text-secondary)] group-hover:text-[var(--color-text-primary)] transition-colors" role="cell">
                    {entry.rank.toString().padStart(2, '0')}
                  </div>
                  
                  <div className="col-span-4 sm:col-span-3 flex items-center gap-3 overflow-hidden" role="cell">
                    <div className="w-6 h-6 hidden sm:flex items-center justify-center shrink-0 text-[10px] font-heading font-bold rounded-full"
                         style={{ background: 'rgba(79,70,229,0.08)', color: 'var(--color-primary)' }}>
                      {entry.username[0]?.toUpperCase()}
                    </div>
                    <span className="font-heading font-bold text-sm tracking-tight truncate text-[var(--color-text-primary)]">
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
                  className="px-4 py-2 font-heading font-bold text-sm border transition-all cursor-pointer disabled:opacity-30 disabled:cursor-not-allowed rounded-[var(--radius-sm)]"
                  style={{
                    borderColor: 'var(--color-border-default)',
                    color: 'var(--color-primary)',
                    background: 'var(--color-bg-card)',
                  }}
                >
                  {t('leaderboard.prev')}
                </button>
                <span className="font-mono text-sm text-[var(--color-text-muted)]">
                  {t('leaderboard.page')} <span className="text-[var(--color-text-primary)] font-bold">{page}</span> / {totalPages}
                </span>
                <button
                  onClick={() => fetchPage(page + 1)}
                  disabled={page >= totalPages}
                  className="px-4 py-2 font-heading font-bold text-sm border transition-all cursor-pointer disabled:opacity-30 disabled:cursor-not-allowed rounded-[var(--radius-sm)]"
                  style={{
                    borderColor: 'var(--color-border-default)',
                    color: 'var(--color-primary)',
                    background: 'var(--color-bg-card)',
                  }}
                >
                  {t('leaderboard.next')}
                </button>
              </div>
            )}

            {/* Total count */}
            <div className="text-center mt-4 text-xs font-mono text-[var(--color-text-muted)]">
              {t('leaderboard.snipersRanked', { count: total })}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
