'use client';

import { EffectComposer, Bloom, Vignette } from '@react-three/postprocessing';
import { BlendFunction } from 'postprocessing';

interface PostProcessingProps {
  bloomIntensity?: number;
  bloomThreshold?: number;
}

export function PostProcessing({
  bloomIntensity = 0.5,
  bloomThreshold = 0.2,
}: PostProcessingProps) {
  return (
    <EffectComposer>
      <Bloom
        luminanceThreshold={bloomThreshold}
        luminanceSmoothing={0.9}
        intensity={bloomIntensity}
        mipmapBlur
      />
      <Vignette
        offset={0.3}
        darkness={0.5}
        blendFunction={BlendFunction.NORMAL}
      />
    </EffectComposer>
  );
}
