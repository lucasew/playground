import React from "react";
import { AbsoluteFill, interpolate, useCurrentFrame } from "remotion";

/**
 * Displays the initial "A long time ago..." style text.
 *
 * This component fades in and out the text "Há muito tempo, em um karaokê muito, muito distante..."
 * setting the humorous tone for the parody.
 *
 * The opacity is interpolated over frames:
 * - 0-20: Fade in
 * - 20-100: Hold
 * - 100-120: Fade out
 */
export const IntroText: React.FC = () => {
  const frame = useCurrentFrame();
  const opacity = interpolate(frame, [0, 20, 100, 120], [0, 1, 1, 0], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });

  return (
    <AbsoluteFill
      style={{
        display: "flex",
        justifyContent: "center",
        alignItems: "center",
        backgroundColor: "black",
      }}
    >
      <div
        style={{
          color: "#4BD5EE",
          fontSize: 48,
          fontWeight: "bold",
          textAlign: "center",
          opacity,
          width: "60%",
          lineHeight: 1.5,
          fontFamily: "sans-serif",
        }}
      >
        Há muito tempo, em um karaokê muito, muito distante...
      </div>
    </AbsoluteFill>
  );
};
