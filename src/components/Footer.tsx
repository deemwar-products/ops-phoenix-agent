"use client";

export function Footer() {
  return (
    <footer className="py-12 px-6 border-t border-surface-border">
      <div className="max-w-6xl mx-auto flex flex-col md:flex-row items-center justify-between gap-6">
        <div className="flex items-center gap-2">
          <span className="text-lg font-semibold tracking-tight">
            ops<span className="text-accent">-</span>phoenix
          </span>
          <span className="text-xs text-muted font-mono">
            by deemwar
          </span>
        </div>

        <div className="flex items-center gap-8 text-sm text-muted">
          <a
            href="https://github.com/deemwar-products/ops-phoenix-agent"
            target="_blank"
            rel="noopener noreferrer"
            className="hover:text-foreground transition-colors"
          >
            GitHub
          </a>
          <a href="#" className="hover:text-foreground transition-colors">
            Documentation
          </a>
          <a href="#" className="hover:text-foreground transition-colors">
            Contact
          </a>
        </div>

        <p className="text-xs text-muted">
          Open source. Built for engineers.
        </p>
      </div>
    </footer>
  );
}
