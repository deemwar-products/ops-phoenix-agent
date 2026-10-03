"use client";

import { useState, useEffect } from "react";

export function ContactModal({ trigger }: { trigger?: React.ReactNode }) {
  const [open, setOpen] = useState(false);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setOpen(false);
    };
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  }, [open]);

  const copyEmail = async () => {
    try {
      await navigator.clipboard.writeText("admin@deemwar.com");
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // clipboard unavailable
    }
  };

  return (
    <>
      {trigger ? (
        <button
          onClick={() => setOpen(true)}
          style={{ padding: 0, border: "none", background: "transparent", display: "block", width: "100%" }}
        >
          {trigger}
        </button>
      ) : (
        <button
          onClick={() => setOpen(true)}
          className="px-5 py-2 text-sm font-semibold rounded-full transition-colors"
          style={{ backgroundColor: "#735c41", color: "#fff" }}
        >
          Contact
        </button>
      )}

      {open && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center p-4"
          style={{ backgroundColor: "rgba(0,0,0,0.5)" }}
          onClick={() => setOpen(false)}
        >
          <div
            className="rounded-2xl p-8 max-w-md w-full relative"
            style={{ backgroundColor: "#fdfcf8", border: "1px solid #e0d9cd" }}
            onClick={(e) => e.stopPropagation()}
          >
            <button
              onClick={() => setOpen(false)}
              className="absolute top-4 right-4 text-xs font-mono hover:opacity-70 transition-opacity"
              style={{ color: "#735c41" }}
            >
              ESC
            </button>

            <p className="text-xs font-semibold uppercase tracking-[0.15em] mb-4 font-mono" style={{ color: "#735c41" }}>
              Contact
            </p>
            <h3 className="text-xl font-bold mb-3" style={{ color: "#1b1c19" }}>
              Talk to us
            </h3>
            <p className="text-sm mb-6" style={{ color: "#6b665e" }}>
              MIT licensed, free forever. Reach out for support or community questions.
            </p>

            <div className="p-4 rounded-xl mb-4" style={{ backgroundColor: "#f7f4ed", border: "1px solid #e0d9cd" }}>
              <p className="text-xs mb-1" style={{ color: "#8b867f" }}>Email</p>
              <div className="flex items-center justify-between gap-2">
                <a
                  href="mailto:admin@deemwar.com"
                  className="text-sm font-mono font-semibold hover:opacity-70 transition-opacity truncate"
                  style={{ color: "#735c41" }}
                >
                  admin@deemwar.com
                </a>
                <button
                  onClick={copyEmail}
                  className="text-[10px] font-mono uppercase tracking-[0.1em] px-2.5 py-1 rounded-md flex-shrink-0 transition-colors"
                  style={{
                    color: copied ? "#2f7d3a" : "#735c41",
                    backgroundColor: copied ? "rgba(63,185,80,0.12)" : "rgba(115,92,65,0.08)",
                  }}
                >
                  {copied ? "Copied" : "Copy"}
                </button>
              </div>
            </div>

            <div className="space-y-2 text-xs" style={{ color: "#6b665e" }}>
              <p>MIT licensed · free forever · bring your own keys</p>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
