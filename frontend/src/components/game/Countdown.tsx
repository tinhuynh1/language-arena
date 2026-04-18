'use client';

import { useEffect, useState } from 'react';

interface CountdownProps {
  ms: number;
  onComplete: () => void;
}

export default function Countdown({ ms, onComplete }: CountdownProps) {
  const totalSeconds = Math.ceil(ms / 1000);
  const [count, setCount] = useState(totalSeconds);
  const [animKey, setAnimKey] = useState(0);

  useEffect(() => {
    if (count <= 0) {
      const goTimer = setTimeout(onComplete, 600);
      return () => clearTimeout(goTimer);
    }

    const timer = setTimeout(() => {
      setCount(prev => prev - 1);
      setAnimKey(prev => prev + 1);
    }, 1000);
    return () => clearTimeout(timer);
  }, [count, onComplete]);

  const progress = count / totalSeconds;
  const circumference = 2 * Math.PI * 90;
  const strokeOffset = circumference * (1 - progress);

  const getColor = () => {
    if (count > 2) return '#00ff88';
    if (count > 1) return '#ffd700';
    if (count > 0) return '#ff3548';
    return '#00d4ff';
  };

  const color = getColor();

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center"
         style={{ background: 'radial-gradient(ellipse at center, rgba(10,14,23,0.95) 0%, #0a0e17 100%)' }}>

      {/* Ambient glow */}
      <div
        className="absolute rounded-full"
        style={{
          width: '300px',
          height: '300px',
          background: `radial-gradient(circle, ${color}15 0%, transparent 70%)`,
          filter: 'blur(40px)',
          transition: 'background 0.3s ease',
        }}
      />

      {/* SVG Progress Ring */}
      <div className="relative" style={{ width: '240px', height: '240px' }}>
        <svg viewBox="0 0 200 200" className="w-full h-full -rotate-90">
          {/* Background ring */}
          <circle
            cx="100" cy="100" r="90"
            fill="none"
            stroke="rgba(255,255,255,0.05)"
            strokeWidth="4"
          />
          {/* Progress ring */}
          <circle
            cx="100" cy="100" r="90"
            fill="none"
            stroke={color}
            strokeWidth="4"
            strokeLinecap="round"
            strokeDasharray={circumference}
            strokeDashoffset={strokeOffset}
            style={{
              transition: 'stroke-dashoffset 1s linear, stroke 0.3s ease',
              filter: `drop-shadow(0 0 8px ${color}80)`,
            }}
          />
          {/* Tick marks */}
          {Array.from({ length: 12 }).map((_, i) => {
            const angle = (i * 30 * Math.PI) / 180;
            const x1 = 100 + 78 * Math.cos(angle);
            const y1 = 100 + 78 * Math.sin(angle);
            const x2 = 100 + 84 * Math.cos(angle);
            const y2 = 100 + 84 * Math.sin(angle);
            return (
              <line
                key={i}
                x1={x1} y1={y1} x2={x2} y2={y2}
                stroke="rgba(255,255,255,0.15)"
                strokeWidth="2"
                strokeLinecap="round"
              />
            );
          })}
        </svg>

        {/* Number */}
        <div
          key={animKey}
          className="absolute inset-0 flex flex-col items-center justify-center"
          style={{ animation: 'countdownPop 0.4s cubic-bezier(0.34, 1.56, 0.64, 1)' }}
        >
          <div
            className="font-heading font-bold leading-none"
            style={{
              fontSize: count > 0 ? '7rem' : '4rem',
              color,
              textShadow: `0 0 30px ${color}60, 0 0 60px ${color}30`,
              transition: 'color 0.3s ease',
            }}
          >
            {count > 0 ? count : 'GO!'}
          </div>
          {count > 0 && (
            <div className="text-xs font-heading uppercase tracking-[0.3em] mt-2"
                 style={{ color: 'rgba(255,255,255,0.3)' }}>
              GET READY
            </div>
          )}
        </div>
      </div>

      {/* Decorative corners */}
      <div className="absolute" style={{ width: '280px', height: '280px' }}>
        {/* Top-left corner */}
        <div className="absolute top-0 left-0 w-6 h-6 border-t-2 border-l-2" style={{ borderColor: `${color}40` }} />
        {/* Top-right corner */}
        <div className="absolute top-0 right-0 w-6 h-6 border-t-2 border-r-2" style={{ borderColor: `${color}40` }} />
        {/* Bottom-left corner */}
        <div className="absolute bottom-0 left-0 w-6 h-6 border-b-2 border-l-2" style={{ borderColor: `${color}40` }} />
        {/* Bottom-right corner */}
        <div className="absolute bottom-0 right-0 w-6 h-6 border-b-2 border-r-2" style={{ borderColor: `${color}40` }} />
      </div>

      {/* Scanline effect */}
      <div className="absolute inset-0 pointer-events-none overflow-hidden opacity-[0.03]">
        <div style={{
          width: '100%',
          height: '200%',
          backgroundImage: 'repeating-linear-gradient(0deg, transparent, transparent 2px, rgba(255,255,255,0.5) 2px, rgba(255,255,255,0.5) 4px)',
          animation: 'scanline 4s linear infinite',
        }} />
      </div>
    </div>
  );
}
