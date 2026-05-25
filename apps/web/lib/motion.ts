import type { Transition, Variants } from 'framer-motion';

export const SPRING_SNAPPY: Transition = { type: 'spring', stiffness: 300, damping: 20 };
export const SPRING_GENTLE: Transition = { type: 'spring', stiffness: 200, damping: 25 };
export const SPRING_BOUNCY: Transition = { type: 'spring', stiffness: 200, damping: 15 };

export const fadeUp: Variants = {
  hidden: { opacity: 0, y: 20 },
  visible: { opacity: 1, y: 0 },
};

export const fadeLeft: Variants = {
  hidden: { opacity: 0, x: -20 },
  visible: { opacity: 1, x: 0 },
};

export const scaleIn: Variants = {
  hidden: { scale: 0, opacity: 0 },
  visible: { scale: 1, opacity: 1 },
};

export const staggerContainer: Variants = {
  hidden: {},
  visible: { transition: { staggerChildren: 0.1 } },
};

export const fadeInUp: Variants = {
  hidden: { opacity: 0, y: 40 },
  visible: { opacity: 1, y: 0 },
};

export const fadeOut: Variants = {
  visible: { opacity: 1 },
  hidden: { opacity: 0.3 },
};