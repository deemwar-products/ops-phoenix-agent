"use client";

import { useEffect, useRef } from "react";

// Opens the narrated video in a modal. The file lives in public/ so Vite
// serves it from the site root: /sre-agent-in-action.mp4. Autoplay with
// sound only happens because the visitor clicked the trigger — browsers block
// otherwise. The poster keeps the button looking alive before the file loads.
export function VideoLightbox() {
  const openRef = useRef<HTMLDivElement | null>(null);
  const videoRef = useRef<HTMLVideoElement | null>(null);
  const triggerRef = useRef<HTMLButtonElement | null>(null);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") close();
    };
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  });

  function close() {
    const el = openRef.current;
    if (!el) return;
    el.style.display = "none";
    // Stop playback so the audio doesn't keep talking behind the page.
    videoRef.current?.pause();
    triggerRef.current?.focus();
  }

  function show() {
    const el = openRef.current;
    if (!el) return;
    el.style.display = "flex";
    videoRef.current?.play().catch(() => {
      // Autoplay refused — the controls are right there, so nothing to do.
    });
  }

  return (
    <>
      <button
        ref={triggerRef}
        onClick={show}
        className="inline-flex items-center gap-2 px-8 py-3.5 font-semibold text-sm rounded-full transition-colors"
        style={{ backgroundColor: "transparent", color: "#735c41", border: "1px solid #d5cfc4" }}
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
          <path d="M8 5.5v13l11-6.5-11-6.5z" />
        </svg>
        Watch how it works
      </button>

      <div
        ref={openRef}
        role="dialog"
        aria-modal="true"
        aria-label="SRE Agent in action video"
        onClick={(e) => {
          // Click outside the dialog box closes; clicks inside do not.
          if (e.target === e.currentTarget) close();
        }}
        className="fixed inset-0 z-50 items-center justify-center p-6"
        style={{ display: "none", backgroundColor: "rgba(15,17,23,0.72)" }}
      >
        <div
          className="relative w-full max-w-5xl rounded-2xl overflow-hidden"
          style={{ backgroundColor: "#0f1117", border: "1px solid rgba(255,255,255,0.12)" }}
        >
          <button
            onClick={close}
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
    </>
  );
}