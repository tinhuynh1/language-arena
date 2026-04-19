'use client';

import { useState, useCallback, useEffect, useRef, useMemo } from 'react';
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

/* Floating particles for atmosphere */
function Particles() {
  const particles = useMemo(() =>
    Array.from({ length: 35 }, (_, i) => ({
      id: i,
      x: Math.random() * 100,
      y: Math.random() * 100,
      size: Math.random() * 3 + 1,
      duration: Math.random() * 15 + 10,
      delay: Math.random() * -20,
      opacity: Math.random() * 0.3 + 0.05,
    })), []);

  return (
    <div className="absolute inset-0 pointer-events-none overflow-hidden z-0">
      {particles.map(p => (
        <div
          key={p.id}
          className="absolute rounded-full"
          style={{
            left: `${p.x}%`,
            top: `${p.y}%`,
            width: `${p.size}px`,
            height: `${p.size}px`,
            background: p.size > 2.5 ? '#00ff88' : '#00d4ff',
            opacity: p.opacity,
            animation: `float ${p.duration}s ease-in-out infinite`,
            animationDelay: `${p.delay}s`,
            filter: p.size > 2 ? `blur(0.5px)` : undefined,
            boxShadow: p.size > 2.5 ? `0 0 ${p.size * 3}px rgba(0,255,136,0.3)` : undefined,
          }}
        />
      ))}
    </div>
  );
}

