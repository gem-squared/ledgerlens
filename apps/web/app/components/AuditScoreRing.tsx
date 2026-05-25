import { scoreTone } from '@/lib/reasons';

interface AuditScoreRingProps {
  score: number;          // 0-100; pass NaN or negative for "skipped"
  label: string;          // e.g. "L1 P-check"
  verdict?: string;       // ALLOW / DENY / SUCCESS / FAILURE / "skipped"
  size?: number;          // px (default 120)
  subtle?: boolean;       // dim the ring (used for skipped gates)
}

const TONE_COLOR: Record<'good' | 'mid' | 'bad', string> = {
  good: 'rgb(16, 185, 129)',  // emerald-500
  mid:  'rgb(245, 158, 11)',  // amber-500
  bad:  'rgb(239, 68, 68)',   // red-500
};

export function AuditScoreRing({
  score,
  label,
  verdict,
  size = 120,
  subtle = false,
}: AuditScoreRingProps) {
  const valid = Number.isFinite(score) && score >= 0;
  const clamped = valid ? Math.max(0, Math.min(100, score)) : 0;
  const tone = valid ? scoreTone(clamped) : 'bad';
  const strokeColor = subtle ? 'rgb(82, 82, 91)' : TONE_COLOR[tone];

  const stroke = 10;
  const r = (size - stroke) / 2;
  const c = size / 2;
  const circumference = 2 * Math.PI * r;
  const dash = (clamped / 100) * circumference;

  return (
    <div className="inline-flex flex-col items-center">
      <div className="relative" style={{ width: size, height: size }}>
        <svg width={size} height={size} className="-rotate-90 transform">
          {/* track */}
          <circle
            cx={c}
            cy={c}
            r={r}
            fill="none"
            stroke="rgb(39, 39, 42)"
            strokeWidth={stroke}
          />
          {/* progress */}
          {valid && (
            <circle
              cx={c}
              cy={c}
              r={r}
              fill="none"
              stroke={strokeColor}
              strokeWidth={stroke}
              strokeDasharray={`${dash} ${circumference - dash}`}
              strokeLinecap="round"
              style={{ transition: 'stroke-dasharray 600ms ease-out' }}
            />
          )}
        </svg>
        <div className="absolute inset-0 flex flex-col items-center justify-center">
          {valid ? (
            <>
              <span className="text-3xl font-bold tracking-tight text-zinc-100">
                {clamped}
              </span>
              <span className="text-[10px] uppercase tracking-wider text-zinc-500">
                / 100
              </span>
            </>
          ) : (
            <span className="text-xs uppercase tracking-wider text-zinc-600">
              skipped
            </span>
          )}
        </div>
      </div>
      <div className="mt-2 text-center">
        <div className="text-xs font-semibold uppercase tracking-wider text-zinc-300">
          {label}
        </div>
        {verdict && (
          <div
            className={`mt-0.5 text-[11px] font-mono uppercase ${
              verdict === 'ALLOW' || verdict === 'SUCCESS'
                ? 'text-emerald-300'
                : verdict === 'DENY' || verdict === 'FAILURE'
                  ? 'text-red-300'
                  : 'text-zinc-500'
            }`}
          >
            {verdict}
          </div>
        )}
      </div>
    </div>
  );
}
