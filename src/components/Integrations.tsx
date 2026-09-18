"use client";

const integrations = [
  {
    name: "Cloud Logging",
    description: "GCP Cloud Logging for structured log ingestion and analysis",
    icon: (
      <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
        <path d="M3 15a4 4 0 004 4h9a5 5 0 10-.1-9.999 5.002 5.002 0 10-9.78 2.096A4.001 4.001 0 003 15z" />
      </svg>
    ),
  },
  {
    name: "Cloud Monitoring",
    description: "GCP Cloud Monitoring for metric collection and anomaly detection",
    icon: (
      <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
        <path d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
      </svg>
    ),
  },
  {
    name: "GitHub",
    description: "PR creation, CI/CD triggers, issue tracking, and auto-merge",
    icon: (
      <svg width="32" height="32" viewBox="0 0 24 24" fill="currentColor">
        <path d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" />
      </svg>
    ),
  },
  {
    name: "Loki + Prometheus",
    description: "Open-source observability stack as an alternative data source",
    icon: (
      <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
        <path d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4" />
      </svg>
    ),
  },
  {
    name: "Docker / Kubernetes",
    description: "Deploy as a container on Cloud Run, GKE, or any Kubernetes cluster",
    icon: (
      <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
        <path d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
      </svg>
    ),
  },
];

export function Integrations() {
  return (
    <section id="integrations" className="py-32 px-6 border-t border-surface-border">
      <div className="max-w-6xl mx-auto">
        {/* Section header */}
        <div className="text-center mb-20">
          <p className="text-sm font-semibold uppercase tracking-widest text-accent mb-4 font-mono">
            Integrations
          </p>
          <h2 className="font-bold text-4xl sm:text-5xl md:text-6xl leading-tight mb-6">
            One container.
            <br />
            <span className="text-accent">Your data never leaves your infra.</span>
          </h2>
          <p className="text-muted text-lg max-w-2xl mx-auto">
            Connect to your existing observability stack. Deploy wherever your
            infrastructure lives. No SaaS dependency — everything runs on your
            machines.
          </p>
        </div>

        {/* Integration grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
          {integrations.map((integration) => (
            <div
              key={integration.name}
              className="flex flex-col items-center text-center p-8 rounded-2xl border border-surface-border bg-surface hover:border-accent/30 transition-colors"
            >
              <div className="text-muted mb-4">{integration.icon}</div>
              <h3 className="font-semibold mb-2">{integration.name}</h3>
              <p className="text-muted text-sm leading-relaxed">
                {integration.description}
              </p>
            </div>
          ))}
        </div>

        {/* Deployment callout */}
        <div className="mt-12 text-center">
          <div className="inline-flex flex-wrap items-center justify-center gap-3 p-4 rounded-xl border border-surface-border bg-surface">
            <span className="text-sm text-muted">Deploy to:</span>
            {["Cloud Run", "GKE", "Docker", "Kubernetes", "ECS"].map(
              (platform) => (
                <span
                  key={platform}
                  className="text-xs font-mono px-3 py-1 rounded-full bg-background border border-surface-border"
                >
                  {platform}
                </span>
              )
            )}
          </div>
        </div>
      </div>
    </section>
  );
}
