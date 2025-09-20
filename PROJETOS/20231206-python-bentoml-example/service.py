from pathlib import Path
from io import BytesIO
import base64

import bentoml
from fastapi.staticfiles import StaticFiles
from fastapi import FastAPI
from PIL import Image


class TransposeRunner(bentoml.Runnable):
    SUPPORTED_RESOURCES = ('cpu',)
    SUPPORTS_CPU_MULTI_THREADING = False
    
    def __init__(self):
        pass

    @bentoml.Runnable.method(batchable=False)
    def transpose(self, image):
        transposed = image.transpose(method=Image.ROTATE_90)
        buf = BytesIO()
        transposed.save(buf, format='JPEG')
        return base64.b64encode(buf.getvalue()).decode('utf-8')

transpose_runner = bentoml.Runner(TransposeRunner, name='transposer')

svc = bentoml.Service(name="teste", runners=[transpose_runner])

@svc.api(input=bentoml.io.Image(), output=bentoml.io.JSON(), route='v1/models/transpose/predict')
async def transpose(image):
    return dict(
        transposed=await transpose_runner.transpose.async_run(image)
    )
    
@svc.api(input=bentoml.io.Text(), output=bentoml.io.JSON(), route='v1/models/teste/predict')
async def teste_rota(text: str) -> any:
    return dict(response=f"Hello, {text}")


router = FastAPI()
svc.mount_asgi_app(router)

router.mount('/demo', StaticFiles(directory=Path(__file__).parent / "static", html=True), name='static')
