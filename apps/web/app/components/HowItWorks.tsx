'use client';

import * as m from 'framer-motion/m';
import { fadeUp, staggerContainer, SPRING_GENTLE } from '@/lib/motion';

const STEPS = [
  { icon: '👨‍⚖️', title: 'Judge Request', desc: 'Asks for trusted web data' },
  { icon: '🤖', title: 'Buyer Agent', desc: 'Extracts intent and spend policy' },
  { icon: '🌐', title: 'Bright Data', desc: 'Fetches public-web evidence' },
  { icon: '📊', title: 'Seller Offer', desc: 'Forms candidate data deal' },
  { icon: '🛡️', title: 'GEM² Gate', desc: 'Audits if claims are grounded' },
  { icon: '💳', title: 'x402 Settle', desc: 'Approves or blocks payment' },
  { icon: '📁', title: 'Audit Bundle', desc: 'Exports replayable record' },
];

export function HowItWorks() {
  return (
    <m.div
      className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 xl:grid-cols-7 gap-4"
      initial="hidden"
      whileInView="visible"
      viewport={{ once: true, amount: 0.2 }}
      variants={staggerContainer}
    >
      {STEPS.map((step, i) => (
        <m.div
          key={step.title}
          className="glass-panel flex flex-col items-center text-center group cursor-default"
          variants={fadeUp}
          transition={SPRING_GENTLE}
        >
          <span className="text-3xl mb-3 transition-transform duration-200 group-hover:scale-110">
            {step.icon}
          </span>
          <h3 className="text-xs font-semibold uppercase tracking-wider text-zinc-200 mb-1">
            {step.title}
          </h3>
          <p className="text-[11px] leading-relaxed text-zinc-400">{step.desc}</p>
          {i < STEPS.length - 1 && (
            <span className="hidden xl:block absolute -right-3 top-1/2 text-zinc-600 text-lg" aria-hidden="true">
              →
            </span>
          )}
        </m.div>
      ))}
    </m.div>
  );
}
