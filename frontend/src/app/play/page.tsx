'use client';

import { useState, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/hooks/useAuth';
import { useGame } from '@/hooks/useGame';
import { useLocale } from '@/i18n/LocaleProvider';
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
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20" /><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z" />
    </svg>
  );
}

function DuelIcon({ color }: { color: string }) {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" /><circle cx="9" cy="7" r="4" /><path d="M23 21v-2a4 4 0 0 0-3-3.87" /><path d="M16 3.13a4 4 0 0 1 0 7.75" />
    </svg>
  );
}

function BattleIcon({ color }: { color: string }) {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M6 9H4.5a2.5 2.5 0 0 1 0-5H6" /><path d="M18 9h1.5a2.5 2.5 0 0 0 0-5H18" /><path d="M4 22h16" /><path d="M10 14.66V17c0 .55-.47.98-.97 1.21C7.85 18.75 7 20 7 22" /><path d="M14 14.66V17c0 .55.47.98.97 1.21C16.15 18.75 17 20 17 22" /><path d="M18 2H6v7a6 6 0 0 0 12 0V2Z" />
    </svg>
  );
}

export default function PlayPage() {
  const { user } = useAuth();
  const game = useGame();
  const router = useRouter();
  const { t } = useLocale();

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
    { value: 'meaning_to_word', label: 'Meaning → Word', desc: 'See meaning, find word' },
    { value: 'word_to_meaning', label: 'Word → Meaning', desc: 'See word, find meaning' },
    { value: 'word_to_ipa', label: 'Word → IPA', desc: 'See word, find IPA' },
    { value: 'definition_to_word', label: 'Definition → Word', desc: 'Read definition, find word' },
  ];

  const ZH_QUIZ_TYPES: { value: QuizType; label: string; desc: string }[] = [
    { value: 'meaning_to_word', label: 'Meaning → Word', desc: 'See meaning, find word' },
    { value: 'word_to_meaning', label: 'Word → Meaning', desc: 'See word, find meaning' },
    { value: 'word_to_pinyin', label: 'Word → Pinyin', desc: 'See word, find pinyin' },
    { value: 'word_to_tone', label: 'Word → Tone', desc: 'See word, pick correct tone' },
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
    solo: { color: '#4F46E5', Icon: SoloIcon },
    duel: { color: '#0D9488', Icon: DuelIcon },
    battle: { color: '#EA580C', Icon: BattleIcon },
  };

  // Idle - Mode/Lang/Level Selection
  if (game.state === 'idle') {
    return (
      <div className="min-h-screen flex items-center justify-center px-4 sm:px-6 relative overflow-hidden">
        {/* Soft blobs */}
        <div className="bg-blob w-[350px] h-[350px] opacity-[0.08]"
             style={{ background: '#4F46E5', top: '-10%', left: '-10%' }} />
        <div className="bg-blob w-[300px] h-[300px] opacity-[0.06]"
             style={{ background: '#0D9488', bottom: '-5%', right: '-10%', animationDelay: '-6s' }} />

        <div className="w-full max-w-2xl text-center relative z-10">
          {/* Title */}
          <div className="animate-fade-in-up">
            <h1 className="font-heading font-bold text-3xl sm:text-4xl mb-3 tracking-tight">
              {t('play.title').replace('{accent}', '')} <span className="text-[var(--color-primary)]">{t('play.titleAccent')}</span>
            </h1>
            <p className="text-[var(--color-text-secondary)] mb-8 sm:mb-10 text-sm sm:text-base">
              {t('play.subtitle')}
            </p>
            {/* Error toast */}
            {game.errorMessage && (
              <div className="mb-6 px-4 py-3 rounded-[var(--radius-sm)] border border-red-200 bg-red-50 text-[var(--color-accent-red)] font-heading text-sm animate-fade-in-up flex items-center justify-center gap-2">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                  <circle cx="12" cy="12" r="10" /><line x1="12" y1="8" x2="12" y2="12" /><line x1="12" y1="16" x2="12.01" y2="16" />
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
                    boxShadow: isSelected ? `0 4px 20px ${cfg.color}20` : undefined,
                    padding: '1.25rem',
                  }}
                >
                  <div className="mb-3 transition-transform duration-200 group-hover:scale-110" style={{ color: isSelected ? cfg.color : 'var(--color-text-muted)' }}>
                    <cfg.Icon color={isSelected ? cfg.color : 'currentColor'} />
                  </div>
                  <div className="font-heading font-bold text-base sm:text-lg mb-1" style={{ color: isSelected ? cfg.color : 'inherit' }}>
                    {mode === 'solo' ? t('play.mode.solo') : mode === 'duel' ? t('play.mode.duel') : t('play.mode.battle')}
                  </div>
                  <p className="text-xs sm:text-sm text-[var(--color-text-muted)]">
                    {mode === 'solo' ? t('play.mode.solo.desc') : mode === 'duel' ? t('play.mode.duel.desc') : t('play.mode.battle.desc')}
                  </p>
                </button>
              );
            })}
          </div>

          {/* Language Selection */}
          <div className="grid grid-cols-2 gap-3 sm:gap-4 mb-6 sm:mb-8 animate-fade-in-up delay-200">
            {[
              { lang: 'en', flagCode: 'gb', label: t('play.lang.en') },
              { lang: 'zh', flagCode: 'cn', label: t('play.lang.zh') },
            ].map(({ lang, flagCode, label }) => {
              const isSelected = selectedLang === lang;
              return (
                <button
                  key={lang}
                  onClick={() => handleLangChange(lang)}
                  className={`card text-center transition-all duration-200 cursor-pointer py-5 ${
                    isSelected ? 'border-2 border-[var(--color-secondary)]' : 'hover:border-[var(--color-text-muted)]'
                  }`}
                  style={{
                    boxShadow: isSelected ? '0 4px 20px rgba(13,148,136,0.12)' : undefined,
                  }}
                >
                  <div className="mb-2 flex justify-center">
                    <img
                      src={`https://flagcdn.com/w80/${flagCode}.png`}
                      alt={label}
                      width={48}
                      height={32}
                      style={{ borderRadius: 6, objectFit: 'cover' }}
                    />
                  </div>
                  <div className="font-heading font-bold text-sm sm:text-base" style={{ color: isSelected ? 'var(--color-secondary)' : 'inherit' }}>
                    {label}
                  </div>
                </button>
              );
            })}
          </div>

          {/* Level Selection */}
          <div className="mb-6 sm:mb-8 animate-fade-in-up delay-300">
            <div className="text-sm font-heading font-medium text-[var(--color-text-muted)] mb-3">
              {t('play.level.label')}
            </div>
            <div className="flex gap-2 sm:gap-3 justify-center flex-wrap">
              {levels.map(l => (
                <button
                  key={l.value}
                  onClick={() => setSelectedLevel(l.value)}
                  className={`px-4 sm:px-5 py-2.5 sm:py-3 font-heading font-bold text-sm sm:text-base transition-all duration-200 border cursor-pointer rounded-[var(--radius-sm)] ${selectedLevel === l.value
                      ? 'border-[var(--color-primary)] text-[var(--color-primary)] bg-[rgba(79,70,229,0.06)]'
                      : 'border-[var(--color-border-default)] text-[var(--color-text-secondary)] hover:border-[var(--color-text-muted)]'
                    }`}
                  style={{
                    boxShadow: selectedLevel === l.value ? '0 2px 8px rgba(79,70,229,0.12)' : undefined,
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
            <div className="text-sm font-heading font-medium text-[var(--color-text-muted)] mb-3">
              {t('play.quiz.label')}
            </div>
            <div className="grid grid-cols-3 gap-2 sm:gap-3">
              {quizTypes.map(q => {
                const isSelected = selectedQuizType === q.value;
                return (
                  <button
                    key={q.value}
                    onClick={() => setSelectedQuizType(q.value)}
                    className={`px-3 sm:px-4 py-3 font-heading font-bold text-sm transition-all duration-200 border text-left cursor-pointer rounded-[var(--radius-sm)] ${isSelected
                        ? 'border-[var(--color-secondary)] text-[var(--color-secondary)] bg-[rgba(13,148,136,0.06)]'
                        : 'border-[var(--color-border-default)] text-[var(--color-text-secondary)] hover:border-[var(--color-text-muted)]'
                      }`}
                    style={{
                      boxShadow: isSelected ? '0 2px 8px rgba(13,148,136,0.12)' : undefined,
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
            <button onClick={handleStart} className="btn-primary text-base sm:text-lg px-10 sm:px-14 py-4 sm:py-5 w-full max-w-sm mx-auto block cursor-pointer">
              {selectedMode === 'battle' ? t('play.btn.createRoom') : selectedMode === 'duel' ? t('play.btn.findOpponent') : t('play.btn.startTraining')}
            </button>

            {selectedMode === 'battle' && (
              <div>
                {!showJoinInput ? (
                  <button
                    onClick={() => setShowJoinInput(true)}
                    className="btn-secondary text-sm px-8 py-3 cursor-pointer"
                  >
                    {t('play.btn.joinWithCode')}
                  </button>
                ) : (
                  <div className="flex gap-2 justify-center items-center max-w-xs mx-auto">
                    <input
                      type="text"
                      className="input-field text-center font-mono font-bold text-lg uppercase tracking-[0.3em]"
                      placeholder={t('play.placeholder.roomCode')}
                      maxLength={6}
                      value={joinCode}
                      onChange={e => setJoinCode(e.target.value.toUpperCase())}
                      onKeyDown={e => e.key === 'Enter' && handleJoinByCode()}
                      autoFocus
                    />
                    <button onClick={handleJoinByCode} className="btn-primary py-3 px-4 cursor-pointer">
                      {t('play.btn.join')}
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
              <div className="w-16 h-16 border-2 border-[var(--color-accent-orange)] border-t-transparent rounded-full animate-spin mx-auto mb-6" />
              <div className="font-heading font-bold text-2xl sm:text-3xl mb-2" style={{ color: '#EA580C' }}>
                {t('play.queuing.connecting')}
              </div>
              <p className="text-sm text-[var(--color-text-muted)]">
                {t('play.queuing.connectingSub')}
              </p>
            </>
          ) : (
            <>
              <div className="w-16 h-16 border-2 border-[var(--color-primary)] border-t-transparent rounded-full animate-spin mx-auto mb-6" />
              <div className="font-heading font-bold text-2xl sm:text-3xl mb-2 text-[var(--color-primary)]">
                {selectedMode === 'duel' ? t('play.queuing.findingOpponent') : selectedMode === 'battle' ? t('play.queuing.creatingRoom') : t('play.queuing.preparingArena')}
              </div>
              <p className="text-sm text-[var(--color-text-muted)]">
                {selectedLevel} • {selectedLang === 'en' ? 'English' : 'Chinese'}
              </p>
            </>
          )}
          <button onClick={handleLeave} className="btn-secondary mt-8 text-sm cursor-pointer">{t('play.queuing.cancel')}</button>
        </div>
      </div>
    );
  }

  // Battle Lobby - Waiting for players
  if (game.state === 'in_lobby') {
    const amIHost = game.isHost || game.hostUsername === user.username;
    return (
      <div className="min-h-screen flex items-center justify-center px-4 sm:px-6 relative overflow-hidden">
        <div className="bg-blob w-[350px] h-[350px] opacity-[0.08]"
             style={{ background: '#0D9488', top: '-10%', right: '-10%' }} />

        <div className="w-full max-w-lg text-center relative z-10">
          {/* Title */}
          <div className="motion-safe:animate-fade-in-up mb-6">
            <h1 className="font-heading font-bold text-3xl sm:text-4xl mb-1 tracking-tight">
              {t('play.lobby.title').replace('{accent}', '')} <span className="text-[var(--color-secondary)]">{t('play.lobby.titleAccent')}</span>
            </h1>
            <p className="text-sm text-[var(--color-text-muted)]">{t('play.lobby.subtitle')}</p>
          </div>

          {/* Room Code Card */}
          <div className="card mb-6 motion-safe:animate-fade-in-up delay-100">
            <div className="text-xs font-heading font-medium text-[var(--color-text-muted)] mb-2">
              {t('play.lobby.shareCode')}
            </div>
            <button
              className="w-full font-mono font-bold text-4xl sm:text-5xl tracking-[0.3em] text-[var(--color-primary)] cursor-pointer transition-all hover:scale-[1.02] focus-visible:ring-2 focus-visible:ring-[var(--color-primary)] focus-visible:outline-none bg-transparent border-none p-0"
              onClick={() => navigator.clipboard?.writeText(game.roomCode)}
              title="Click to copy room code"
              aria-label={`Room code: ${game.roomCode}. Click to copy.`}
            >
              {game.roomCode}
            </button>
            <div className="text-xs text-[var(--color-text-muted)] mt-2 opacity-60">{t('play.lobby.clickToCopy')}</div>
          </div>

          {/* Player List */}
          <div className="card mb-6 motion-safe:animate-fade-in-up delay-200">
            <div className="flex items-center justify-between mb-3">
              <span className="text-xs font-heading font-medium text-[var(--color-text-muted)]">
                {t('play.lobby.players')}
              </span>
              <span className="font-mono font-bold text-sm text-[var(--color-primary)]">
                {game.playerCount}
              </span>
            </div>
            <div className="space-y-2">
              {game.players.map(name => {
                const isPlayerHost = name === game.hostUsername;
                return (
                  <div key={name} className="flex items-center gap-3 px-3 py-2.5 rounded-[var(--radius-sm)] transition-all" style={{
                    background: isPlayerHost ? 'rgba(13,148,136,0.06)' : 'var(--color-bg-secondary)',
                    border: `1px solid ${isPlayerHost ? 'rgba(13,148,136,0.2)' : 'var(--color-border-default)'}`,
                  }}>
                    <div className="w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold font-heading shrink-0" style={{
                      background: isPlayerHost ? 'rgba(13,148,136,0.1)' : 'rgba(79,70,229,0.08)',
                      color: isPlayerHost ? 'var(--color-secondary)' : 'var(--color-primary)',
                    }}>
                      {name[0]?.toUpperCase()}
                    </div>
                    <span className="font-heading text-sm font-bold flex-1 text-left" style={{
                      color: isPlayerHost ? 'var(--color-secondary)' : 'var(--color-text-primary)',
                    }}>
                      {name}
                    </span>
                    {isPlayerHost && (
                      <span className="flex items-center gap-1 px-2 py-0.5 text-[10px] font-heading font-medium rounded-full" style={{
                        background: 'rgba(13,148,136,0.1)',
                        color: 'var(--color-secondary)',
                        border: '1px solid rgba(13,148,136,0.2)',
                      }}>
                        <svg width="10" height="10" viewBox="0 0 24 24" fill="currentColor" stroke="none">
                          <path d="M12 2L15.09 8.26L22 9.27L17 14.14L18.18 21.02L12 17.77L5.82 21.02L7 14.14L2 9.27L8.91 8.26L12 2Z" />
                        </svg>
                        {t('play.lobby.host')}
                      </span>
                    )}
                    {name === user.username && !isPlayerHost && (
                      <span className="text-[10px] font-heading text-[var(--color-text-muted)] opacity-50">
                        {t('play.lobby.you')}
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
                {game.playerCount < 2 ? t('play.lobby.waiting') : t('play.lobby.startGame', { count: game.playerCount })}
              </button>
            ) : (
              <button
                disabled
                className="text-base px-10 py-3.5 font-heading font-bold border-2 cursor-not-allowed opacity-60 rounded-[var(--radius-sm)]"
                style={{
                  borderColor: 'rgba(13,148,136,0.2)',
                  color: 'var(--color-secondary)',
                  background: 'rgba(13,148,136,0.05)',
                }}
              >
                {t('play.lobby.waitingHost')}
              </button>
            )}
            <button onClick={handleLeave} className="btn-secondary py-3.5 px-6 cursor-pointer">
              {t('play.lobby.leave')}
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
          <div className="font-heading font-bold text-3xl sm:text-4xl mb-2 text-[var(--color-primary)]">
            {game.mode === 'duel' ? t('play.matched.matchFound') : game.mode === 'battle' ? t('play.matched.gameStarting') : t('play.matched.arenaReady')}
          </div>
          {game.mode === 'duel' && game.opponent && (
            <div className="text-lg text-[var(--color-text-secondary)] mb-6">
              vs <span className="font-bold text-[var(--color-secondary)]">{game.opponent}</span>
            </div>
          )}
          {game.mode === 'battle' && (
            <div className="text-lg text-[var(--color-text-secondary)] mb-6">
              {t('play.matched.players', { count: game.playerCount })}
            </div>
          )}
          <button
            onClick={handleReady}
            disabled={readySent}
            className={`text-lg px-12 py-4 font-heading font-bold border-2 transition-all cursor-pointer rounded-[var(--radius-sm)] ${readySent
                ? 'border-[var(--color-primary)] text-[var(--color-primary)] bg-[rgba(79,70,229,0.06)] opacity-80 cursor-not-allowed'
                : 'btn-primary'
              }`}
          >
            {readySent ? t('play.matched.waiting') : t('play.matched.ready')}
          </button>
          {readySent && (
            <p className="text-sm text-[var(--color-text-muted)] mt-3 animate-pulse">
              {t('play.matched.waitingOpponent')}
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
      <div className="h-[calc(100vh-4.5rem)] relative bg-[var(--color-bg-primary)]">
        <GameCanvas
          targets={game.targets}
          question={game.question}
          quizType={game.quizType}
          round={game.round}
          totalRounds={game.totalRounds}
          timeMs={game.timeMs}
          myCorrect={game.myCorrect}
          opponentCorrect={game.opponentCorrect}
          opponent={game.opponent}
          mode={game.mode}
          lastReactionMs={game.lastReactionMs}
          lastIsCorrect={game.lastIsCorrect}
          onHit={(target, _ms) => game.hitTarget(target.id)}
          claimedTargets={game.claimedTargets}
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
      <div className="min-h-screen bg-[var(--color-bg-primary)]">
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
