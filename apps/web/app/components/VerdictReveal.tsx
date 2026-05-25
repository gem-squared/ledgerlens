'use client';

import * as m from 'framer-motion/m';
import { SPRING_BOUNCY } from '@/lib/motion';

const particles = Array.from({ length: 8 }, (_, i) => {
  const angle = (i / 8) * Math.PI * 2;
  return { x: Math.cos(angle) * 60, y: Math.sin(angle) * 60 };
});

interface VerdictRevealProps {
  verdict: 'approved' | 'blocked';
}

export function VerdictReveal({ verdict }: VerdictRevealProps) {
  const isApproved = verdict === 'approved';
  const color = isApproved ? '#10b981' : '#ef4444';

  return (
    <div className="relative flex flex-col items-center justify-center py-10">
      {/* Radial glow */}
      <m.div
        className="absolute w-32 h-32 rounded-full"
        style={{ background: `radial-gradient(circle, ${color}40, transparent 70%)` }}
        initial={{ scale: 0, opacity: 0 }}
        animate={{ scale: 1.5, opacity: 1 }}
        transition={{ duration: 0.5, ease: 'easeOut' }}
      />

      {/* Particles (approved only) */}
      {isApproved && particles.map((p, i) => (
        <m.div
          key={i}
          className="absolute w-1.5 h-1.5 rounded-full bg-emerald-400"
          initial={{ x: 0, y: 0, opacity: 1, scale: 1 }}
          animate={{ x: p.x, y: p.y, opacity: 0, scale: 0 }}
          transition={{ delay: 0.3 + i * 0.03, duration: 0.7, ease: 'easeOut' }}
        />
      ))}

      {/* Icon */}
      <m.div
        className="relative z-10 text-5xl"
        initial={isApproved ? { scale: 0, opacity: 0 } : { scale: 2, opacity: 0, x: 0 }}
        animate={isApproved ? { scale: 1, opacity: 1 } : { scale: 1, opacity: 1, x: [-4, 4, -3, 3, 0] }}
        transition={isApproved ? SPRING_BOUNCY : { duration: 0.4 }}
      >
        {isApproved ? '✅' : '❌'}
      </m.div>

      {/* Verdict text */}
      <m.p
        className="relative z-10 mt-3 text-lg font-bold uppercase tracking-wider"
        style={{ color }}
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.6, duration: 0.3 }}
      >
        {isApproved ? 'Payment Approved' : 'Payment Blocked'}
      </m.p>
    </div>
  );
}
