// Package paymentgate implements LedgerLens's x402 protocol SIMULATION mode
// settlement layer and orchestrates the end-to-end Evidence → Memory → L1 →
// F → L2 → L3 → Settle pipeline.
//
// ⚠ SIMULATION ONLY (v2.2 lock — see Docs/Bright-Data-winning-strategy.md §8):
// the hackathon repo ships exactly one Settler implementation — SimulatedSettler.
// No private keys, no Coinbase account, no Base Sepolia dependency, no tx_hash.
// The schemas.Settler interface (internal/schemas/types.go) is the swap point
// for a post-hackathon CoinbaseX402Settler.
package paymentgate
