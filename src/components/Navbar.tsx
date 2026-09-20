"use client";

import { useState, useEffect } from "react";

export function Navbar() {
  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 20);
    window.addEventListener("scroll", onScroll);
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  const links = [
    { href: "#what-it-does", label: "What it does" },
    { href: "#install", label: "Install" },
    { href: "#problem", label: "Problem" },
    { href: "#how-it-works", label: "How it works" },
    { href: "#pricing", label: "Pricing" },
  ];

  return (
    <nav
      className="fixed top-0 left-0 right-0 z-50 transition-all duration-300"
      style={{
        backgroundColor: scrolled ? "rgba(253,252,248,0.9)" : "transparent",
        backdropFilter: scrolled ? "blur(12px)" : "none",
        borderBottom: scrolled ? "1px solid #e8e3d9" : "1px solid transparent",
      }}
    >
      <div className="max-w-6xl mx-auto px-6 h-16 flex items-center justify-between">
        <a href="#" className="flex items-center gap-1.5">
          <span className="text-lg font-bold tracking-tight" style={{ color: "#1b1c19" }}>
            SRE<span style={{ color: "#735c41" }}> Agent</span>
          </span>
        </a>

        <div className="hidden md:flex items-center gap-8 text-sm font-medium" style={{ color: "#6b665e" }}>
          {links.map((l) => (
            <a
              key={l.href}
              href={l.href}
              className="hover:opacity-70 transition-opacity"
            >
              {l.label}
            </a>
          ))}
        </div>

        <a
          href="https://deemwar.com/contact"
          className="px-5 py-2 text-sm font-semibold rounded-full transition-colors"
          style={{ backgroundColor: "#735c41", color: "#fff" }}
        >
          Talk to us
        </a>
      </div>
    </nav>
  );
}
