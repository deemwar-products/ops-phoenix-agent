"use client";

const capabilities = [
  {
    title: "Detect",
    description:
      "Z-score baselines on your metrics catch anomalies before they become alerts. Learns your traffic patterns — no threshold tuning.",
  },
  {
    title: "Analyze + Fix",
    description:
      "Reads logs from Cloud Logging, Loki, or Prometheus, correlates with recent deploys and traces, and writes a real code diff — not a runbook suggestion.",
  },
  {
    title: "Deploy + Verify",
    description:
      "Opens a PR with the fix for your review — nothing deploys without human approval. Monitors key metrics post-deploy and rolls back automatically if the fix didn't work.",
  },
];

export function WhatItDoes() {
  return (
    <section id="what-it-does" className="py-32 px-6" style={{ backgroundColor: "#fdfcf8" }}>
      <div className="max-w-6xl mx-auto">
        <div className="mb-16">
          <p
            className="text-xs font-semibold uppercase tracking-[0.2em] mb-4 font-mono"
            style={{ color: "#735c41" }}
          >
            What it does
          </p>
          <h2 className="text-3xl sm:text-4xl md:text-5xl leading-tight mb-4" style={{ color: "#1b1c19" }}>
            Three jobs. All of them on your own machine.
          </h2>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-12">
          {capabilities.map((cap) => (
            <div key={cap.title}>
              <h3
                className="text-xl font-bold mb-3"
                style={{ color: "#1b1c19" }}
              >
                {cap.title}
              </h3>
              <p className="text-sm leading-relaxed" style={{ color: "#6b665e" }}>
                {cap.description}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
