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
          onClick={() => copy("sre-agent detect --time-window 5m\nsre-agent analyze\nsre-agent fix")}
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
    <section id="two-surfaces" className="py-32 px-6" style={{ backgroundColor: "#0f1117" }}>
      <div className="max-w-6xl mx-auto">
        <div className="mb-12">
          <p
            className="text-xs font-semibold uppercase tracking-[0.2em] mb-4 font-mono"
            style={{ color: "#735c41" }}
          >
            Two interfaces. One engine.
          </p>
          <h2 className="text-3xl sm:text-4xl md:text-5xl leading-tight mb-4" style={{ color: "#f0ede6" }}>
            Run it yourself. Or let an AI run it for you.
          </h2>
          <p className="text-base sm:text-lg leading-relaxed max-w-2xl" style={{ color: "#8b867f" }}>
            The same sre-agent binary powers both interfaces. Same commands, same output,
            same fix quality.
          </p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* CLI */}
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.2em] mb-4 font-mono" style={{ color: "#735c41" }}>
              CLI — human-driven
            </p>
            <p className="text-sm mb-4" style={{ color: "#8b867f" }}>
              Run from your terminal. Full control over every flag and step. Pipe it into
              your scripts, cron, or CI pipeline.
            </p>
            <ColoredTerminalPanel label="terminal">
              <span style={{ color: "#735c41" }}>$</span> sre-agent detect --time-window 5m{'\n'}
              <span style={{ color: "#8b949e" }}>
                {'›'} Found 2 errors. P1: 5xx spike /api/resumes{'\n\n'}
              </span>
              <span style={{ color: "#735c41" }}>$</span> sre-agent analyze{'\n'}
              <span style={{ color: "#8b949e" }}>
                {'›'} Root cause: token refresh regression in auth-service v2.3.1{'\n\n'}
              </span>
              <span style={{ color: "#735c41" }}>$</span> sre-agent fix{'\n'}
              <span style={{ color: "#3fb950" }}>
                {'›'} PR #482 opened. CI passed. Team merges and deploys via their pipeline.{'\n'}
              </span>
            </ColoredTerminalPanel>
          </div>

          {/* Agent skill */}
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.2em] mb-4 font-mono" style={{ color: "#735c41" }}>
              Agent skill — AI-driven
            </p>
            <p className="text-sm mb-4" style={{ color: "#8b867f" }}>
              Claude Code, Codex, Cursor, or any agent that reads SKILL.md can drive
              sre-agent. Give it a goal, it picks the right commands.
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
                <span style={{ color: "#ff7b72" }}>You:</span> Check production for errors and fix anything critical.{'\n\n'}
                <span style={{ color: "#8b949e" }}>{'(calls sre-agent skill)'}</span> {"→"} <span style={{ color: "#58a6ff" }}>detect --time-window 1h</span>{'\n\n'}
                <span style={{ color: "#c9d1d9" }}>→ Found 3 errors. P1: 5xx spike /api/resumes.{'\n\n'}</span>
                <span style={{ color: "#8b949e" }}>{'(calls sre-agent skill)'}</span> {"→"} <span style={{ color: "#58a6ff" }}>analyze</span>{'\n\n'}
                <span style={{ color: "#c9d1d9" }}>→ Root cause: token refresh regression. Fix confidence: 0.91.{'\n\n'}</span>
                <span style={{ color: "#8b949e" }}>{'(calls sre-agent skill)'}</span> {"→"} <span style={{ color: "#58a6ff" }}>fix</span>{'\n\n'}
                <span style={{ color: "#3fb950" }}>→ PR #482 opened. Team reviews, merges, deploys.{'\n'}</span>
              </pre>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
