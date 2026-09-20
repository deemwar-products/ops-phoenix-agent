"use client";

import { useState, useCallback } from "react";

function useCopyToClipboard() {
  const [copied, setCopied] = useState(false);
  const copy = useCallback(async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // clipboard unavailable
    }
  }, []);
  return { copied, copy };
}

function ColoredTerminalPanel({ label, children }: { label: string; children: React.ReactNode }) {
  const { copied, copy } = useCopyToClipboard();

  return (
    <div
      className="rounded-xl overflow-hidden"
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
            {label}
          </span>
        </div>
        <button
          onClick={() => copy("sre-agent detect --source grafana --duration 5m\nsre-agent analyze --finding 0 --deep\nsre-agent fix --finding 0 --approve")}
          className="text-[10px] font-mono uppercase tracking-[0.1em] px-2.5 py-1 rounded-md transition-colors"
          style={{
            color: copied ? "#3fb950" : "rgba(255,255,255,0.25)",
            backgroundColor: copied ? "rgba(63,185,80,0.12)" : "transparent",
          }}
        >
          {copied ? "Copied" : "Copy"}
        </button>
      </div>
      <pre className="p-5 text-sm font-mono leading-7 overflow-x-auto" style={{ color: "#c9d1d9" }}>
        {children}
      </pre>
    </div>
  );
}

export function TwoSurfaces() {
  return (
    <section id="two-surfaces" className="py-32 px-6" style={{ backgroundColor: "#fdfcf8" }}>
      <div className="max-w-6xl mx-auto">
        <div className="mb-12">
          <p
            className="text-xs font-semibold uppercase tracking-[0.2em] mb-4 font-mono"
            style={{ color: "#735c41" }}
          >
            See it in action
          </p>
          <h2 className="text-3xl sm:text-4xl md:text-5xl leading-tight mb-4" style={{ color: "#1b1c19" }}>
            Same engine. Two surfaces.
          </h2>
          <p className="text-base sm:text-lg leading-relaxed max-w-2xl" style={{ color: "#6b665e" }}>
            Run SRE Agent from your terminal or let any AI agent drive it. The
            output — and the incident resolution — is identical.
          </p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Terminal */}
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.2em] mb-4 font-mono" style={{ color: "#735c41" }}>
              From the terminal
            </p>
            <ColoredTerminalPanel label="terminal">
              <span style={{ color: "#735c41" }}>$</span> sre-agent detect --source grafana --duration 5m{`\n`}
              <span style={{ color: "#8b949e" }}>{'›'} Scanning Cloud Logging for anomalies...</span>{`\n`}
              <span style={{ color: "#8b949e" }}>{'›'} 2 anomalies found. P1: 5xx spike on /api/resumes</span>{`\n\n`}
              <span style={{ color: "#735c41" }}>$</span> sre-agent analyze --finding 0 --deep{`\n`}
              <span style={{ color: "#8b949e" }}>{'›'} Root cause: token refresh regression in auth-service v2.3.1</span>{`\n\n`}
              <span style={{ color: "#735c41" }}>$</span> sre-agent fix --finding 0 --approve{`\n`}
              <span style={{ color: "#3fb950" }}>{'›'} PR opened: #482 revert token validation path</span>{`\n`}
              <span style={{ color: "#3fb950" }}>{'›'} CI passed. Awaiting human approval to deploy.</span>
            </ColoredTerminalPanel>
          </div>

          {/* Agent skill */}
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.2em] mb-4 font-mono" style={{ color: "#735c41" }}>
              From Claude or Codex (agent skill)
            </p>
            <div
              className="rounded-xl overflow-hidden"
              style={{ backgroundColor: "#0f1117", border: "1px solid rgba(255,255,255,0.08)" }}
            >
              <div className="flex items-center gap-2 px-4 py-3" style={{ borderBottom: "1px solid rgba(255,255,255,0.06)" }}>
                <span className="w-3 h-3 rounded-full" style={{ backgroundColor: "#ff5f57" }} />
                <span className="w-3 h-3 rounded-full" style={{ backgroundColor: "#febc2e" }} />
                <span className="w-3 h-3 rounded-full" style={{ backgroundColor: "#28c840" }} />
                <span className="ml-3 text-xs font-mono" style={{ color: "rgba(255,255,255,0.25)" }}>
                  agent session
                </span>
              </div>
              <pre className="p-5 text-sm font-mono leading-7 overflow-x-auto" style={{ color: "#c9d1d9" }}>
                <span style={{ color: "#ff7b72" }}>You:</span>{" "}Run a full detection pass on prod, analyze
                the top finding, and if confidence is above 80% open a PR.{`\n\n`}
                <span style={{ color: "#8b949e" }}>{'(calls sre-agent skill)'}</span> →{" "}
                <span style={{ color: "#58a6ff" }}>detect --source grafana --duration 5m</span>{`\n\n`}
                <span style={{ color: "#c9d1d9" }}>→ 2 anomalies found. P1: 5xx spike on /api/resumes (confidence 0.91)</span>{`\n\n`}
                <span style={{ color: "#8b949e" }}>{'(calls sre-agent skill)'}</span> →{" "}
                <span style={{ color: "#58a6ff" }}>analyze --finding 0 --deep</span>{`\n\n`}
                <span style={{ color: "#c9d1d9" }}>→ Root cause: token refresh regression in auth-service v2.3.1. Opening PR…</span>
              </pre>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
