'use client';

import { useState, useCallback, useEffect, useRef } from 'react';
import type { Target } from '@/hooks/useWebSocket';

interface GameCanvasProps {
  targets: Target[];
  question: string;
  round: number;
  totalRounds: number;
  timeMs: number;
  myScore: number;
  opponentScore: number;
  opponent: string;
  mode: 'solo' | 'duel' | 'battle';
  lastReactionMs: number;
  onHit: (targetId: string) => number;
}

export default function GameCanvas({
  targets,
  question,
  round,
  totalRounds,
  timeMs,
  myScore,
  opponentScore,
  opponent,
  mode,
  lastReactionMs,
  onHit,
}: GameCanvasProps) {
  const [hitTargets, setHitTargets] = useState<Set<string>>(new Set());
  const [answered, setAnswered] = useState(false);
  const [showPopup, setShowPopup] = useState<{ x: number; y: number; text: string; correct: boolean } | null>(null);
  const [timeLeft, setTimeLeft] = useState(timeMs);
  const timerRef = useRef<NodeJS.Timeout | null>(null);

  useEffect(() => {
    setHitTargets(new Set());
    setAnswered(false);
    setTimeLeft(timeMs);
    setShowPopup(null);

    timerRef.current = setInterval(() => {
      setTimeLeft(prev => {
        if (prev <= 0) {
          if (timerRef.current) clearInterval(timerRef.current);
          return 0;
        }
        return prev - 100;
      });
    }, 100);

    return () => {
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, [round, timeMs]);

  const handleTargetClick = useCallback((target: Target, e: React.MouseEvent) => {
    if (answered) return;

    const reactionMs = onHit(target.id);

    setHitTargets(prev => new Set(prev).add(target.id));
    setAnswered(true);

    const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
    setShowPopup({
      x: rect.left + rect.width / 2,
      y: rect.top,
      text: target.correct ? `+${Math.max(100, Math.round((timeMs - reactionMs) * 1000 / timeMs))}` : '-50',
      correct: target.correct,
    });

    setTimeout(() => setShowPopup(null), 800);
  }, [hitTargets, onHit, timeMs]);

  const timePercent = (timeLeft / timeMs) * 100;
  const timeColor = timePercent > 50 ? '#00ff88' : timePercent > 25 ? '#ffd700' : '#ff3548';

  return (
    <div className="relative w-full h-full min-h-[400px] sm:min-h-[500px] crosshair-cursor select-none overflow-hidden">
      {/* HUD Top Bar */}
      <div className="absolute top-0 left-0 right-0 z-20 flex items-center justify-between px-3 sm:px-6 py-2 sm:py-3"
           style={{ background: 'linear-gradient(180deg, rgba(10,14,23,0.95) 0%, transparent 100%)' }}>
        {/* Score P1 */}
        <div className="flex items-center gap-2 sm:gap-3">
          <div className="text-xs sm:text-sm text-[var(--color-text-secondary)] font-heading uppercase tracking-wider">You</div>
          <div className="text-xl sm:text-3xl font-heading font-bold text-glow" style={{ color: '#00ff88' }}>{myScore}</div>
        </div>

        {/* Round info */}
        <div className="text-center">
          <div className="text-[10px] sm:text-xs text-[var(--color-text-muted)] font-heading uppercase tracking-widest">Round</div>
          <div className="text-lg sm:text-2xl font-heading font-bold">{round}<span className="text-[var(--color-text-muted)]">/{totalRounds}</span></div>
        </div>

        {/* Score P2 / Mode badge */}
        {mode === 'duel' && (
          <div className="flex items-center gap-2 sm:gap-3">
            <div className="text-xl sm:text-3xl font-heading font-bold" style={{ color: '#ff6b35' }}>{opponentScore}</div>
            <div className="text-xs sm:text-sm text-[var(--color-text-secondary)] font-heading uppercase tracking-wider">{opponent}</div>
          </div>
        )}
        {mode === 'battle' && (
          <div className="flex items-center gap-2">
            <div className="text-xs sm:text-sm text-[var(--color-text-muted)] font-heading uppercase tracking-wider px-2 sm:px-3 py-1 border border-[var(--color-accent-cyan)] bg-[rgba(0,212,255,0.05)]" style={{ borderRadius: '2px', color: '#00d4ff' }}>
              BATTLE
            </div>
          </div>
        )}
        {mode === 'solo' && <div className="w-16 sm:w-24" />}
      </div>

      {/* Timer Bar */}
      <div className="absolute top-12 sm:top-16 left-0 right-0 z-20 h-1 bg-[var(--color-bg-card)]">
        <div
          className="h-full transition-all duration-100 ease-linear"
          style={{ width: `${timePercent}%`, backgroundColor: timeColor }}
        />
      </div>

      {/* Question */}
      <div className="absolute top-14 sm:top-20 left-1/2 -translate-x-1/2 z-20 text-center">
        <div className="text-[10px] sm:text-xs text-[var(--color-text-muted)] font-heading uppercase tracking-widest mb-1">Find the word for</div>
        <div className="text-lg sm:text-2xl font-heading font-bold text-[var(--color-accent-cyan)] px-3 sm:px-4 py-1.5 sm:py-2 border border-[var(--color-accent-cyan)] bg-[rgba(0,212,255,0.05)]"
             style={{ borderRadius: '2px' }}>
          {question}
        </div>
      </div>

      {/* Reaction time */}
      {lastReactionMs > 0 && (
        <div className="absolute top-28 sm:top-36 left-1/2 -translate-x-1/2 z-20">
          <span className="font-mono text-xs sm:text-sm" style={{ color: lastReactionMs < 1000 ? '#00ff88' : lastReactionMs < 2000 ? '#ffd700' : '#ff3548' }}>
            {lastReactionMs}ms
          </span>
        </div>
      )}

      {/* Game Area - Targets */}
      <div className="absolute inset-0 pt-32 sm:pt-40 pb-4">
        {targets.map(target => (
          <button
            key={target.id}
            onClick={(e) => handleTargetClick(target, e)}
            disabled={answered}
            className={`absolute px-2.5 sm:px-4 py-2 sm:py-3 font-heading font-bold text-sm sm:text-lg border-2 transition-all
              ${hitTargets.has(target.id)
                ? 'target-hit opacity-0 pointer-events-none'
                : 'target-spawn target-pulse hover:scale-110 cursor-crosshair'
              }`}
            style={{
              left: `${target.x}%`,
              top: `${target.y}%`,
              transform: 'translate(-50%, -50%)',
              backgroundColor: hitTargets.has(target.id)
                ? 'transparent'
                : 'rgba(26, 35, 50, 0.9)',
              borderColor: hitTargets.has(target.id)
                ? 'transparent'
                : '#00ff88',
              color: '#e8ecf1',
              borderRadius: '2px',
              minWidth: '60px',
              textAlign: 'center',
            }}
          >
            {target.word}
          </button>
        ))}
      </div>

      {/* Score popup */}
      {showPopup && (
        <div
          className="fixed z-50 score-popup font-heading font-bold text-xl sm:text-2xl pointer-events-none"
          style={{
            left: showPopup.x,
            top: showPopup.y,
            color: showPopup.correct ? '#00ff88' : '#ff3548',
            transform: 'translateX(-50%)',
          }}
        >
          {showPopup.text}
        </div>
      )}

      {/* Grid overlay for HUD feel */}
      <div className="absolute inset-0 pointer-events-none opacity-5"
           style={{
             backgroundImage: `
               linear-gradient(rgba(0,255,136,0.1) 1px, transparent 1px),
               linear-gradient(90deg, rgba(0,255,136,0.1) 1px, transparent 1px)
             `,
             backgroundSize: '40px 40px',
           }}
      />
    </div>
  );
}
