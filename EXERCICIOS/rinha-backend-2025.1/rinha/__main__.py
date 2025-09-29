import os
import uvicorn
import logging
from . import app

logger = logging.getLogger(__name__)

def main():
    uvicorn.run('rinha:app', host="0.0.0.0", port=int(os.environ.get('PORT', 9999)), reload=True)
    

if __name__ == "__main__":
    main()
