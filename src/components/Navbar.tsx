"use client";

import { useState, useEffect } from "react";
import { ContactModal } from "./ContactModal";

export function Navbar() {
  const [scrolled, setScrolled] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 20);
    window.addEventListener("scroll", onScroll);
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  useEffect(() => {
    if (!menuOpen) return;
    const onKey = (e: KeyboardEvent) => { if (e.key === "Escape") setMenuOpen(false); };
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  }, [menuOpen]);

  const navLinkClass = "block py-2 text-sm font-medium transition-opacity hover:opacity-70";

  const links = [
    { href: "#how-it-works", label: "How it works" },
    { href: "#two-surfaces", label: "Two surfaces" },
    { href: "#install", label: "Install" },
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
            <a key={l.href} href={l.href} className="hover:opacity-70 transition-opacity">{l.label}</a>
          ))}
        </div>

        <div className="hidden md:block">
          <ContactModal
            trigger={
              <span className="inline-flex items-center gap-2 px-5 py-2 text-sm font-semibold rounded-full transition-colors" style={{ backgroundColor: "#735c41", color: "#fff" }}>
                Contact
              </span>
            }
          />
        </div>

        <button
          onClick={() => setMenuOpen(!menuOpen)}
          className="md:hidden flex flex-col justify-center items-center w-8 h-8 gap-1.5"
          aria-label="Toggle menu"
        >
          <span className="block w-5 h-0.5 transition-all duration-300" style={{ backgroundColor: "#1b1c19", transform: menuOpen ? "rotate(45deg) translate(2px, 2px)" : "none" }} />
          <span className="block w-5 h-0.5 transition-all duration-300" style={{ backgroundColor: "#1b1c19", opacity: menuOpen ? 0 : 1 }} />
          <span className="block w-5 h-0.5 transition-all duration-300" style={{ backgroundColor: "#1b1c19", transform: menuOpen ? "rotate(-45deg) translate(2px, -2px)" : "none" }} />
        </button>
      </div>

      {menuOpen && (
        <div className="md:hidden" style={{ backgroundColor: "#fdfcf8", borderBottom: "1px solid #e8e3d9" }}>
          <div className="px-6 py-4 flex flex-col gap-1">
            {links.map((l) => (
              <a
                key={l.href}
                href={l.href}
                className={navLinkClass}
                style={{ color: "#6b665e" }}
                onClick={() => setMenuOpen(false)}
              >
                {l.label}
              </a>
            ))}
            <div className="pt-2">
              <ContactModal
                trigger={
                  <span className="block w-full text-center py-2.5 rounded-full text-sm font-semibold" style={{ backgroundColor: "#735c41", color: "#fff" }}>
                    Contact
                  </span>
                }
              />
            </div>
          </div>
        </div>
      )}
    </nav>
  );
}
