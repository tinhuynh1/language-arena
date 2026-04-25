'use client';

import { useState, useCallback, useEffect, useRef, useMemo } from 'react';
import type { Target, QuizType } from '@/hooks/useWebSocket';
import { useLocale } from '@/i18n/LocaleProvider';
import type { TranslationKey } from '@/i18n';

interface GameCanvasProps {
  targets: Target[];
  question: string;
  quizType: QuizType;
  round: number;
  totalRounds: number;
  timeMs: number;
  myCorrect: number;
  opponentCorrect: number;
  opponent: string;
  mode: 'solo' | 'duel' | 'battle';
  lastReactionMs: number;
  lastIsCorrect: boolean | null;
  onHit: (target: Target, reactionMs: number) => void;
  claimedTargets: Record<string, string>;
}

const QUIZ_LABEL_KEYS: Record<QuizType, TranslationKey> = {
  meaning_to_word: 'game.quiz.meaningToWord',
  word_to_meaning: 'game.quiz.wordToMeaning',
  word_to_ipa: 'game.quiz.wordToIpa',
  word_to_pinyin: 'game.quiz.wordToPinyin',
  word_to_tone: 'game.quiz.wordToTone',
  definition_to_word: 'game.quiz.definitionToWord',
};

/* Floating decorative dots */
function FloatingDots() {
  const dots = useMemo(() =>
    Array.from({ length: 20 }, (_, i) => ({
      id: i,
      x: Math.random() * 100,
      y: Math.random() * 100,
      size: Math.random() * 4 + 2,
      duration: Math.random() * 15 + 10,
      delay: Math.random() * -20,
      opacity: Math.random() * 0.15 + 0.03,
    })), []);

  return (
    <div className="absolute inset-0 pointer-events-none overflow-hidden z-0">
      {dots.map(p => (
        <div
          key={p.id}
          className="absolute rounded-full motion-reduce:!animate-none"
          style={{
            left: `${p.x}%`,
            top: `${p.y}%`,
            width: `${p.size}px`,
            height: `${p.size}px`,
            background: p.size > 4 ? 'var(--color-primary)' : 'var(--color-secondary)',
            opacity: p.opacity,
            animation: `float ${p.duration}s ease-in-out infinite`,
            animationDelay: `${p.delay}s`,
          }}
        />
      ))}
    </div>
  );
}

