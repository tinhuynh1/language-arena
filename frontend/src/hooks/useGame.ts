'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import {
  useWebSocket,
  type WSMessage,
  type Target,
  type QuizType,
  type RoundStartData,
  type ScoreUpdateData,
  type GameOverData,
  type MatchFoundData,
  type RoomCreatedData,
  type PlayerJoinedData,
  type LeaderboardPlayer,
  type LiveLeaderboardData,
  type HostChangedData,
  type GameStateSyncData,
} from './useWebSocket';

export type GameState = 'idle' | 'queuing' | 'creating_room' | 'in_lobby' | 'matched' | 'countdown' | 'playing' | 'round_end' | 'game_over';

interface GameStore {
  state: GameState;
  mode: 'solo' | 'duel' | 'battle';
  language: string;
  level: string;
  quizType: QuizType;
  roomId: string;
  roomCode: string;
  opponent: string;
  playerCount: number;
  players: string[];
  round: number;
  totalRounds: number;
  question: string;
  targets: Target[];
  myScore: number;
  opponentScore: number;
  timeMs: number;
  countdownMs: number;
  lastReactionMs: number;
  gameOverData: GameOverData | null;
  liveLeaderboard: LeaderboardPlayer[];
  errorMessage: string;
  isHost: boolean;
  hostUsername: string;
}

const initialStore: GameStore = {
  state: 'idle',
  mode: 'solo',
  language: 'en',
  level: 'A1',
  quizType: 'meaning_to_word',
  roomId: '',
  roomCode: '',
  opponent: '',
  playerCount: 0,
  players: [],
  round: 0,
  totalRounds: 10,
  question: '',
  targets: [],
  myScore: 0,
  opponentScore: 0,
  timeMs: 5000,
  countdownMs: 0,
  lastReactionMs: 0,
  gameOverData: null,
  liveLeaderboard: [],
  errorMessage: '',
  isHost: false,
  hostUsername: '',
};

