'use client';

import { useRef } from 'react';
import * as m from 'framer-motion/m';
import { useScroll, useTransform } from 'framer-motion';
import { fadeInUp, SPRING_GENTLE } from '@/lib/motion';

interface ScrollSectionProps {
  id: string;
  children: React.ReactNode;
  speed?: number;
  className?: string;
}

export function ScrollSection({ id, children, speed = 1.0, className = '' }: ScrollSectionProps) {
  const ref = useRef<HTMLElement>(null);
  const { scrollYProgress } = useScroll({ target: ref, offset: ['start end', 'end start'] });
  const speedOffset = (speed - 1) * 100;
  const y = useTransform(scrollYProgress, [0, 1], [0, speedOffset]);

  return (
    <m.section
      ref={ref}
      id={id}
      className={`relative py-16 md:py-24 ${className}`}
      initial="hidden"
      whileInView="visible"
      viewport={{ once: false, amount: 0.3 }}
      variants={fadeInUp}
      transition={{ duration: 0.6, ease: 'easeOut' }}
      style={{ y: y as unknown as number }}
      aria-label={id.replace(/-/g, ' ')}
    >
      {children}
    </m.section>
  );
}
