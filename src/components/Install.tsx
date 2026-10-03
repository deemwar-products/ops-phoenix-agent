"use client";

import { useState, useCallback } from "react";

function CopyButton({ text, dark }: { text: string; dark?: boolean }) {
  const [copied, setCopied] = useState(false);
  const copy = useCallback(async (t: string) => {
    try {
      await navigator.clipboard.writeText(t);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // clipboard unavailable
    }
  }, []);
  if (dark) {
    return (
      <button
        onClick={() => copy(text)}
        className="text-[10px] font-mono uppercase tracking-[0.1em] px-2.5 py-1 rounded-md transition-colors"
        style={{
          color: copied ? "#3fb950" : "rgba(255,255,255,0.25)",
          backgroundColor: copied ? "rgba(63,185,80,0.12)" : "transparent",
        }}
      >
        {copied ? "Copied" : "Copy"}
      </button>
    );
  }
  return (
    <button
      onClick={() => copy(text)}
      className="text-[10px] font-mono uppercase tracking-[0.1em] px-2.5 py-1 rounded-md transition-colors"
      style={{
        color: copied ? "#2f7d3a" : "#735c41",
        backgroundColor: copied ? "rgba(115,92,65,0.10)" : "transparent",
      }}
    >
      {copied ? "Copied" : "Copy"}
    </button>
  );
}

export function Install() {
  return (
    <section id="install" className="py-32 px-6" style={{ backgroundColor: "#fdfcf8" }}>
      <div className="max-w-6xl mx-auto">
        <div className="mb-10">
          <p
            className="text-xs font-semibold uppercase tracking-[0.2em] mb-4 font-mono"
            style={{ color: "#735c41" }}
          >
            Install
          </p>
          <h2 className="text-3xl sm:text-4xl md:text-5xl leading-tight mb-3" style={{ color: "#1b1c19" }}>
            One command. That&apos;s it.
          </h2>
          <p className="text-base" style={{ color: "#6b665e" }}>
            One binary. No runtime. Runs on macOS, Linux, Windows — inside a container,
            on a VM, or on your laptop.
          </p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 items-stretch">
          <div>
            <div
              className="rounded-xl overflow-hidden mb-4"
              style={{ backgroundColor: "#0f1117", border: "1px solid rgba(255,255,255,0.08)" }}
            >
              <div
                className="flex items-center justify-between px-4 py-3"
                style={{ borderBottom: "1px solid rgba(255,255,255,0.06)" }}
              >
                <div className="flex items-center gap-2">
                  <span className="w-3 h-3 rounded-full" style={{ backgroundColor: "#ff5f57" }} />
                  <span className="w-3 h-3 rounded-full" style={{ backgroundColor: "#febc2e" }} />
                  <span className="w-3 h-3 rounded-full" style={{ backgroundColor: "#28c840" }} />
                  <span className="ml-3 text-xs font-mono" style={{ color: "rgba(255,255,255,0.25)" }}>
                    any platform with Go
                  </span>
                </div>
                <CopyButton text="go install github.com/deemwar-products/ops-phoenix-agent/sre-agent/cmd/sre-agent@latest" dark />
              </div>
              <pre className="p-5 text-sm font-mono leading-6" style={{ color: "#c9d1d9" }}>
                go install github.com/deemwar-products/ops-phoenix-agent/sre-agent/cmd/sre-agent@latest
              </pre>
            </div>
          </div>

          <div>
            <div
              className="rounded-xl overflow-hidden mb-4"
              style={{ backgroundColor: "#0f1117", border: "1px solid rgba(255,255,255,0.08)" }}
            >
              <div
                className="flex items-center justify-between px-4 py-3"
                style={{ borderBottom: "1px solid rgba(255,255,255,0.06)" }}
              >
                <div className="flex items-center gap-2">
                  <span className="w-3 h-3 rounded-full" style={{ backgroundColor: "#ff5f57" }} />
                  <span className="w-3 h-3 rounded-full" style={{ backgroundColor: "#febc2e" }} />
                  <span className="w-3 h-3 rounded-full" style={{ backgroundColor: "#28c840" }} />
                  <span className="ml-3 text-xs font-mono" style={{ color: "rgba(255,255,255,0.25)" }}>
                    from source
                  </span>
                </div>
                <CopyButton text="git clone https://github.com/deemwar-products/ops-phoenix-agent.git && cd ops-phoenix-agent/sre-agent && go build ./cmd/sre-agent" dark />
              </div>
              <pre className="p-5 text-sm font-mono leading-6" style={{ color: "#c9d1d9" }}>
                git clone https://github.com/deemwar-products/ops-phoenix-agent.git{'\n'}cd ops-phoenix-agent/sre-agent && go build ./cmd/sre-agent
              </pre>
            </div>
          </div>
        </div>

        <div className="mt-6">
          <p className="text-xs" style={{ color: "#8b867f" }}>
            Requires Go 1.24+. brew, npm, pip, and winget packages coming soon.
          </p>
        </div>

        <div className="mt-12">
          <p
            className="text-xs font-semibold uppercase tracking-[0.2em] mb-6 font-mono"
            style={{ color: "#735c41" }}
          >
            First run
          </p>

          <div
            className="rounded-xl overflow-hidden"
            style={{ border: "1px solid #e0d9cd", backgroundColor: "#ffffff" }}
          >
            <div
              className="px-6 py-4"
              style={{ borderBottom: "1px solid #e0d9cd", backgroundColor: "#f7f4ed" }}
            >
              <p className="text-xs font-mono uppercase tracking-[0.15em]" style={{ color: "#735c41" }}>
                Configure in under a minute
              </p>
            </div>

            <div className="px-6 py-6 space-y-6">
              {[
                {
                  label: "Grafana URL",
                  example: "https://observability.yourco.com",
                  note: "Grafana Cloud works too (https://your-project.grafana.net).",
                },
                {
                  label: "Grafana token",
                  example: "glsa_xxxxxxxxxxxxxxxxxxxxxxxxxxxx",
                  note: 'Service account token with Viewer permissions. Create at Grafana → Administration → Service Accounts.',
                },
                {
                  label: "Container patterns",
                  example: "myapp-*, my-service-*",
                  note: "Glob patterns to match your container names in Loki (e.g. nginx-*, api-*).",
                },
                {
                  label: "GitHub repo",
                  example: "yourco/your-product",
                  note: "The repo where the agent opens PRs. gh CLI must be authenticated.",
                },
                {
                  label: "AI API key",
                  example: "ANTHROPIC_API_KEY=sk-ant-...",
                  note: "Claude API key. Set as environment variable or enter during setup.",
                },
              ].map((item) => (
                <div key={item.label}>
                  <p className="text-sm font-semibold mb-1" style={{ color: "#1b1c19" }}>
                    {item.label}
                  </p>
                  <div className="flex items-center justify-between">
                    <pre
                      className="text-xs font-mono leading-6 px-4 py-3 rounded-lg overflow-x-auto flex-1"
                      style={{ backgroundColor: "#f7f4ed", color: "#735c41" }}
                    >
                      {item.example}
                    </pre>
                    <CopyButton text={item.example} />
                  </div>
                  <p className="text-xs mt-2" style={{ color: "#8b867f" }}>
                    {item.note}
                  </p>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
