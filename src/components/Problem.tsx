"use client";

const pains = [
  {
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
        <path d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
      </svg>
    ),
    title: "Alert Fatigue",
    description:
      "50+ alerts per day, and most of them are noise. Engineers learn to ignore the bell — until the one that actually matters gets buried.",
  },
  {
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
        <path d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
    ),
    title: "30+ Min Investigations",
    description:
      "Every incident starts the same way: open 5 dashboards, grep through logs, check recent deploys, cross-reference with traces. Same investigation, different day.",
  },
  {
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
        <path d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
      </svg>
    ),
    title: "2 AM Pages",
    description:
      "Your on-call engineer gets paged at 2 AM for an issue that could have waited until morning — or been fixed before anyone noticed.",
  },
  {
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
        <path d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
      </svg>
    ),
    title: "Dashboards That Don't Answer",
    description:
      "You have Grafana, Cloud Monitoring, Datadog — a dozen dashboards full of metrics. None of them tell you why the page just fired.",
  },
];

export function Problem() {
  return (
    <section id="problem" className="py-32 px-6">
      <div className="max-w-6xl mx-auto">
        {/* Section header */}
        <div className="text-center mb-20">
          <p className="text-sm font-semibold uppercase tracking-widest text-accent mb-4 font-mono">
            The problem
          </p>
          <h2 className="font-display text-4xl sm:text-5xl md:text-6xl leading-tight mb-6">
            You have more signal<br />
            than sense.
          </h2>
          <p className="text-muted text-lg max-w-2xl mx-auto">
            Your observability stack tells you something broke. It doesn&apos;t
            tell you why, what to do about it, or whether the fix worked. That
            part still requires a human.
          </p>
        </div>

        {/* Pain cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {pains.map((pain) => (
            <div
              key={pain.title}
              className="group p-8 rounded-2xl border border-surface-border bg-surface hover:border-accent/30 transition-colors"
            >
              <div className="text-accent mb-4">{pain.icon}</div>
              <h3 className="text-lg font-semibold mb-2">{pain.title}</h3>
              <p className="text-muted text-sm leading-relaxed">
                {pain.description}
              </p>
            </div>
          ))}
        </div>

        {/* Divider */}
        <div className="mt-20 text-center">
          <p className="text-muted text-lg">
            None of these are "add more monitoring" problems. They&apos;re{" "}
            <span className="text-foreground font-medium">
              not enough intelligence
            </span>{" "}
            problems.
          </p>
        </div>
      </div>
    </section>
  );
}
