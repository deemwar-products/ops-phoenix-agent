"use client";

import { useEffect, useRef, useState } from "react";

// Dedicated "See it in action" section. Shows the real PR page as a poster;
// clicking it opens the narrated video in a modal. Self-contained so it
// doesn't depend on the hero's VideoLightbox.
export function DemoSection() {
  const [open, setOpen] = useState(false);
  const videoRef = useRef<HTMLVideoElement | null>(null);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setOpen(false);
    };
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  }, []);

  return (
    <section id="demo" className="py-32 px-6" style={{ backgroundColor: "#fdfcf8" }}>
      <div className="max-w-4xl mx-auto text-center">
        <p
          className="text-xs font-semibold uppercase tracking-[0.2em] mb-4 font-mono"
          style={{ color: "#735c41" }}
        >
          Watch it work
        </p>
        <h2 className="text-3xl sm:text-4xl md:text-5xl leading-tight mb-8" style={{ color: "#1b1c19" }}>
          See the agent in action
        </h2>

        <div
          className="relative rounded-2xl overflow-hidden cursor-pointer group"
          style={{ border: "1px solid #e0d9cd" }}
          onClick={() => setOpen(true)}
        >
          <img
            src="/sre-agent-in-action-poster.png"
            alt="SRE Agent pull request"
            className="w-full transition-transform duration-300 group-hover:scale-[1.01]"
          />
          <div
            className="absolute inset-0 flex items-center justify-center"
            style={{ backgroundColor: "rgba(15,17,23,0.35)" }}
          >
            <div
              className="w-20 h-20 rounded-full flex items-center justify-center transition-transform duration-300 group-hover:scale-110"
              style={{ backgroundColor: "#735c41" }}
            >
              <svg width="28" height="28" viewBox="0 0 24 24" fill="#fff" aria-hidden="true">
                <path d="M8 5.5v13l11-6.5-11-6.5z" />
              </svg>
            </div>
          </div>
        </div>

        <p className="mt-4 text-sm" style={{ color: "#8b867f" }}>
          75 seconds · real commands · real pull request
        </p>
      </div>

      {open && (
        <div
          role="dialog"
          aria-modal="true"
          aria-label="SRE Agent demo video"
          onClick={(e) => {
            if (e.target === e.currentTarget) setOpen(false);
          }}
          className="fixed inset-0 z-50 items-center justify-center p-6"
          style={{ display: "flex", backgroundColor: "rgba(15,17,23,0.72)" }}
        >
          <div
            className="relative w-full max-w-5xl rounded-2xl overflow-hidden"
            style={{ backgroundColor: "#0f1117", border: "1px solid rgba(255,255,255,0.12)" }}
          >
            <button
              onClick={() => setOpen(false)}
              aria-label="Close video"
              className="absolute top-3 right-3 z-10 w-9 h-9 flex items-center justify-center rounded-full text-sm"
              style={{ backgroundColor: "rgba(255,255,255,0.12)", color: "#fff" }}
            >
              ✕
            </button>
            <video
              ref={videoRef}
              src="/sre-agent-in-action.mp4"
              poster="/sre-agent-in-action-poster.png"
              controls
              autoPlay
              playsInline
              preload="metadata"
              className="w-full"
              style={{ display: "block" }}
            />
          </div>
        </div>
      )}
    </section>
  );
}
