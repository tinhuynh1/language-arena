'use client';

import { useEffect, useState } from 'react';

interface CountdownProps {
  ms: number;
  onComplete: () => void;
}

export default function Countdown({ ms, onComplete }: CountdownProps) {
  const [count, setCount] = useState(Math.ceil(ms / 1000));

  useEffect(() => {
    if (count <= 0) {
      onComplete();
      return;
    }

    const timer = setTimeout(() => setCount(prev => prev - 1), 1000);
    return () => clearTimeout(timer);
  }, [count, onComplete]);

  return (
    <div className="flex items-center justify-center min-h-[500px]">
      <div className="relative">
        <div
          className="text-[12rem] font-heading font-bold text-glow leading-none"
          style={{
            color: count > 2 ? '#00ff88' : count > 1 ? '#ffd700' : '#ff3548',
            animation: 'targetSpawn 0.5s cubic-bezier(0.34, 1.56, 0.64, 1)',
          }}
          key={count}
        >
          {count > 0 ? count : 'GO!'}
        </div>

        {/* Pulse rings */}
        <div
          className="absolute inset-0 rounded-full"
          style={{
            border: `2px solid ${count > 2 ? '#00ff88' : count > 1 ? '#ffd700' : '#ff3548'}`,
            opacity: 0.3,
            animation: 'targetPulse 1s ease-in-out infinite',
          }}
        />
      </div>
    </div>
  );
}
