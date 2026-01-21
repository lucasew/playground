import React from "react";
import { AbsoluteFill, interpolate, useCurrentFrame, useVideoConfig } from "remotion";
import { lyrics } from "../lyrics";

export const Crawl: React.FC = () => {
  const frame = useCurrentFrame();
  const { height, width } = useVideoConfig();

  const translateY = interpolate(frame, [0, 2500], [height, -height * 4], {
    extrapolateLeft: "clamp",
  });

  return (
    <AbsoluteFill
      style={{
        backgroundColor: "transparent",
        perspective: 400,
        display: "flex",
        justifyContent: "center",
        overflow: "hidden",
      }}
    >
      <div
        style={{
          width: width * 0.8,
          transform: `rotateX(25deg) translateY(${translateY}px)`,
          color: "#FFE81F",
          textAlign: "justify",
          fontSize: 48,
          fontWeight: "bold",
          lineHeight: 1.5,
          fontFamily: "sans-serif",
        }}
      >
        <div style={{ textAlign: "center", marginBottom: 100 }}>
          <div style={{ fontSize: 32, marginBottom: 20 }}>EVIDÊNCIAS</div>
          <div style={{ fontSize: 64, marginBottom: 40 }}>Chitãozinho & Xororó</div>
        </div>
        {lyrics.map((line, i) => (
          <p key={i} style={{ marginBottom: 40 }}>
            {line}
          </p>
        ))}
      </div>
      {/* Fade out effect at the top */}
      <div
        style={{
          position: "absolute",
          top: 0,
          left: 0,
          right: 0,
          height: height * 0.4,
          background: "linear-gradient(to bottom, black 0%, transparent 100%)",
          pointerEvents: "none",
        }}
      />
    </AbsoluteFill>
  );
};
