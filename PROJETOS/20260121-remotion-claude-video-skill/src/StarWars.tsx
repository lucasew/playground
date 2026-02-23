import React from "react";
import { AbsoluteFill, Sequence, useVideoConfig } from "remotion";
import { Starfield } from "./components/Starfield";
import { IntroText } from "./components/IntroText";
import { Logo } from "./components/Logo";
import { Crawl } from "./components/Crawl";

/**
 * Main composition for the Star Wars intro parody.
 *
 * Orchestrates the sequence of events using Remotion's `Sequence` component:
 * 1. `Starfield`: Background active throughout the entire video.
 * 2. `IntroText`: "A long time ago..." style text (frames 0-150).
 * 3. `Logo`: Zooming "EVIDÊNCIAS" logo (frames 150-300).
 * 4. `Crawl`: Scrolling lyrics text (starts at frame 300).
 *
 * The background color is explicitly set to black using `AbsoluteFill`.
 */
export const StarWars: React.FC = () => {
  const { durationInFrames } = useVideoConfig();

  return (
    <AbsoluteFill style={{ backgroundColor: "black" }}>
      <Starfield />
      
      <Sequence from={0} durationInFrames={150}>
        <IntroText />
      </Sequence>

      <Sequence from={150} durationInFrames={150}>
        <Logo />
      </Sequence>

      <Sequence from={300} durationInFrames={durationInFrames - 300}>
        <Crawl />
      </Sequence>
    </AbsoluteFill>
  );
};
