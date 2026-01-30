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
      title: "자동 히스토리 수집",
      description:
        "Chrome Extension이 백그라운드에서 자동으로 브라우징 히스토리를 추적합니다. 별도의 입력 없이 모든 방문 기록이 안전하게 수집됩니다.",
      visual: <AutoCollectVisual />,
      bgClassName: "section-light",
    },
    {
      title: "AI 기반 태그 추출 & 관계 그래프",
      description:
        "강력한 AI가 방문한 페이지들을 분석하여 핵심 키워드를 추출하고, 페이지 간의 연관성을 파악하여 지식 그래프를 자동으로 구축합니다.",
      visual: <AIAnalysisVisual />,
      bgClassName: "section-gray",
    },
    {
      title: "3D 인터랙티브 마인드맵",
      description:
        "수집된 데이터가 아름다운 3D 마인드맵으로 시각화됩니다. 마우스로 회전하고 노드를 클릭하여 인사이트를 탐색하세요.",
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
