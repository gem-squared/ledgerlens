import type { Config } from 'tailwindcss';

const config: Config = {
  content: ['./app/**/*.{ts,tsx}', './components/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        gateApproved: '#10b981', // green-500
        gateBlocked: '#ef4444',  // red-500
        gateEscalated: '#f59e0b',// amber-500
        simBadge: '#6366f1',     // indigo-500 — SIMULATION MODE pill
      },
    },
  },
  plugins: [],
};

export default config;
