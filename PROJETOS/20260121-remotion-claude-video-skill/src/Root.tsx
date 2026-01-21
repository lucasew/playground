import { Composition } from "remotion";
import { StarWars } from "./StarWars";

export const RemotionRoot: React.FC = () => {
  return (
    <>
      <Composition
        id="StarWars"
        component={StarWars}
        durationInFrames={3000}
        fps={30}
        width={1920}
        height={1080}
      />
    </>
  );
};
