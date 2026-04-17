'use client';

import Link from 'next/link';
import { useAuth } from '@/hooks/useAuth';
import { useEffect, useState } from 'react';
import { api } from '@/lib/api';

export default function HomePage() {
  const { user } = useAuth();
  const [onlineCount, setOnlineCount] = useState(0);

  useEffect(() => {
    api.online().then(d => setOnlineCount(d.online)).catch(() => {});
  }, []);

  return (
    <div className="relative overflow-hidden">
      {/* Hero */}
      <section className="relative min-h-[90vh] flex flex-col items-center justify-center px-6">
        {/* Grid background */}
        <div className="absolute inset-0 opacity-[0.03]" style={{
          backgroundImage: `
            linear-gradient(rgba(0,255,136,0.3) 1px, transparent 1px),
            linear-gradient(90deg, rgba(0,255,136,0.3) 1px, transparent 1px)
          `,
          backgroundSize: '60px 60px',
        }} />

        {/* Decorative crosshair */}
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[300px] h-[300px] opacity-5 pointer-events-none">
          <div className="absolute top-0 left-1/2 w-px h-full bg-[var(--color-accent-neon)]" />
          <div className="absolute top-1/2 left-0 w-full h-px bg-[var(--color-accent-neon)]" />
          <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-16 h-16 border border-[var(--color-accent-neon)] rounded-full" />
        </div>

        <div className="relative z-10 text-center max-w-3xl">
          {/* Badge */}
          <div className="inline-flex items-center gap-2 px-3 py-1 mb-8 border border-[var(--color-border-default)] text-xs font-heading uppercase tracking-widest text-[var(--color-text-muted)]"
               style={{ borderRadius: '2px' }}>
            <span className="w-2 h-2 rounded-full bg-[var(--color-accent-neon)] animate-pulse" />
            {onlineCount > 0 ? `${onlineCount} players online` : 'Ready to play'}
          </div>

          {/* Title */}
          <h1 className="font-heading font-bold text-7xl md:text-8xl tracking-tight leading-none mb-6">
            <span className="block text-glow" style={{ color: '#00ff88' }}>LINGO</span>
            <span className="block text-[var(--color-text-primary)]">SNIPER</span>
          </h1>

          {/* Subtitle */}
          <p className="text-lg text-[var(--color-text-secondary)] max-w-xl mx-auto mb-10 leading-relaxed">
            Train your reflex aim while mastering foreign vocabulary.
            Click the correct word target before time runs out.
            Challenge friends in real-time 1v1 duels.
          </p>

          {/* CTA */}
          <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
            <Link href={user ? "/play" : "/login"} className="btn-primary text-lg px-8 py-4">
              {user ? "START TRAINING" : "SIGN UP & PLAY"}
            </Link>
            <Link href="/leaderboard" className="btn-secondary text-lg px-8 py-4">
              LEADERBOARD
            </Link>
          </div>
        </div>
      </section>

      {/* Features */}
      <section className="py-24 px-6 border-t border-[var(--color-border-default)]">
        <div className="max-w-5xl mx-auto">
          <h2 className="font-heading font-bold text-3xl text-center mb-16 uppercase tracking-wider">
            How It <span style={{ color: '#00ff88' }}>Works</span>
          </h2>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            {/* Feature 1 */}
            <div className="card group">
              <div className="w-12 h-12 flex items-center justify-center border border-[var(--color-accent-neon)] text-[var(--color-accent-neon)] font-heading font-bold text-xl mb-4"
                   style={{ borderRadius: '2px' }}>
                01
              </div>
              <h3 className="font-heading font-bold text-xl mb-2 uppercase">Choose Language</h3>
              <p className="text-sm text-[var(--color-text-secondary)] leading-relaxed">
                Select English or Chinese vocabulary sets. Three difficulty levels from beginner to advanced.
              </p>
            </div>

            {/* Feature 2 */}
            <div className="card group">
              <div className="w-12 h-12 flex items-center justify-center border border-[var(--color-accent-orange)] text-[var(--color-accent-orange)] font-heading font-bold text-xl mb-4"
                   style={{ borderRadius: '2px' }}>
                02
              </div>
              <h3 className="font-heading font-bold text-xl mb-2 uppercase">Aim & Click</h3>
              <p className="text-sm text-[var(--color-text-secondary)] leading-relaxed">
                Word targets spawn randomly. Read the meaning, find and click the correct word as fast as you can.
              </p>
            </div>

            {/* Feature 3 */}
            <div className="card group">
              <div className="w-12 h-12 flex items-center justify-center border border-[var(--color-accent-cyan)] text-[var(--color-accent-cyan)] font-heading font-bold text-xl mb-4"
                   style={{ borderRadius: '2px' }}>
                03
              </div>
              <h3 className="font-heading font-bold text-xl mb-2 uppercase">1v1 Duel</h3>
              <p className="text-sm text-[var(--color-text-secondary)] leading-relaxed">
                Real-time multiplayer duels via WebSocket. Same targets, same timer. Fastest and most accurate player wins.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Stats teaser */}
      <section className="py-16 px-6 border-t border-[var(--color-border-default)]">
        <div className="max-w-3xl mx-auto grid grid-cols-3 gap-8 text-center">
          <div>
            <div className="text-4xl font-heading font-bold" style={{ color: '#00ff88' }}>50+</div>
            <div className="text-xs text-[var(--color-text-muted)] font-heading uppercase tracking-wider mt-1">Vocabulary Words</div>
          </div>
          <div>
            <div className="text-4xl font-heading font-bold" style={{ color: '#ff6b35' }}>2</div>
            <div className="text-xs text-[var(--color-text-muted)] font-heading uppercase tracking-wider mt-1">Languages</div>
          </div>
          <div>
            <div className="text-4xl font-heading font-bold" style={{ color: '#00d4ff' }}>{"<50ms"}</div>
            <div className="text-xs text-[var(--color-text-muted)] font-heading uppercase tracking-wider mt-1">WS Latency</div>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="py-8 px-6 border-t border-[var(--color-border-default)] text-center">
        <p className="text-xs text-[var(--color-text-muted)] font-heading">
          LINGO SNIPER — Built with Go + Next.js + WebSocket
        </p>
      </footer>
    </div>
  );
}
