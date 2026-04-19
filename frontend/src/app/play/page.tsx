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

  // Idle - Mode/Lang/Level Selection
  if (game.state === 'idle') {
    return (
      <div className="min-h-screen flex items-center justify-center px-4 sm:px-6">
        <div className="w-full max-w-2xl text-center">
          <h1 className="font-heading font-bold text-2xl sm:text-4xl mb-2 uppercase tracking-wider">
            Select <span style={{ color: '#00ff88' }}>Mission</span>
          </h1>
          <p className="text-[var(--color-text-muted)] mb-6 sm:mb-8 text-xs sm:text-sm">Choose mode, language, and level</p>

          {/* Mode Selection */}
          <div className="grid grid-cols-3 gap-2 sm:gap-3 mb-4 sm:mb-6">
            <button
              onClick={() => setSelectedMode('solo')}
              className={`card text-left transition-all ${selectedMode === 'solo' ? 'border-[var(--color-accent-neon)]' : ''}`}
            >
              <div className="font-heading font-bold text-sm sm:text-lg mb-1 uppercase" style={{ color: selectedMode === 'solo' ? '#00ff88' : 'inherit' }}>
                Solo
              </div>
              <p className="text-xs text-[var(--color-text-muted)]">Practice alone</p>
            </button>
            <button
              onClick={() => setSelectedMode('duel')}
              className={`card text-left transition-all ${selectedMode === 'duel' ? 'border-[var(--color-accent-orange)]' : ''}`}
            >
              <div className="font-heading font-bold text-sm sm:text-lg mb-1 uppercase" style={{ color: selectedMode === 'duel' ? '#ff6b35' : 'inherit' }}>
                1v1 Duel
              </div>
              <p className="text-xs text-[var(--color-text-muted)]">Real-time match</p>
            </button>
            <button
              onClick={() => setSelectedMode('battle')}
              className={`card text-left transition-all ${selectedMode === 'battle' ? 'border-[var(--color-accent-cyan)]' : ''}`}
            >
              <div className="font-heading font-bold text-sm sm:text-lg mb-1 uppercase" style={{ color: selectedMode === 'battle' ? '#00d4ff' : 'inherit' }}>
                Battle
              </div>
              <p className="text-xs text-[var(--color-text-muted)]">Up to 100 players</p>
            </button>
          </div>

          {/* Language Selection */}
          <div className="grid grid-cols-2 gap-2 sm:gap-3 mb-4 sm:mb-6">
            <button
              onClick={() => handleLangChange('en')}
              className={`card text-center transition-all ${selectedLang === 'en' ? 'border-[var(--color-accent-cyan)]' : ''}`}
            >
              <div className="text-2xl mb-1">🇬🇧</div>
              <div className="font-heading font-bold uppercase text-sm" style={{ color: selectedLang === 'en' ? '#00d4ff' : 'inherit' }}>
                English (CEFR)
              </div>
            </button>
            <button
              onClick={() => handleLangChange('zh')}
              className={`card text-center transition-all ${selectedLang === 'zh' ? 'border-[var(--color-accent-cyan)]' : ''}`}
            >
              <div className="text-2xl mb-1">🇨🇳</div>
              <div className="font-heading font-bold uppercase text-sm" style={{ color: selectedLang === 'zh' ? '#00d4ff' : 'inherit' }}>
                Chinese (HSK)
              </div>
            </button>
          </div>

          {/* Level Selection */}
          <div className="mb-4 sm:mb-6">
            <div className="text-xs font-heading uppercase tracking-widest text-[var(--color-text-muted)] mb-3">
              Vocabulary Level
            </div>
            <div className="flex gap-2 justify-center flex-wrap">
              {levels.map(l => (
                <button
                  key={l.value}
                  onClick={() => setSelectedLevel(l.value)}
                  className={`px-4 py-2 font-heading font-bold text-sm uppercase transition-all border ${selectedLevel === l.value
                      ? 'border-[var(--color-accent-neon)] text-[var(--color-accent-neon)] bg-[rgba(0,255,136,0.1)]'
                      : 'border-[var(--color-border-default)] text-[var(--color-text-secondary)] hover:border-[var(--color-text-muted)]'
                    }`}
                  style={{ borderRadius: '2px' }}
                >
                  <div>{l.label}</div>
                  <div className="text-[10px] font-normal lowercase opacity-70">{l.desc}</div>
                </button>
              ))}
            </div>
          </div>

          {/* Quiz Type Selection */}
          <div className="mb-8">
            <div className="text-xs font-heading uppercase tracking-widest text-[var(--color-text-muted)] mb-3">
              Quiz Type
            </div>
            <div className="grid grid-cols-3 gap-2">
              {quizTypes.map(q => (
                <button
                  key={q.value}
                  onClick={() => setSelectedQuizType(q.value)}
                  className={`px-3 py-2 font-heading font-bold text-sm uppercase transition-all border text-left ${selectedQuizType === q.value
                      ? 'border-[var(--color-accent-orange)] text-[var(--color-accent-orange)] bg-[rgba(255,107,53,0.1)]'
                      : 'border-[var(--color-border-default)] text-[var(--color-text-secondary)] hover:border-[var(--color-text-muted)]'
                    }`}
                  style={{ borderRadius: '2px' }}
                >
                  <div>{q.label}</div>
                  <div className="text-[10px] font-normal normal-case opacity-70">{q.desc}</div>
                </button>
              ))}
            </div>
          </div>

          {/* Action Buttons */}
          <div className="space-y-3">
            <button onClick={handleStart} className="btn-primary text-base sm:text-lg px-8 sm:px-12 py-3 sm:py-4 w-full max-w-xs mx-auto block">
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
              <div className="w-16 h-16 border-2 border-[var(--color-accent-orange)] border-t-transparent rounded-full animate-spin mx-auto mb-6" />
              <div className="font-heading font-bold text-xl sm:text-2xl mb-2 uppercase" style={{ color: '#ff6b35' }}>
                Connecting...
              </div>
              <p className="text-xs sm:text-sm text-[var(--color-text-muted)]">
                Establishing secure connection
              </p>
            </>
          ) : (
            <>
              <div className="w-16 h-16 border-2 border-[var(--color-accent-neon)] border-t-transparent rounded-full animate-spin mx-auto mb-6" />
              <div className="font-heading font-bold text-xl sm:text-2xl mb-2 uppercase">
                {selectedMode === 'duel' ? 'Finding Opponent...' : selectedMode === 'battle' ? 'Creating Room...' : 'Preparing Arena...'}
              </div>
              <p className="text-xs sm:text-sm text-[var(--color-text-muted)]">
                {selectedLevel} • {selectedLang === 'en' ? 'English' : 'Chinese'}
              </p>
            </>
          )}
          <button onClick={handleLeave} className="btn-secondary mt-6 sm:mt-8 text-sm">CANCEL</button>
        </div>
      </div>
    );
  }

  // Battle Lobby - Waiting for players
  if (game.state === 'in_lobby') {
    return (
      <div className="min-h-screen flex items-center justify-center px-6">
        <div className="w-full max-w-md text-center">
          <div className="font-heading font-bold text-2xl sm:text-3xl mb-2 uppercase" style={{ color: '#00d4ff' }}>
            BATTLE ROOM
          </div>

          {/* Room Code */}
          <div className="card mb-6 py-6">
            <div className="text-xs font-heading uppercase tracking-widest text-[var(--color-text-muted)] mb-2">
              Share this code with friends
            </div>
            <div
              className="font-mono font-bold text-3xl sm:text-5xl tracking-[0.2em] sm:tracking-[0.3em] text-glow cursor-pointer"
              style={{ color: '#00ff88' }}
              onClick={() => navigator.clipboard?.writeText(game.roomCode)}
              title="Click to copy"
            >
              {game.roomCode}
            </div>
            <div className="text-xs text-[var(--color-text-muted)] mt-2">Click to copy</div>
          </div>

          {/* Player Count */}
          <div className="card mb-6">
            <div className="flex items-center justify-between mb-3">
              <span className="text-sm font-heading uppercase tracking-wider text-[var(--color-text-muted)]">Players</span>
              <span className="font-mono font-bold text-lg" style={{ color: '#00ff88' }}>
                {game.playerCount}/100
              </span>
            </div>
            {game.players.length > 0 && (
              <div className="flex flex-wrap gap-2">
                {game.players.map(name => (
                  <span key={name} className="px-2 py-1 text-xs font-heading bg-[var(--color-bg-primary)] border border-[var(--color-border-default)]"
                    style={{ borderRadius: '2px' }}>
                    {name}
                  </span>
                ))}
              </div>
            )}
          </div>

          {/* Start Button (Host only) */}
          <div className="flex gap-3 justify-center">
            <button
              onClick={() => game.startGame()}
              className="btn-primary text-lg px-10 py-4"
              disabled={game.playerCount < 1}
            >
              START GAME ({game.playerCount} players)
            </button>
            <button onClick={handleLeave} className="btn-secondary py-4 px-6">
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
          <div className="font-heading font-bold text-3xl mb-2 uppercase" style={{ color: '#00ff88' }}>
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
            className={`text-lg px-12 py-4 font-heading font-bold uppercase tracking-wider border-2 transition-all ${readySent
                ? 'border-[var(--color-accent-neon)] text-[var(--color-accent-neon)] bg-[rgba(0,255,136,0.1)] opacity-80 cursor-not-allowed'
                : 'btn-primary'
              }`}
            style={{ borderRadius: '2px' }}
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
      <div className="h-[calc(100vh-3.5rem)] sm:h-[calc(100vh-4rem)] relative" style={{ background: 'var(--color-bg-primary)' }}>
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
