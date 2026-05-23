// Package agent implements LedgerLens's Judge Request Mode — the buyer-agent
// + seller-offer-agent pipeline that runs upstream of the Trust Gate.
//
// Pipeline (Slice 1 of v2 upgrade):
//
//   ① intent.Extract(judgeQuery)      → BuyerIntent              (Anthropic Haiku 4.5)
//   ② brightdata.SERP(intent.SearchTerms) + Unlocker(top hits) → []EvidenceReceipt
//   ③ synthesize.Synthesize(intent, receipts) → SellerOffer      (Anthropic Haiku 4.5)
//   ④ orchestrator.Run(buyerReq, offer, receipts) → DecisionPacket + Settlement + Bundle
//   ⑤ report.Compose(...)              → FinalReport             (template-driven, no LLM)
//
// The "agents" are LLM pipelines + tool orchestration — they read public web
// evidence (Bright Data) and construct structured records (BuyerIntent, SellerOffer)
// that the rest of the system audits. This is the agentic-commerce thesis:
// autonomous decisions, but not autonomous spend — payment requires the gate.
//
// LIVE latency: 20-45 seconds per request. The wait is the product:
//
//	"Fast agents are dangerous if they spend before verification.
//	 LedgerLens deliberately waits."
package agent
