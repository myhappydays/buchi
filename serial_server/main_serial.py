
from contextlib import asynccontextmanager
from fastapi import FastAPI, HTTPException
from serial_manager import SerialManager
import uvicorn

@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup
    if not serial_manager.find_and_open_port():
        print("CRITICAL: Could not connect to BUCHI device. The application will run without device communication.")
    yield
    # Shutdown
    serial_manager.close_port()

app = FastAPI(lifespan=lifespan)

serial_manager = SerialManager(baud_rate=115200)

@app.get("/write-url")
async def write_url_endpoint(url: str):
    if not url:
        raise HTTPException(status_code=400, detail="URL is required")
    
    try:
        response = serial_manager.write_url(url)
        return {"status": "success", "response": response}
    except ConnectionError as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/validate-token")
async def validate_token_endpoint(token: str):
    try:
        is_valid = serial_manager.validate_token(token)
        if not is_valid:
            raise HTTPException(status_code=403, detail="Invalid Token")
        return {"status": "ok", "token_validated": True}
    except ConnectionError as e:
        raise HTTPException(status_code=500, detail=str(e))

if __name__ == "__main__":
    uvicorn.run(app, host="127.0.0.1", port=28245)

