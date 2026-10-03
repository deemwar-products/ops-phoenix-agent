import { Navbar } from "@/components/Navbar";
import { Hero } from "@/components/Hero";
import { Problem } from "@/components/Problem";
import { HowItWorks } from "@/components/HowItWorks";
import { DemoSection } from "@/components/DemoSection";
import { TwoSurfaces } from "@/components/TwoSurfaces";
import { Install } from "@/components/Install";
import { DataSources } from "@/components/DataSources";
import { Pricing } from "@/components/Pricing";
import { CTA } from "@/components/CTA";
import { Footer } from "@/components/Footer";

export function App() {
  return (
    <>
      <Navbar />
      <main>
        <Hero />
        <Problem />
        <HowItWorks />
        <DemoSection />
        <TwoSurfaces />
        <Install />
        <DataSources />
        <Pricing />
        <CTA />
      </main>
      <Footer />
    </>
  );
}
