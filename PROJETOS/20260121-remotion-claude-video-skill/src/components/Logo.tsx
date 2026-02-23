import React from "react";
import { AbsoluteFill, interpolate, useCurrentFrame } from "remotion";

/**
 * Displays the main title logo "EVIDÊNCIAS" with a zoom-out effect.
 *
 * Simulates the iconic Star Wars logo receding into the distance.
 *
 * Effects:
 * - Scale: Zooms out from 2x to 0x (disappearing into the distance).
 * - Opacity: Fades in quickly, stays visible, then fades out as it disappears.
 */
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
