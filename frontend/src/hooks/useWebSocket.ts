'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import { getWsUrl } from '@/lib/api';

export type WSMessageType =
  | 'join_queue' | 'create_room' | 'join_room' | 'start_game' | 'ready' | 'target_hit' | 'leave_room'
  | 'queue_joined' | 'match_found' | 'room_created' | 'player_joined' | 'player_left'
  | 'countdown' | 'round_start' | 'score_update' | 'live_leaderboard'
  | 'round_end' | 'game_over' | 'opponent_left' | 'error';

export interface WSMessage {
  type: WSMessageType;
  data?: unknown;
}

export type QuizType =
  | 'meaning_to_word'
  | 'word_to_meaning'
  | 'word_to_ipa'
  | 'word_to_pinyin';

export interface Target {
  id: string;
  word: string;
  meaning: string;
  label: string;
  x: number;
  y: number;
  correct: boolean;
}

export interface RoundStartData {
  round: number;
  total: number;
  question: string;
  targets: Target[];
  time_ms: number;
}

export interface ScoreUpdateData {
  you: number;
  opponent: number;
  last_hit_by?: string;
  reaction_ms?: number;
}

export interface LeaderboardPlayer {
  rank: number;
  username: string;
  score: number;
}

export interface LiveLeaderboardData {
  round: number;
  players: LeaderboardPlayer[];
}

export interface GameOverData {
  winner: string;
  your_score: number;
  opponent_score: number;
  stats: {
    total_rounds: number;
    avg_reaction_ms: number;
    accuracy: number;
  };
  ranking?: LeaderboardPlayer[];
}

export interface MatchFoundData {
  room_id: string;
  opponent: string;
  player_count: number;
  mode: string;
}

export interface RoomCreatedData {
  room_code: string;
  room_id: string;
  language: string;
  level: string;
}

export interface PlayerJoinedData {
  username: string;
  player_count: number;
  players: string[];
}

type MessageHandler = (msg: WSMessage) => void;

export function useWebSocket() {
  const wsRef = useRef<WebSocket | null>(null);
  const [connected, setConnected] = useState(false);
  const handlersRef = useRef<Set<MessageHandler>>(new Set());

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return;

    const ws = new WebSocket(getWsUrl());
    wsRef.current = ws;

    ws.onopen = () => {
      setConnected(true);
      console.log('[WS] Connected');
    };

    ws.onmessage = (event) => {
      try {
        const msg: WSMessage = JSON.parse(event.data);
        handlersRef.current.forEach(handler => handler(msg));
      } catch (err) {
        console.error('[WS] Parse error:', err);
      }
    };

    ws.onclose = () => {
      setConnected(false);
      console.log('[WS] Disconnected');
    };

    ws.onerror = (err) => {
      console.error('[WS] Error:', err);
    };
  }, []);

  const disconnect = useCallback(() => {
    wsRef.current?.close();
    wsRef.current = null;
    setConnected(false);
  }, []);

  const send = useCallback((msg: WSMessage) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(msg));
    }
  }, []);

  const addHandler = useCallback((handler: MessageHandler) => {
    handlersRef.current.add(handler);
    return () => { handlersRef.current.delete(handler); };
  }, []);

  useEffect(() => {
    return () => { disconnect(); };
  }, [disconnect]);

  return { connected, connect, disconnect, send, addHandler };
}
