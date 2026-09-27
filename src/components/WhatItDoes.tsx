"use client";

const capabilities = [
  {
    title: "Detect",
    description:
      "Queries Loki, Grafana, or Cloud Monitoring for errors in your container logs. Deduplicates by pattern, sorts by frequency, and surfaces the top finding in seconds — not minutes.",
  },
  {
    title: "Analyze + Fix",
    description:
      "Sends the error summary to Claude, which correlates it with recent deploys and traces, identifies the root cause, and writes a unified diff that targets it directly.",
  },
  {
    title: "PR + Deploy",
    description:
      "Opens a PR with the fix for your review. Nothing deploys without human approval. Post-deploy, it monitors key metrics and rolls back if the fix didn't hold.",
  },
];

export function WhatItDoes() {
  return (
    <section id="what-it-does" className="py-32 px-6" style={{ backgroundColor: "#0f1117" }}>
      <div className="max-w-6xl mx-auto">
        <div className="mb-16">
          <p
            className="text-xs font-semibold uppercase tracking-[0.2em] mb-4 font-mono"
            style={{ color: "#735c41" }}
          >
            What it does
          </p>
          <h2 className="text-3xl sm:text-4xl md:text-5xl leading-tight mb-4" style={{ color: "#f0ede6" }}>
            Three steps. All on your machine.
          </h2>
          <p className="text-base sm:text-lg leading-relaxed max-w-2xl" style={{ color: "#8b867f" }}>
            No cloud dependency. No data leaves your infrastructure. The agent runs locally,
            reads your logs, and writes fixes to your repo.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-12">
          {capabilities.map((cap) => (
            <div key={cap.title}>
              <div
                className="text-xs font-mono mb-3 tracking-wider"
                style={{ color: "#735c41" }}
              >
                {cap.title === "Detect" && "01 — QUERY"}
                {cap.title === "Analyze + Fix" && "02 — AI + CODE"}
                {cap.title === "PR + Deploy" && "03 — SHIP"}
              </div>
              <h3
                className="text-xl font-bold mb-3"
                style={{ color: "#f0ede6" }}
              >
                {cap.title}
              </h3>
              <p className="text-sm leading-relaxed" style={{ color: "#8b867f" }}>
                {cap.description}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