/* Crosshair SVG icon for HUD */
function CrosshairIcon({ size = 16, color = 'currentColor' }: { size?: number; color?: string }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round">
      <circle cx="12" cy="12" r="10" opacity="0.4" />
      <circle cx="12" cy="12" r="4" />
      <line x1="12" y1="2" x2="12" y2="6" />
      <line x1="12" y1="18" x2="12" y2="22" />
      <line x1="2" y1="12" x2="6" y2="12" />
      <line x1="18" y1="12" x2="22" y2="12" />
    </svg>
  );
}

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


  const isTimeUp = timeLeft <= 0;

  const handleTargetClick = useCallback((target: Target, e: React.MouseEvent) => {
    if (answered || isTimeUp) return;

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
  }, [hitTargets, onHit, timeMs, isTimeUp]);

  const timePercent = (timeLeft / timeMs) * 100;
  const timeColor = timePercent > 50 ? '#00ff88' : timePercent > 25 ? '#ffd700' : '#ff3548';
  const timeSeconds = Math.ceil(timeLeft / 1000);
  const isUrgent = timePercent < 25;

  return (
    <div className="relative w-full h-full min-h-[400px] sm:min-h-[500px] crosshair-cursor select-none overflow-hidden">

      {/* ── Background Layers ─────────────────────────── */}

      {/* Radial gradient background */}
      <div className="absolute inset-0" style={{
        background: 'radial-gradient(ellipse at 50% 30%, rgba(0,255,136,0.03) 0%, transparent 50%), radial-gradient(ellipse at 80% 70%, rgba(0,212,255,0.02) 0%, transparent 40%), var(--color-bg-primary)',
      }} />

      {/* Animated particles */}
      <Particles />

      {/* Grid overlay */}
      <div className="absolute inset-0 pointer-events-none z-0" style={{
        opacity: 0.04,
        backgroundImage: `
          linear-gradient(rgba(0,255,136,0.2) 1px, transparent 1px),
          linear-gradient(90deg, rgba(0,255,136,0.2) 1px, transparent 1px)
        `,
        backgroundSize: '60px 60px',
      }} />

      {/* Vignette */}
      <div className="absolute inset-0 pointer-events-none z-[2]" style={{
        background: 'radial-gradient(ellipse at center, transparent 40%, rgba(8, 12, 20, 0.7) 100%)',
      }} />

      {/* Scanline effect */}
      <div className="absolute inset-0 pointer-events-none overflow-hidden z-[2]" style={{ opacity: 0.015 }}>
        <div style={{
          width: '100%',
          height: '200%',
          backgroundImage: 'repeating-linear-gradient(0deg, transparent, transparent 2px, rgba(0,255,136,0.4) 2px, rgba(0,255,136,0.4) 4px)',
          animation: 'scanline 6s linear infinite',
        }} />
      </div>

      {/* ── HUD Layer ─────────────────────────────────── */}

      {/* Top HUD Bar - Glassmorphism panel */}
      <div className="absolute top-0 left-0 right-0 z-30" style={{
        background: 'linear-gradient(180deg, rgba(8,12,20,0.95) 0%, rgba(8,12,20,0.8) 70%, transparent 100%)',
      }}>
        <div className="flex items-center justify-between px-4 sm:px-8 py-3 sm:py-4">

          {/* P1 Score Panel */}
          <div className="flex items-center gap-3 sm:gap-4 px-4 sm:px-5 py-2 sm:py-2.5 rounded-sm" style={{
            background: 'rgba(0, 255, 136, 0.06)',
            border: '1px solid rgba(0, 255, 136, 0.15)',
            backdropFilter: 'blur(8px)',
          }}>
            <CrosshairIcon size={18} color="#00ff88" />
            <div>
              <div className="text-[10px] sm:text-xs text-[var(--color-text-muted)] font-heading uppercase tracking-widest leading-none">You</div>
              <div className="text-2xl sm:text-4xl font-heading font-bold text-glow leading-none mt-0.5" style={{ color: '#00ff88' }}>{myScore}</div>
            </div>
          </div>

          {/* Center - Round + Timer */}
          <div className="flex flex-col items-center gap-1">
            <div className="flex items-center gap-3 px-5 sm:px-6 py-1.5 sm:py-2 rounded-sm" style={{
              background: 'rgba(255,255,255,0.03)',
              border: '1px solid rgba(255,255,255,0.06)',
              backdropFilter: 'blur(8px)',
            }}>
              <div className="text-xs sm:text-sm text-[var(--color-text-muted)] font-heading uppercase tracking-widest">Round</div>
              <div className="text-xl sm:text-3xl font-heading font-bold">
                {round}<span className="text-[var(--color-text-muted)] text-base sm:text-xl">/{totalRounds}</span>
              </div>
            </div>
            {/* Circular mini-timer */}
            <div className="flex items-center gap-2">
              <svg width="20" height="20" viewBox="0 0 24 24" className="-rotate-90">
                <circle cx="12" cy="12" r="10" fill="none" stroke="rgba(255,255,255,0.06)" strokeWidth="2" />
                <circle
                  cx="12" cy="12" r="10"
                  fill="none"
                  stroke={timeColor}
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeDasharray={2 * Math.PI * 10}
                  strokeDashoffset={2 * Math.PI * 10 * (1 - timePercent / 100)}
                  style={{ transition: 'stroke-dashoffset 0.1s linear', filter: isUrgent ? `drop-shadow(0 0 4px ${timeColor})` : undefined }}
                />
              </svg>
              <span className={`font-mono text-sm font-bold ${isUrgent ? 'animate-pulse' : ''}`} style={{ color: timeColor }}>
                {timeSeconds}s
              </span>
            </div>
          </div>

          {/* P2 / Mode badge */}
          {mode === 'duel' && (
            <div className="flex items-center gap-3 sm:gap-4 px-4 sm:px-5 py-2 sm:py-2.5 rounded-sm" style={{
              background: 'rgba(255, 107, 53, 0.06)',
              border: '1px solid rgba(255, 107, 53, 0.15)',
              backdropFilter: 'blur(8px)',
            }}>
              <div className="text-right">
                <div className="text-[10px] sm:text-xs text-[var(--color-text-muted)] font-heading uppercase tracking-widest leading-none">{opponent}</div>
                <div className="text-2xl sm:text-4xl font-heading font-bold leading-none mt-0.5" style={{ color: '#ff6b35', textShadow: '0 0 12px rgba(255,107,53,0.4)' }}>{opponentScore}</div>
              </div>
              <CrosshairIcon size={18} color="#ff6b35" />
            </div>
          )}
          {mode === 'battle' && (
            <div className="flex items-center gap-2 px-4 py-2 rounded-sm" style={{
              background: 'rgba(0, 212, 255, 0.06)',
              border: '1px solid rgba(0, 212, 255, 0.15)',
              backdropFilter: 'blur(8px)',
            }}>
              <CrosshairIcon size={16} color="#00d4ff" />
              <div className="text-sm font-heading uppercase tracking-wider font-bold" style={{ color: '#00d4ff' }}>
                BATTLE
              </div>
            </div>
          )}
          {mode === 'solo' && (
            <div className="flex items-center gap-2 px-4 py-2 rounded-sm opacity-50" style={{
              border: '1px solid rgba(255,255,255,0.06)',
            }}>
              <div className="text-sm font-heading uppercase tracking-wider text-[var(--color-text-muted)]">SOLO</div>
            </div>
          )}
        </div>

        {/* Timer Bar — full width */}
        <div className="h-1 sm:h-1.5 bg-[rgba(255,255,255,0.03)]" style={{ margin: '0 1rem' }}>
          <div
            className="h-full transition-all duration-100 ease-linear rounded-full"
            style={{
              width: `${timePercent}%`,
              background: isUrgent
                ? `linear-gradient(90deg, ${timeColor}, ${timeColor}cc)`
                : `linear-gradient(90deg, ${timeColor}, ${timeColor}99)`,
              boxShadow: `0 0 ${isUrgent ? '16' : '8'}px ${timeColor}60`,
            }}
          />
        </div>
      </div>

      {/* ── Question Area ─────────────────────────────── */}
      <div className="absolute top-[6.5rem] sm:top-[7.5rem] left-1/2 -translate-x-1/2 z-30 text-center max-w-[90vw]">
        <div className="text-xs sm:text-sm text-[var(--color-text-muted)] font-heading uppercase tracking-[0.2em] mb-2">
          {QUIZ_LABELS[quizType]}
        </div>
        <div className="relative inline-block">
          {/* Animated border glow */}
          <div className="absolute -inset-[2px] rounded-sm opacity-60" style={{
            background: 'linear-gradient(135deg, #00d4ff, #00ff88, #00d4ff)',
            backgroundSize: '200% 200%',
            animation: 'shimmer 3s linear infinite',
            filter: 'blur(1px)',
          }} />
          <div className="relative text-xl sm:text-3xl font-heading font-bold px-5 sm:px-8 py-2.5 sm:py-3.5 rounded-sm"
               style={{
                 background: 'rgba(8, 12, 20, 0.95)',
                 color: '#00d4ff',
                 textShadow: '0 0 20px rgba(0,212,255,0.4)',
               }}>
            {question}
          </div>
        </div>

        {/* Reaction time badge — below question, centered */}
        {lastReactionMs > 0 && (
          <div className="flex justify-center mt-3">
            <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full" style={{
              background: lastReactionMs < 1000 ? 'rgba(0,255,136,0.1)' : lastReactionMs < 2000 ? 'rgba(255,215,0,0.1)' : 'rgba(255,53,72,0.1)',
              border: `1px solid ${lastReactionMs < 1000 ? 'rgba(0,255,136,0.2)' : lastReactionMs < 2000 ? 'rgba(255,215,0,0.2)' : 'rgba(255,53,72,0.2)'}`,
            }}>
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke={lastReactionMs < 1000 ? '#00ff88' : lastReactionMs < 2000 ? '#ffd700' : '#ff3548'} strokeWidth="2">
                <circle cx="12" cy="12" r="10" />
                <polyline points="12 6 12 12 16 14" />
              </svg>
              <span className="font-mono text-sm font-bold" style={{ color: lastReactionMs < 1000 ? '#00ff88' : lastReactionMs < 2000 ? '#ffd700' : '#ff3548' }}>
                {lastReactionMs}ms
              </span>
            </div>
          </div>
        )}
      </div>

      {/* ── Targets ───────────────────────────────────── */}
      <div className="absolute top-[12rem] sm:top-[14rem] left-0 right-0 bottom-6 z-10">
        {targets.map(target => {
          const isHit = hitTargets.has(target.id);
          return (
            <button
              key={target.id}
              onClick={(e) => handleTargetClick(target, e)}
              disabled={answered || isTimeUp}
              className={`absolute transition-all duration-200 group focus-visible:ring-2 focus-visible:ring-[#00ff88] focus-visible:outline-none rounded-sm
                ${isHit
                  ? 'target-hit opacity-0 pointer-events-none'
                  : 'target-spawn cursor-crosshair'
                }`}
              style={{
                left: `${target.x}%`,
                top: `${target.y}%`,
                transform: 'translate(-50%, -50%)',
              }}
            >
              {!isHit && (
                <div className="relative">
                  {/* Outer glow ring */}
                  <div className="absolute -inset-1 rounded-md opacity-0 group-hover:opacity-100 transition-opacity duration-200" style={{
                    background: 'linear-gradient(135deg, rgba(0,255,136,0.3), rgba(0,212,255,0.3))',
                    filter: 'blur(6px)',
                  }} />
                  {/* Target card */}
                  <div className="relative px-4 sm:px-6 py-3 sm:py-4 rounded-md font-heading font-bold text-base sm:text-lg transition-all duration-200 group-hover:scale-105"
                       style={{
                         background: 'linear-gradient(135deg, rgba(16, 28, 44, 0.95), rgba(10, 18, 32, 0.98))',
                         border: '1px solid rgba(0, 255, 136, 0.35)',
                         color: '#e8ecf1',
                         minWidth: '80px',
                         textAlign: 'center',
                         boxShadow: '0 0 15px rgba(0, 255, 136, 0.1), 0 4px 20px rgba(0,0,0,0.4), inset 0 1px 0 rgba(255,255,255,0.04)',
                         animation: 'targetPulse 2s ease-in-out infinite',
                         backdropFilter: 'blur(8px)',
                       }}>
                    {/* Top accent line */}
                    <div className="absolute top-0 left-2 right-2 h-px" style={{
                      background: 'linear-gradient(90deg, transparent, rgba(0,255,136,0.5), transparent)',
                    }} />
                    {/* Corner decorations */}
                    <div className="absolute top-0 left-0 w-2 h-2 border-t border-l border-[rgba(0,255,136,0.4)]" />
                    <div className="absolute top-0 right-0 w-2 h-2 border-t border-r border-[rgba(0,255,136,0.4)]" />
                    <div className="absolute bottom-0 left-0 w-2 h-2 border-b border-l border-[rgba(0,255,136,0.4)]" />
                    <div className="absolute bottom-0 right-0 w-2 h-2 border-b border-r border-[rgba(0,255,136,0.4)]" />
                    {target.label || target.word}
                  </div>
                </div>
              )}
            </button>
          );
        })}
      </div>

      {/* ── Score popup ───────────────────────────────── */}
      {showPopup && (
        <div
          className="fixed z-50 score-popup pointer-events-none flex flex-col items-center"
          style={{
            left: showPopup.x,
            top: showPopup.y,
            transform: 'translateX(-50%)',
          }}
        >
          <div className="font-heading font-bold text-3xl sm:text-4xl"
               style={{
                 color: showPopup.correct ? '#00ff88' : '#ff3548',
                 textShadow: showPopup.correct
                   ? '0 0 20px rgba(0,255,136,0.7), 0 0 40px rgba(0,255,136,0.3)'
                   : '0 0 20px rgba(255,53,72,0.7), 0 0 40px rgba(255,53,72,0.3)',
               }}>
            {showPopup.text}
          </div>
          <div className="text-xs font-heading uppercase tracking-widest mt-1"
               style={{ color: showPopup.correct ? 'rgba(0,255,136,0.5)' : 'rgba(255,53,72,0.5)' }}>
            {showPopup.correct ? 'HIT' : 'MISS'}
          </div>
        </div>
      )}

      {/* ── Corner HUD decorations ────────────────────── */}
      {/* Bottom-left - Round progress dots */}
      <div className="absolute bottom-4 left-4 sm:left-6 z-20 flex items-center gap-1.5">
        {Array.from({ length: totalRounds }, (_, i) => (
          <div
            key={i}
            className="w-2 h-2 rounded-full transition-all duration-300"
            style={{
              background: i < round ? '#00ff88' : i === round - 1 ? '#00ff88' : 'rgba(255,255,255,0.1)',
              boxShadow: i < round ? '0 0 6px rgba(0,255,136,0.4)' : undefined,
            }}
          />
        ))}
      </div>

      {/* Bottom-right - Mode & Level indicator */}
      <div className="absolute bottom-4 right-4 sm:right-6 z-20">
        <div className="text-[10px] font-mono uppercase tracking-widest text-[var(--color-text-muted)] opacity-50">
          {quizType.replace(/_/g, ' ')}
        </div>
      </div>
    </div>
  );
}
