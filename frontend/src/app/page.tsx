'use client';

import Link from 'next/link';
import { useAuth } from '@/hooks/useAuth';
import { useEffect, useState } from 'react';
import { api } from '@/lib/api';
import { useLocale } from '@/i18n/LocaleProvider';

function BookIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20" /><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z" />
    </svg>
  );
}

function UsersIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" /><circle cx="9" cy="7" r="4" /><path d="M23 21v-2a4 4 0 0 0-3-3.87" /><path d="M16 3.13a4 4 0 0 1 0 7.75" />
    </svg>
  );
}

function TrophyIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M6 9H4.5a2.5 2.5 0 0 1 0-5H6" /><path d="M18 9h1.5a2.5 2.5 0 0 0 0-5H18" /><path d="M4 22h16" /><path d="M10 14.66V17c0 .55-.47.98-.97 1.21C7.85 18.75 7 20 7 22" /><path d="M14 14.66V17c0 .55.47.98.97 1.21C16.15 18.75 17 20 17 22" /><path d="M18 2H6v7a6 6 0 0 0 12 0V2Z" />
    </svg>
  );
}

export default function HomePage() {
  const { user } = useAuth();
  const { t } = useLocale();
  const [onlineCount, setOnlineCount] = useState(0);

  useEffect(() => {
    api.online().then(d => setOnlineCount(d.online)).catch(() => {});
  }, []);

  const MODES = [
    {
      id: 'solo',
      label: t('home.mode.solo.label'),
      color: '#4F46E5',
      desc: t('home.mode.solo.desc'),
      Icon: BookIcon,
    },
    {
      id: 'duel',
      label: t('home.mode.duel.label'),
      color: '#0D9488',
      desc: t('home.mode.duel.desc'),
      Icon: UsersIcon,
    },
    {
      id: 'battle',
      label: t('home.mode.battle.label'),
      color: '#EA580C',
      desc: t('home.mode.battle.desc'),
      Icon: TrophyIcon,
    },
  ];

  const QUIZ_TYPES = [
    { label: t('home.quiz.meaningToWord.label'), desc: t('home.quiz.meaningToWord.desc') },
    { label: t('home.quiz.wordToMeaning.label'), desc: t('home.quiz.wordToMeaning.desc') },
    { label: t('home.quiz.wordToIpa.label'), desc: t('home.quiz.wordToIpa.desc') },
    { label: t('home.quiz.wordToPinyin.label'), desc: t('home.quiz.wordToPinyin.desc') },
  ];

  const HOW_STEPS = [
    { num: '01', color: '#4F46E5', title: t('home.how.step1.title'), desc: t('home.how.step1.desc') },
    { num: '02', color: '#0D9488', title: t('home.how.step2.title'), desc: t('home.how.step2.desc') },
    { num: '03', color: '#0EA5E9', title: t('home.how.step3.title'), desc: t('home.how.step3.desc') },
  ];

  return (
    <div className="relative overflow-hidden">
      {/* ── Hero ─────────────────────────────────────────────── */}
      <section className="relative min-h-[100vh] flex flex-col items-center justify-center px-6 pt-20">
        {/* Soft gradient blobs */}
        <div className="bg-blob w-[500px] h-[500px] opacity-[0.12]"
             style={{ background: '#4F46E5', top: '-15%', left: '-10%' }} />
        <div className="bg-blob w-[400px] h-[400px] opacity-[0.08]"
             style={{ background: '#0D9488', bottom: '5%', right: '-10%', animationDelay: '-5s' }} />

        <div className="relative z-10 text-center max-w-3xl animate-fade-in-up">
          {/* Live badge */}
          <div className="inline-flex items-center gap-2 mb-8 badge" aria-label={`${onlineCount} learners online`}>
            <span className="w-2 h-2 rounded-full bg-[var(--color-secondary)] animate-pulse" aria-hidden="true" />
            {onlineCount > 0 ? t('home.badge.online', { count: onlineCount }) : t('home.badge.ready')}
          </div>

          {/* Headline */}
          <h1 className="font-heading font-bold leading-tight tracking-tight mb-6">
            <span className="block text-[clamp(3rem,8vw,5.5rem)] text-gradient-primary">LinguaLeap</span>
          </h1>

          <p className="text-base sm:text-lg text-[var(--color-text-secondary)] max-w-2xl mx-auto mb-12 leading-relaxed animate-fade-in-up delay-200">
            {t('home.subtitle')}
          </p>

          {/* CTA row */}
          <div className="flex flex-col sm:flex-row gap-4 justify-center items-center animate-fade-in-up delay-300">
            <Link href={user ? '/play' : '/login'} className="btn-primary text-base px-10 py-4">
              {user ? t('home.cta.play') : t('home.cta.start')}
            </Link>
            <Link href="/leaderboard" className="btn-secondary text-base px-10 py-4">
              {t('home.cta.leaderboard')}
            </Link>
          </div>

          {/* Mini stats row */}
          <div className="flex items-center justify-center gap-10 mt-14 animate-fade-in-up delay-400">
            {[['380+', t('home.stat.words')], ['2', t('home.stat.languages')], ['4', t('home.stat.quizModes')]].map(([val, label]) => (
              <div key={label} className="text-center">
                <div className="text-2xl font-heading font-bold stat-shimmer">{val}</div>
                <div className="text-xs font-heading text-[var(--color-text-muted)] mt-1">{label}</div>
              </div>
            ))}
          </div>
        </div>

        {/* Scroll hint */}
        <div className="absolute bottom-8 left-1/2 -translate-x-1/2 flex flex-col items-center gap-2 opacity-30" aria-hidden="true">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="var(--color-text-muted)" strokeWidth="2" strokeLinecap="round">
            <path d="M12 5v14M5 12l7 7 7-7" />
          </svg>
        </div>
      </section>

      {/* ── Study Modes ───────────────────────────────────────── */}
      <section className="py-24 px-6 border-t border-[var(--color-border-default)]" aria-labelledby="modes-heading">
        <div className="max-w-5xl mx-auto">
          <div className="text-center mb-16">
            <div className="badge mb-4">{t('home.modes.badge')}</div>
            <h2 id="modes-heading" className="font-heading font-bold text-3xl sm:text-4xl tracking-tight">
              {t('home.modes.title').replace('{accent}', '')} <span className="text-[var(--color-primary)]">{t('home.modes.titleAccent')}</span>
            </h2>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {MODES.map((mode) => (
              <div key={mode.id} className="card group cursor-pointer">
                <div className="w-12 h-12 flex items-center justify-center rounded-[var(--radius-md)] mb-4 transition-transform duration-200 group-hover:scale-110"
                     style={{ background: `${mode.color}12`, color: mode.color }}>
                  <mode.Icon />
                </div>
                <h3 className="font-heading font-bold text-lg mb-2" style={{ color: mode.color }}>
                  {mode.label}
                </h3>
                <p className="text-sm text-[var(--color-text-secondary)] leading-relaxed">{mode.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── How It Works ─────────────────────────────────────── */}
      <section className="py-24 px-6 bg-[var(--color-bg-secondary)]" aria-labelledby="how-heading">
        <div className="max-w-5xl mx-auto">
          <div className="text-center mb-16">
            <div className="badge mb-4">{t('home.how.badge')}</div>
            <h2 id="how-heading" className="font-heading font-bold text-3xl sm:text-4xl tracking-tight">
              {t('home.how.title').replace('{accent}', '')} <span className="text-[var(--color-primary)]">{t('home.how.titleAccent')}</span>
            </h2>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-8">
            {HOW_STEPS.map(({ num, color, title, desc }) => (
              <div key={num} className="card group">
                <div className="w-10 h-10 flex items-center justify-center rounded-full font-heading font-bold text-sm mb-5"
                     style={{ background: `${color}15`, color }}>
                  {num}
                </div>
                <h3 className="font-heading font-bold text-lg mb-2">{title}</h3>
                <p className="text-sm text-[var(--color-text-secondary)] leading-relaxed">{desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── Quiz Types ───────────────────────────────────────── */}
      <section className="py-24 px-6 border-t border-[var(--color-border-default)]" aria-labelledby="quiz-heading">
        <div className="max-w-4xl mx-auto">
          <div className="text-center mb-16">
            <div className="badge mb-4">{t('home.quiz.badge')}</div>
            <h2 id="quiz-heading" className="font-heading font-bold text-3xl sm:text-4xl tracking-tight">
              {t('home.quiz.title').replace('{accent}', '')} <span className="text-[var(--color-primary)]">{t('home.quiz.titleAccent')}</span>
            </h2>
            <p className="text-sm text-[var(--color-text-muted)] mt-4 max-w-xl mx-auto">
              {t('home.quiz.subtitle')}
            </p>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {QUIZ_TYPES.map(({ label, desc }, i) => (
              <div key={label} className="card flex items-start gap-4">
                <div className="w-8 h-8 flex-shrink-0 flex items-center justify-center rounded-full bg-[rgba(79,70,229,0.08)] text-[var(--color-primary)] font-heading font-bold text-xs">
                  {String(i + 1).padStart(2, '0')}
                </div>
                <div>
                  <div className="font-heading font-bold text-sm mb-1 text-[var(--color-text-primary)]">{label}</div>
                  <div className="text-sm text-[var(--color-text-secondary)]">{desc}</div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── Final CTA ────────────────────────────────────────── */}
      <section className="py-24 px-6 bg-[var(--color-bg-secondary)]">
        <div className="max-w-2xl mx-auto text-center">
          <h2 className="font-heading font-bold text-3xl sm:text-5xl tracking-tight mb-6">
            {t('home.cta2.title').replace('{accent}', '')} <span className="text-[var(--color-primary)]">{t('home.cta2.titleAccent')}</span>?
          </h2>
          <p className="text-[var(--color-text-secondary)] mb-10 leading-relaxed">
            {t('home.cta2.subtitle')}
          </p>
          <Link href={user ? '/play' : '/login'} className="btn-primary text-base px-12 py-4">
            {user ? t('home.cta2.play') : t('home.cta2.start')}
          </Link>
        </div>
      </section>

      {/* ── Footer ───────────────────────────────────────────── */}
      <footer className="py-8 px-6 border-t border-[var(--color-border-default)]">
        <div className="max-w-7xl mx-auto flex flex-col sm:flex-row items-center justify-between gap-4">
          <div className="font-heading font-bold text-sm text-[var(--color-text-muted)] tracking-tight">
            <span className="text-[var(--color-primary)]">Lingua</span>Leap
          </div>
          <p className="text-xs text-[var(--color-text-muted)]">
            {t('home.footer.builtWith')}
          </p>
          <div className="flex items-center gap-4 text-xs font-heading text-[var(--color-text-muted)]">
            <Link href="/play" className="hover:text-[var(--color-primary)] transition-colors">{t('home.footer.play')}</Link>
            <Link href="/leaderboard" className="hover:text-[var(--color-primary)] transition-colors">{t('home.footer.ranks')}</Link>
          </div>
        </div>
      </footer>
    </div>
  );
}
