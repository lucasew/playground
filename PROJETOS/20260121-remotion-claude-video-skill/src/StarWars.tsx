import React from "react";
import { AbsoluteFill, Sequence, useVideoConfig } from "remotion";
import { Starfield } from "./components/Starfield";
import { IntroText } from "./components/IntroText";
import { Logo } from "./components/Logo";
import { Crawl } from "./components/Crawl";

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
