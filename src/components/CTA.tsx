"use client";

export function CTA() {
  return (
    <section id="cta" className="py-32 px-6" style={{ backgroundColor: "#fdfcf8" }}>
      <div className="max-w-3xl mx-auto text-center">
        <h2 className="text-3xl sm:text-4xl md:text-5xl leading-tight mb-6" style={{ color: "#1b1c19" }}>
          Not ready to commit?
          <br />
          <span style={{ color: "#735c41" }}>Talk to us.</span>
        </h2>
        <p className="text-base sm:text-lg leading-relaxed mb-10" style={{ color: "#6b665e" }}>
          SRE Agent is in private beta. Tell us about your infra and we&apos;ll
          set you up — no card, no commitment.
        </p>
        <a
          href="https://deemwar.com/contact"
          className="inline-flex items-center gap-2 px-8 py-3.5 rounded-full text-sm font-semibold transition-colors"
          style={{ backgroundColor: "#735c41", color: "#fff" }}
        >
          Talk to us
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
            <path d="M5 12h14M12 5l7 7-7 7" />
          </svg>
        </a>
      </div>
    </section>
  );
}
