import React from "react";
import { AbsoluteFill, interpolate, useCurrentFrame } from "remotion";

export const Logo: React.FC = () => {
  const frame = useCurrentFrame();

  const scale = interpolate(frame, [0, 150], [2, 0], {
    extrapolateRight: "clamp",
  });

  const opacity = interpolate(frame, [0, 10, 130, 150], [0, 1, 1, 0], {
    extrapolateRight: "clamp",
  });

  return (
    <AbsoluteFill
      style={{
        display: "flex",
        justifyContent: "center",
        alignItems: "center",
      }}
    >
      <div
        style={{
          color: "#FFE81F",
          fontSize: 180,
          fontWeight: 900,
          textAlign: "center",
          transform: `scale(${scale})`,
          opacity,
          fontFamily: "sans-serif",
          letterSpacing: -5,
        }}
      >
        EVIDÊNCIAS
      </div>
    </AbsoluteFill>
  );
};
