"use client";

import { useState, useCallback } from "react";
import { TerminalPanel } from "./TerminalPanel";

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
            Checksum verified. macOS, Linux and Windows.
          </h2>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 items-stretch">
          <TerminalPanel label="Homebrew" platform="macOS / Linux">
{`brew install deemwar-products/tap/sre-agent`}
          </TerminalPanel>

          <TerminalPanel label="Go install" platform="any platform with Go">
{`go install github.com/deemwar-products/sre-agent/cmd/sre-agent@latest`}
          </TerminalPanel>

          <TerminalPanel label="Windows" platform="PowerShell + winget">
{`winget install DeemwarProducts.SREAgent`}
          </TerminalPanel>
        </div>

        <div className="mt-6">
          <TerminalPanel label="Windows (Go toolchain)" platform="cmd / PowerShell">
{`go install github.com/deemwar-products/sre-agent/cmd/sre-agent@latest`}
          </TerminalPanel>
        </div>

        <div className="mt-12">
          <p
            className="text-xs font-semibold uppercase tracking-[0.2em] mb-6 font-mono"
            style={{ color: "#735c41" }}
          >
            First-run setup
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
                Run sre-agent init — here is what it will ask for
              </p>
            </div>

            <div className="px-6 py-6 space-y-6">
              {[
                {
                  label: "Grafana URL",
                  hint: "The base URL of your Grafana instance.",
                  example: "https://grafana.your-company.com",
                  note: "Must be reachable from the machine running SRE Agent. Grafana Cloud URLs work too (e.g., https://your-project.grafana.net).",
                },
                {
                  label: "Grafana API key",
                  hint: "A Grafana service account token with at least Viewer + Alerting read permissions.",
                  example: "glsa_xxxxxxxxxxxxxxxxxxxxxxxxxxxx",
                  note: 'Create one at Grafana → Administration → Service Accounts → New API Key. Scope: "Viewer" for read-only, "Editor" if you want SRE Agent to annotate dashboards.',
                },
                {
                  label: "Primary data source",
                  hint: "Where SRE Agent pulls metrics and logs from.",
                  example: "grafana  (or: loki / prometheus / cloud-logging)",
                  note: "Grafana is the default — it acts as a unified query layer for Loki, Prometheus, and Cloud Monitoring behind it. Choose a direct source only if you are not running Grafana.",
                },
                {
                  label: "Mode",
                  hint: "How aggressive the fix pipeline is.",
                  example: "guided  (or: autonomous)",
                  note: 'Guided — opens a PR and waits for your approval before deploying. Autonomous — deploys automatically after CI passes (not recommended for production without a rollback policy).',
                },
                {
                  label: "Notification channel (optional)",
                  hint: "Where to send incident summaries after resolution.",
                  example: "slack://#sre-alerts  (or: email / none)",
                  note: "Requires the corresponding integration to be set up. Leave blank to skip notifications.",
                },
              ].map((item) => (
                <div key={item.label}>
                  <p className="text-sm font-semibold mb-1" style={{ color: "#1b1c19" }}>
                    {item.label}
                  </p>
                  <p className="text-xs mb-2" style={{ color: "#6b665e" }}>
                    {item.hint}
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

        <p className="mt-8 text-sm" style={{ color: "#8b867f" }}>
          Requires Go 1.22+ (or the pre-built binary via Homebrew / winget). Run{" "}
          <code
            className="font-mono text-xs px-1.5 py-0.5 rounded"
            style={{ backgroundColor: "#f0ede6", color: "#735c41" }}
          >
            sre-agent init
          </code>{" "}
          to get started.
        </p>
      </div>
    </section>
  );
}
