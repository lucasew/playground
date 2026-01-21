import React, { useMemo } from "react";
import { AbsoluteFill, useVideoConfig } from "remotion";

const NUM_STARS = 200;

export const Starfield: React.FC = () => {
  const { width, height } = useVideoConfig();
  
  const stars = useMemo(() => {
    return Array.from({ length: NUM_STARS }).map((_, i) => ({
      x: Math.random() * width,
      y: Math.random() * height,
      size: Math.random() * 2 + 1,
      opacity: Math.random(),
    }));
  }, [width, height]);

  return (
    <AbsoluteFill style={{ backgroundColor: "black" }}>
      {stars.map((star, i) => (
        <div
          key={i}
          style={{
            position: "absolute",
            left: star.x,
            top: star.y,
            width: star.size,
            height: star.size,
            backgroundColor: "white",
            borderRadius: "50%",
            opacity: star.opacity,
          }}
        />
      ))}
    </AbsoluteFill>
  );
};
