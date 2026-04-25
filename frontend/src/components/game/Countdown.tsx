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
  const color = '#4F46E5';

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center"
         style={{ background: 'radial-gradient(ellipse at center, rgba(250,251,254,0.98) 0%, #F0F4F8 100%)' }}>

      {/* Soft glow behind circle */}
      <div
        className="absolute rounded-full"
        style={{
          width: '350px',
          height: '350px',
          background: `radial-gradient(circle, rgba(79,70,229,0.08) 0%, transparent 70%)`,
          filter: 'blur(60px)',
        }}
      />

      {/* Circle Container */}
      <div className="relative" style={{ width: '280px', height: '280px' }}>

        {/* Progress ring */}
        <svg viewBox="0 0 280 280" className="absolute inset-0 w-full h-full -rotate-90">
          <circle
            cx="140" cy="140" r="120"
            fill="none"
            stroke="var(--color-border-default)"
            strokeWidth="3"
          />
          <circle
            cx="140" cy="140" r="120"
            fill="none"
            stroke={color}
            strokeWidth="3"
            strokeLinecap="round"
            strokeDasharray={2 * Math.PI * 120}
            strokeDashoffset={2 * Math.PI * 120 * (1 - progress)}
            style={{
              transition: 'stroke-dashoffset 1s linear',
            }}
          />
        </svg>

        {/* Number display */}
        <div
          key={animKey}
          className="absolute inset-0 flex flex-col items-center justify-center"
          style={{ animation: 'countdownPop 0.5s cubic-bezier(0.34, 1.56, 0.64, 1)' }}
        >
          <div
            className="font-heading font-bold leading-none"
            style={{
              fontSize: count > 0 ? '7rem' : '3.5rem',
              color: count > 0 ? color : 'var(--color-secondary)',
            }}
          >
            {count > 0 ? count : t('countdown.go')}
          </div>
          {count > 0 && (
            <div
              className="text-sm font-heading mt-4 font-medium text-[var(--color-text-muted)]"
              style={{ letterSpacing: '0.1em' }}
            >
              {t('countdown.lockOn')}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
