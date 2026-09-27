"use client";

export function Footer() {
  return (
    <footer
      className="py-12 px-6"
      style={{ borderTop: "1px solid #e8e3d9", backgroundColor: "#fdfcf8" }}
    >
      <div className="max-w-6xl mx-auto flex flex-col md:flex-row items-center justify-between gap-6">
        <div className="flex items-center gap-1.5">
          <span className="text-base font-bold tracking-tight" style={{ color: "#1b1c19" }}>
            SRE<span style={{ color: "#735c41" }}> Agent</span>
          </span>
        </div>

        <div className="flex items-center gap-8 text-sm" style={{ color: "#6b665e" }}>
          <a href="https://github.com/deemwar-products/ops-phoenix-agent" className="hover:opacity-70 transition-opacity" target="_blank" rel="noopener noreferrer">
            GitHub
          </a>
          <a href="mailto:admin@deemwar.com" className="hover:opacity-70 transition-opacity">
            Contact
          </a>
          <a href="https://deemwar.com" className="hover:opacity-70 transition-opacity">
            Deemwar
          </a>
        </div>

        <p className="text-xs" style={{ color: "#8b867f" }}>
          &copy; 2026 Deemwar Products. MIT licensed.
        </p>
      </div>
    </footer>
  );
}
