'use client';

import { useEffect, useState } from 'react';

const NAV_ITEMS = [
  { id: 'how-it-works', label: 'How It Works' },
  { id: 'evidence', label: 'Evidence & Audit' },
  { id: 'settlement', label: 'Settlement' },
] as const;

export function StickyNav() {
  const [visible, setVisible] = useState(false);
  const [active, setActive] = useState('');

  useEffect(() => {
    const sections = NAV_ITEMS.map(n => document.getElementById(n.id)).filter(Boolean) as HTMLElement[];
    const hero = document.getElementById('hero');

    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.target === hero) {
            setVisible(!entry.isIntersecting);
          } else if (entry.isIntersecting) {
            setActive(entry.target.id);
          }
        }
      },
      { threshold: 0.3 }
    );

    if (hero) observer.observe(hero);
    sections.forEach(s => observer.observe(s));
    return () => observer.disconnect();
  }, []);

  function scrollTo(id: string) {
    document.getElementById(id)?.scrollIntoView({ behavior: 'smooth' });
  }

  return (
    <nav
      role="navigation"
      aria-label="Section navigation"
      className={`fixed top-0 left-0 right-0 z-50 transition-all duration-300 ${
        visible
          ? 'opacity-100 bg-white/[0.05] backdrop-blur-md border-b border-white/[0.08]'
          : 'opacity-0 pointer-events-none'
      }`}
    >
      <div className="mx-auto max-w-6xl flex items-center justify-center gap-1 px-4 py-3">
        {NAV_ITEMS.map((item) => (
          <button
            key={item.id}
            type="button"
            onClick={() => scrollTo(item.id)}
            className={`px-3 py-1.5 rounded-lg text-xs font-medium transition-all duration-200 ${
              active === item.id
                ? 'text-indigo-400 bg-indigo-500/10 border-b-2 border-indigo-400'
                : 'text-zinc-400 hover:text-zinc-200 hover:bg-white/[0.05]'
            }`}
          >
            {item.label}
          </button>
        ))}
      </div>
    </nav>
  );
}