export function useGame() {
  const [store, setStore] = useState<GameStore>(initialStore);
  const ws = useWebSocket();
  const roundStartTimeRef = useRef<number>(0);

  const updateStore = useCallback((update: Partial<GameStore>) => {
    setStore(prev => ({ ...prev, ...update }));
  }, []);

  useEffect(() => {
    const removeHandler = ws.addHandler((msg: WSMessage) => {
      switch (msg.type) {
        case 'queue_joined':
          updateStore({ state: 'queuing' });
          break;

        case 'room_created': {
          const data = msg.data as RoomCreatedData;
          updateStore({
            state: 'in_lobby',
            roomCode: data.room_code,
            roomId: data.room_id,
            playerCount: 1,
            isHost: true,
          });
          break;
        }

        case 'player_joined': {
          const data = msg.data as PlayerJoinedData;
          updateStore({
            playerCount: data.player_count,
            players: data.players,
            hostUsername: data.host || '',
          });
          break;
        }

        case 'player_left': {
          const data = msg.data as PlayerJoinedData;
          updateStore({
            playerCount: data.player_count,
            players: data.players,
            hostUsername: data.host || '',
          });
          break;
        }

        case 'match_found': {
          const data = msg.data as MatchFoundData;
          updateStore({
            state: data.mode === 'battle' ? 'in_lobby' : 'matched',
            roomId: data.room_id,
            opponent: data.opponent || '',
            playerCount: data.player_count || 2,
            isHost: data.is_host || false,
            hostUsername: data.host || '',
          });
          break;
        }

        case 'countdown': {
          const data = msg.data as { ms: number };
          updateStore({ state: 'countdown', countdownMs: data.ms });
          break;
        }

        case 'round_start': {
          const data = msg.data as RoundStartData;
          roundStartTimeRef.current = Date.now();
          updateStore({
            state: 'playing',
            round: data.round,
            totalRounds: data.total,
            question: data.question,
            targets: data.targets,
            timeMs: data.time_ms,
          });
          break;
        }

        case 'score_update': {
          const data = msg.data as ScoreUpdateData;
          updateStore({
            myScore: data.you,
            opponentScore: data.opponent,
            lastReactionMs: data.reaction_ms || 0,
          });
          break;
        }

        case 'live_leaderboard': {
          const data = msg.data as LiveLeaderboardData;
          updateStore({ liveLeaderboard: data.players });
          break;
        }

        case 'round_end': {
          const data = msg.data as { result: string; next_in_ms: number };
          updateStore({ state: 'round_end' });
          break;
        }

        case 'game_over': {
          const data = msg.data as GameOverData;
          updateStore({ state: 'game_over', gameOverData: data });
          break;
        }

        case 'opponent_left':
          updateStore({ state: 'game_over', gameOverData: null });
          break;

        case 'host_changed': {
          const data = msg.data as HostChangedData;
          updateStore({
            hostUsername: data.new_host,
          });
          break;
        }

        case 'game_state_sync': {
          const data = msg.data as GameStateSyncData;
          const stateMap: Record<string, GameState> = {
            waiting: 'in_lobby',
            countdown: 'countdown',
            playing: 'playing',
            round_end: 'round_end',
            finished: 'game_over',
          };
          roundStartTimeRef.current = Date.now() - (data.elapsed_ms || 0);
          updateStore({
            state: stateMap[data.state] || 'playing',
            mode: (data.mode as 'solo' | 'duel' | 'battle') || 'solo',
            roomCode: data.room_code,
            round: data.round,
            totalRounds: data.total_rounds,
            question: data.question,
            targets: data.targets,
            timeMs: data.time_ms,
            myScore: data.your_score,
            opponentScore: data.opponent_score,
            players: data.players,
            playerCount: data.players?.length || 0,
          });
          console.log('[Game] Reconnected! Restored state:', data.state, 'round:', data.round);
          break;
        }

        case 'error': {
          const errorStr = typeof msg.data === 'string' ? msg.data : 'Unknown error';
          console.error('[Game] Error:', errorStr);
          // Reset to idle so the user can try again
          updateStore({ state: 'idle', errorMessage: errorStr });
          break;
        }
      }
    });

    return removeHandler;
  }, [ws, updateStore]);

  const joinGame = useCallback((mode: 'solo' | 'duel', language: string, level: string, quizType: QuizType) => {
    ws.connect();
    updateStore({ ...initialStore, mode, language, level, quizType, state: 'queuing', errorMessage: '' });

    setTimeout(() => {
      ws.send({ type: 'join_queue', data: { mode, language, level, quiz_type: quizType } });
    }, 500);
  }, [ws, updateStore]);

  const createRoom = useCallback((language: string, level: string, quizType: QuizType) => {
    ws.connect();
    updateStore({ ...initialStore, mode: 'battle', language, level, quizType, state: 'creating_room', errorMessage: '' });

    setTimeout(() => {
      ws.send({ type: 'create_room', data: { language, level, quiz_type: quizType } });
    }, 500);
  }, [ws, updateStore]);

  const joinRoom = useCallback((roomCode: string) => {
    ws.connect();
    updateStore({ ...initialStore, mode: 'battle', state: 'queuing', roomCode, errorMessage: '' });

    setTimeout(() => {
      ws.send({ type: 'join_room', data: { room_code: roomCode } });
    }, 500);
  }, [ws, updateStore]);

  const startGame = useCallback(() => {
    ws.send({ type: 'start_game' });
  }, [ws]);

  const ready = useCallback(() => {
    ws.send({ type: 'ready' });
  }, [ws]);

  const hitTarget = useCallback((targetId: string) => {
    const reactionMs = Date.now() - roundStartTimeRef.current;
    ws.send({
      type: 'target_hit',
      data: { target_id: targetId, reaction_ms: reactionMs },
    });
    return reactionMs;
  }, [ws]);

  const leaveGame = useCallback(() => {
    ws.send({ type: 'leave_room' });
    ws.disconnect();
    setStore(initialStore);
  }, [ws]);

  return {
    ...store,
    connected: ws.connected,
    joinGame,
    createRoom,
    joinRoom,
    startGame,
    ready,
    hitTarget,
    leaveGame,
  };
}
