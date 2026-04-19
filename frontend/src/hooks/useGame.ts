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
          });
          break;
        }

        case 'player_joined': {
          const data = msg.data as PlayerJoinedData;
          updateStore({
            playerCount: data.player_count,
            players: data.players,
          });
          break;
        }

        case 'player_left': {
          const data = msg.data as PlayerJoinedData;
          updateStore({
            playerCount: data.player_count,
            players: data.players,
          });
          break;
        }

        case 'match_found': {
          const data = msg.data as MatchFoundData;
          updateStore({
            state: 'matched',
            roomId: data.room_id,
            opponent: data.opponent || '',
            playerCount: data.player_count || 2,
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

        case 'round_end':
          updateStore({ state: 'round_end' });
          break;

        case 'game_over': {
          const data = msg.data as GameOverData;
          updateStore({ state: 'game_over', gameOverData: data });
          break;
        }

        case 'opponent_left':
          updateStore({ state: 'game_over', gameOverData: null });
          break;

        case 'error':
          console.error('[Game] Error:', msg.data);
          break;
      }
    });

    return removeHandler;
  }, [ws, updateStore]);

  const joinGame = useCallback((mode: 'solo' | 'duel', language: string, level: string, quizType: QuizType) => {
    ws.connect();
    updateStore({ ...initialStore, mode, language, level, quizType, state: 'queuing' });
    ws.send({ type: 'join_queue', data: { mode, language, level, quiz_type: quizType } });
  }, [ws, updateStore]);

  const createRoom = useCallback((language: string, level: string, quizType: QuizType) => {
    ws.connect();
    updateStore({ ...initialStore, mode: 'battle', language, level, quizType, state: 'creating_room' });
    ws.send({ type: 'create_room', data: { language, level, quiz_type: quizType } });
  }, [ws, updateStore]);

  const joinRoom = useCallback((roomCode: string) => {
    ws.connect();
    updateStore({ ...initialStore, mode: 'battle', state: 'queuing', roomCode });
    ws.send({ type: 'join_room', data: { room_code: roomCode } });
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
