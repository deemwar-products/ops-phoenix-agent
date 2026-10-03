"use client";

import { ContactModal } from "./ContactModal";

export function CTA() {
  return (
    <section id="cta" className="py-32 px-6" style={{ backgroundColor: "#0f1117" }}>
      <div className="max-w-3xl mx-auto text-center">
        <h2 className="text-3xl sm:text-4xl md:text-5xl leading-tight mb-6" style={{ color: "#f0ede6" }}>
          Free to use. Support when you need it.
        </h2>
        <p className="text-base sm:text-lg leading-relaxed mb-10" style={{ color: "#8b867f" }}>
          Install in under a minute — no signup, no card. Bring your own keys
          and run it on your machine or in your own cloud. Need a hand? The
          setup package ($300, one-time) gets it running against your stack.
        </p>
        <div className="flex flex-col items-center gap-4">
          <ContactModal
            trigger={
              <span className="inline-flex items-center gap-2 px-8 py-3.5 rounded-full text-sm font-semibold cursor-pointer" style={{ backgroundColor: "#735c41", color: "#fff" }}>
                admin@deemwar.com
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                  <path d="M5 12h14M12 5l7 7-7 7" />
                </svg>
              </span>
            }
          />
          <p className="text-xs" style={{ color: "#6b665e" }}>
            Free to use · Setup package: $300 one-time
          </p>
        </div>
      </div>
    </section>
  );
}
