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
      // clipboard unavailable — silent fail
    }
  }, []);

  return { copied, copy };
}

interface TerminalPanelProps {
  label: string;
  platform?: string;
  children: string;
}

export function TerminalPanel({ label, platform, children }: TerminalPanelProps) {
  const { copied, copy } = useCopyToClipboard();

  return (
    <div
      className="rounded-xl overflow-hidden flex flex-col"
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
        {platform && (
          <span className="text-[10px] font-mono uppercase tracking-[0.12em]" style={{ color: "rgba(255,255,255,0.2)" }}>
            {platform}
          </span>
        )}
        <button
          onClick={() => copy(children.trim())}
          className="text-[10px] font-mono uppercase tracking-[0.1em] px-2.5 py-1 rounded-md transition-colors"
          style={{
            color: copied ? "#3fb950" : "rgba(255,255,255,0.25)",
            backgroundColor: copied ? "rgba(63,185,80,0.12)" : "transparent",
          }}
        >
          {copied ? "Copied" : "Copy"}
        </button>
      </div>
      <pre className="p-5 text-sm font-mono leading-7 overflow-x-auto flex-1" style={{ color: "#c9d1d9" }}>
        {children}
      </pre>
    </div>
  );
}
