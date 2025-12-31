'use client';

import { MindmapCanvas } from '@/components/mindmap/MindmapCanvas';
import { TestSphere } from '@/components/mindmap/TestSphere';

export default function Test3DPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">3D 렌더링 테스트</h1>
        <p className="text-gray-500 mt-1">
          React Three Fiber 환경 설정 확인용 페이지입니다.
        </p>
      </div>

      <div className="bg-white rounded-xl shadow-sm p-4">
        <h2 className="text-lg font-medium mb-4">테스트 Canvas</h2>
        <MindmapCanvas className="h-[500px]">
          <TestSphere />
        </MindmapCanvas>
      </div>

      <div className="bg-white rounded-xl shadow-sm p-4">
        <h2 className="text-lg font-medium mb-2">조작 방법</h2>
        <ul className="text-sm text-gray-600 space-y-1">
          <li>• 마우스 드래그: 회전</li>
          <li>• 스크롤: 줌</li>
          <li>• 우클릭 드래그: 이동 (Pan)</li>
          <li>• 구체 호버: 색상 변경</li>
        </ul>
      </div>
    </div>
  );
}