export default function GameCanvas({
  targets,
  question,
  quizType,
  round,
  totalRounds,
  timeMs,
  myCorrect,
  opponentCorrect,
  opponent,
  mode,
  lastReactionMs,
  lastIsCorrect,
  onHit,
  claimedTargets,
}: GameCanvasProps) {
  const { t } = useLocale();
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

    onHit(target, 0);

    setHitTargets(prev => new Set(prev).add(target.id));
    setAnswered(true);

    const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
    setShowPopup({
      x: rect.left + rect.width / 2,
      y: rect.top,
      text: target.correct ? '✓' : '✗',
      correct: target.correct,
    });

    setTimeout(() => setShowPopup(null), 800);
  }, [hitTargets, onHit, isTimeUp]);

  const timePercent = (timeLeft / timeMs) * 100;
  const timeColor = timePercent > 50 ? 'var(--color-secondary)' : timePercent > 25 ? 'var(--color-accent-gold)' : 'var(--color-accent-red)';
  const timeColorRaw = timePercent > 50 ? '#0D9488' : timePercent > 25 ? '#F59E0B' : '#DC2626';
  const timeSeconds = Math.ceil(timeLeft / 1000);
  const isUrgent = timePercent < 25;

  return (
    <div className="relative w-full h-full min-h-[400px] sm:min-h-[500px] cursor-pointer select-none overflow-hidden bg-[var(--color-bg-primary)]">

      {/* ── Background ─────────────────────────────────── */}
      <div className="absolute inset-0" style={{
        background: 'radial-gradient(ellipse at 50% 30%, rgba(79,70,229,0.03) 0%, transparent 50%), radial-gradient(ellipse at 80% 70%, rgba(13,148,136,0.02) 0%, transparent 40%), var(--color-bg-primary)',
      }} />
      <FloatingDots />

      {/* ── HUD Layer ─────────────────────────────────── */}
      <div className="absolute top-0 left-0 right-0 z-30" style={{
        background: 'linear-gradient(180deg, rgba(250,251,254,0.98) 0%, rgba(250,251,254,0.85) 70%, transparent 100%)',
      }}>
        <div className="flex items-center justify-between px-4 sm:px-8 py-3 sm:py-4">

          {/* P1 Score Panel */}
          <div className="flex items-center gap-3 sm:gap-4 px-4 sm:px-5 py-2 sm:py-2.5 rounded-[var(--radius-sm)]" style={{
            background: 'rgba(79, 70, 229, 0.06)',
            border: '1px solid rgba(79, 70, 229, 0.12)',
          }}>
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="var(--color-primary)" strokeWidth="2" strokeLinecap="round"><circle cx="12" cy="12" r="10" /><path d="M9 12l2 2 4-4" /></svg>
            <div>
              <div className="text-[10px] sm:text-xs text-[var(--color-text-muted)] font-heading leading-none">{t('game.hud.you')}</div>
              <div className="text-2xl sm:text-4xl font-heading font-bold leading-none mt-0.5 text-[var(--color-primary)]">{myCorrect}<span className="text-base sm:text-xl text-[var(--color-text-muted)]">/{totalRounds}</span></div>
            </div>
          </div>

          {/* Center - Round + Timer */}
          <div className="flex flex-col items-center gap-1">
            <div className="flex items-center gap-3 px-5 sm:px-6 py-1.5 sm:py-2 rounded-[var(--radius-sm)]" style={{
              background: 'var(--color-bg-card)',
              border: '1px solid var(--color-border-default)',
              boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
            }}>
              <div className="text-xs sm:text-sm text-[var(--color-text-muted)] font-heading">{t('game.hud.round')}</div>
              <div className="text-xl sm:text-3xl font-heading font-bold">
                {round}<span className="text-[var(--color-text-muted)] text-base sm:text-xl">/{totalRounds}</span>
              </div>
            </div>
            {/* Circular mini-timer */}
            <div className="flex items-center gap-2">
              <svg width="20" height="20" viewBox="0 0 24 24" className="-rotate-90">
                <circle cx="12" cy="12" r="10" fill="none" stroke="var(--color-border-default)" strokeWidth="2" />
                <circle
                  cx="12" cy="12" r="10"
                  fill="none"
                  stroke={timeColorRaw}
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeDasharray={2 * Math.PI * 10}
                  strokeDashoffset={2 * Math.PI * 10 * (1 - timePercent / 100)}
                  style={{ transition: 'stroke-dashoffset 0.1s linear' }}
                />
              </svg>
              <span className={`font-mono text-sm font-bold ${isUrgent ? 'animate-pulse' : ''}`} style={{ color: timeColorRaw }}>
                {timeSeconds}s
              </span>
            </div>
          </div>

          {/* P2 / Mode badge */}
          {mode === 'duel' && (
            <div className="flex items-center gap-3 sm:gap-4 px-4 sm:px-5 py-2 sm:py-2.5 rounded-[var(--radius-sm)]" style={{
              background: 'rgba(13, 148, 136, 0.06)',
              border: '1px solid rgba(13, 148, 136, 0.12)',
            }}>
              <div className="text-right">
                <div className="text-[10px] sm:text-xs text-[var(--color-text-muted)] font-heading leading-none">{opponent}</div>
                <div className="text-2xl sm:text-4xl font-heading font-bold leading-none mt-0.5 text-[var(--color-secondary)]">{opponentCorrect}<span className="text-base sm:text-xl text-[var(--color-text-muted)]">/{totalRounds}</span></div>
              </div>
            </div>
          )}
          {mode === 'battle' && (
            <div className="flex items-center gap-2 px-4 py-2 rounded-[var(--radius-sm)]" style={{
              background: 'rgba(234, 88, 12, 0.06)',
              border: '1px solid rgba(234, 88, 12, 0.12)',
            }}>
              <div className="text-sm font-heading font-bold text-[var(--color-accent-orange)]">
                {t('game.hud.battle')}
              </div>
            </div>
          )}
          {mode === 'solo' && (
            <div className="flex items-center gap-2 px-4 py-2 rounded-[var(--radius-sm)] opacity-50" style={{
              border: '1px solid var(--color-border-default)',
            }}>
              <div className="text-sm font-heading text-[var(--color-text-muted)]">{t('game.hud.solo')}</div>
            </div>
          )}
        </div>

        {/* Timer Bar */}
        <div className="h-1 sm:h-1.5 bg-[var(--color-bg-secondary)]" style={{ margin: '0 1rem' }}>
          <div
            className="h-full transition-all duration-100 ease-linear rounded-full"
            style={{
              width: `${timePercent}%`,
              background: timeColorRaw,
              boxShadow: isUrgent ? `0 0 8px ${timeColorRaw}60` : undefined,
            }}
          />
        </div>
      </div>

      {/* ── Question Area ─────────────────────────────── */}
      <div className="absolute top-[6.5rem] sm:top-[7.5rem] left-1/2 -translate-x-1/2 z-30 text-center max-w-[90vw]">
        <div className="text-xs sm:text-sm text-[var(--color-text-muted)] font-heading mb-2">
          {t(QUIZ_LABEL_KEYS[quizType])}
        </div>
        <div className="relative inline-block">
          <div className="text-xl sm:text-3xl font-heading font-bold px-5 sm:px-8 py-2.5 sm:py-3.5 rounded-[var(--radius-md)] text-[var(--color-primary)]"
               style={{
                 background: 'var(--color-bg-card)',
                 border: '2px solid var(--color-primary)',
                 boxShadow: '0 4px 20px rgba(79,70,229,0.12)',
               }}>
            {question}
          </div>
        </div>

        {/* Reaction time badge */}
        {lastReactionMs > 0 && (
          <div className="flex justify-center mt-3">
            <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full" style={{
              background: lastReactionMs < 1000 ? 'rgba(13,148,136,0.08)' : lastReactionMs < 2000 ? 'rgba(245,158,11,0.08)' : 'rgba(220,38,38,0.08)',
              border: `1px solid ${lastReactionMs < 1000 ? 'rgba(13,148,136,0.2)' : lastReactionMs < 2000 ? 'rgba(245,158,11,0.2)' : 'rgba(220,38,38,0.2)'}`,
            }}>
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke={lastReactionMs < 1000 ? '#0D9488' : lastReactionMs < 2000 ? '#F59E0B' : '#DC2626'} strokeWidth="2">
                <circle cx="12" cy="12" r="10" /><polyline points="12 6 12 12 16 14" />
              </svg>
              <span className="font-mono text-sm font-bold" style={{ color: lastReactionMs < 1000 ? '#0D9488' : lastReactionMs < 2000 ? '#F59E0B' : '#DC2626' }}>
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
          const isClaimed = !!claimedTargets[target.id];
          const shouldHide = isHit || isClaimed;

          return (
            <button
              key={target.id}
              onClick={(e) => handleTargetClick(target, e)}
              disabled={answered || isTimeUp || shouldHide}
              className={`absolute transition-all duration-200 group focus-visible:ring-2 focus-visible:ring-[var(--color-primary)] focus-visible:outline-none rounded-[var(--radius-sm)]
                ${shouldHide
                  ? 'target-hit opacity-0 pointer-events-none'
                  : 'target-spawn cursor-pointer'
                }`}
              style={{
                left: `${target.x}%`,
                top: `${target.y}%`,
                transform: 'translate(-50%, -50%)',
              }}
            >
              {!shouldHide && (
                <div className="relative">
                  <div className="relative px-4 sm:px-6 py-3 sm:py-4 rounded-[var(--radius-md)] font-heading font-bold text-base sm:text-lg transition-all duration-200 group-hover:scale-105 group-hover:shadow-lg"
                       style={{
                         background: 'var(--color-bg-card)',
                         border: '1.5px solid var(--color-border-default)',
                         color: 'var(--color-text-primary)',
                         minWidth: '80px',
                         textAlign: 'center',
                         boxShadow: '0 2px 10px rgba(0,0,0,0.08)',
                       }}>
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
               style={{ color: showPopup.correct ? '#0D9488' : '#DC2626' }}>
            {showPopup.text}
          </div>
          <div className="text-xs font-heading mt-1"
               style={{ color: showPopup.correct ? 'rgba(13,148,136,0.6)' : 'rgba(220,38,38,0.6)' }}>
            {showPopup.correct ? t('game.hit') : t('game.miss')}
          </div>
        </div>
      )}

      {/* ── Bottom HUD ────────────────────────────────── */}
      {/* Round progress dots */}
      <div className="absolute bottom-4 left-4 sm:left-6 z-20 flex items-center gap-1.5">
        {Array.from({ length: totalRounds }, (_, i) => (
          <div
            key={i}
            className="w-2 h-2 rounded-full transition-all duration-300"
            style={{
              background: i < round ? 'var(--color-primary)' : 'var(--color-border-default)',
            }}
          />
        ))}
      </div>

      {/* Quiz type indicator */}
      <div className="absolute bottom-4 right-4 sm:right-6 z-20">
        <div className="text-[10px] font-mono text-[var(--color-text-muted)] opacity-50">
          {quizType.replace(/_/g, ' ')}
        </div>
      </div>
    </div>
  );
}
