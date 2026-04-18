'use client';

import Link from 'next/link';
import { useAuth } from '@/hooks/useAuth';
import { useEffect, useState } from 'react';
import { api } from '@/lib/api';

const MODES = [
  {
    id: 'solo',
    label: 'Solo',
    color: '#00ff88',
    desc: 'Train alone at your own pace. Perfect for building vocabulary and improving reaction time.',
    icon: '◎',
  },
  {
    id: 'duel',
    label: '1v1 Duel',
    color: '#ff6b35',
    desc: 'Face another player in real-time. Same targets, same clock — fastest and most accurate wins.',
    icon: '⊕',
  },
  {
    id: 'battle',
    label: 'Battle Royale',
    color: '#00d4ff',
    desc: 'Up to 8 players compete simultaneously. Climb the live leaderboard round by round.',
    icon: '⊞',
  },
];

const QUIZ_TYPES = [
  { label: 'Meaning → Word', desc: 'See the definition, click the right word.' },
  { label: 'Word → Meaning', desc: 'See the word, pick the correct definition.' },
  { label: 'Word → IPA', desc: 'English pronunciation training with IPA.' },
  { label: 'Word → Pinyin', desc: 'Chinese pronunciation with Pinyin romanization.' },
];

export default function HomePage() {
  const { user } = useAuth();
  const [onlineCount, setOnlineCount] = useState(0);

  useEffect(() => {
    api.online().then(d => setOnlineCount(d.online)).catch(() => {});
  }, []);

  return (
    <div className="relative overflow-hidden">
      {/* ── Hero ─────────────────────────────────────────────── */}
      <section className="relative min-h-[100vh] flex flex-col items-center justify-center px-6 pt-20">
        {/* Ambient orbs */}
        <div className="orb w-[600px] h-[600px] opacity-[0.07] animate-float"
             style={{ background: '#00ff88', top: '-10%', left: '-15%', animationDelay: '0s' }} />
        <div className="orb w-[400px] h-[400px] opacity-[0.05] animate-float"
             style={{ background: '#00d4ff', bottom: '5%', right: '-10%', animationDelay: '-4s' }} />
        <div className="orb w-[300px] h-[300px] opacity-[0.04] animate-float"
             style={{ background: '#a855f7', top: '30%', right: '5%', animationDelay: '-8s' }} />

        {/* Grid */}
        <div className="absolute inset-0 opacity-[0.025]" aria-hidden="true" style={{
          backgroundImage: `linear-gradient(rgba(0,255,136,0.4) 1px, transparent 1px),
                            linear-gradient(90deg, rgba(0,255,136,0.4) 1px, transparent 1px)`,
          backgroundSize: '60px 60px',
        }} />

        <div className="relative z-10 text-center max-w-4xl animate-fade-in-up">
          {/* Live badge */}
          <div className="inline-flex items-center gap-2 mb-8 badge" aria-label={`${onlineCount} players online`}>
            <span className="w-1.5 h-1.5 rounded-full bg-[var(--color-accent-neon)] animate-pulse" aria-hidden="true" />
            {onlineCount > 0 ? `${onlineCount} players online` : 'Ready to play'}
          </div>

          {/* Headline */}
          <h1 className="font-heading font-bold leading-none tracking-tight mb-6">
            <span className="block text-[clamp(4rem,10vw,8rem)] text-gradient-neon text-glow">LINGO</span>
            <span className="block text-[clamp(4rem,10vw,8rem)] text-[var(--color-text-primary)]">SNIPER</span>
          </h1>

          <p className="text-base sm:text-lg text-[var(--color-text-secondary)] max-w-2xl mx-auto mb-12 leading-relaxed animate-fade-in-up delay-200">
            Master vocabulary through reflex-driven gameplay.
            Read the cue, target the answer, beat your opponent — all in under a second.
          </p>

          {/* CTA row */}
          <div className="flex flex-col sm:flex-row gap-4 justify-center items-center animate-fade-in-up delay-300">
            <Link href={user ? '/play' : '/login'} className="btn-primary text-base px-10 py-4">
              {user ? 'Enter Arena' : 'Start Playing Free'}
            </Link>
            <Link href="/leaderboard" className="btn-secondary text-base px-10 py-4">
              Leaderboard
            </Link>
          </div>

          {/* Mini stats row */}
          <div className="flex items-center justify-center gap-8 mt-14 animate-fade-in-up delay-400">
            {[['380+', 'Words'], ['2', 'Languages'], ['4', 'Quiz Modes']].map(([val, label]) => (
              <div key={label} className="text-center">
                <div className="text-2xl font-heading font-bold stat-shimmer">{val}</div>
                <div className="text-[10px] font-heading uppercase tracking-widest text-[var(--color-text-muted)] mt-0.5">{label}</div>
              </div>
            ))}
          </div>
        </div>

        {/* Scroll hint */}
        <div className="absolute bottom-8 left-1/2 -translate-x-1/2 flex flex-col items-center gap-2 opacity-40" aria-hidden="true">
          <div className="w-px h-10 bg-[var(--color-accent-neon)]" style={{ animation: 'scanline 2s ease-in-out infinite' }} />
        </div>
      </section>

      {/* ── Game Modes ───────────────────────────────────────── */}
      <section className="py-24 px-6 border-t border-[var(--color-border-default)]" aria-labelledby="modes-heading">
        <div className="max-w-5xl mx-auto">
          <div className="text-center mb-16">
            <div className="badge mb-4">Game Modes</div>
            <h2 id="modes-heading" className="font-heading font-bold text-3xl sm:text-4xl uppercase tracking-wider">
              Pick Your <span className="text-gradient-neon">Battle</span>
            </h2>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {MODES.map((mode) => (
              <div key={mode.id} className="card group relative overflow-hidden">
                {/* Accent bar */}
                <div className="absolute top-0 left-0 right-0 h-[2px] transition-all duration-300 group-hover:opacity-100 opacity-60"
                     style={{ background: mode.color }} />
                <div className="text-3xl mb-4" style={{ color: mode.color }} aria-hidden="true">{mode.icon}</div>
                <h3 className="font-heading font-bold text-xl uppercase tracking-wider mb-2" style={{ color: mode.color }}>
                  {mode.label}
                </h3>
                <p className="text-sm text-[var(--color-text-secondary)] leading-relaxed">{mode.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── How It Works ─────────────────────────────────────── */}
      <section className="py-24 px-6 border-t border-[var(--color-border-default)]" aria-labelledby="how-heading">
        <div className="max-w-5xl mx-auto">
          <div className="text-center mb-16">
            <div className="badge mb-4">Gameplay</div>
            <h2 id="how-heading" className="font-heading font-bold text-3xl sm:text-4xl uppercase tracking-wider">
              How It <span className="text-gradient-neon">Works</span>
            </h2>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-8">
            {[
              { num: '01', color: '#00ff88', title: 'Choose Your Setup', desc: 'Select language (English or Chinese), difficulty level, and quiz type. Your opponent is matched automatically.' },
              { num: '02', color: '#ff6b35', title: 'Aim & Fire', desc: 'Word targets appear across the canvas. Read the prompt and click the correct target before time expires.' },
              { num: '03', color: '#00d4ff', title: 'Score & Climb', desc: 'Faster, more accurate answers earn more points. Track your best reaction time and climb the global rankings.' },
            ].map(({ num, color, title, desc }) => (
              <div key={num} className="card group">
                <div className="w-11 h-11 flex items-center justify-center border font-heading font-bold text-base mb-5 transition-all duration-200 group-hover:scale-110"
                     style={{ borderColor: color, color, borderRadius: '2px' }}>
                  {num}
                </div>
                <h3 className="font-heading font-bold text-lg uppercase tracking-wider mb-2">{title}</h3>
                <p className="text-sm text-[var(--color-text-secondary)] leading-relaxed">{desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── Quiz Types ───────────────────────────────────────── */}
      <section className="py-24 px-6 border-t border-[var(--color-border-default)]" aria-labelledby="quiz-heading">
        <div className="max-w-4xl mx-auto">
          <div className="text-center mb-16">
            <div className="badge mb-4">Quiz Types</div>
            <h2 id="quiz-heading" className="font-heading font-bold text-3xl sm:text-4xl uppercase tracking-wider">
              4 Ways to <span className="text-gradient-neon">Train</span>
            </h2>
            <p className="text-sm text-[var(--color-text-muted)] mt-4 max-w-xl mx-auto">
              Each quiz type challenges a different aspect of vocabulary mastery. Rotate between them for complete fluency.
            </p>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {QUIZ_TYPES.map(({ label, desc }, i) => (
              <div key={label} className="card flex items-start gap-4">
                <div className="w-8 h-8 flex-shrink-0 flex items-center justify-center border border-[var(--color-accent-neon)] text-[var(--color-accent-neon)] font-heading font-bold text-xs"
                     style={{ borderRadius: '2px' }}>
                  {String(i + 1).padStart(2, '0')}
                </div>
                <div>
                  <div className="font-heading font-bold text-sm uppercase tracking-wider mb-1 text-[var(--color-text-primary)]">{label}</div>
                  <div className="text-sm text-[var(--color-text-secondary)]">{desc}</div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── Final CTA ────────────────────────────────────────── */}
      <section className="py-24 px-6 border-t border-[var(--color-border-default)]">
        <div className="max-w-2xl mx-auto text-center">
          <h2 className="font-heading font-bold text-3xl sm:text-5xl uppercase tracking-wider mb-6">
            Ready to <span className="text-gradient-neon">Snipe</span>?
          </h2>
          <p className="text-[var(--color-text-secondary)] mb-10 leading-relaxed">
            Free to play. No downloads. Compete in real-time from your browser.
          </p>
          <Link href={user ? '/play' : '/login'} className="btn-primary text-base px-12 py-4">
            {user ? 'Play Now' : 'Create Free Account'}
          </Link>
        </div>
      </section>

      {/* ── Footer ───────────────────────────────────────────── */}
      <footer className="py-8 px-6 border-t border-[var(--color-border-default)]">
        <div className="max-w-7xl mx-auto flex flex-col sm:flex-row items-center justify-between gap-4">
          <div className="font-heading font-bold text-sm text-[var(--color-text-muted)] tracking-wider">
            <span className="text-[var(--color-accent-neon)]">LINGO</span> SNIPER
          </div>
          <p className="text-xs text-[var(--color-text-muted)] font-heading">
            Built with Go · Next.js · WebSocket · PostgreSQL
          </p>
          <div className="flex items-center gap-4 text-xs font-heading uppercase tracking-wider text-[var(--color-text-muted)]">
            <Link href="/play" className="hover:text-[var(--color-accent-neon)] transition-colors">Play</Link>
            <Link href="/leaderboard" className="hover:text-[var(--color-accent-neon)] transition-colors">Ranks</Link>
          </div>
        </div>
      </footer>
    </div>
  );
}
