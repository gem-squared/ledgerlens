// TS mirror of internal/trustgate/auditgate/reasons.go.
// Parses the audit gate's flat `reasons: string[]` into a structured shape
// so the UI can render a hierarchical disclosure: ring → dimensions → rules.

export interface DimensionScore {
  name: string;
  score: number;
  note: string;
}

export interface RuleFinding {
  index: number;
  rule: string;
  verdict: 'PASS' | 'FAIL';
  reason: string;
}

export interface EEFFlag {
  tag: '⊢' | '⊨' | '⊬' | '⊥';
  text: string;
}

export interface EvidenceUse {
  index: number | null;
  excerpt: string;
  ruleRefs: number[];
}

export interface ParsedReasons {
  typeFinding: string;
  rules: RuleFinding[];
  sptCodes: string[];
  eefFlags: EEFFlag[];
  evidence: EvidenceUse[];
  dimensions: DimensionScore[];
}

const RX_TYPE     = /^\[TYPE\]\s+(.+)$/;
const RX_RULE     = /^\[RULE-(\d+)\]\s+(.+?)\s*[—-]\s*(PASS|FAIL):\s*(.+)$/;
const RX_SPT      = /^\[SPT-(S→T|L→G|Δe→∫de)\]\s+(.+)$/;
const RX_EEF      = /^\[EEF-([⊢⊨⊬⊥])\]\s+(.+)$/;
const RX_EVIDENCE = /^\[EVIDENCE-(\d+|⊥)\]\s+(.+)$/;
const RX_DIM      = /^\[DIM-([A-Za-z_-]+)\]\s+(\d+)\/100\s*(?:[—:-]\s*(.+))?$/;
const RX_RULE_REF = /RULE-(\d+)/g;

export function parseReasons(reasons: string[]): ParsedReasons {
  const out: ParsedReasons = {
    typeFinding: '',
    rules: [],
    sptCodes: [],
    eefFlags: [],
    evidence: [],
    dimensions: [],
  };
  for (const raw of reasons ?? []) {
    const line = raw.trim();
    let m: RegExpMatchArray | null;
    if ((m = line.match(RX_TYPE))) {
      out.typeFinding = m[1];
    } else if ((m = line.match(RX_RULE))) {
      out.rules.push({
        index: parseInt(m[1], 10),
        rule: m[2].trim(),
        verdict: m[3] as 'PASS' | 'FAIL',
        reason: m[4].trim(),
      });
    } else if ((m = line.match(RX_SPT))) {
      out.sptCodes.push(m[1]);
    } else if ((m = line.match(RX_EEF))) {
      out.eefFlags.push({ tag: m[1] as EEFFlag['tag'], text: m[2] });
    } else if ((m = line.match(RX_EVIDENCE))) {
      const refs: number[] = [];
      let r: RegExpExecArray | null;
      const rx = new RegExp(RX_RULE_REF.source, 'g');
      while ((r = rx.exec(m[2])) !== null) refs.push(parseInt(r[1], 10));
      out.evidence.push({
        index: m[1] === '⊥' ? null : parseInt(m[1], 10),
        excerpt: m[2],
        ruleRefs: refs,
      });
    } else if ((m = line.match(RX_DIM))) {
      out.dimensions.push({
        name: m[1],
        score: parseInt(m[2], 10),
        note: m[3] ?? '',
      });
    }
  }
  return out;
}

export function dimensionLabel(name: string): string {
  return name.charAt(0).toUpperCase() + name.slice(1).replace(/[-_]/g, ' ');
}

export function scoreTone(score: number): 'good' | 'mid' | 'bad' {
  if (score >= 75) return 'good';
  if (score >= 40) return 'mid';
  return 'bad';
}
