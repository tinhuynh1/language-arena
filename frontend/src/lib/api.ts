const API_URL = process.env.NEXT_PUBLIC_API_URL || '';
const WS_URL = process.env.NEXT_PUBLIC_WS_URL || '';

export interface APIResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
}

export interface User {
  id: string;
  username: string;
  email: string;
  total_score: number;
  games_played: number;
  best_reaction_ms: number;
  created_at: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface Vocabulary {
  id: string;
  word: string;
  meaning: string;
  language: string;
  difficulty: number;
  category: string;
}

export interface LeaderboardEntry {
  rank: number;
  user_id: string;
  username: string;
  total_score: number;
  games_played: number;
  best_reaction_ms: number;
}

function getToken(): string | null {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem('token');
}

async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const token = getToken();
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options?.headers as Record<string, string>),
  };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_URL}${path}`, { ...options, headers });
  const data: APIResponse<T> = await res.json();

  if (!data.success) {
    throw new Error(data.error || 'Request failed');
  }
  return data.data as T;
}

export const api = {
  auth: {
    register: (body: { username: string; email: string; password: string }) =>
      apiFetch<AuthResponse>('/api/v1/auth/register', { method: 'POST', body: JSON.stringify(body) }),
    login: (body: { email: string; password: string }) =>
      apiFetch<AuthResponse>('/api/v1/auth/login', { method: 'POST', body: JSON.stringify(body) }),
  },
  vocab: {
    get: (lang: string) => apiFetch<Vocabulary[]>(`/api/v1/vocab?lang=${lang}`),
  },
  leaderboard: {
    get: (limit = 20) => apiFetch<LeaderboardEntry[]>(`/api/v1/leaderboard?limit=${limit}`),
  },
  stats: {
    me: () => apiFetch<{ user: User; recent_games: unknown[] }>('/api/v1/stats/me'),
  },
  online: () => apiFetch<{ online: number }>('/api/v1/online'),
};

export function getWsUrl(): string {
  const token = getToken();
  if (WS_URL) {
    return `${WS_URL}/api/v1/ws/game?token=${token}`;
  }
  // Auto-detect from current page (works behind any proxy/tunnel)
  const proto = typeof window !== 'undefined' && window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const host = typeof window !== 'undefined' ? window.location.host : 'localhost';
  return `${proto}//${host}/api/v1/ws/game?token=${token}`;
}

export { API_URL, WS_URL };
