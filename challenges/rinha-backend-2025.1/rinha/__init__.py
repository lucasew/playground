from fastapi import FastAPI, Response, status
from pydantic import BaseModel
import contextlib
import logging
from redis.asyncio import Redis
from argparse import ArgumentParser
import os
import time
import json
import requests
from datetime import datetime
import asyncio

logger = logging.getLogger(__name__)

def parse_args():
    parser = ArgumentParser()
    parser.add_argument('--redis-url', default=os.environ.get('REDIS_URL', 'redis://localhost:6379'))
    parser.add_argument('--verbose', action='store_true')
    args = parser.parse_args()
    logging.basicConfig(level=logging.DEBUG if args.verbose else logging.INFO)
    return args

@contextlib.asynccontextmanager
async def lifespan(app: FastAPI):
    args = parse_args()
    logger.info("Conectando no redis")
    app.redis = Redis.from_url(args.redis_url, decode_responses=True)
    await app.redis.publish('startup', "subiu") # teste
    app.session = requests.Session()
    yield
    await app.redis.aclose(True)
    app.session.close()

class PaymentProcessor():
    def __init__(self, host):
        self.host = host

    @property
    def key_requests(self):
        return self.host + "__requests"

    @property
    def key_amount(self):
        return self.host + "__amount"

    @property
    def key_healthy(self):
        return self.host + "__healthy"

    async def summary(self, app):
        return {
            'totalRequests': int(await app.redis.get(self.key_requests) or 0),
            'totalAmount': float(await app.redis.get(self.key_amount) or 0.0)
        }

    async def purge(self, app):
        await app.redis.delete(self.key_requests)
        await app.redis.delete(self.key_amount)
        await app.session.post(self.host + "/admin/purge-payments")

    async def handle_payment(self, app, correlationId: str, amount: float):
        r = app.session.post(self.host + "/payments", data=json.dumps({
               'correlationId': correlationId,
               'amount': amount,
               'requestedAt': datetime.now().isoformat()
           }), headers={
                   'Content-Type': 'application/json'
               })
        # logger.info(f'upstream status: {r.status} {await r.text()}')
        if r.status_code == status.HTTP_200_OK:
            await asyncio.gather(
                app.redis.incr(self.key_requests),
                app.redis.incrbyfloat(self.key_amount, amount)
             )
            return True
        else:
            await app.redis.set(self.key_healthy, 'false', ex=1)
            return False

    async def healthcheck(self, app):
        r = app.session.get(self.host + '/payments/service-health')
        return r.json()

    async def is_ok(self, app):
        ret = await app.redis.get(self.key_healthy)
        if ret is None:
            return True
        return json.loads(ret)

            

DEFAULT_PROCESSOR = PaymentProcessor("http://payment-processor-default:8080") 
FALLBACK_PROCESSOR = PaymentProcessor("http://payment-processor-fallback:8080")
PROCESSORS = [DEFAULT_PROCESSOR, FALLBACK_PROCESSOR]

app = FastAPI(lifespan=lifespan)

@app.get('/time')
async def get_time():
    ret = await app.redis.get('last_time')
    if ret is None:
        ret = json.dumps({'time': time.time()})
        await app.redis.set('last_time', ret, ex=5)
    return json.loads(ret)['time']

class Payment(BaseModel):
    correlationId: str
    amount: float

@app.post('/payments')
async def handle_payment(data: Payment, response: Response):
    response.status_code = status.HTTP_500_INTERNAL_SERVER_ERROR
    for processor in PROCESSORS:
        if await processor.is_ok(app):
            if not await processor.handle_payment(app, data.correlationId, data.amount):
                continue
            response.status_code = status.HTTP_200_OK
            return response
    return response
    
@app.get('/payments-summary')
async def payments_summary():
    return {
        'default': await DEFAULT_PROCESSOR.summary(app),
        'fallback': await FALLBACK_PROCESSOR.summary(app)
    }

@app.get('/debug/health')
async def debug_health():
    return {
        'default': {
            'bruto': await DEFAULT_PROCESSOR.healthcheck(app),
            'is_ok': await DEFAULT_PROCESSOR.is_ok(app)
        },
        'fallback': {
            'bruto': await FALLBACK_PROCESSOR.healthcheck(app),
            'is_ok': await FALLBACK_PROCESSOR.is_ok(app)
        }
    }

@app.get('/reset')
async def reset():
    await asyncio.gather(
        DEFAULT_PROCESSOR.purge(app),
        FALLBACK_PROCESSOR.purge(app)
     )
