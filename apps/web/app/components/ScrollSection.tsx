'use client';

import { useRef } from 'react';
import { motion, useScroll, useTransform, useInView } from 'framer-motion';

interface ScrollSectionProps {
  id: string;
  children: React.ReactNode;
  speed?: number;
  className?: string;
}

export function ScrollSection({ id, children, speed = 1.0, className = '' }: ScrollSectionProps) {
  const ref = useRef<HTMLDivElement>(null);
  const inView = useInView(ref, { once: true, margin: '-50px' });
  const { scrollYProgress } = useScroll({ target: ref, offset: ['start end', 'end start'] });
  const speedOffset = (speed - 1) * 150;
  const y = useTransform(scrollYProgress, [0, 1], [0, speedOffset]);

  return (
    <section
      id={id}
      className={`relative py-16 md:py-24 ${className}`}
      aria-label={id.replace(/-/g, ' ')}
    >
      <motion.div
        ref={ref}
        style={{ y }}
        className={`transition-all duration-700 ease-out ${
          inView ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-10'
        }`}
      >
        {children}
      </motion.div>
    </section>
  );
}
