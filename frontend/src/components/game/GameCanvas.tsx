'use client';

import { useState, useCallback, useEffect, useRef } from 'react';
import type { Target, QuizType } from '@/hooks/useWebSocket';

interface GameCanvasProps {
  targets: Target[];
  question: string;
  quizType: QuizType;
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

const QUIZ_LABELS: Record<QuizType, string> = {
  meaning_to_word: 'Find the word for',
  word_to_meaning: 'Find the meaning of',
  word_to_ipa: 'Find the IPA for',
  word_to_pinyin: 'Find the pinyin for',
};

export default function GameCanvas({
  targets,
  question,
  quizType,
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
    <div className="relative w-full h-full min-h-[400px] sm:min-h-[500px] crosshair-cursor select-none overflow-hidden vignette">
      {/* HUD Top Bar */}
      <div className="absolute top-0 left-0 right-0 z-20 flex items-center justify-between px-4 sm:px-8 py-3 sm:py-4"
           style={{ background: 'linear-gradient(180deg, rgba(8,12,20,0.97) 0%, rgba(8,12,20,0.7) 70%, transparent 100%)' }}>
        {/* Score P1 */}
        <div className="flex items-center gap-3 sm:gap-4">
          <div className="text-sm sm:text-base text-[var(--color-text-secondary)] font-heading uppercase tracking-wider">You</div>
          <div className="text-2xl sm:text-4xl font-heading font-bold text-glow" style={{ color: '#00ff88' }}>{myScore}</div>
        </div>

        {/* Round info */}
        <div className="text-center">
          <div className="text-xs sm:text-sm text-[var(--color-text-muted)] font-heading uppercase tracking-widest">Round</div>
          <div className="text-xl sm:text-3xl font-heading font-bold">{round}<span className="text-[var(--color-text-muted)]">/{totalRounds}</span></div>
        </div>

        {/* Score P2 / Mode badge */}
        {mode === 'duel' && (
          <div className="flex items-center gap-3 sm:gap-4">
            <div className="text-2xl sm:text-4xl font-heading font-bold" style={{ color: '#ff6b35' }}>{opponentScore}</div>
            <div className="text-sm sm:text-base text-[var(--color-text-secondary)] font-heading uppercase tracking-wider">{opponent}</div>
          </div>
        )}
        {mode === 'battle' && (
          <div className="flex items-center gap-2">
            <div className="text-sm font-heading uppercase tracking-wider px-3 sm:px-4 py-1.5 border border-[var(--color-accent-cyan)] bg-[rgba(0,212,255,0.05)]" style={{ borderRadius: '3px', color: '#00d4ff' }}>
              BATTLE
            </div>
          </div>
        )}
        {mode === 'solo' && <div className="w-20 sm:w-28" />}
      </div>

      {/* Timer Bar */}
      <div className="absolute top-14 sm:top-[4.5rem] left-0 right-0 z-20 h-1.5 bg-[var(--color-bg-card)]">
        <div
          className="h-full transition-all duration-100 ease-linear"
          style={{
            width: `${timePercent}%`,
            backgroundColor: timeColor,
            boxShadow: timePercent < 25 ? `0 0 12px ${timeColor}80` : undefined,
          }}
        />
      </div>

      {/* Question */}
      <div className="absolute top-16 sm:top-22 left-1/2 -translate-x-1/2 z-20 text-center">
        <div className="text-xs sm:text-sm text-[var(--color-text-muted)] font-heading uppercase tracking-widest mb-1.5">{QUIZ_LABELS[quizType]}</div>
        <div className="text-xl sm:text-3xl font-heading font-bold text-[var(--color-accent-cyan)] px-4 sm:px-6 py-2 sm:py-3 border-2 border-[var(--color-accent-cyan)] bg-[rgba(0,212,255,0.06)]"
             style={{
               borderRadius: '3px',
               boxShadow: '0 0 20px rgba(0, 212, 255, 0.15)',
             }}>
          {question}
        </div>
      </div>

      {/* Reaction time */}
      {lastReactionMs > 0 && (
        <div className="absolute top-32 sm:top-40 left-1/2 -translate-x-1/2 z-20">
          <span className="font-mono text-sm sm:text-base font-bold" style={{ color: lastReactionMs < 1000 ? '#00ff88' : lastReactionMs < 2000 ? '#ffd700' : '#ff3548' }}>
            {lastReactionMs}ms
          </span>
        </div>
      )}

      {/* Game Area - Targets */}
      <div className="absolute inset-0 pt-36 sm:pt-44 pb-4">
        {targets.map(target => (
          <button
            key={target.id}
            onClick={(e) => handleTargetClick(target, e)}
            disabled={answered}
            className={`absolute px-3 sm:px-5 py-2.5 sm:py-3.5 font-heading font-bold text-base sm:text-lg border-2 transition-all
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
                : 'rgba(16, 24, 40, 0.92)',
              borderColor: hitTargets.has(target.id)
                ? 'transparent'
                : '#00ff88',
              color: '#e8ecf1',
              borderRadius: '3px',
              minWidth: '70px',
              textAlign: 'center',
            }}
          >
            {target.label || target.word}
          </button>
        ))}
      </div>

      {/* Score popup */}
      {showPopup && (
        <div
          className="fixed z-50 score-popup font-heading font-bold text-2xl sm:text-3xl pointer-events-none"
          style={{
            left: showPopup.x,
            top: showPopup.y,
            color: showPopup.correct ? '#00ff88' : '#ff3548',
            transform: 'translateX(-50%)',
            textShadow: showPopup.correct
              ? '0 0 16px rgba(0,255,136,0.6)'
              : '0 0 16px rgba(255,53,72,0.6)',
          }}
        >
          {showPopup.text}
        </div>
      )}

      {/* Grid overlay for HUD feel */}
      <div className="absolute inset-0 pointer-events-none opacity-[0.03]"
           style={{
             backgroundImage: `
               linear-gradient(rgba(0,255,136,0.15) 1px, transparent 1px),
               linear-gradient(90deg, rgba(0,255,136,0.15) 1px, transparent 1px)
             `,
             backgroundSize: '50px 50px',
           }}
      />
    </div>
  );
}
