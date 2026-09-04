package agent

import (
	"context"
	"strings"
	"time"
)

// Mock is a placeholder AgentService so the end-to-end chat workflow
// works before the real agent is implemented. It returns believable,
// migration-themed canned replies based on simple keyword matching.
type Mock struct{}

func NewMock() *Mock { return &Mock{} }

func (m *Mock) Run(ctx context.Context, req Request) (<-chan Event, error) {
	ch := make(chan Event, 4)
	go func() {
		defer close(ch)

		reply := mockReplyFor(req.Message)

		select {
		case <-ctx.Done():
			return
		case ch <- Event{Type: EventThinking, Data: map[string]string{"content": "Analyzing Zscaler configuration..."}}:
		}

		// Simulate a brief tool call so the demo feels alive.
		select {
		case <-ctx.Done():
			return
		case ch <- Event{Type: EventToolStart, Data: map[string]string{"tool": "zscaler_config_parser"}}:
		}
		time.Sleep(300 * time.Millisecond)
		select {
		case <-ctx.Done():
			return
		case ch <- Event{Type: EventToolResult, Data: map[string]string{"tool": "zscaler_config_parser", "status": "ok"}}:
		}

		select {
		case <-ctx.Done():
			return
		case ch <- Event{Type: EventMessage, Data: map[string]string{"content": reply}}:
		}

		select {
		case <-ctx.Done():
			return
		case ch <- Event{Type: EventDone}:
		}
	}()
	return ch, nil
}

func mockReplyFor(message string) string {
	msg := strings.ToLower(strings.TrimSpace(message))
	if msg == "" {
		return "I didn't catch a question. Could you rephrase what you'd like me to analyze in the Zscaler configuration?"
	}

	switch {
	case containsAny(msg, "summary", "summarize", "overview", "summarise"):
		return "Here's a preliminary summary of the uploaded Zscaler configuration:\n\n" +
			"- **Identity**: one primary ZIA tenant with Cloud Firewall and ZPA modules enabled.\n" +
			"- **Policies**: 42 access policies, 8 firewall rules, 12 DNS filtering rules.\n" +
			"- **Locations**: 3 sub-locations configured with bandwidth controls.\n" +
			"- **Notable**: several legacy IP-based rules that may need re-evaluation for the target platform.\n\n" +
			"This is a mock analysis for demonstration purposes — the real agent will parse the actual config."

	case containsAny(msg, "nat", "network address translation"):
		return "I found 6 NAT rules in the configuration:\n\n" +
			"| # | Source | Translated | Interface |\n|---|--------|------------|-----------|\n" +
			"| 1 | 10.0.0.0/24 | 203.0.113.10 | outside |\n" +
			"| 2 | 10.0.1.0/24 | 203.0.113.11 | outside |\n" +
			"| 3 | 10.0.2.0/24 | 203.0.113.12 | outside |\n\n" +
			"Two rules reference deprecated public IPs that should be reviewed before migration.\n\n" +
			"_Note: this is mock output for the POC._"

	case containsAny(msg, "policy", "policies", "rule", "rules", "firewall"):
		return "I analyzed the firewall and access policies. Key observations:\n\n" +
			"- **3 policies** use `any` as the destination — tighten these before migrating.\n" +
			"- **2 policies** overlap with broader rules above them (shadowed rules).\n" +
			"- **1 policy** references a disabled application group.\n" +
			"- DNS filtering rules look clean and map cleanly to the target platform.\n\n" +
			"Would you like me to list the shadowed rules?\n\n" +
			"_Note: this is mock output for the POC._"

	case containsAny(msg, "issue", "issues", "problem", "problems", "risk", "risks", "concern"):
		return "Potential migration risks identified:\n\n" +
			"1. **Legacy IPsec VPN** — the config references a legacy IPsec tunnel that the target platform deprecates. Plan a replacement (e.g., ZTNA).\n" +
			"2. **Hardcoded IPs** — 4 policies use hardcoded public IPs instead of FQDN groups.\n" +
			"3. **Broad allow rules** — 3 `any`-destination rules exceed the target's least-privilege baseline.\n" +
			"4. **Logging gaps** — 2 firewall rules have logging disabled.\n\n" +
			"_Note: this is mock output for the POC._"

	case containsAny(msg, "vpn", "ipsec", "tunnel", "ztna"):
		return "VPN/IPsec review:\n\n" +
			"- 1 active IPsec tunnel to a branch site (legacy, phase-1 IKEv1).\n" +
			"- Recommend migrating to ZTNA or IKEv2 before cutover.\n" +
			"- No clientless VPN users detected — good candidate for ZTNA replacement.\n\n" +
			"_Note: this is mock output for the POC._"

	case containsAny(msg, "dns", "domain", "fqdn"):
		return "DNS/FQDN review:\n\n" +
			"- 12 DNS filtering rules, all using Zscaler's cloud DNS resolver.\n" +
			"- 3 FQDN groups referenced by access policies.\n" +
			"- No conflicting DNS sinkhole rules detected.\n\n" +
			"_Note: this is mock output for the POC._"

	case containsAny(msg, "convert", "conversion", "translate", "transform", "target"):
		return "Conversion readiness:\n\n" +
			"- ~85% of policies map directly to the target platform's object model.\n" +
			"- The remaining 15% need manual review (mostly legacy IP-based rules).\n" +
			"- Estimated effort: 2–3 hours of rule-by-rule review.\n\n" +
			"Run `Convert configuration` when you're ready for a dry-run export.\n\n" +
			"_Note: this is mock output for the POC._"

	case containsAny(msg, "hello", "hi", "hey", "help"):
		return "Hi! I'm the Zscaler migration assistant. I can help you:\n\n" +
			"- Summarize the uploaded configuration\n" +
			"- Identify migration risks and shadowed rules\n" +
			"- Review NAT, VPN, DNS, and firewall policies\n" +
			"- Assess conversion readiness\n\n" +
			"What would you like to look at first?\n\n" +
			"_Note: this is mock output for the POC._"
	}

	return "I've reviewed the configuration in the context of your question: \"" + strings.TrimSpace(message) + "\".\n\n" +
		"Preliminary findings:\n" +
		"- The configuration parses cleanly with no structural errors.\n" +
		"- A few rules warrant review before migration (legacy IPs, broad allows).\n" +
		"- Overall the config is in good shape for a migration dry-run.\n\n" +
		"Try asking me to \"summarize the configuration\", \"list NAT rules\", or \"what issues do you see?\".\n\n" +
		"_Note: this is mock output for the POC._"
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
