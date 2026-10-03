"use client";

const integrations = [
  {
    name: "Grafana / Loki",
    category: "Observability",
    status: "done" as const,
    note: "Self-hosted Grafana, Grafana Cloud, or Loki direct. Grafana token with Viewer permissions — no admin access needed.",
    required: true,
  },
  {
    name: "GitHub",
    category: "Source control",
    status: "done" as const,
    note: "gh CLI auth or token. Agent opens PRs and reads workflow status. Requires repo write access only.",
    required: true,
  },
  {
    name: "Anthropic (Claude)",
    category: "AI",
    status: "done" as const,
    note: "API key for Claude. Root-cause analysis and fix generation. Any model supported — defaults to claude-sonnet.",
    required: true,
  },
];

export function DataSources() {
  return (
    <section id="data-sources" className="py-32 px-6" style={{ backgroundColor: "#fdfcf8" }}>
      <div className="max-w-6xl mx-auto">
        <div className="mb-12 max-w-3xl">
          <p
            className="text-xs font-semibold uppercase tracking-[0.2em] mb-4 font-mono"
            style={{ color: "#735c41" }}
          >
            Integrations
          </p>
          <h2 className="text-3xl sm:text-4xl md:text-5xl leading-tight mb-4" style={{ color: "#1b1c19" }}>
            Three things. That's all you need.
          </h2>
          <p className="text-base sm:text-lg leading-relaxed" style={{ color: "#6b665e" }}>
            The agent needs a log source, a code repo, and an AI key. Nothing else.
            Docker, Kubernetes, VMs, bare metal — if logs flow into Loki, the agent sees them.
          </p>
        </div>

        <div
          className="rounded-2xl overflow-hidden mb-16"
          style={{ border: "1px solid #e0d9cd", backgroundColor: "#ffffff" }}
        >
          <div
            className="grid grid-cols-12 gap-4 px-6 py-4 text-xs font-mono uppercase tracking-[0.15em]"
            style={{ borderBottom: "1px solid #e0d9cd", color: "#735c41", backgroundColor: "#f7f4ed" }}
          >
            <div className="col-span-5 sm:col-span-4">Integration</div>
            <div className="col-span-5 sm:col-span-5">Note</div>
            <div className="col-span-2 text-right">Required</div>
          </div>

          {integrations.map((s, idx) => (
            <div
              key={s.name}
              className="grid grid-cols-12 gap-4 px-6 py-5 items-center"
              style={{
                borderBottom: idx === integrations.length - 1 ? "none" : "1px solid #f0ede6",
              }}
            >
              <div className="col-span-12 sm:col-span-4">
                <p className="text-base font-semibold" style={{ color: "#1b1c19" }}>
                  {s.name}
                </p>
              </div>
              <div className="col-span-7 sm:col-span-5 text-sm" style={{ color: "#6b665e" }}>
                {s.note}
              </div>
              <div className="col-span-5 sm:col-span-3 flex sm:justify-end">
                {s.required ? (
                  <span className="text-xs font-mono" style={{ color: "#735c41" }}>Required</span>
                ) : (
                  <span className="text-xs" style={{ color: "#8b867f" }}>Optional</span>
                )}
              </div>
            </div>
          ))}
        </div>

        {/* Security section */}
        <div className="max-w-3xl mb-16">
          <p
            className="text-xs font-semibold uppercase tracking-[0.2em] mb-4 font-mono"
            style={{ color: "#735c41" }}
          >
            Security
          </p>
          <h3 className="text-xl sm:text-2xl font-bold mb-4" style={{ color: "#1b1c19" }}>
            All secrets stay on your machine.
          </h3>
          <p className="text-sm leading-relaxed mb-6" style={{ color: "#6b665e" }}>
            The agent never sends credentials, tokens, or log data to any external service
            except the APIs you explicitly configure.
          </p>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {[
              {
                title: "Grafana token",
                body: "Stored in OS keychain (macOS) or encrypted file (Linux). Never written to config. Config stores the Grafana URL only — token is resolved at runtime.",
              },
              {
                title: "GitHub token",
                body: "Resolved from gh CLI, env var, or keychain — in that order. Agent never persists it. Config only stores the repo name and base branch.",
              },
              {
                title: "AI API key",
                body: "Resolved from env var or keychain. Sent only to the AI provider you configured (Anthropic, OpenAI, etc.). Never logged or cached.",
              },
              {
                title: "Log data",
                body: "Error patterns are sent to the AI provider for analysis. No raw logs, no PII, no request bodies — just the error message text.",
              },
            ].map((item) => (
              <div
                key={item.title}
                className="p-5 rounded-xl"
                style={{ backgroundColor: "#f7f4ed", border: "1px solid #e8e3d9" }}
              >
                <p className="text-sm font-semibold mb-2" style={{ color: "#1b1c19" }}>
                  {item.title}
                </p>
                <p className="text-xs leading-relaxed" style={{ color: "#6b665e" }}>
                  {item.body}
                </p>
              </div>
            ))}
          </div>
        </div>

        {/* Testing section */}
        <div>
          <p
            className="text-xs font-semibold uppercase tracking-[0.2em] mb-4 font-mono"
            style={{ color: "#735c41" }}
          >
            Testing
          </p>
          <h3 className="text-xl sm:text-2xl font-bold mb-4" style={{ color: "#1b1c19" }}>
            Verify it works before you trust it.
          </h3>
          <p className="text-sm leading-relaxed mb-8" style={{ color: "#6b665e" }}>
            Every command supports dry-run mode. No PRs are created, no deploys triggered.
            Test the full pipeline in under a minute.
          </p>

          <div
            className="rounded-xl overflow-hidden"
            style={{ backgroundColor: "#0f1117", border: "1px solid rgba(255,255,255,0.08)" }}
          >
            <div
              className="flex items-center gap-2 px-4 py-3"
              style={{ borderBottom: "1px solid rgba(255,255,255,0.06)" }}
            >
              <span className="w-3 h-3 rounded-full" style={{ backgroundColor: "#ff5f57" }} />
              <span className="w-3 h-3 rounded-full" style={{ backgroundColor: "#febc2e" }} />
              <span className="w-3 h-3 rounded-full" style={{ backgroundColor: "#28c840" }} />
              <span className="ml-3 text-xs font-mono" style={{ color: "rgba(255,255,255,0.25)" }}>
                terminal
              </span>
            </div>
            <pre className="p-5 text-sm font-mono leading-7 overflow-x-auto" style={{ color: "#c9d1d9" }}>
              <span style={{ color: "#735c41" }}>$</span> sre-agent status{'\n'}
              <span style={{ color: "#8b949e" }}>
                {'›'} Observability: grafana_self_hosted @ https://observability.yourco.com{'\n'}
                {'›'} GitHub: yourco/product (base: main){'\n'}
                {'›'} AI: anthropic / claude-sonnet-4-20250514{'\n'}
                {'›'} Credentials: github ✓  anthropic ✓  grafana ✓{'\n\n'}
              </span>
              <span style={{ color: "#735c41" }}>$</span> sre-agent fix --dry-run{'\n'}
              <span style={{ color: "#8b949e" }}>
                {'›'} DRY RUN — no PR created{'\n'}
                {'›'} Would open PR: revert token validation path{'\n'}
                {'›'} Tests: passed{'\n'}
              </span>
            </pre>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mt-6">
            {[
              { label: "1. status", body: "Verify config + credentials + backend connectivity." },
              { label: "2. detect", body: "Query logs for errors. No side effects." },
              { label: "3. analyze", body: "Send to AI for root-cause. No PR created." },
            ].map((step) => (
              <div key={step.label} className="p-4 rounded-xl" style={{ border: "1px solid #e0d9cd" }}>
                <p className="text-sm font-mono mb-1" style={{ color: "#735c41" }}>{step.label}</p>
                <p className="text-xs" style={{ color: "#6b665e" }}>{step.body}</p>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}
