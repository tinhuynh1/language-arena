'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import { getWsUrl } from '@/lib/api';

export type WSMessageType =
  | 'join_queue' | 'create_room' | 'join_room' | 'start_game' | 'ready' | 'target_hit' | 'leave_room'
  | 'queue_joined' | 'match_found' | 'room_created' | 'player_joined' | 'player_left'
  | 'countdown' | 'round_start' | 'score_update' | 'live_leaderboard' | 'target_claimed'
  | 'round_end' | 'game_over' | 'opponent_left' | 'host_changed' | 'game_state_sync' | 'error';

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
  your_correct: number;
  your_avg_ms: number;
  opponent_correct?: number;
  opponent_avg_ms?: number;
  last_hit_by?: string;
  reaction_ms?: number;
  is_correct?: boolean;
}

export interface LeaderboardPlayer {
  rank: number;
  username: string;
  correct_count: number;
  avg_reaction_ms: number;
}

export interface LiveLeaderboardData {
  round: number;
  players: LeaderboardPlayer[];
}

export interface GameOverData {
  winner: string;
  your_correct: number;
  your_avg_ms: number;
  opponent_correct: number;
  opponent_avg_ms: number;
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
  is_host: boolean;
  host: string;
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
  host: string;
}

export interface HostChangedData {
  new_host: string;
}

export interface TargetClaimedData {
  target_id: string;
  claimed_by: string;
}

export interface GameStateSyncData {
  room_code: string;
  mode: string;
  state: string;
  round: number;
  total_rounds: number;
  question: string;
  targets: Target[];
  time_ms: number;
  elapsed_ms: number;
  your_correct: number;
  your_avg_ms: number;
  opponent_correct: number;
  opponent_avg_ms: number;
  players: string[];
}

type MessageHandler = (msg: WSMessage) => void;

const MAX_RECONNECT_ATTEMPTS = 5;
const RECONNECT_BASE_DELAY_MS = 1000;

export function useWebSocket() {
  const wsRef = useRef<WebSocket | null>(null);
  const [connected, setConnected] = useState(false);
  const handlersRef = useRef<Set<MessageHandler>>(new Set());
  const msgQueue = useRef<WSMessage[]>([]);
  const reconnectAttempts = useRef(0);
  const reconnectTimer = useRef<ReturnType<typeof setTimeout>>(undefined);
  const intentionalClose = useRef(false);

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN || wsRef.current?.readyState === WebSocket.CONNECTING) return;

    intentionalClose.current = false;
    const url = getWsUrl();
    console.log('[WS] Connecting to:', url.replace(/token=.*/, 'token=***'));

    const ws = new WebSocket(url);
    wsRef.current = ws;

    // Connection timeout: if WS doesn't open within 8s, force reconnect
    const connectTimeout = setTimeout(() => {
      if (ws.readyState !== WebSocket.OPEN) {
        console.warn('[WS] Connection timeout, retrying...');
        ws.close();
      }
    }, 8000);

    ws.onopen = () => {
      clearTimeout(connectTimeout);
      setConnected(true);
      reconnectAttempts.current = 0;
      console.log('[WS] Connected');

      // Flush queued messages
      while (msgQueue.current.length > 0) {
        const msg = msgQueue.current.shift();
        if (msg) {
          console.log('[WS] Sending queued msg:', msg.type);
          ws.send(JSON.stringify(msg));
        }
      }
    };

    ws.onmessage = (event) => {
      try {
        const msg: WSMessage = JSON.parse(event.data);
        console.log('[WS] Received:', msg.type);
        handlersRef.current.forEach(handler => handler(msg));
      } catch (err) {
        console.error('[WS] Parse error:', err);
      }
    };

    ws.onclose = (event) => {
      clearTimeout(connectTimeout);
      setConnected(false);
      console.log('[WS] Disconnected, code:', event.code, 'reason:', event.reason);

      // Auto-reconnect on unexpected disconnect (covers F5, network drop, etc.)
      if (!intentionalClose.current && reconnectAttempts.current < MAX_RECONNECT_ATTEMPTS) {
        reconnectAttempts.current++;
        const delay = RECONNECT_BASE_DELAY_MS * Math.pow(1.5, reconnectAttempts.current - 1);
        console.log(`[WS] Reconnecting in ${Math.round(delay)}ms (attempt ${reconnectAttempts.current}/${MAX_RECONNECT_ATTEMPTS})...`);
        reconnectTimer.current = setTimeout(() => {
          connect();
        }, delay);
      }
    };

    ws.onerror = (err) => {
      console.error('[WS] Error:', err);
    };
  }, []);

  const disconnect = useCallback(() => {
    intentionalClose.current = true;
    clearTimeout(reconnectTimer.current);
    reconnectAttempts.current = 0;
    msgQueue.current = [];
    wsRef.current?.close();
    wsRef.current = null;
    setConnected(false);
  }, []);

  const send = useCallback((msg: WSMessage) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      console.log('[WS] Sending:', msg.type);
      wsRef.current.send(JSON.stringify(msg));
    } else {
      console.log('[WS] Queued:', msg.type);
      msgQueue.current.push(msg);
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
