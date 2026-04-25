'use client';

import { useEffect, useState } from 'react';
import { useLocale } from '@/i18n/LocaleProvider';

interface CountdownProps {
  ms: number;
  onComplete: () => void;
}

export default function Countdown({ ms, onComplete }: CountdownProps) {
  const { t } = useLocale();
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
  const color = '#00ff88';
  const dimColor = 'rgba(0, 255, 136, 0.15)';

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center"
         style={{ background: 'radial-gradient(ellipse at center, rgba(10,14,23,0.97) 0%, #080c14 100%)' }}>

      {/* Ambient glow behind scope */}
      <div
        className="absolute rounded-full"
        style={{
          width: '400px',
          height: '400px',
          background: `radial-gradient(circle, ${dimColor} 0%, transparent 70%)`,
          filter: 'blur(60px)',
          animation: 'scopePulse 2s ease-in-out infinite',
        }}
      />

      {/* Scope Container */}
      <div className="relative" style={{ width: '320px', height: '320px' }}>

        {/* Outer rotating ring with dashes */}
        <svg viewBox="0 0 320 320" className="absolute inset-0 w-full h-full" style={{ animation: 'scopeRotate 8s linear infinite' }}>
          <circle
            cx="160" cy="160" r="150"
            fill="none"
            stroke={color}
            strokeWidth="1"
            strokeDasharray="8 12"
            opacity="0.2"
          />
        </svg>

        {/* Progress ring (depleting) */}
        <svg viewBox="0 0 320 320" className="absolute inset-0 w-full h-full -rotate-90">
          {/* Background ring */}
          <circle
            cx="160" cy="160" r="130"
            fill="none"
            stroke="rgba(255,255,255,0.04)"
            strokeWidth="3"
          />
          {/* Active progress */}
          <circle
            cx="160" cy="160" r="130"
            fill="none"
            stroke={color}
            strokeWidth="3"
            strokeLinecap="round"
            strokeDasharray={2 * Math.PI * 130}
            strokeDashoffset={2 * Math.PI * 130 * (1 - progress)}
            style={{
              transition: 'stroke-dashoffset 1s linear',
              filter: `drop-shadow(0 0 6px ${color}80)`,
            }}
          />
        </svg>

        {/* Crosshair lines */}
        <svg viewBox="0 0 320 320" className="absolute inset-0 w-full h-full" style={{ animation: 'crosshairBlink 2s ease-in-out infinite' }}>
          {/* Horizontal lines (gap in center) */}
          <line x1="30" y1="160" x2="120" y2="160" stroke={color} strokeWidth="1" opacity="0.5" />
          <line x1="200" y1="160" x2="290" y2="160" stroke={color} strokeWidth="1" opacity="0.5" />
          {/* Vertical lines (gap in center) */}
          <line x1="160" y1="30" x2="160" y2="120" stroke={color} strokeWidth="1" opacity="0.5" />
          <line x1="160" y1="200" x2="160" y2="290" stroke={color} strokeWidth="1" opacity="0.5" />

          {/* Small center cross */}
          <line x1="150" y1="160" x2="170" y2="160" stroke={color} strokeWidth="2" opacity="0.8" />
          <line x1="160" y1="150" x2="160" y2="170" stroke={color} strokeWidth="2" opacity="0.8" />

          {/* Tick marks at cardinal points */}
          <line x1="160" y1="30" x2="160" y2="40" stroke={color} strokeWidth="2" opacity="0.4" />
          <line x1="160" y1="280" x2="160" y2="290" stroke={color} strokeWidth="2" opacity="0.4" />
          <line x1="30" y1="160" x2="40" y2="160" stroke={color} strokeWidth="2" opacity="0.4" />
          <line x1="280" y1="160" x2="290" y2="160" stroke={color} strokeWidth="2" opacity="0.4" />

          {/* Diamond markers at 45deg increments */}
          {[45, 135, 225, 315].map(deg => {
            const rad = (deg * Math.PI) / 180;
            const cx = 160 + 110 * Math.cos(rad);
            const cy = 160 + 110 * Math.sin(rad);
            return (
              <rect key={deg} x={cx - 3} y={cy - 3} width={6} height={6}
                    fill={color} opacity="0.3"
                    transform={`rotate(45, ${cx}, ${cy})`} />
            );
          })}
        </svg>

        {/* Inner scope circle */}
        <svg viewBox="0 0 320 320" className="absolute inset-0 w-full h-full">
          <circle
            cx="160" cy="160" r="80"
            fill="none"
            stroke={color}
            strokeWidth="1"
            opacity="0.15"
          />
          <circle
            cx="160" cy="160" r="50"
            fill="none"
            stroke={color}
            strokeWidth="1"
            strokeDasharray="4 8"
            opacity="0.1"
          />
        </svg>

        {/* Number display */}
        <div
          key={animKey}
          className="absolute inset-0 flex flex-col items-center justify-center"
          style={{ animation: 'numberDrop 0.5s cubic-bezier(0.34, 1.56, 0.64, 1)' }}
        >
          <div
            className="font-heading font-bold leading-none"
            style={{
              fontSize: count > 0 ? '8rem' : '4.5rem',
              color: count > 0 ? color : '#00d4ff',
              textShadow: count > 0
                ? `0 0 40px ${color}50, 0 0 80px ${color}25, 0 0 120px ${color}10`
                : '0 0 40px rgba(0,212,255,0.5)',
            }}
          >
            {count > 0 ? count : t('countdown.go')}
          </div>
          {count > 0 && (
            <div
              className="text-xs font-heading uppercase mt-4 font-bold"
              style={{
                color: 'rgba(0, 255, 136, 0.5)',
                animation: 'lockOnPulse 1.5s ease-in-out infinite',
                letterSpacing: '0.4em',
              }}
            >
              {t('countdown.lockOn')}
            </div>
          )}
        </div>
      </div>

      {/* Corner brackets */}
      <div className="absolute" style={{ width: '360px', height: '360px' }}>
        <div className="absolute top-0 left-0 w-8 h-8 border-t-2 border-l-2" style={{ borderColor: `${color}30` }} />
        <div className="absolute top-0 right-0 w-8 h-8 border-t-2 border-r-2" style={{ borderColor: `${color}30` }} />
        <div className="absolute bottom-0 left-0 w-8 h-8 border-b-2 border-l-2" style={{ borderColor: `${color}30` }} />
        <div className="absolute bottom-0 right-0 w-8 h-8 border-b-2 border-r-2" style={{ borderColor: `${color}30` }} />
      </div>

      {/* Scanline effect */}
      <div className="absolute inset-0 pointer-events-none overflow-hidden opacity-[0.02]">
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
