'use client';

import type { GameOverData, LeaderboardPlayer } from '@/hooks/useWebSocket';
import { useLocale } from '@/i18n/LocaleProvider';

interface GameOverScreenProps {
  data: GameOverData | null;
  mode: 'solo' | 'duel' | 'battle';
  username: string;
  onPlayAgain: () => void;
  onLeave: () => void;
}

export default function GameOverScreen({ data, mode, username, onPlayAgain, onLeave }: GameOverScreenProps) {
  const { t } = useLocale();

  if (!data) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[500px] gap-6">
        <div className="text-4xl font-heading font-bold text-[var(--color-accent-orange)]">
          {t('gameover.opponentLeft')}
        </div>
        <button onClick={onLeave} className="btn-primary">{t('gameover.backToLobby')}</button>
      </div>
    );
  }

  const isWinner = data.winner === username || (mode === 'solo');
  const reactionColor = data.stats.avg_reaction_ms < 1000 ? '#00ff88' : data.stats.avg_reaction_ms < 2000 ? '#ffd700' : '#ff3548';

  // Find player rank in battle mode
  let myRank = 0;
  if (mode === 'battle' && data.ranking) {
    const entry = data.ranking.find(p => p.username === username);
    if (entry) myRank = entry.rank;
  }

  const titleText = () => {
    if (mode === 'solo') return t('gameover.missionComplete');
    if (mode === 'battle') {
      if (myRank === 1) return t('gameover.champion');
      if (myRank === 2) return t('gameover.runnerUp');
      if (myRank === 3) return t('gameover.thirdPlace');
      if (myRank <= 10) return t('gameover.top10');
      return t('gameover.battleOver');
    }
    if (isWinner) return t('gameover.victory');
    if (data.winner === 'draw') return t('gameover.draw');
    return t('gameover.defeat');
  };

  const titleColor = () => {
    if (mode === 'solo') return '#00ff88';
    if (mode === 'battle') {
      if (myRank === 1) return '#ffd700';
      if (myRank === 2) return '#c0c0c0';
      if (myRank === 3) return '#cd7f32';
      return '#00d4ff';
    }
    return isWinner ? '#00ff88' : '#ff3548';
  };

  return (
    <div className="flex flex-col items-center justify-center min-h-[500px] gap-8 px-4 py-12">
      {/* Title */}
      <div className="text-center">
        <div className="text-6xl font-heading font-bold text-glow mb-2"
             style={{ color: titleColor() }}>
          {titleText()}
        </div>
        <div className="text-lg text-[var(--color-text-secondary)] font-heading">
          {mode === 'duel' && data.winner !== 'draw' && t('gameover.winner', { name: data.winner })}
          {mode === 'battle' && myRank > 0 && t('gameover.yourRank', { rank: myRank })}
        </div>
      </div>

      {/* Correct Answers */}
      <div className="flex items-center gap-12">
        <div className="text-center">
          <div className="text-sm text-[var(--color-text-muted)] font-heading uppercase tracking-wider mb-1">{t('gameover.correct')}</div>
          <div className="text-5xl font-heading font-bold" style={{ color: '#00ff88' }}>{data.your_correct}</div>
        </div>
        {mode === 'duel' && (
          <>
            <div className="text-3xl font-heading text-[var(--color-text-muted)]">VS</div>
            <div className="text-center">
              <div className="text-sm text-[var(--color-text-muted)] font-heading uppercase tracking-wider mb-1">{t('gameover.opponent')}</div>
              <div className="text-5xl font-heading font-bold" style={{ color: '#ff6b35' }}>{data.opponent_correct}</div>
            </div>
          </>
        )}
      </div>

      {/* Stats */}
      <div className="grid grid-cols-3 gap-6 w-full max-w-md">
        <div className="card text-center">
          <div className="text-xs text-[var(--color-text-muted)] font-heading uppercase tracking-wider mb-1">{t('gameover.rounds')}</div>
          <div className="text-2xl font-heading font-bold">{data.stats.total_rounds}</div>
        </div>
        <div className="card text-center">
          <div className="text-xs text-[var(--color-text-muted)] font-heading uppercase tracking-wider mb-1">{t('gameover.avgReaction')}</div>
          <div className="text-2xl font-heading font-bold font-mono" style={{ color: reactionColor }}>
            {data.stats.avg_reaction_ms}ms
          </div>
        </div>
        <div className="card text-center">
          <div className="text-xs text-[var(--color-text-muted)] font-heading uppercase tracking-wider mb-1">{t('gameover.accuracy')}</div>
          <div className="text-2xl font-heading font-bold">{data.stats.accuracy}%</div>
        </div>
      </div>

      {/* Battle Ranking */}
      {mode === 'battle' && data.ranking && data.ranking.length > 0 && (
        <div className="w-full max-w-md">
          <div className="text-xs font-heading uppercase tracking-widest text-[var(--color-text-muted)] mb-3 text-center">
            {t('gameover.finalRanking')}
          </div>
          <div className="space-y-1">
            {data.ranking.map((p) => {
              const rankColor = p.rank === 1 ? '#ffd700' : p.rank === 2 ? '#c0c0c0' : p.rank === 3 ? '#cd7f32' : 'var(--color-text-muted)';
              const isMe = p.username === username;
              return (
                <div
                  key={p.username}
                  className="flex items-center justify-between px-4 py-3"
                  style={{
                    background: isMe ? 'rgba(0, 255, 136, 0.1)' : 'var(--color-bg-card)',
                    borderLeft: isMe ? '3px solid #00ff88' : '3px solid transparent',
                    borderRadius: '2px',
                  }}
                >
                  <div className="flex items-center gap-3">
                    <span className="font-mono font-bold text-lg w-8" style={{ color: rankColor }}>#{p.rank}</span>
                    <span className="font-heading font-bold" style={{ color: isMe ? '#00ff88' : 'var(--color-text-primary)' }}>
                      {p.username} {isMe && t('gameover.you')}
                    </span>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="font-mono font-bold text-lg" style={{ color: '#00ff88' }}>{p.correct_count}</span>
                    {p.avg_reaction_ms > 0 && (
                      <span className="font-mono text-sm" style={{ color: 'var(--color-text-muted)' }}>{p.avg_reaction_ms}ms</span>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* Actions */}
      <div className="flex gap-4">
        <button onClick={onPlayAgain} className="btn-primary">{t('gameover.playAgain')}</button>
        <button onClick={onLeave} className="btn-secondary">{t('gameover.backToLobby')}</button>
      </div>
    </div>
  );
}
