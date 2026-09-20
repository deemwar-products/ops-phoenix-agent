import { Navbar } from "@/components/Navbar";
import { Hero } from "@/components/Hero";
import { WhatItDoes } from "@/components/WhatItDoes";
import { DataSources } from "@/components/DataSources";
import { Install } from "@/components/Install";
import { Problem } from "@/components/Problem";
import { HowItWorks } from "@/components/HowItWorks";
import { TwoSurfaces } from "@/components/TwoSurfaces";
import { Pricing } from "@/components/Pricing";
import { CTA } from "@/components/CTA";
import { Footer } from "@/components/Footer";

export function App() {
  return (
    <>
      <Navbar />
      <main>
        <Hero />
        <WhatItDoes />
        <DataSources />
        <Install />
        <Problem />
        <HowItWorks />
        <TwoSurfaces />
        <Pricing />
        <CTA />
      </main>
      <Footer />
    </>
  );
}
