"use client";

import { FeaturePanel } from "./FeaturePanel";
import { AutoCollectVisual } from "./visuals/AutoCollectVisual";
import { AIAnalysisVisual } from "./visuals/AIAnalysisVisual";
import { MindmapVisual } from "./visuals/MindmapVisual";

/**
 * FeaturesSection - Container for 3 feature panels
 *
 * Features:
 * - 3 full-page sticky sections stacked vertically
 * - Each with center-stage visual and scroll animations
 * - Alternating background colors for variety
 */
export function FeaturesSection() {
  const features = [
    {
      title: "Automatic History Collection",
      description:
        "The Chrome Extension automatically tracks your browsing history in the background. All visit records are collected securely without any manual input.",
      visual: <AutoCollectVisual />,
      bgClassName: "section-light",
    },
    {
      title: "AI-Powered Tag Extraction & Relationship Graph",
      description:
        "Powerful AI analyzes visited pages to extract key keywords and identifies relationships between pages to automatically build a knowledge graph.",
      visual: <AIAnalysisVisual />,
      bgClassName: "section-gray",
    },
    {
      title: "3D Interactive Mindmap",
      description:
        "Collected data is visualized as a beautiful 3D mindmap. Rotate with your mouse and click nodes to explore insights.",
      visual: <MindmapVisual />,
      bgClassName: "section-lighter",
    },
  ];

  return (
    <>
      {features.map((feature, index) => (
        <FeaturePanel
          key={index}
          title={feature.title}
          description={feature.description}
          visual={feature.visual}
          bgClassName={feature.bgClassName}
          zIndex={40 + index * 10}
        />
      ))}
    </>
  );
}
