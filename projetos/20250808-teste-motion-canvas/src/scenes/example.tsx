import {Camera, Circle, Layout, Rect, makeScene2D} from '@motion-canvas/2d';
import {all, createRef, loop} from '@motion-canvas/core';

export default makeScene2D(function* (view) {
  // Create your animations here

  const circle = createRef<Circle>();
  const rect = createRef<Rect>();
  const camera = createRef<Camera>()

  const size = Math.min(view.width(), view.height())*.9
  view.add(
    <Camera ref={camera}>
      <Layout  alignItems='center' justifyContent='center'>
        <Circle ref={circle} scale={0} height={size} width={size} fill={'blue'} />
        <Rect ref={rect} scale={1} height={size} width={size} fill={'red'} />
      </Layout>
    </Camera>
  )

  yield *loop(10, () => all(
    loop(2, () => rect().scale(.25, 1).to(1, 1)),
    loop(2, () => circle().scale(1, 1).to(.25, 1)),
    camera().rotation(0, 0).to(90, 1).to(180, 1).to(270, 1).to(360, 1).to(0, 0)
  ))
  
});
