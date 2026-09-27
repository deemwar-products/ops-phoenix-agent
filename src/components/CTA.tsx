"use client";

export function CTA() {
  return (
    <section id="cta" className="py-32 px-6" style={{ backgroundColor: "#0f1117" }}>
      <div className="max-w-3xl mx-auto text-center">
        <h2 className="text-3xl sm:text-4xl md:text-5xl leading-tight mb-6" style={{ color: "#f0ede6" }}>
          Free to use. Support when you need it.
        </h2>
        <p className="text-base sm:text-lg leading-relaxed mb-10" style={{ color: "#8b867f" }}>
          Install the Open plan in under a minute. No signup, no card, no cloud.
          For support, custom integrations, or enterprise SLAs, reach out directly.
        </p>
        <div className="flex flex-col items-center gap-4">
          <a
            href="mailto:admin@deemwar.com?subject=SRE%20Agent%20inquiry"
            className="inline-flex items-center gap-2 px-8 py-3.5 rounded-full text-sm font-semibold transition-colors"
            style={{ backgroundColor: "#735c41", color: "#fff" }}
          >
            admin@deemwar.com
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
              <path d="M5 12h14M12 5l7 7-7 7" />
            </svg>
          </a>
          <p className="text-xs" style={{ color: "#6b665e" }}>
            Support: $100/month · Enterprise: custom pricing
          </p>
        </div>
      </div>
    </section>
  );
}
