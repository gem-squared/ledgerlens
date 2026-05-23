import './globals.css';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'LedgerLens — No grounded claim, no payment.',
  description:
    'x402-native Agent-to-Agent Payments with a Bright Data web-evidence layer and a GEM² Trust Gate before settlement. Simulation mode for public demo safety.',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="min-h-screen antialiased">{children}</body>
    </html>
  );
}
