'use client';

import { useState, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/hooks/useAuth';
import { useGame } from '@/hooks/useGame';
import type { QuizType } from '@/hooks/useWebSocket';
import GameCanvas from '@/components/game/GameCanvas';
import Countdown from '@/components/game/Countdown';
import GameOverScreen from '@/components/game/GameOverScreen';
import LiveLeaderboard from '@/components/game/LiveLeaderboard';

const EN_LEVELS = [
  { value: 'A1', label: 'A1', desc: 'Beginner' },
  { value: 'A2', label: 'A2', desc: 'Elementary' },
  { value: 'B1', label: 'B1', desc: 'Intermediate' },
  { value: 'B2', label: 'B2', desc: 'Upper Intermediate' },
  { value: 'C1', label: 'C1', desc: 'Advanced' },
  { value: 'C2', label: 'C2', desc: 'Mastery' },
];

const ZH_LEVELS = [
  { value: 'HSK1', label: 'HSK1', desc: 'Basic' },
  { value: 'HSK2', label: 'HSK2', desc: 'Elementary' },
  { value: 'HSK3', label: 'HSK3', desc: 'Intermediate' },
  { value: 'HSK4', label: 'HSK4', desc: 'Upper Intermediate' },
  { value: 'HSK5', label: 'HSK5', desc: 'Advanced' },
];

/* SVG icons for mode cards */
function SoloIcon({ color }: { color: string }) {
  return (
    <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="12" r="10" />
      <circle cx="12" cy="12" r="6" />
      <circle cx="12" cy="12" r="2" />
    </svg>
  );
}

function DuelIcon({ color }: { color: string }) {
  return (
    <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z" />
      <line x1="8" y1="13" x2="16" y2="13" />
      <line x1="12" y1="9" x2="12" y2="17" />
    </svg>
  );
}

function BattleIcon({ color }: { color: string }) {
  return (
    <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
    </svg>
  );
}

export default function PlayPage() {
  const { user } = useAuth();
  const game = useGame();
  const router = useRouter();

  const [selectedMode, setSelectedMode] = useState<'solo' | 'duel' | 'battle'>('solo');
  const [selectedLang, setSelectedLang] = useState<string>('en');
  const [selectedLevel, setSelectedLevel] = useState<string>('A1');
  const [selectedQuizType, setSelectedQuizType] = useState<QuizType>('meaning_to_word');
  const [joinCode, setJoinCode] = useState('');
  const [showJoinInput, setShowJoinInput] = useState(false);
  const [readySent, setReadySent] = useState(false);

  if (!user) {
    router.push('/login');
    return null;
  }

  const levels = selectedLang === 'en' ? EN_LEVELS : ZH_LEVELS;

  const EN_QUIZ_TYPES: { value: QuizType; label: string; desc: string }[] = [
    { value: 'meaning_to_word', label: 'Meaning → Word', desc: 'See meaning, shoot word' },
    { value: 'word_to_meaning', label: 'Word → Meaning', desc: 'See word, shoot meaning' },
    { value: 'word_to_ipa', label: 'Word → IPA', desc: 'See word, shoot IPA' },
  ];

  const ZH_QUIZ_TYPES: { value: QuizType; label: string; desc: string }[] = [
    { value: 'meaning_to_word', label: 'Meaning → Word', desc: 'See meaning, shoot word' },
    { value: 'word_to_meaning', label: 'Word → Meaning', desc: 'See word, shoot meaning' },
    { value: 'word_to_pinyin', label: 'Word → Pinyin', desc: 'See word, shoot pinyin' },
  ];

  const quizTypes = selectedLang === 'en' ? EN_QUIZ_TYPES : ZH_QUIZ_TYPES;

  const handleLangChange = (lang: string) => {
    setSelectedLang(lang);
    setSelectedLevel(lang === 'en' ? 'A1' : 'HSK1');
    setSelectedQuizType('meaning_to_word');
  };

  const handleStart = () => {
    if (selectedMode === 'battle') {
      game.createRoom(selectedLang, selectedLevel, selectedQuizType);
    } else {
      game.joinGame(selectedMode, selectedLang, selectedLevel, selectedQuizType);
    }
  };

  const handleJoinByCode = () => {
    if (joinCode.trim().length >= 4) {
      game.joinRoom(joinCode.trim().toUpperCase());
    }
  };

  const handleReady = useCallback(() => {
    game.ready();
    setReadySent(true);
  }, [game]);

  const handlePlayAgain = () => {
    game.leaveGame();
    setReadySent(false);
    setTimeout(() => {
      if (selectedMode === 'battle') {
        game.createRoom(selectedLang, selectedLevel, selectedQuizType);
      } else {
        game.joinGame(selectedMode, selectedLang, selectedLevel, selectedQuizType);
      }
    }, 300);
  };

  const handleLeave = () => {
    game.leaveGame();
  };

  const MODE_CONFIG = {
    solo: { color: '#00ff88', glowColor: 'rgba(0,255,136,0.15)', Icon: SoloIcon },
    duel: { color: '#ff6b35', glowColor: 'rgba(255,107,53,0.15)', Icon: DuelIcon },
    battle: { color: '#00d4ff', glowColor: 'rgba(0,212,255,0.15)', Icon: BattleIcon },
  };

  // Idle - Mode/Lang/Level Selection
  if (game.state === 'idle') {
    return (
      <div className="min-h-screen flex items-center justify-center px-4 sm:px-6 relative overflow-hidden">
        {/* Background grid */}
        <div className="absolute inset-0 opacity-[0.02]" aria-hidden="true" style={{
          backgroundImage: `linear-gradient(rgba(0,255,136,0.4) 1px, transparent 1px),
                            linear-gradient(90deg, rgba(0,255,136,0.4) 1px, transparent 1px)`,
          backgroundSize: '60px 60px',
        }} />

        {/* Ambient orbs */}
        <div className="orb w-[400px] h-[400px] opacity-[0.06]"
             style={{ background: '#00ff88', top: '-10%', left: '-10%' }} />
        <div className="orb w-[300px] h-[300px] opacity-[0.04]"
             style={{ background: '#00d4ff', bottom: '-5%', right: '-10%', animationDelay: '-6s' }} />

        <div className="w-full max-w-2xl text-center relative z-10">
          {/* Title */}
          <div className="animate-fade-in-up">
            <h1 className="font-heading font-bold text-3xl sm:text-5xl mb-3 uppercase tracking-wider">
              Select <span className="text-gradient-neon text-glow">Mission</span>
            </h1>
            <p className="text-[var(--color-text-secondary)] mb-8 sm:mb-10 text-sm sm:text-base">
              Choose your mode, language, and difficulty level
            </p>
            {/* Error toast */}
            {game.errorMessage && (
              <div className="mb-6 px-4 py-3 rounded-sm border border-[var(--color-accent-red)] bg-[rgba(255,53,72,0.08)] text-[var(--color-accent-red)] font-heading text-sm uppercase tracking-wider animate-fade-in-up flex items-center justify-center gap-2">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                  <circle cx="12" cy="12" r="10" />
                  <line x1="12" y1="8" x2="12" y2="12" />
                  <line x1="12" y1="16" x2="12.01" y2="16" />
                </svg>
                {game.errorMessage}
              </div>
            )}
          </div>

          {/* Mode Selection */}
          <div className="grid grid-cols-3 gap-3 sm:gap-4 mb-6 sm:mb-8 animate-fade-in-up delay-100">
            {(['solo', 'duel', 'battle'] as const).map(mode => {
              const cfg = MODE_CONFIG[mode];
              const isSelected = selectedMode === mode;
              return (
                <button
                  key={mode}
                  onClick={() => setSelectedMode(mode)}
                  className={`card text-left transition-all duration-200 cursor-pointer relative overflow-hidden group ${
                    isSelected ? 'border-2' : 'border hover:border-[var(--color-text-muted)]'
                  }`}
                  style={{
                    borderColor: isSelected ? cfg.color : undefined,
                    boxShadow: isSelected ? `0 0 24px ${cfg.glowColor}, inset 0 0 20px ${cfg.glowColor}` : undefined,
                    padding: '1.25rem',
                  }}
                >
                  {/* Accent bar top */}
                  {isSelected && (
                    <div className="absolute top-0 left-0 right-0 h-[2px]" style={{ background: cfg.color, boxShadow: `0 0 10px ${cfg.color}` }} />
                  )}
                  <div className="mb-3 transition-transform duration-200 group-hover:scale-110" style={{ color: isSelected ? cfg.color : 'var(--color-text-muted)' }}>
                    <cfg.Icon color={isSelected ? cfg.color : 'currentColor'} />
                  </div>
                  <div className="font-heading font-bold text-base sm:text-lg mb-1 uppercase tracking-wide" style={{ color: isSelected ? cfg.color : 'inherit' }}>
                    {mode === 'solo' ? 'Solo' : mode === 'duel' ? '1v1 Duel' : 'Battle'}
                  </div>
                  <p className="text-xs sm:text-sm text-[var(--color-text-muted)]">
                    {mode === 'solo' ? 'Practice alone' : mode === 'duel' ? 'Real-time match' : 'Up to 100 players'}
                  </p>
                </button>
              );
            })}
          </div>

          {/* Language Selection */}
          <div className="grid grid-cols-2 gap-3 sm:gap-4 mb-6 sm:mb-8 animate-fade-in-up delay-200">
            {[
              { lang: 'en', flag: '🇬🇧', label: 'English (CEFR)' },
              { lang: 'zh', flag: '🇨🇳', label: 'Chinese (HSK)' },
            ].map(({ lang, flag, label }) => {
              const isSelected = selectedLang === lang;
              return (
                <button
                  key={lang}
                  onClick={() => handleLangChange(lang)}
                  className={`card text-center transition-all duration-200 cursor-pointer py-5 ${
                    isSelected ? 'border-2 border-[var(--color-accent-cyan)]' : 'hover:border-[var(--color-text-muted)]'
                  }`}
                  style={{
                    boxShadow: isSelected ? '0 0 24px rgba(0,212,255,0.15), inset 0 0 20px rgba(0,212,255,0.08)' : undefined,
                  }}
                >
                  <div className="text-3xl sm:text-4xl mb-2">{flag}</div>
                  <div className="font-heading font-bold uppercase text-sm sm:text-base tracking-wide" style={{ color: isSelected ? '#00d4ff' : 'inherit' }}>
                    {label}
                  </div>
                </button>
              );
            })}
          </div>

          {/* Level Selection */}
          <div className="mb-6 sm:mb-8 animate-fade-in-up delay-300">
            <div className="text-xs sm:text-sm font-heading uppercase tracking-widest text-[var(--color-text-muted)] mb-3">
              Vocabulary Level
            </div>
            <div className="flex gap-2 sm:gap-3 justify-center flex-wrap">
              {levels.map(l => (
                <button
                  key={l.value}
                  onClick={() => setSelectedLevel(l.value)}
                  className={`px-4 sm:px-5 py-2.5 sm:py-3 font-heading font-bold text-sm sm:text-base uppercase transition-all duration-200 border cursor-pointer ${selectedLevel === l.value
                      ? 'border-[var(--color-accent-neon)] text-[var(--color-accent-neon)] bg-[rgba(0,255,136,0.1)]'
                      : 'border-[var(--color-border-default)] text-[var(--color-text-secondary)] hover:border-[var(--color-text-muted)]'
                    }`}
                  style={{
                    borderRadius: '3px',
                    boxShadow: selectedLevel === l.value ? '0 0 12px rgba(0,255,136,0.15)' : undefined,
                  }}
                >
                  <div>{l.label}</div>
                  <div className="text-[10px] sm:text-xs font-normal lowercase opacity-70 mt-0.5">{l.desc}</div>
                </button>
              ))}
            </div>
          </div>

          {/* Quiz Type Selection */}
          <div className="mb-8 sm:mb-10 animate-fade-in-up delay-400">
            <div className="text-xs sm:text-sm font-heading uppercase tracking-widest text-[var(--color-text-muted)] mb-3">
              Quiz Type
            </div>
            <div className="grid grid-cols-3 gap-2 sm:gap-3">
              {quizTypes.map(q => {
                const isSelected = selectedQuizType === q.value;
                return (
                  <button
                    key={q.value}
                    onClick={() => setSelectedQuizType(q.value)}
                    className={`px-3 sm:px-4 py-3 font-heading font-bold text-sm uppercase transition-all duration-200 border text-left cursor-pointer ${isSelected
                        ? 'border-[var(--color-accent-orange)] text-[var(--color-accent-orange)] bg-[rgba(255,107,53,0.12)]'
                        : 'border-[var(--color-border-default)] text-[var(--color-text-secondary)] hover:border-[var(--color-text-muted)]'
                      }`}
                    style={{
                      borderRadius: '3px',
                      boxShadow: isSelected ? '0 0 12px rgba(255,107,53,0.15)' : undefined,
                    }}
                  >
                    <div>{q.label}</div>
                    <div className="text-[10px] sm:text-xs font-normal normal-case opacity-70 mt-0.5">{q.desc}</div>
                  </button>
                );
              })}
            </div>
          </div>

          {/* Action Buttons */}
          <div className="space-y-4 animate-fade-in-up delay-400">
            <button onClick={handleStart} className="btn-primary text-base sm:text-lg px-10 sm:px-14 py-4 sm:py-5 w-full max-w-sm mx-auto block">
              {selectedMode === 'battle' ? 'CREATE ROOM' : selectedMode === 'duel' ? 'FIND OPPONENT' : 'START TRAINING'}
            </button>

            {selectedMode === 'battle' && (
              <div>
                {!showJoinInput ? (
                  <button
                    onClick={() => setShowJoinInput(true)}
                    className="btn-secondary text-sm px-8 py-3"
                  >
                    JOIN WITH CODE
                  </button>
                ) : (
                  <div className="flex gap-2 justify-center items-center max-w-xs mx-auto">
                    <input
                      type="text"
                      className="input-field text-center font-mono font-bold text-lg uppercase tracking-[0.3em]"
                      placeholder="ROOM CODE"
                      maxLength={6}
                      value={joinCode}
                      onChange={e => setJoinCode(e.target.value.toUpperCase())}
                      onKeyDown={e => e.key === 'Enter' && handleJoinByCode()}
                      autoFocus
                    />
                    <button onClick={handleJoinByCode} className="btn-primary py-3 px-4">
                      JOIN
                    </button>
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </div>
    );
  }

  // Queuing / Creating Room
  if (game.state === 'queuing' || game.state === 'creating_room') {
    return (
      <div className="min-h-screen flex items-center justify-center px-4">
        <div className="text-center">
          {!game.connected ? (
            <>
              <div className="w-20 h-20 border-2 border-[var(--color-accent-orange)] border-t-transparent rounded-full animate-spin mx-auto mb-6" />
              <div className="font-heading font-bold text-2xl sm:text-3xl mb-2 uppercase" style={{ color: '#ff6b35' }}>
                Connecting...
              </div>
              <p className="text-sm text-[var(--color-text-muted)]">
                Establishing secure connection
              </p>
            </>
          ) : (
            <>
              <div className="w-20 h-20 border-2 border-[var(--color-accent-neon)] border-t-transparent rounded-full animate-spin mx-auto mb-6" />
              <div className="font-heading font-bold text-2xl sm:text-3xl mb-2 uppercase text-glow">
                {selectedMode === 'duel' ? 'Finding Opponent...' : selectedMode === 'battle' ? 'Creating Room...' : 'Preparing Arena...'}
              </div>
              <p className="text-sm text-[var(--color-text-muted)]">
                {selectedLevel} • {selectedLang === 'en' ? 'English' : 'Chinese'}
              </p>
            </>
          )}
          <button onClick={handleLeave} className="btn-secondary mt-8 text-sm">CANCEL</button>
        </div>
      </div>
    );
  }

  // Battle Lobby - Waiting for players
  if (game.state === 'in_lobby') {
    const amIHost = game.isHost || game.hostUsername === user.username;
    return (
      <div className="min-h-screen flex items-center justify-center px-4 sm:px-6 relative overflow-hidden">
        {/* Background grid */}
        <div className="absolute inset-0 opacity-[0.02]" aria-hidden="true" style={{
          backgroundImage: `linear-gradient(rgba(0,212,255,0.4) 1px, transparent 1px),
                            linear-gradient(90deg, rgba(0,212,255,0.4) 1px, transparent 1px)`,
          backgroundSize: '60px 60px',
        }} />

        {/* Ambient orbs */}
        <div className="orb w-[400px] h-[400px] opacity-[0.06]"
             style={{ background: '#00d4ff', top: '-10%', right: '-10%' }} />
        <div className="orb w-[300px] h-[300px] opacity-[0.04]"
             style={{ background: '#00ff88', bottom: '-5%', left: '-10%', animationDelay: '-6s' }} />

        <div className="w-full max-w-lg text-center relative z-10">
          {/* Title */}
          <div className="animate-fade-in-up mb-6">
            <h1 className="font-heading font-bold text-3xl sm:text-4xl mb-1 uppercase tracking-wider">
              Battle <span className="text-glow-cyan" style={{ color: '#00d4ff' }}>Room</span>
            </h1>
            <p className="text-sm text-[var(--color-text-muted)]">Waiting for players to join</p>
          </div>

          {/* Room Code Card */}
          <div className="mb-6 px-6 py-5 rounded-sm animate-fade-in-up delay-100" style={{
            background: 'rgba(0, 212, 255, 0.04)',
            border: '1px solid rgba(0, 212, 255, 0.15)',
            backdropFilter: 'blur(8px)',
          }}>
            <div className="text-xs font-heading uppercase tracking-[0.25em] text-[var(--color-text-muted)] mb-2">
              Share this code
            </div>
            <button
              className="w-full font-mono font-bold text-4xl sm:text-5xl tracking-[0.3em] text-glow cursor-pointer transition-all hover:scale-[1.02] focus-visible:ring-2 focus-visible:ring-[#00ff88] focus-visible:outline-none rounded-sm bg-transparent border-none p-0"
              style={{ color: '#00ff88' }}
              onClick={() => navigator.clipboard?.writeText(game.roomCode)}
              title="Click to copy room code"
              aria-label={`Room code: ${game.roomCode}. Click to copy.`}
            >
              {game.roomCode}
            </button>
            <div className="text-xs text-[var(--color-text-muted)] mt-2 opacity-60">Click to copy</div>
          </div>

          {/* Player List */}
          <div className="mb-6 px-5 py-4 rounded-sm animate-fade-in-up delay-200" style={{
            background: 'rgba(255,255,255,0.02)',
            border: '1px solid rgba(255,255,255,0.06)',
          }}>
            <div className="flex items-center justify-between mb-3">
              <span className="text-xs font-heading uppercase tracking-wider text-[var(--color-text-muted)]">
                Players
              </span>
              <span className="font-mono font-bold text-sm" style={{ color: '#00ff88' }}>
                {game.playerCount}
              </span>
            </div>
            <div className="space-y-2">
              {game.players.map(name => {
                const isPlayerHost = name === game.hostUsername;
                return (
                  <div key={name} className="flex items-center gap-3 px-3 py-2.5 rounded-sm transition-all" style={{
                    background: isPlayerHost ? 'rgba(0, 212, 255, 0.06)' : 'rgba(255,255,255,0.02)',
                    border: `1px solid ${isPlayerHost ? 'rgba(0, 212, 255, 0.2)' : 'rgba(255,255,255,0.05)'}`,
                  }}>
                    {/* Avatar circle */}
                    <div className="w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold font-heading shrink-0" style={{
                      background: isPlayerHost ? 'rgba(0, 212, 255, 0.15)' : 'rgba(0, 255, 136, 0.1)',
                      color: isPlayerHost ? '#00d4ff' : '#00ff88',
                      border: `1px solid ${isPlayerHost ? 'rgba(0,212,255,0.3)' : 'rgba(0,255,136,0.2)'}`,
                    }}>
                      {name[0]?.toUpperCase()}
                    </div>
                    <span className="font-heading text-sm font-bold flex-1 text-left" style={{
                      color: isPlayerHost ? '#00d4ff' : 'var(--color-text-primary)',
                    }}>
                      {name}
                    </span>
                    {/* Host badge */}
                    {isPlayerHost && (
                      <span className="flex items-center gap-1 px-2 py-0.5 text-[10px] font-heading uppercase tracking-wider rounded-sm" style={{
                        background: 'rgba(0, 212, 255, 0.1)',
                        color: '#00d4ff',
                        border: '1px solid rgba(0, 212, 255, 0.2)',
                      }}>
                        <svg width="10" height="10" viewBox="0 0 24 24" fill="#00d4ff" stroke="none">
                          <path d="M12 2L15.09 8.26L22 9.27L17 14.14L18.18 21.02L12 17.77L5.82 21.02L7 14.14L2 9.27L8.91 8.26L12 2Z" />
                        </svg>
                        HOST
                      </span>
                    )}
                    {name === user.username && !isPlayerHost && (
                      <span className="text-[10px] font-heading uppercase tracking-wider text-[var(--color-text-muted)] opacity-50">
                        YOU
                      </span>
                    )}
                  </div>
                );
              })}
            </div>
          </div>

          {/* Actions */}
          <div className="flex gap-3 justify-center animate-fade-in-up delay-300">
            {amIHost ? (
              <button
                onClick={() => game.startGame()}
                className="btn-primary text-base px-10 py-3.5 cursor-pointer"
                disabled={game.playerCount < 2}
              >
                {game.playerCount < 2 ? 'WAITING FOR PLAYERS...' : `START GAME (${game.playerCount} players)`}
              </button>
            ) : (
              <button
                disabled
                className="text-base px-10 py-3.5 font-heading font-bold uppercase tracking-wider border-2 cursor-not-allowed opacity-60"
                style={{
                  borderColor: 'rgba(0,212,255,0.2)',
                  color: '#00d4ff',
                  background: 'rgba(0,212,255,0.05)',
                  borderRadius: '3px',
                }}
              >
                WAITING FOR HOST TO START...
              </button>
            )}
            <button onClick={handleLeave} className="btn-secondary py-3.5 px-6 cursor-pointer">
              LEAVE
            </button>
          </div>
        </div>
      </div>
    );
  }

  // Matched - Ready check (solo/duel)
  if (game.state === 'matched') {
    return (
      <div className="min-h-screen flex items-center justify-center px-6">
        <div className="text-center">
          <div className="font-heading font-bold text-3xl sm:text-4xl mb-2 uppercase text-glow" style={{ color: '#00ff88' }}>
            {game.mode === 'duel' ? 'MATCH FOUND' : game.mode === 'battle' ? 'GAME STARTING' : 'ARENA READY'}
          </div>
          {game.mode === 'duel' && game.opponent && (
            <div className="text-lg text-[var(--color-text-secondary)] mb-6">
              vs <span className="font-bold" style={{ color: '#ff6b35' }}>{game.opponent}</span>
            </div>
          )}
          {game.mode === 'battle' && (
            <div className="text-lg text-[var(--color-text-secondary)] mb-6">
              {game.playerCount} players
            </div>
          )}
          <button
            onClick={handleReady}
            disabled={readySent}
            className={`text-lg px-12 py-4 font-heading font-bold uppercase tracking-wider border-2 transition-all cursor-pointer ${readySent
                ? 'border-[var(--color-accent-neon)] text-[var(--color-accent-neon)] bg-[rgba(0,255,136,0.1)] opacity-80 cursor-not-allowed'
                : 'btn-primary'
              }`}
            style={{ borderRadius: '3px' }}
          >
            {readySent ? '✓ WAITING...' : 'READY'}
          </button>
          {readySent && (
            <p className="text-sm text-[var(--color-text-muted)] mt-3 animate-pulse">
              Waiting for opponent to be ready...
            </p>
          )}
        </div>
      </div>
    );
  }

  // Countdown
  if (game.state === 'countdown') {
    return <Countdown ms={game.countdownMs} onComplete={() => { }} />;
  }

  // Playing
  if (game.state === 'playing' || game.state === 'round_end') {
    return (
      <div className="h-[calc(100vh-4.5rem)] relative" style={{ background: 'var(--color-bg-primary)' }}>
        <GameCanvas
          targets={game.targets}
          question={game.question}
          quizType={game.quizType}
          round={game.round}
          totalRounds={game.totalRounds}
          timeMs={game.timeMs}
          myScore={game.myScore}
          opponentScore={game.opponentScore}
          opponent={game.opponent}
          mode={game.mode}
          lastReactionMs={game.lastReactionMs}
          onHit={game.hitTarget}
        />
        {game.mode === 'battle' && (
          <LiveLeaderboard players={game.liveLeaderboard} round={game.round} />
        )}
      </div>
    );
  }

  // Game Over
  if (game.state === 'game_over') {
    return (
      <div className="min-h-screen" style={{ background: 'var(--color-bg-primary)' }}>
        <GameOverScreen
          data={game.gameOverData}
          mode={game.mode}
          username={user.username}
          onPlayAgain={handlePlayAgain}
          onLeave={handleLeave}
        />
      </div>
    );
  }

  return null;
}
