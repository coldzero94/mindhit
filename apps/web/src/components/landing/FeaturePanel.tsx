"use client";

import { ReactNode } from "react";
import { FullPageSection } from "./animations/FullPageSection";
import { ScrollReveal } from "./animations/ScrollReveal";

interface FeaturePanelProps {
  title: string;
  description: string;
  visual: ReactNode;
  /**
   * Background color class (section-light, section-gray, etc.)
   */
  bgClassName?: string;
  /**
   * Z-index for stacking order
   */
  zIndex?: number;
}

/**
 * FeaturePanel - iPhone-style single feature with center-stage visual
 *
 * Features:
 * - Full-page sticky section
 * - Center visual with scroll-linked scale/rotate
 * - Single-column layout (title → visual → description)
 */
export function FeaturePanel({
  title,
  description,
  visual,
  bgClassName = "section-light",
  zIndex,
}: FeaturePanelProps) {
  return (
    <FullPageSection className={bgClassName} zIndex={zIndex}>
      <div className="container mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex flex-col items-center justify-center min-h-screen py-20">
          {/* Title */}
          <h3 className="text-3xl sm:text-4xl md:text-5xl font-bold text-center text-foreground mb-12 sm:mb-16">
            {title}
          </h3>

          {/* Center Visual with Scroll Animation */}
          <div className="my-8 sm:my-12">
            <ScrollReveal
              scaleRange={[0.85, 1, 0.95]}
              rotateRange={[0, 0, 2]}
              className="w-full max-w-md mx-auto"
            >
              {visual}
            </ScrollReveal>
          </div>

          {/* Description */}
          <p className="text-lg sm:text-xl text-center text-muted-foreground max-w-2xl mt-8 sm:mt-12">
            {description}
          </p>
        </div>
      </div>
    </FullPageSection>
  );
}
