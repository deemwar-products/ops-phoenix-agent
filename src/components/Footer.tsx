"use client";

import { useState, useEffect } from "react";

export function Footer() {
  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 20);
    window.addEventListener("scroll", onScroll);
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  const scrollTo = (id: string) => {
    const el = document.getElementById(id);
    if (!el) return;
    const navHeight = 64;
    const top = el.getBoundingClientRect().top + window.scrollY - navHeight;
    window.scrollTo({ top, behavior: "smooth" });
  };

  return (
    <footer
      className="py-12 px-6"
      style={{
        borderTop: "1px solid #e8e3d9",
        backgroundColor: scrolled ? "rgba(253,252,248,0.9)" : "#fdfcf8",
        backdropFilter: scrolled ? "blur(12px)" : "none",
        transition: "background-color 0.3s, backdrop-filter 0.3s",
      }}
    >
      <div className="max-w-6xl mx-auto flex flex-col md:flex-row items-center justify-between gap-6">
        <div className="flex items-center gap-1.5">
          <span className="text-base font-bold tracking-tight" style={{ color: "#1b1c19" }}>
            SRE<span style={{ color: "#735c41" }}> Agent</span>
          </span>
        </div>

        <div className="flex items-center gap-8 text-sm" style={{ color: "#6b665e" }}>
          <button onClick={() => scrollTo("install")} className="hover:opacity-70 transition-opacity">Install</button>
          <button onClick={() => scrollTo("how-it-works")} className="hover:opacity-70 transition-opacity">Docs</button>
          <button onClick={() => scrollTo("pricing")} className="hover:opacity-70 transition-opacity">Pricing</button>
        </div>

        <p className="text-xs" style={{ color: "#8b867f" }}>
          &copy; 2026 Deemwar Products
        </p>
      </div>
    </footer>
  );
}
